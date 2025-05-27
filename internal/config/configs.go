package configs

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	configData Config = Config{}
	configLock sync.RWMutex
)

type Config struct {
	Server  Server `yaml:"server"`
	Paths   Paths  `yaml:"paths"`
	Core    Core   `yaml:"core"`
	Logging Logs   `yaml:"logging"`
}

func GetConfig() Config {
	configLock.RLock()
	defer configLock.RUnlock()

	return configData
}

// Loads the configurations from the config.yaml
func LoadConfig(configPath string) (*Config, error) {

	if configPath == `` {
		configPath = `_data/config.yaml`
	}

	data, err := os.ReadFile(configPath)
	//Just puke, I dont want to run w/out config
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse the configuration file %w", err)
	}
	configLock.Lock()
	defer configLock.Unlock()
	configData = config

	return &config, nil
}
