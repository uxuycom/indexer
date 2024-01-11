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

package utils

import (
	"github.com/uxuycom/indexer/model"
	"math/big"
)

// ServerConfig Config
type ServerConfig struct {
	//Port       string `json:"port"`
	FromBlock         uint64 `json:"from_block"`
	ScanLimit         uint64 `json:"scan_limit"`
	DelayedScanNumber uint64 `json:"delayed_scan_number"`
}

type SqliteConfig struct {
	Database string `json:"database"`
}

type ChainConfig struct {
	ChainName  string           `json:"chain_name"`
	Rpc        string           `json:"rpc"`
	UserName   string           `json:"username"`
	PassWord   string           `json:"password"`
	ChainGroup model.ChainGroup `json:"chain_group"`
}

type IndexFilter struct {
	Whitelist *struct {
		Ticks     []string `json:"ticks"`
		Protocols []string `json:"protocols"`
	} `json:"whitelist"`
	EventTopics []string `json:"event_topics"`
}

// DatabaseConfig database config
type DatabaseConfig struct {
	Type      string `json:"type"`
	Dsn       string `json:"dsn"`
	EnableLog bool   `json:"enable_log"`
}

// RouterResult
type HttpResult struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
}

type BaseParams struct {
	P  string `json:"p"`
	Op string `json:"op"`
}

type NewParams struct {
	P              string `json:"p"`
	Op             string `json:"op"`
	Tick           string `json:"tick"`
	Max            string `json:"max"`
	Amt            string `json:"amt"`
	Lim            string `json:"lim"`
	Dec            int64  `json:"dec"`
	Burn           string `json:"burn"`
	Func           string `json:"func"`
	ReceiveAddress string `json:"receive_address"`
	ToAddress      string `json:"to_address"`
	RateFee        string `json:"rate_fee"`
	Repeat         int64  `json:"repeat"`
}

type Drc20Params struct {
	Tick      string `json:"tick"`
	Limit     uint64 `json:"limit"`
	OffSet    uint64 `json:"offset"`
	Completed uint64 `json:"completed"`
}

type SwapParams struct {
	Op            string `json:"op"`
	Tick0         string `json:"tick0"`
	Tick1         string `json:"tick1"`
	Amt0          string `json:"amt0"`
	Amt1          string `json:"amt1"`
	Amt0Min       string `json:"amt0_min"`
	Amt1Min       string `json:"amt1_min"`
	Liquidity     string `json:"liquidity"`
	Path          string `json:"path"`
	HolderAddress string `json:"holder_address"`
}

type WDogeParams struct {
	Op            string `json:"op"`
	Tick          string `json:"tick"`
	Amt           string `json:"amt"`
	HolderAddress string `json:"holder_address"`
}

// Models
type Cardinals struct {
	OrderId        string   `json:"order_id"`
	P              string   `json:"p"`
	Op             string   `json:"op"`
	Tick           string   `json:"tick"`
	Amt            *big.Int `json:"amt"`
	Max            *big.Int `json:"max"`
	Lim            *big.Int `json:"lim"`
	Dec            int64    `json:"dec"`
	Burn           string   `json:"burn"`
	Func           string   `json:"func"`
	Repeat         int64    `json:"repeat"`
	Drc20TxHash    string   `json:"drc20_tx_hash"`
	BlockNumber    int64    `json:"block_number"`
	BlockHash      string   `json:"block_hash"`
	ReceiveAddress string   `json:"receive_address"`
	ToAddress      string   `json:"to_address"`
	FeeAddress     string   `json:"fee_address"`
	OrderStatus    int64    `json:"order_status"`
	ErrInfo        string   `json:"err_info"`
	CreateDate     string   `json:"create_date"`
}

// SWAP
type SwapInfo struct {
	OrderId         string   `json:"order_id"`
	Op              string   `json:"op"`
	Tick            string   `json:"tick"`
	Tick0           string   `json:"tick0"`
	Tick1           string   `json:"tick1"`
	Amt0            *big.Int `json:"amt0"`
	Amt1            *big.Int `json:"amt1"`
	Amt0Min         *big.Int `json:"amt0_min"`
	Amt1Min         *big.Int `json:"amt1_min"`
	Amt0Out         *big.Int `json:"amt0_out"`
	Amt1Out         *big.Int `json:"amt1_out"`
	Path            []string `json:"path"`
	Liquidity       *big.Int `json:"liquidity"`
	HolderAddress   string   `json:"holder_address"`
	FeeAddress      string   `json:"fee_address"`
	SwapTxHash      string   `json:"swap_tx_hash"`
	SwapBlockNumber int64    `json:"swap_block_number"`
	SwapBlockHash   string   `json:"swap_block_hash"`
	OrderStatus     int64    `json:"order_status"`
	UpdateDate      string   `json:"update_date"`
	CreateDate      string   `json:"create_date"`
}

// swap_liquidity
type SwapLiquidity struct {
	Tick            string   `json:"tick"`
	Tick0           string   `json:"tick0"`
	Tick1           string   `json:"tick1"`
	Amt0            *big.Int `json:"amt0"`
	Amt1            *big.Int `json:"amt1"`
	Path            string   `json:"path"`
	LiquidityTotal  *big.Int `json:"liquidity_total"`
	ReservesAddress string   `json:"reserves_address"`
	HolderAddress   string   `json:"holder_address"`
}

// WDOGE
type WDogeInfo struct {
	OrderId          string   `json:"order_id"`
	Op               string   `json:"op"`
	Tick             string   `json:"tick"`
	Amt              *big.Int `json:"amt"`
	HolderAddress    string   `json:"holder_address"`
	FeeAddress       string   `json:"fee_address"`
	WDogeTxHash      string   `json:"wdoge_tx_hash"`
	WDogeBlockNumber int64    `json:"wdoge_block_number"`
	WDogeBlockHash   string   `json:"wdoge_block_hash"`
	UpdateDate       string   `json:"update_date"`
	CreateDate       string   `json:"create_date"`
}

// cardinals_revert
type CardinalsRevert struct {
	Tick        string   `json:"tick"`
	FromAddress string   `json:"from_address"`
	ToAddress   string   `json:"to_address"`
	Amt         *big.Int `json:"amt"`
	BlockNumber int64    `json:"block_number"`
}
