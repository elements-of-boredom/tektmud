package users

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	configs "tektmud/internal/config"
	"tektmud/internal/logger"

	"golang.org/x/crypto/argon2"
	"gopkg.in/yaml.v3"
)

type UserManager struct {
	indexPath string
	usersDir  string
	mu        sync.RWMutex
	userIndex *UserIndex
	users     map[uint64]*UserRecord
}

type HashSalt struct {
	Version uint16
	Salt    []byte
	Hash    []byte
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
	//lowercase the name to prevent casing duplicates
	lusername := strings.ToLower(username)
	um.mu.RLock()
	userId, exists := um.userIndex.UsersByName[lusername]
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

	if user, err := um.GetUserById(userId); err != nil {
		return false
	} else {
		return um.comparePassword(input, user.Password)
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

	encryptedPassword, err := um.encryptedPassword(password)
	if err != nil {
		return nil, fmt.Errorf("encryption of user '%s' password failed", username)
	}

	user := &UserRecord{
		Id:       um.userIndex.NextUserId,
		Username: username,
		Password: encryptedPassword,
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

// Generate a hash of the password provided.
func (um *UserManager) encryptedPassword(password string) (string, error) {
	var err error

	//generate a random salt.
	salt, err := randomSecret(16)
	if err != nil {
		return "", err
	}

	encoded := hashPassword(password, salt)
	return encoded, nil
}

func hashPassword(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, 5, 7168, 1, 32)
	hs := &HashSalt{
		Version: uint16(1),
		Hash:    hash,
		Salt:    salt,
	}
	result := make([]byte, 50) //2 + 16 + 32
	binary.LittleEndian.PutUint16(result[0:2], hs.Version)
	copy(result[2:18], hs.Salt)
	copy(result[18:50], hs.Hash)

	encoded := base64.StdEncoding.EncodeToString(result)
	return encoded
}

func (um *UserManager) comparePassword(input string, existing string) bool {

	existingHS, err := base64.StdEncoding.DecodeString(existing)
	if err != nil {
		logger.Error("Unable to decode existing password", "error", err)
		return false
	}

	oldHs := &HashSalt{
		Version: binary.LittleEndian.Uint16(existingHS[0:2]),
		Salt:    existingHS[2:18],
		Hash:    existingHS[18:50],
	}

	//Ok we have our salt, now
	encrypted := hashPassword(input, oldHs.Salt)
	return encrypted == existing

}

func randomSecret(length uint32) ([]byte, error) {
	secret := make([]byte, length)

	_, err := rand.Read(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}
