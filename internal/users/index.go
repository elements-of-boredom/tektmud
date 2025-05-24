package users

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	configs "tektmud/internal/config"
	"tektmud/internal/util"
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
	Filename   string
	headerData IndexHeader
	NextUserId uint64 //The next userId to assign (can be different than IndexHeader.RecordCount)
}

func NewUserIndex(indexName string) *UserIndex {
	c := configs.GetConfig()
	filename := util.FilePath(c.Paths.RootDataDir, `/`, c.Paths.UserData, `/`, indexName)
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

//TODO: Rebuild
//scan all files in user dir and rebuild index
