package main

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"open-indexer/client"
	"open-indexer/devents"
	"open-indexer/protocol"
	"open-indexer/xylog"

	"net/http"
	_ "net/http/pprof"
	"open-indexer/config"
	"open-indexer/dcache"
	"open-indexer/explorer"
	"open-indexer/storage"
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
	if cfg.ProfileEnabled {
		go func() {
			_ = http.ListenAndServe("0.0.0.0:6060", nil)
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
	flag.StringVar(&flagConfig, "config", "config.json", "config file")
	flag.Parse()
}
