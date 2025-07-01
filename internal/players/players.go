package players

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

type PlayerManager struct {
	indexPath   string
	playersDir  string
	mu          sync.RWMutex
	playerIndex *PlayerIndex
	players     map[uint64]*PlayerRecord
}

type HashSalt struct {
	Version uint16
	Salt    []byte
	Hash    []byte
}

var (
	players map[uint64]*PlayerRecord = make(map[uint64]*PlayerRecord)
)

// Creates a new UserManager Instance
func NewPlayerManager(indexPath, playersDir string) (*PlayerManager, error) {
	pm := &PlayerManager{
		indexPath:  indexPath,
		playersDir: playersDir,
		players:    players,
	}
	//Ensure the directory exists
	if err := os.MkdirAll(playersDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create players directory: %w", err)
	}

	if err := pm.loadBinaryIndex(); err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	return pm, nil
}

func (pm *PlayerManager) loadBinaryIndex() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	idx := NewPlayerIndex(pm.indexPath)
	if !idx.Exists() {
		//Create it.
		idx.Create()
		idx.Rebuild()
	}
	//At the start when we load the index, lets force it to rebuild.
	idx.Rebuild()
	pm.playerIndex = idx
	return nil
}

func (pm *PlayerManager) GetPlayerByUsername(username string) (*PlayerRecord, error) {
	//first lets see if they are in our active cache
	//lowercase the name to prevent casing duplicates
	lusername := strings.ToLower(username)
	pm.mu.RLock()
	playerId, exists := pm.playerIndex.PlayersByName[lusername]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("player '%s' not found", username)
	}

	return pm.GetPlayerById(playerId)
}

func (pm *PlayerManager) GetPlayerById(playerId uint64) (*PlayerRecord, error) {

	pm.mu.RLock()
	player, exists := pm.players[playerId]
	pm.mu.RUnlock()
	if exists {
		return player, nil
	}

	config := configs.GetConfig()
	playerFilePath := filepath.Join(config.Paths.RootDataDir, config.Paths.PlayerData, fmt.Sprintf("%d.yaml", playerId))

	data, err := os.ReadFile(playerFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read player file: %w", err)
	}

	var playerRecord PlayerRecord
	if err := yaml.Unmarshal(data, &playerRecord); err != nil {
		return nil, fmt.Errorf("failed to parse player file: %w", err)
	}
	//Validate the character - This sets up
	//their character for use in game.
	if !playerRecord.Char.Validate() {
		return nil, fmt.Errorf("failed to validate the player file: %w", err)
	}
	//add it to our cache
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.players[playerRecord.Id] = &playerRecord
	return &playerRecord, nil
}

func GetByCharacterName(characterName string) *PlayerRecord {
	if len(characterName) <= 0 {
		logger.Warn("Something asked for a character name of 0 length")
		return nil
	}

	for _, u := range players {
		if strings.EqualFold(u.Char.Name, characterName) {
			return u
		}
	}

	return nil
}

func (pm *PlayerManager) PasswordMeetsMinimums(input string, username string) bool {

	return len(input) > 5 &&
		username != input
}

func (pm *PlayerManager) ValidatePassword(input string, playerId uint64) bool {

	if player, err := pm.GetPlayerById(playerId); err != nil {
		return false
	} else {
		return pm.comparePassword(input, player.Password)
	}
}

func (pm *PlayerManager) CreatePlayer(username string, password string, email string) (*PlayerRecord, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	c := configs.GetConfig()

	//make sure our player doens't already exist
	if _, exists := pm.playerIndex.PlayersByName[username]; exists {
		return nil, fmt.Errorf("player '%s' already exists", username)
	}

	encryptedPassword, err := pm.encryptedPassword(password)
	if err != nil {
		return nil, fmt.Errorf("encryption of player '%s' password failed", username)
	}

	pr := &PlayerRecord{
		Id:       pm.playerIndex.NextPlayerId,
		Username: username,
		Password: encryptedPassword,
		Roles:    []string{RoleUser},
	}

	//save the player file
	playerFilePath := filepath.Join(c.Paths.RootDataDir, c.Paths.PlayerData, fmt.Sprintf("%d.yaml", pr.Id))
	if err := pm.savePlayerFile(pr, playerFilePath); err != nil {
		return nil, fmt.Errorf("failed to save the player file: %w", err)
	}

	//Update the index
	pm.playerIndex.PlayersByName[pr.Username] = pr.Id
	pm.playerIndex.headerData.RecordCount++
	pm.playerIndex.NextPlayerId++
	//If we can't update the index, we are going to have problems
	//roll back the player save and error out
	if err := pm.playerIndex.Save(); err != nil {
		os.Remove(playerFilePath)
		delete(pm.playerIndex.PlayersByName, username)
		pm.playerIndex.headerData.RecordCount--
		pm.playerIndex.NextPlayerId--
		return nil, fmt.Errorf("failed to update index during create player: %w", err)
	}

	return pr, nil
}

// UpdatePlayer updates an existing player
func (pm *PlayerManager) UpdatePlayer(player *PlayerRecord) error {
	pm.mu.RLock()
	playerId, exists := pm.playerIndex.PlayersByName[player.Username]
	pm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("player '%s' not found", player.Username)
	}

	if playerId != player.Id {
		return fmt.Errorf("player Id mismatch")
	}

	playerFilePath := filepath.Join(pm.playersDir, fmt.Sprintf("%d.yaml", player.Id))
	return pm.savePlayerFile(player, playerFilePath)
}

// savePlayerFile saves a player to a YAML file
func (pm *PlayerManager) savePlayerFile(player *PlayerRecord, path string) error {
	data, err := yaml.Marshal(player)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Generate a hash of the password provided.
func (pm *PlayerManager) encryptedPassword(password string) (string, error) {
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

func (pm *PlayerManager) comparePassword(input string, existing string) bool {

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
