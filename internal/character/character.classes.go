package character

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	configs "tektmud/internal/config"
	"tektmud/internal/logger"

	"gopkg.in/yaml.v3"
)

var (
	classesById   map[int]*CharacterClass    = make(map[int]*CharacterClass)
	classesByName map[string]*CharacterClass = make(map[string]*CharacterClass)
)

type CharacterClass struct {
	Id   int    `yaml:"id"`
	Name string `yaml:"name"`
	//Description string `yaml:"description"`
}

func InitializeClassData() error {
	c := configs.GetConfig()
	filePath := filepath.Join(c.Paths.RootDataDir, c.Paths.Classes)

	dirEntries, err := os.ReadDir(filePath)
	if err != nil {
		return fmt.Errorf("failed to read class data directory %s, %w", filePath, err)
	}

	for _, file := range dirEntries {
		if filepath.Ext(file.Name()) == ".yaml" {
			err := loadClass(filepath.Join(filePath, file.Name()))
			if err != nil {
				logger.Error("error loading class file", "file", file.Name(), "err", err)
			}
		}
	}

	return nil
}

func loadClass(classFile string) error {
	data, err := os.ReadFile(classFile)
	if err != nil {
		return fmt.Errorf("failed to read class file: %w", err)
	}

	var cc CharacterClass
	if err := yaml.Unmarshal(data, &cc); err != nil {
		return fmt.Errorf("failed to parse class file: %w", err)
	}

	mu.Lock()
	classesById[cc.Id] = &cc
	classesByName[strings.ToLower(cc.Name)] = &cc
	mu.Unlock()
	return nil
}

func GetClassById(id int) *CharacterClass {
	c, exists := classesById[id]
	if !exists {
		return nil
	}
	return c
}

func GetClassNameById(id int) string {
	c, exists := classesById[id]
	if !exists {
		return "Unknown ClassId"
	}
	return c.Name
}

func GetClassByName(name string) *CharacterClass {
	c, exists := classesByName[strings.ToLower(name)]
	if !exists {
		return nil
	}
	return c
}
