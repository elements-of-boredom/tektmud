package users

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	configs "tektmud/internal/config"

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
	Username string
	UserId   uint64
}

type UserIndex struct {
	Filename    string
	headerData  IndexHeader
	NextUserId  uint64 //The next userId to assign (can be different than IndexHeader.RecordCount)
	UsersByName map[string]uint64
}

func NewUserIndex(indexName string) *UserIndex {
	c := configs.GetConfig()
	filename := filepath.Join(c.Paths.RootDataDir, c.Paths.UserData, indexName)
	idx := &UserIndex{Filename: filename}

	return idx
}

func (idx *UserIndex) Exists() bool {
	_, err := os.Stat(idx.Filename)
	return err == nil
}

func (idx *UserIndex) Delete() {
	if idx.Exists() {
		os.Remove(idx.Filename)
	}
}

func (idx *UserIndex) Create() error {
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

func (idx *UserIndex) Save() error {
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

	//Sort usernames for consistent output
	usernames := make([]string, 0, len(idx.UsersByName))
	for username := range idx.UsersByName {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)

	for _, username := range usernames {
		userId := idx.UsersByName[username]

		//Write the UserId (8bytes)
		if err := binary.Write(file, binary.LittleEndian, userId); err != nil {
			return fmt.Errorf("failed to write userId: %w", err)
		}

		//Write the username length (2 bytes) - technically capped at 65535 chars, system will cap lower
		usernameLen := uint16(len(username))
		if err := binary.Write(file, binary.LittleEndian, usernameLen); err != nil {
			return fmt.Errorf("failed to write usernameLen: %w", err)
		}

		//Write the username
		if _, err := file.Write([]byte(username)); err != nil {
			return fmt.Errorf("failed to write username: %w", err)
		}
	}
	file.Close()

	//now copy over the old index
	if err := os.Rename(tmpIndexPath, idx.Filename); err != nil {
		return err
	}
	return nil
}

func (idx *UserIndex) Rebuild() error {
	mu.Lock()
	defer mu.Unlock()

	//scan the user directory for any user files. rebuild the index
	//then save it.
	file, err := os.Open(idx.Filename)
	if err != nil {
		return fmt.Errorf("failed to open index for rebuild")
	}
	defer file.Close()

	users := idx.scanForUserFiles()
	idx.headerData.RecordCount = uint64(len(users))
	idx.UsersByName = users

	data := IndexHeader{IndexVersion: IndexVersion, RecordCount: uint64(len(users))}
	//Write header (16 bytes)
	if err := binary.Write(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("failed to write header %w", err)
	}

	//Sort usernames for consistent output
	usernames := make([]string, 0, len(users))
	for username := range users {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)

	for _, username := range usernames {
		userId := users[username]

		//Write the UserId (8bytes)
		if err := binary.Write(file, binary.LittleEndian, userId); err != nil {
			return fmt.Errorf("failed to write userId: %w", err)
		}

		//Write the username length (2 bytes) - technically capped at 65535 chars, system will cap lower
		usernameLen := uint16(len(username))
		if err := binary.Write(file, binary.LittleEndian, usernameLen); err != nil {
			return fmt.Errorf("failed to write usernameLen: %w", err)
		}

		//Write the username
		if _, err := file.Write([]byte(username)); err != nil {
			return fmt.Errorf("failed to write username: %w", err)
		}
	}

	//Calcluate the next userId to be used.
	for _, val := range users {
		if val > idx.NextUserId {
			idx.NextUserId = val
		}
	}
	idx.NextUserId++

	return nil
}

func (idx *UserIndex) UserIdByName(name string) (uint64, bool) {

	id, exists := idx.UsersByName[name]
	return id, exists

}

func (idx *UserIndex) scanForUserFiles() map[string]uint64 {
	c := configs.GetConfig()
	dir := filepath.Join(c.Paths.RootDataDir, c.Paths.UserData)

	dirInfos, err := os.ReadDir(dir)
	if err != nil {
		slog.Error("unable to read files in index directory.", "err", err)
		return make(map[string]uint64)
	}

	userData := make(map[string]uint64)

	for _, fileInfo := range dirInfos {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".yaml") {
			data, err := os.ReadFile(filepath.Join(dir, fileInfo.Name()))
			if err != nil {
				slog.Error("unable to open file:", "filename", fileInfo.Name())
			}
			var userRecord UserRecord
			if err := yaml.Unmarshal(data, &userRecord); err != nil {
				slog.Error("unable to deserialize file:", "filename", fileInfo.Name())
			}
			userData[userRecord.Username] = userRecord.Id
		}
	}
	return userData
}
