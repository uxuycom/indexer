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

import "time"

type ChainGroup string

const (
	EvmChainGroup ChainGroup = "evm"
	BtcChainGroup ChainGroup = "btc"
)

const (
	ChainBTC  string = "btc"
	ChainAVAX string = "avalanche"
)

type ChainInfo struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	ChainId    int64     `json:"chain_id" gorm:"column:chain_id"`
	Chain      string    `json:"chain" gorm:"column:chain"`
	OuterChain string    `json:"outer_chain" gorm:"column:outer_chain"`
	Name       string    `json:"name" gorm:"column:name"`
	Logo       string    `json:"logo" gorm:"column:logo"`
	NetworkId  int64     `json:"network_id" gorm:"column:network_id"`
	Ext        string    `json:"ext" gorm:"column:ext"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`
}

type ChainInfoExt struct {
	TickCount    int64 `json:"tick_count"`
	AddressCount int64 `json:"address_count"`
	DeployCount  int64 `json:"deploy_count"`
	MintCount    int64 `json:"mint_count"`
	ChainInfo
}

func (ChainInfo) TableName() string {
	return "chain_info"
}
