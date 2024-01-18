// Copyright (c) 2023-2024 The UXUY Developer Team
// License:
// MIT License

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE

package config

import (
	"encoding/json"
	"github.com/uxuycom/indexer/model"
	"log"
	"os"
	"path/filepath"
)

type ScanConfig struct {
	StartBlock        uint64 `json:"start_block"`
	BlockBatchWorkers uint64 `json:"block_batch_workers"`
	TxBatchWorkers    uint64 `json:"tx_batch_workers"`
	DelayedBlockNum   uint64 `json:"delayed_block_num"`
}

type ChainConfig struct {
	ChainId    int              `json:"chain_id"`
	ChainName  string           `json:"chain_name"`
	ChainRPC   string           `json:"rpc"`
	OrdRpc     string           `json:"ord_rpc"`
	Testnet    bool             `json:"testnet"`
	UserName   string           `json:"username"`
	PassWord   string           `json:"password"`
	ChainGroup model.ChainGroup `json:"chain_group"`
}

type IndexFilter struct {
	Whitelist *struct {
		Ticks     []string `json:"ticks"`
		Protocols []string `json:"protocols"`
	} `json:"whitelist"`
	EventTopics []string `json:"event_topics"`
}

// DatabaseConfig database config
type DatabaseConfig struct {
	Type      string `json:"type"`
	Dsn       string `json:"dsn"`
	EnableLog bool   `json:"enable_log"`
}

type ProfileConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type Config struct {
	Scan     ScanConfig     `json:"scan"`
	Chain    *ChainConfig   `json:"chain"`
	LogLevel string         `json:"log_level"`
	LogPath  string         `json:"log_path"`
	Filters  *IndexFilter   `json:"filters"`
	Database DatabaseConfig `json:"database"`
	Profile  *ProfileConfig `json:"profile"`
}

type JsonRcpConfig struct {
	RpcListen     []string       `json:"rpclisten"`
	RpcMaxClients int64          `json:"rpcmaxclients"`
	LogLevel      string         `json:"log_level"`
	LogPath       string         `json:"log_path"`
	Database      DatabaseConfig `json:"database"`
	Profile       *ProfileConfig `json:"profile"`
	CacheStore    *CacheConfig   `json:"cache_store"`
}

type CacheConfig struct {
	Started     bool   `json:"started"`
	MaxCapacity int64  `json:"max_capacity"`
	Duration    uint32 `json:"duration"`
}

func LoadConfig(cfg *Config, filePath string) {
	// Default config.
	configFileName := "config.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}

	configFileName, _ = filepath.Abs(configFileName)
	log.Printf("Loading config: %v", configFileName)

	if filePath != "" {
		configFileName = filePath
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("File error: ", err.Error())
	}
	defer func() {
		_ = configFile.Close()
	}()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		log.Fatal("Config error: ", err.Error())
	}
}

func LoadJsonRpcConfig(cfg *JsonRcpConfig, filePath string) {
	// Default config.
	configFileName := "config_jsonrpc.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}

	configFileName, _ = filepath.Abs(configFileName)
	log.Printf("Loading config: %v", configFileName)

	if filePath != "" {
		configFileName = filePath
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("File error: ", err.Error())
	}
	defer func() {
		_ = configFile.Close()
	}()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		log.Fatal("Config error: ", err.Error())
	}
}

func (cfg *JsonRcpConfig) GetConfig() *JsonRcpConfig {
	return cfg
}

func (cfg *Config) GetConfig() *Config {
	return cfg
}
