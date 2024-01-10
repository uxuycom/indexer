package model

import "time"

type BlockStatus struct {
	Chain       string    `json:"chain" gorm:"column:chain"`               // chain name
	BlockHash   string    `json:"block_hash" gorm:"column:block_hash"`     // block hash
	BlockNumber uint64    `json:"block_number" gorm:"column:block_number"` // block height
	BlockTime   time.Time `json:"block_time" gorm:"column:block_time"`     // block time
}

func (BlockStatus) TableName() string {
	return "block"
}
