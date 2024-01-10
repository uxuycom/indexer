package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	UTXOStatusUnspent = 1
	UTXOStatusSpent   = 2
)

type Balances struct {
	ID        uint64          `gorm:"primaryKey" json:"id"`
	SID       uint64          `json:"sid"  gorm:"column:sid"`
	Chain     string          `json:"chain" gorm:"column:chain"`
	Protocol  string          `json:"protocol" gorm:"column:protocol"`
	Address   string          `json:"address" gorm:"column:address"`
	Tick      string          `json:"tick" gorm:"column:tick"`
	Available decimal.Decimal `json:"available" gorm:"column:available;type:decimal(36,18)"` // available balance = overall balance - transferable balance
	Balance   decimal.Decimal `json:"balance" gorm:"column:balance;type:decimal(36,18)"`     // overall balance
	CreatedAt time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

func (Balances) TableName() string {
	return "balances"
}

type UTXO struct {
	ID        uint64          `gorm:"primaryKey" json:"id"`
	Sn        string          `json:"sn" gorm:"column:sn"`
	Chain     string          `json:"chain" gorm:"column:chain"`
	Protocol  string          `json:"protocol" gorm:"column:protocol"`
	Address   string          `json:"address" gorm:"column:address"`
	Tick      string          `json:"tick" gorm:"column:tick"`
	Amount    decimal.Decimal `json:"amount" gorm:"column:amount;type:decimal(36,18)"` // amount
	RootHash  string          `json:"root_hash" gorm:"column:root_hash"`
	TxHash    string          `json:"tx_hash" gorm:"column:tx_hash"`
	Status    int8            `json:"status" gorm:"column:status"` // tx status
	CreatedAt time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

func (UTXO) TableName() string {
	return "utxos"
}

type BalanceInscription struct {
	Chain        string          `json:"chain"`
	Protocol     string          `json:"protocol"`
	Tick         string          `json:"tick"`
	Address      string          `json:"address"`
	Balance      decimal.Decimal `json:"balance"`
	DeployHash   string          `json:"deploy_hash"`
	TransferType int8            `json:"transfer_type"`
}
