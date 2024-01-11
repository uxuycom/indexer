package storage

import (
	"errors"
	"github.com/ethereum/go-ethereum/log"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"open-indexer/utils"
)

func NewSqliteClient(cfg *utils.DatabaseConfig, gormCfg *gorm.Config) (*DBClient, error) {
	if cfg == nil {
		return nil, errors.New("invalid configuration file")
	}
	if gormCfg == nil {
		return nil, errors.New("invalid configuration file")
	}
	db, err := gorm.Open(sqlite.Open(cfg.Dsn), gormCfg)
	if err != nil {
		log.Error("connect to sqlite failed", "err", err)
		return nil, err
	}

	conn := &DBClient{
		SqlDB: db,
	}

	log.Info("connect to sqlite success")
	return conn, nil
}
