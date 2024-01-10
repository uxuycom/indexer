package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"log"
	"open-indexer/config"
	"open-indexer/jsonrpc"
	"open-indexer/storage"
	"open-indexer/xylog"
	"os"
	"os/signal"
	"syscall"
)

var (
	cfg        config.Config
	flagConfig string
)

func main() {

	// init args
	initArgs()

	config.LoadConfig(&cfg, flagConfig)

	logLevel, _ := logrus.ParseLevel(cfg.LogLevel)
	xylog.InitLog(logLevel, cfg.LogPath)

	//db client
	dbc, err := storage.NewDbClient(&cfg.Database)
	if err != nil {
		log.Fatalf("initialize db client err:%v", err)
		return
	}

	//init server
	server, err := jsonrpc.NewRPCServer(dbc)
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
	flag.StringVar(&flagConfig, "config", "config.json", "config file")
	flag.Parse()
}
