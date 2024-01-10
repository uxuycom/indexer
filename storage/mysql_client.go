package storage

import (
	"github.com/ethereum/go-ethereum/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"open-indexer/utils"
)

func NewMysqlClient(cfg *utils.DatabaseConfig, gormCfg *gorm.Config) (*DBClient, error) {
	db, err := gorm.Open(mysql.Open(cfg.Dsn), gormCfg)
	if err != nil {
		log.Error("connect to mysql failed", "err", err)
		return nil, err
	}
	conn := &DBClient{
		SqlDB: db,
	}
	return conn, nil
}
