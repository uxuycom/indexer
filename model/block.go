package model

import (
	"time"
)

type Block struct {
	ChainId     int64     `json:"chain_id" gorm:"column:chain_id"`
	Chain       string    `json:"chain" gorm:"column:chain"`
	BlockHash   string    `json:"block_hash" gorm:"column:block_hash"`
	BlockNumber string    `json:"block_number" gorm:"column:block_number"`
	BlockTime   time.Time `json:"block_time" gorm:"column:block_time"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Block) TableName() string {
	return "block"
}
