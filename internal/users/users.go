package users

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	configs "tektmud/internal/config"

	"gopkg.in/yaml.v3"
)

type UserManager struct {
	indexPath string
	usersDir  string
	mu        sync.RWMutex
	userIndex *UserIndex
	users     map[uint64]*UserRecord
}

// Creates a new UserManager Instance
func NewUserManager(indexPath, usersDir string) (*UserManager, error) {
	um := &UserManager{
		indexPath: indexPath,
		usersDir:  usersDir,
		users:     make(map[uint64]*UserRecord),
	}
	//Ensure the directory exists
	if err := os.MkdirAll(usersDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create users directory: %w", err)
	}

	if err := um.loadBinaryIndex(); err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	return um, nil
}

func (um *UserManager) loadBinaryIndex() error {
	um.mu.Lock()
	defer um.mu.Unlock()

	idx := NewUserIndex(um.indexPath)
	if !idx.Exists() {
		//Create it.
		idx.Create()
		idx.Rebuild()
	}
	//At the start when we load the index, lets force it to rebuild.
	idx.Rebuild()
	um.userIndex = idx
	return nil
}

func (um *UserManager) GetUserByUsername(username string) (*UserRecord, error) {
	//first lets see if they are in our active cache
	um.mu.RLock()
	userId, exists := um.userIndex.UsersByName[username]
	um.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("user '%s' not found", username)
	}

	return um.GetUserById(userId)
}

func (um *UserManager) GetUserById(userId uint64) (*UserRecord, error) {

	um.mu.RLock()
	user, exists := um.users[userId]
	um.mu.RUnlock()
	if exists {
		return user, nil
	}

	config := configs.GetConfig()
	userPath := filepath.Join(config.Paths.RootDataDir, config.Paths.UserData, fmt.Sprintf("%d.yaml", userId))

	data, err := os.ReadFile(userPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read user file: %w", err)
	}

	var userRecord UserRecord
	if err := yaml.Unmarshal(data, &userRecord); err != nil {
		return nil, fmt.Errorf("failed to parse user file: %w", err)
	}
	//add it to our cache
	um.mu.Lock()
	defer um.mu.Unlock()
	um.users[userRecord.Id] = &userRecord
	return &userRecord, nil
}

func (um *UserManager) PasswordMeetsMinimums(input string, username string) bool {

	return len(input) > 5 &&
		username != input
}

func (um *UserManager) ValidatePassword(input string, userId uint64) bool {
	//TODO: Encryption
	if user, err := um.GetUserById(userId); err != nil {
		return false
	} else {
		return user.Password == input
	}
}

func (um *UserManager) CreateUser(username string, password string, email string) (*UserRecord, error) {
	um.mu.Lock()
	defer um.mu.Unlock()

	c := configs.GetConfig()

	//make sure our user doens't already exist
	if _, exists := um.userIndex.UsersByName[username]; exists {
		return nil, fmt.Errorf("user '%s' already exists", username)
	}

	user := &UserRecord{
		Id:       um.userIndex.NextUserId,
		Username: username,
		Password: password, //TODO: Encrypt
		Roles:    []string{RoleUser},
	}

	//save the user file
	userPath := filepath.Join(c.Paths.RootDataDir, c.Paths.UserData, fmt.Sprintf("%d.yaml", user.Id))
	if err := um.saveUserFile(user, userPath); err != nil {
		return nil, fmt.Errorf("failed to save the user file: %w", err)
	}

	//Update the index
	um.userIndex.UsersByName[user.Username] = user.Id
	um.userIndex.headerData.RecordCount++
	um.userIndex.NextUserId++
	//If we can't update the index, we are going to have problems
	//roll back the user save and error out
	if err := um.userIndex.Save(); err != nil {
		os.Remove(userPath)
		delete(um.userIndex.UsersByName, username)
		um.userIndex.headerData.RecordCount--
		um.userIndex.NextUserId--
		return nil, fmt.Errorf("failed to update index during create user: %w", err)
	}

	return user, nil
}

// saveUserFile saves a user to a YAML file
func (um *UserManager) saveUserFile(user *UserRecord, path string) error {
	data, err := yaml.Marshal(user)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
