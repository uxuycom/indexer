package config

import (
	"encoding/json"
	"log"
	"open-indexer/utils"
	"os"
	"path/filepath"
)

type Config struct {
	Server         utils.ServerConfig   `json:"server"`
	Chain          utils.ChainConfig    `json:"chain"`
	LogLevel       string               `json:"log_level"`
	LogPath        string               `json:"log_path"`
	Filters        *utils.IndexFilter   `json:"filters"`
	Ticks          []string             `json:"tick_whitelist"`
	Database       utils.DatabaseConfig `json:"database"`
	ProfileEnabled bool                 `json:"profile_enabled"`
}

func LoadConfig(cfg *Config, filep string) {

	// Default config.
	configFileName := "config.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}

	configFileName, _ = filepath.Abs(configFileName)
	log.Printf("Loading config: %v", configFileName)

	if filep != "" {
		configFileName = filep
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("File error: ", err.Error())
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		log.Fatal("Config error: ", err.Error())
	}
}

func (cfg *Config) GetConfig() *Config {
	return cfg
}
