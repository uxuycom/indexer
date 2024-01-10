package storage

import (
	"github.com/ethereum/go-ethereum/log"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"open-indexer/utils"
)

func NewSqliteClient(cfg *utils.DatabaseConfig, gormCfg *gorm.Config) (*DBClient, error) {
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
