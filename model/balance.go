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
	Available decimal.Decimal `json:"available" gorm:"column:available;type:decimal(38,18)"` // available balance = overall balance - transferable balance
	Balance   decimal.Decimal `json:"balance" gorm:"column:balance;type:decimal(38,18)"`     // overall balance
	CreatedAt time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

func (Balances) TableName() string {
	return "balances"
}

type UTXO struct {
	ID        uint64          `gorm:"primaryKey" json:"id"`
	SN        string          `json:"sn" gorm:"column:sn"`
	Chain     string          `json:"chain" gorm:"column:chain"`
	Protocol  string          `json:"protocol" gorm:"column:protocol"`
	Address   string          `json:"address" gorm:"column:address"`
	Tick      string          `json:"tick" gorm:"column:tick"`
	Amount    decimal.Decimal `json:"amount" gorm:"column:amount;type:decimal(38,18)"` // amount
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
