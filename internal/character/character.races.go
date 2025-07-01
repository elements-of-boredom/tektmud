package character

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	configs "tektmud/internal/config"
	"tektmud/internal/logger"

	"gopkg.in/yaml.v3"
)

var (
	RacesById   map[int]*Race    = make(map[int]*Race)
	racesByName map[string]*Race = make(map[string]*Race)

	mu sync.Mutex
)

type Race struct {
	Id      int         `yaml:"id"`
	Name    string      `yaml:"name"`
	Stats   Stats       `yaml:"stats"`
	Resists Resistances `yaml:"resistances"`
	BuffIds []int       `yaml:"buff_ids"`
}

func InitializeRaceData() error {
	c := configs.GetConfig()
	filePath := filepath.Join(c.Paths.RootDataDir, c.Paths.Races)

	dirEntries, err := os.ReadDir(filePath)
	if err != nil {
		return fmt.Errorf("failed to read races data directory %s, %w", filePath, err)
	}

	for _, file := range dirEntries {
		if filepath.Ext(file.Name()) == ".yaml" {
			err := loadRace(filepath.Join(filePath, file.Name()))
			if err != nil {
				logger.Error("error loading race file", "file", file.Name(), "err", err)
			}
		}
	}

	return nil
}

func loadRace(raceFile string) error {
	data, err := os.ReadFile(raceFile)
	if err != nil {
		return fmt.Errorf("failed to read race file: %w", err)
	}

	var race Race
	if err := yaml.Unmarshal(data, &race); err != nil {
		return fmt.Errorf("failed to parse user file: %w", err)
	}

	mu.Lock()
	RacesById[race.Id] = &race
	racesByName[strings.ToLower(race.Name)] = &race
	mu.Unlock()
	return nil
}

// Normalizes the race name to avoid casing issues
func GetRaceByName(name string) *Race {
	race, exists := racesByName[strings.ToLower(name)]
	if !exists {
		return nil
	}
	return race
}

func GetRaceNameById(id int) string {
	race, exists := RacesById[id]
	if !exists {
		return "Invalid RaceId"
	}
	return race.Name
}

func GetStatsForRace(name string) *Stats {
	race := GetRaceByName(name)
	if race == nil {
		return nil
	}
	return &race.Stats
}
