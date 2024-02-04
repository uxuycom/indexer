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
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/uxuycom/indexer/config"
	"github.com/uxuycom/indexer/jsonrpc"
	"github.com/uxuycom/indexer/storage"
	"github.com/uxuycom/indexer/xylog"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	cfg        config.JsonRcpConfig
	flagConfig string
)

func main() {

	// init args
	initArgs()

	config.LoadJsonRpcConfig(&cfg, flagConfig)

	fmt.Printf("start with config = %v", cfg)
	logLevel, _ := logrus.ParseLevel(cfg.LogLevel)
	xylog.InitLog(logLevel, cfg.LogPath)

	//db client
	dbc, err := storage.NewDbClient(&cfg.Database)
	if err != nil {
		log.Fatalf("initialize db client err:%v", err)
		return
	}
	//init server
	server, err := jsonrpc.NewRPCServer(dbc, cfg.CacheStore)
	if err != nil {
		log.Fatalf("server init err[%v]", err)
	}

	//start server
	server.Start()

	// openapi
	jsonrpc.CreateOpenApi()

	//register terminate signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-server.RequestedProcessShutdown():
		log.Println("server quit with command")
		return
	case sig := <-signalCh:
		log.Println("server quit with signal:", sig.String())
		return
	}
}

func initArgs() {
	flag.StringVar(&flagConfig, "config", "config_jsonrpc.json", "config file")
	flag.Parse()
}
