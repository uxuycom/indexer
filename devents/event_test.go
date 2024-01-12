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

package devents

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/uxuycom/indexer/config"
	"github.com/uxuycom/indexer/storage"
	"os"
	"path/filepath"
	"testing"
)

func LoadConfig(cfg *config.Config, filePath string) error {
	// Default config.
	configFileName := "../config.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}

	configFileName, _ = filepath.Abs(configFileName)

	if filePath != "" {
		configFileName = filePath
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		return fmt.Errorf("file open error: %s", err.Error())
	}
	defer func() {
		_ = configFile.Close()
	}()

	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		return fmt.Errorf("config error: %s", err.Error())
	}
	return nil
}

func TestDBLockGetAndRelease(t *testing.T) {
	// testing mysql client variable
	var flagConfig string
	flag.StringVar(&flagConfig, "config", "../config.json", "config file")
	flag.Parse()

	var cfg config.Config
	if err := LoadConfig(&cfg, flagConfig); err != nil {
		t.Logf("load config failed & ignore this test case. err:%v", err)
		return
	}

	dbClient, err := storage.NewDbClient(&cfg.Database)
	if err != nil {
		t.Log("init db client failed & ignore this test case")
		return
	}

	// lock
	ok, err := dbClient.GetLock()
	if err != nil {
		t.Log("get lock failed & ignore this test case")
		return
	}
	assert.Truef(t, ok, "get lock should success but failed")

	for i := 0; i < 10; i++ {
		dbClientN, errN := storage.NewDbClient(&cfg.Database)
		if errN != nil {
			t.Log("init db client failed & ignore this test case:", i)
			continue
		}

		okN, errN := dbClientN.GetLock()
		if errN != nil {
			t.Log("get lock error & ignore this test case:", i)
			continue
		}
		assert.Falsef(t, okN, "get lock should failed but success")
	}

	// release
	cnt, err := dbClient.ReleaseLock()
	if err != nil {
		t.Log("release lock failed & ignore this test case")
		return
	}
	assert.Equal(t, int64(1), cnt, "return cnt should be 1")

	// release again
	cnt, err = dbClient.ReleaseLock()
	if err != nil {
		t.Log("release lock failed & ignore this test case")
		return
	}
	assert.Equal(t, int64(0), cnt, "return cnt should be 0")
}
