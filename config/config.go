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
	"github.com/spf13/viper"
	"github.com/uxuycom/indexer/model"
	"log"
	"path/filepath"
)

type ScanConfig struct {
	StartBlock        uint64 `json:"start_block" mapstructure:"start_block"`
	BlockBatchWorkers uint64 `json:"block_batch_workers" mapstructure:"block_batch_workers"`
	TxBatchWorkers    uint64 `json:"tx_batch_workers" mapstructure:"tx_batch_workers"`
	DelayedBlockNum   uint64 `json:"delayed_block_num" mapstructure:"delayed_block_num"`
}

type ChainConfig struct {
	ChainName  string           `json:"chain_name" mapstructure:"chain_name"`
	Rpc        string           `json:"rpc"`
	UserName   string           `json:"username"`
	PassWord   string           `json:"password"`
	ChainGroup model.ChainGroup `json:"chain_group" mapstructure:"chain_group"`
}

type StatConfig struct {
	AddressStartId uint64 `json:"address_start_id" mapstructure:"address_start_id"`
	BalanceStartId uint64 `json:"balance_start_id" mapstructure:"balance_start_id"`
}

type IndexFilter struct {
	Whitelist *struct {
		Ticks     []string `json:"ticks"`
		Protocols []string `json:"protocols"`
	} `json:"whitelist"`
	EventTopics []string `json:"event_topics" mapstructure:"event_topics"`
}

// DatabaseConfig database config
type DatabaseConfig struct {
	Type      string `json:"type"`
	Dsn       string `json:"dsn"`
	EnableLog bool   `json:"enable_log" mapstructure:"enable_log"`
}

type ProfileConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type Config struct {
	Scan     ScanConfig     `json:"scan"`
	Chain    ChainConfig    `json:"chain"`
	LogLevel string         `json:"log_level" mapstructure:"log_level"`
	LogPath  string         `json:"log_path" mapstructure:"log_path"`
	Filters  *IndexFilter   `json:"filters"`
	Database DatabaseConfig `json:"database"`
	Profile  *ProfileConfig `json:"profile"`
	Stat     *StatConfig    `json:"stat"`
}

type RpcConfig struct {
	LogLevel     string         `json:"log_level" mapstructure:"log_level"`
	LogPath      string         `json:"log_path" mapstructure:"log_path"`
	Database     DatabaseConfig `json:"database"`
	Profile      *ProfileConfig `json:"profile"`
	CacheStore   *CacheConfig   `json:"cache_store" mapstructure:"cache_store"`
	DebugLevel   string         `json:"debug_level" mapstructure:"debug_level"`
	DisableTLS   bool           `json:"notls" description:"Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`
	RPCCert      string         `json:"rpccert" description:"File containing the certificate file"`
	RPCKey       string         `json:"rpckey" description:"File containing the certificate key"`
	RPCLimitPass string         `json:"rpclimitpass" default-mask:"-" description:"Password for limited RPC connections"`
	RPCLimitUser string         `json:"rpclimituser" description:"Username for limited RPC connections"`
	RPCListeners []string       `json:"rpclisten" mapstructure:"rpclisten" description:"Add an interface/port to listen for RPC
connections (default port: 6583, testnet: 16583)"`
	RPCMaxClients        int    `json:"rpcmaxclients" description:"Max number of RPC clients for standard connections"`
	RPCMaxConcurrentReqs int    `json:"rpcmaxconcurrentreqs" description:"Max number of concurrent RPC requests that may be processed concurrently"`
	RPCMaxWebsockets     int    `json:"rpcmaxwebsockets" description:"Max number of RPC websocket connections"`
	RPCQuirks            bool   `json:"rpcquirks" description:"Mirror some JSON-RPC quirks of Bitcoin Core -- NOTE: Discouraged unless interoperability issues need to be worked around"`
	RPCPass              string `json:"rpcpass" default-mask:"-" description:"Password for RPC connections"`
	RPCUser              string `json:"rpcuser" description:"Username for RPC connections"`
}

type CacheConfig struct {
	Started     bool   `json:"started"`
	MaxCapacity int64  `json:"max_capacity" mapstructure:"max_capacity"`
	Duration    uint32 `json:"duration"`
}

func LoadConfig(cfg *Config, configFile string) {
	UnmarshalConfig(configFile, cfg)
}

func LoadJsonRpcConfig(cfg *RpcConfig, configFile string) {
	UnmarshalConfig(configFile, cfg)
}

func UnmarshalConfig(configFile string, cfg interface{}) {
	fileName := filepath.Base(configFile)
	viper.SetConfigFile(fileName)
	viper.SetConfigType("json")

	dir := filepath.Dir(configFile)
	viper.AddConfigPath(dir)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Read file error, error:%v", err.Error())
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unmarshal config fail! error:%v ", err)
	}
	viper.WatchConfig()
}
func (cfg *RpcConfig) GetConfig() *RpcConfig {
	return cfg
}

func (cfg *Config) GetConfig() *Config {
	return cfg
}
