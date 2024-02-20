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
	TransferTypeHash    = 1
	TransferTypeBalance = 2
)

type ChainInfo struct {
	ChainId         string `json:"chain_id"`
	ChainName       string `json:"chain_name"`
	ProtocolOnChain string `json:"protocol_on_chain"`
}

type Inscriptions struct {
	ID           uint32          `gorm:"primaryKey" json:"id"` // ID
	SID          uint32          `json:"sid"  gorm:"column:sid"`
	Chain        string          `json:"chain" gorm:"column:chain"`
	Protocol     string          `json:"protocol" gorm:"column:protocol"`
	Tick         string          `json:"tick" gorm:"column:tick"`
	Name         string          `json:"name" gorm:"column:name"`
	LimitPerMint decimal.Decimal `gorm:"column:limit_per_mint;type:decimal(38,18)" json:"limit_per_mint"`
	DeployBy     string          `json:"deploy_by" gorm:"column:deploy_by"`
	TotalSupply  decimal.Decimal `gorm:"column:total_supply;type:decimal(38,18)" json:"total_supply"`
	DeployHash   string          `json:"deploy_hash" gorm:"column:deploy_hash"`
	DeployTime   time.Time       `json:"deploy_time" gorm:"column:deploy_time"`
	TransferType int8            `json:"transfer_type" gorm:"column:transfer_type"`
	CreatedAt    time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"column:updated_at"`
	Decimals     int8            `json:"decimals" gorm:"column:decimals"`
}

func (Inscriptions) TableName() string {
	return "inscriptions"
}

// InscriptionsStats inscriptions statistics
type InscriptionsStats struct {
	ID                uint32          `gorm:"primaryKey" json:"id"`
	SID               uint32          `json:"sid"  gorm:"column:sid"`
	Chain             string          `json:"chain" gorm:"column:chain"`
	Protocol          string          `json:"protocol" gorm:"column:protocol"`
	Tick              string          `json:"tick" gorm:"column:tick"`
	Minted            decimal.Decimal `gorm:"column:minted;type:decimal(38,18)" json:"minted"`
	MintCompletedTime *time.Time      `gorm:"column:mint_completed_time" json:"mint_completed_time"`
	MintFirstBlock    uint64          `gorm:"column:mint_first_block" json:"mint_first_block"`
	MintLastBlock     uint64          `gorm:"column:mint_last_block" json:"mint_last_block"`
	LastSN            uint64          `gorm:"column:last_sn" json:"last_sn"`
	Holders           uint64          `gorm:"column:holders" json:"holders"`
	TxCnt             uint64          `gorm:"column:tx_cnt" json:"tx_cnt"`
	CreatedAt         time.Time       `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"column:updated_at" json:"updated_at"`
}

func (InscriptionsStats) TableName() string {
	return "inscriptions_stats"
}

type AllChain struct {
	Chain string `json:"chain"`
	count uint32 `json:"count"`
}

func (AllChain) TableName() string {
	return "inscriptions"
}

type InscriptionOverView struct {
	ID           uint32          `gorm:"primaryKey" json:"id"`
	Chain        string          `json:"chain" gorm:"column:chain"`
	Protocol     string          `json:"protocol" gorm:"column:protocol"`
	Tick         string          `json:"tick" gorm:"column:tick"`
	Name         string          `json:"name" gorm:"column:name"`
	LimitPerMint decimal.Decimal `gorm:"column:limit_per_mint;type:decimal(38,18)" json:"limit_per_mint"`
	DeployBy     string          `json:"deploy_by" gorm:"column:deploy_by"`
	TotalSupply  decimal.Decimal `gorm:"column:total_supply;type:decimal(38,18)" json:"total_supply"`
	DeployHash   string          `json:"deploy_hash" gorm:"column:deploy_hash"`
	DeployTime   time.Time       `json:"deploy_time" gorm:"column:deploy_time"`
	TransferType int8            `json:"transfer_type" gorm:"column:transfer_type"`
	CreatedAt    time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"column:updated_at"`
	Decimals     int8            `json:"decimals" gorm:"column:decimals"`
	Holders      uint64          `json:"holders" gorm:"column:holders"`
	Minted       decimal.Decimal `gorm:"column:minted;type:decimal(38,18)" json:"minted"`
	TxCnt        uint64          `gorm:"column:tx_cnt" json:"tx_cnt"`
}

type InscriptionBrief struct {
	Chain         string `json:"chain"`
	Protocol      string `json:"protocol"`
	Tick          string `json:"tick"`
	DeployBy      string `json:"deploy_by"`
	DeployHash    string `json:"deploy_hash"`
	TotalSupply   string `json:"total_supply"`
	MintedPercent string `json:"minted_percent"`
	LimitPerMint  string `json:"limit_per_mint"`
	Holders       uint64 `json:"holders"`
	TransferType  int8   `json:"transfer_type"`
	Status        uint32 `json:"status"`
	Minted        string `json:"minted"`
	TxCnt         uint64 `json:"tx_cnt"`
	CreatedAt     uint32 `json:"created_at"`
}

type UserInscription struct {
	Chain         string `json:"chain"`
	Protocol      string `json:"protocol"`
	Tick          string `json:"tick"`
	TotalSupply   string `json:"total_supply"`
	MintedPercent string `json:"minted_percent"`
	LimitPerMint  string `json:"limit_per_mint"`
	Holders       uint64 `json:"holders"`
	Status        uint32 `json:"status"`
	Minted        string `json:"minted"`
	CreatedAt     uint32 `json:"created_at"`
	Address       string `json:"address"`
}

const (
	MintStatusProcessing uint32 = 1
	MintStatusAllMinted  uint32 = 2
)
