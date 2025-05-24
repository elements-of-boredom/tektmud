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
	nameToId  map[string]uint64 //username -> userId
	users     map[uint64]*UserRecord
}

// Creates a new UserManager Instance
func NewUserManager(indexPath, usersDir string) (*UserManager, error) {
	um := &UserManager{
		indexPath: indexPath,
		usersDir:  usersDir,
		nameToId:  make(map[string]uint64),
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
	}
	um.userIndex = idx
	return nil
}

func (um *UserManager) GetUserByUsername(username string) (*UserRecord, error) {
	//first lets see if they are in our active cache
	um.mu.RLock()

	//TODO: This needs to properly leverage the index.
	userId, exists := um.nameToId[username]
	um.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("user '%s' not found", username)
	}

	return um.GetUserById(userId)
}

func (um *UserManager) GetUserById(userId uint64) (*UserRecord, error) {
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
	return &userRecord, nil
}
