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

// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC commands that are supported by
// a chain server.

package jsonrpc

import "math/big"

// EmptyCmd defines the empty JSON-RPC command.
type EmptyCmd struct{}

// FindAllInscriptionsCmd defines the inscription JSON-RPC command.
type FindAllInscriptionsCmd struct {
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Chain    string `json:"chain"`
	Protocol string `json:"protocol"`
	Tick     string `json:"tick"`
	DeployBy string `json:"deploy_by"`
	Sort     int    `json:"sort"`
}

type FindAllInscriptionsResponse struct {
	Inscriptions interface{} `json:"inscriptions"`
	Total        int64       `json:"total"`
	Limit        int         `json:"limit"`
	Offset       int         `json:"offset"`
}

type InscriptionInfo struct {
	Chain        string `json:"chain"`
	Protocol     string `json:"protocol"`
	Tick         string `json:"tick"`
	Name         string `json:"name"`
	LimitPerMint string `json:"limit_per_mint"`
	DeployBy     string `json:"deploy_by"`
	TotalSupply  string `json:"total_supply"`
	DeployHash   string `json:"deploy_hash"`
	DeployTime   uint32 `json:"deploy_time"`
	TransferType int8   `json:"transfer_type"`
	CreatedAt    uint32 `json:"created_at"`
	UpdatedAt    uint32 `json:"updated_at"`
	Decimals     int8   `json:"decimals"`
}

// FindInscriptionTickCmd defines the inscription JSON-RPC command.
type FindInscriptionTickCmd struct {
	Chain    string
	Protocol string
	Tick     string
}

type FindInscriptionTickResponse struct {
	Tick interface{} `json:"tick"`
}

// FindUserTransactionsCmd defines the inscription JSON-RPC command.
type FindUserTransactionsCmd struct {
	Limit    int
	Offset   int
	Address  string
	Chain    string
	Protocol string
	Tick     string
	Event    int8
}

type AddressTransaction struct {
	Chain     string `json:"chain"`
	Protocol  string `json:"protocol"`
	Tick      string `json:"tick"`
	Address   string `json:"address"`
	TxHash    string `json:"tx_hash"`
	Amount    string `json:"amount"`
	Event     int8   `json:"event"`
	Operate   string `json:"operate"`
	Status    int8   `json:"status"`
	CreatedAt uint32 `json:"created_at"`
	UpdatedAt uint32 `json:"updated_at"`
}

type FindUserTransactionsResponse struct {
	Transactions interface{} `json:"transactions"`
	Total        int64       `json:"total"`
	Limit        int         `json:"limit"`
	Offset       int         `json:"offset"`
}

// FindUserBalancesCmd defines the inscription JSON-RPC command.
type FindUserBalancesCmd struct {
	Limit    int
	Offset   int
	Address  string
	Chain    string
	Protocol string
	Tick     string
}

type FindUserBalanceCmd struct {
	Address  string
	Chain    string
	Protocol string
	Tick     string
}

type BalanceInfo struct {
	Chain        string `json:"chain"`
	Protocol     string `json:"protocol"`
	Tick         string `json:"tick"`
	Address      string `json:"address"`
	Balance      string `json:"balance"`
	DeployHash   string `json:"deploy_hash"`
	TransferType int8   `json:"transfer_type"`
}

type BalanceBrief struct {
	Tick         string       `json:"tick"`
	Balance      string       `json:"balance"`
	TransferType int8         `json:"transfer_type"`
	Utxos        []*UTXOBrief `json:"utxos,omitempty"`
}

type UTXOBrief struct {
	Tick     string `json:"tick"`
	Amount   string `json:"amount"`
	RootHash string `json:"root_hash"`
}

type FindUserBalancesResponse struct {
	Inscriptions interface{} `json:"inscriptions"`
	Total        int64       `json:"total"`
	Limit        int         `json:"limit"`
	Offset       int         `json:"offset"`
}

type FindUserBalanceResponse struct {
	Balance interface{} `json:"balance"`
}

type FindTickHoldersCmd struct {
	Limit    int
	Offset   int
	Chain    string
	Protocol string
	Tick     string
}

type FindTickHoldersResponse struct {
	Holders interface{} `json:"holders"`
	Total   int64       `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
}

type LastBlockNumberCmd struct {
	Chain string
}

type LastBlockNumberResponse struct {
	BlockNumber *big.Int `json:"block_number"`
}

type TxOperateCmd struct {
	Chain     string
	InputData string
}

type TxOperateResponse struct {
	Operate    string `json:"operate"`
	Protocol   string `json:"protocol"`
	Tick       string `json:"tick"`
	DeployHash string `json:"deploy_hash"`
}

type GetTxByHashCmd struct {
	Chain  string
	TxHash string
}

type TransactionInfo struct {
	Protocol string `json:"protocol"`
	Tick     string `json:"tick"`
	DeployBy string `json:"deploy_by"`
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
}

type GetTxByHashResponse struct {
	IsInscription bool             `json:"is_inscription"`
	Transaction   *TransactionInfo `json:"transaction,omitempty"`
}

func init() {
	// No special flags for commands in this file.
	flags := UsageFlag(0)

	MustRegisterCmd("inscription.All", (*FindAllInscriptionsCmd)(nil), flags)
	MustRegisterCmd("inscription.Tick", (*FindInscriptionTickCmd)(nil), flags)
	MustRegisterCmd("address.Transactions", (*FindUserTransactionsCmd)(nil), flags)
	MustRegisterCmd("address.Balances", (*FindUserBalancesCmd)(nil), flags)
	MustRegisterCmd("address.Balance", (*FindUserBalanceCmd)(nil), flags)
	MustRegisterCmd("tick.Holders", (*FindTickHoldersCmd)(nil), flags)
	MustRegisterCmd("block.LastNumber", (*LastBlockNumberCmd)(nil), flags)
	MustRegisterCmd("tool.InscriptionTxOperate", (*TxOperateCmd)(nil), flags)
	MustRegisterCmd("transaction.Info", (*GetTxByHashCmd)(nil), flags)
}
