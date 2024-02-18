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

package main

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/uxuycom/indexer/client"
	"github.com/uxuycom/indexer/config"
	"github.com/uxuycom/indexer/dcache"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/explorer"
	"github.com/uxuycom/indexer/protocol"
	"github.com/uxuycom/indexer/storage"
	"github.com/uxuycom/indexer/task"
	"github.com/uxuycom/indexer/xylog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	cfg        config.Config
	flagConfig string
)

func main() {
	// init
	runtime.GOMAXPROCS(runtime.NumCPU())

	// init args
	initArgs()

	// load configs
	config.LoadConfig(&cfg, flagConfig)

	// enable profile
	if cfg.Profile != nil && cfg.Profile.Enabled {
		go func() {
			listen := cfg.Profile.Listen
			if listen == "" {
				listen = ":6060"
			}
			if err := http.ListenAndServe(listen, nil); err != nil {
				xylog.Logger.Infof("start profile err:%v", err)
			}
		}()
	}

	// set log debug level
	if lv, err := logrus.ParseLevel(cfg.LogLevel); err == nil {
		xylog.InitLog(lv, cfg.LogPath)
	}

	dbClient, err := storage.NewDbClient(&cfg.Database)
	if err != nil {
		xylog.Logger.Fatalf("db init err:%v", err)
	}
	rpcClient, err := client.NewRPCClient(cfg.Chain.Rpc, cfg.Chain.ChainGroup)
	if err != nil {
		xylog.Logger.Fatalf("initialize rpc client err:%v", err)
	}

	dCache := dcache.NewManager(dbClient, cfg.Chain.ChainName)

	// init task
	task.InitTask(dbClient)

	// init protocols
	protocol.InitProtocols(dCache)

	// Listen for SIGINT and SIGTERM signals
	quit := make(chan os.Signal, 1)
	dEvent := devents.NewDEvents(context.TODO(), dbClient)
	exp := explorer.NewExplorer(rpcClient, dbClient, &cfg, dCache, dEvent, quit)
	go exp.Scan()
	go exp.Index()
	go exp.FlushDB()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// notify service stopped
	exp.Stop()
	xylog.Logger.Infof("service stopped")
}

func initArgs() {

	pflag.StringVarP(&flagConfig, "config", "c", "config.json", "config file")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
}
