package players

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	configs "tektmud/internal/config"
	"tektmud/internal/logger"

	"gopkg.in/yaml.v3"
)

var (
	mu sync.RWMutex
)

const (
	IndexVersion = 1
)

type IndexHeader struct {
	IndexVersion uint64 //The current version of our index header
	RecordCount  uint64 //The number of index records we are holding

}

type IndexEntry struct {
	Playername string
	PlayerId   uint64
}

type PlayerIndex struct {
	Filename      string
	headerData    IndexHeader
	NextPlayerId  uint64 //The next playerId to assign (can be different than IndexHeader.RecordCount)
	PlayersByName map[string]uint64
}

func NewPlayerIndex(indexName string) *PlayerIndex {
	c := configs.GetConfig()
	filename := filepath.Join(c.Paths.RootDataDir, c.Paths.PlayerData, indexName)
	idx := &PlayerIndex{Filename: filename}

	return idx
}

func (idx *PlayerIndex) Exists() bool {
	_, err := os.Stat(idx.Filename)
	return err == nil
}

func (idx *PlayerIndex) Delete() {
	if idx.Exists() {
		os.Remove(idx.Filename)
	}
}

func (idx *PlayerIndex) Create() error {
	mu.Lock()
	defer mu.Unlock()

	//Just as a safety.
	idx.Delete()

	file, err := os.Create(idx.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data := IndexHeader{IndexVersion: IndexVersion, RecordCount: 0}
	//Write header (16 bytes)
	if err := binary.Write(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("failed to write header %w", err)
	}
	idx.headerData = data
	return nil
}

func (idx *PlayerIndex) Save() error {
	//Going to do a "safe" save by first creating a tmp index then moving it
	tmpIndexPath := idx.Filename + `.tmp`
	file, err := os.Create(tmpIndexPath)
	if err != nil {
		return err
	}

	//Write the header (16 bytes)
	if err := binary.Write(file, binary.LittleEndian, idx.headerData); err != nil {
		return err
	}

	//Sort playernames for consistent output
	playernames := make([]string, 0, len(idx.PlayersByName))
	for playername := range idx.PlayersByName {
		playernames = append(playernames, playername)
	}
	sort.Strings(playernames)

	for _, playername := range playernames {
		playerId := idx.PlayersByName[playername]

		//Write the PlayerId (8bytes)
		if err := binary.Write(file, binary.LittleEndian, playerId); err != nil {
			return fmt.Errorf("failed to write playerId: %w", err)
		}

		//Write the playername length (2 bytes) - technically capped at 65535 chars, system will cap lower
		playernameLen := uint16(len(playername))
		if err := binary.Write(file, binary.LittleEndian, playernameLen); err != nil {
			return fmt.Errorf("failed to write playernameLen: %w", err)
		}

		//Write the playername
		if _, err := file.Write([]byte(playername)); err != nil {
			return fmt.Errorf("failed to write playername: %w", err)
		}
	}
	file.Close()

	//now copy over the old index
	if err := os.Rename(tmpIndexPath, idx.Filename); err != nil {
		return err
	}
	return nil
}

func (idx *PlayerIndex) Rebuild() error {
	mu.Lock()
	defer mu.Unlock()

	//scan the players directory for any player files. rebuild the index
	//then save it.
	file, err := os.Open(idx.Filename)
	if err != nil {
		return fmt.Errorf("failed to open index for rebuild")
	}
	defer file.Close()

	players := idx.scanForPlayerFiles()
	idx.headerData.RecordCount = uint64(len(players))
	idx.PlayersByName = players
	idx.NextPlayerId = idx.getNextAvailablePlayerId()

	data := IndexHeader{IndexVersion: IndexVersion, RecordCount: uint64(len(players))}
	//Write header (16 bytes)
	if err := binary.Write(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("failed to write header %w", err)
	}

	//Sort playernames for consistent output
	playernames := make([]string, 0, len(players))
	for player := range players {
		playernames = append(playernames, player)
	}
	sort.Strings(playernames)

	for _, playername := range playernames {
		playerId := players[playername]

		//Write the PlayerId (8bytes)
		if err := binary.Write(file, binary.LittleEndian, playerId); err != nil {
			return fmt.Errorf("failed to write playerId: %w", err)
		}

		//Write the playername length (2 bytes) - technically capped at 65535 chars, system will cap lower
		playernameLen := uint16(len(playername))
		if err := binary.Write(file, binary.LittleEndian, playernameLen); err != nil {
			return fmt.Errorf("failed to write playernameLen: %w", err)
		}

		//Write the playername
		if _, err := file.Write([]byte(playername)); err != nil {
			return fmt.Errorf("failed to write playername: %w", err)
		}
	}

	//Calcluate the next playerId to be used.
	for _, val := range players {
		if val > idx.NextPlayerId {
			idx.NextPlayerId = val
		}
	}
	idx.NextPlayerId++

	return nil
}

func (idx *PlayerIndex) getNextAvailablePlayerId() uint64 {
	if len(idx.PlayersByName) > 0 {
		var maxPlayerId uint64 = 0
		for _, i := range idx.PlayersByName {
			if i > maxPlayerId {
				maxPlayerId = i
			}
		}
		return maxPlayerId + 1
	}
	return 1 //Start at 1, so we can use 0 for "system"
}

func (idx *PlayerIndex) PlayerIdByName(name string) (uint64, bool) {

	id, exists := idx.PlayersByName[name]
	return id, exists

}

func (idx *PlayerIndex) scanForPlayerFiles() map[string]uint64 {
	c := configs.GetConfig()
	dir := filepath.Join(c.Paths.RootDataDir, c.Paths.PlayerData)

	dirInfos, err := os.ReadDir(dir)
	if err != nil {
		logger.Error("unable to read files in index directory.", "err", err)
		return make(map[string]uint64)
	}

	playerData := make(map[string]uint64)

	for _, fileInfo := range dirInfos {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".yaml") {
			data, err := os.ReadFile(filepath.Join(dir, fileInfo.Name()))
			if err != nil {
				logger.Error("unable to open file:", "filename", fileInfo.Name())
			}
			var playerRecord PlayerRecord
			if err := yaml.Unmarshal(data, &playerRecord); err != nil {
				logger.Error("unable to deserialize file:", "filename", fileInfo.Name())
			}
			playerData[playerRecord.Username] = playerRecord.Id
		}
	}
	return playerData
}
