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

package devents

import (
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/client/xycommon"
)

const (
	OperateDeploy           string = "deploy"
	OperateMint             string = "mint"
	OperateTransfer         string = "transfer"
	OperateInscribeTransfer string = "inscribe_transfer"
	OperateList             string = "list"
	OperateDelist           string = "delist"
	OperateExchange         string = "exchange"
)

type MetaData struct {
	Chain    string
	Protocol string `json:"p"`
	Operate  string `json:"op"`
	Tick     string `json:"tick"`
	Data     string
}

func (original *MetaData) Copy() *MetaData {
	return &MetaData{
		Chain:    original.Chain,
		Protocol: original.Protocol,
		Operate:  original.Operate,
		Tick:     original.Tick,
		Data:     original.Data,
	}
}

type Deploy struct {
	Name      string
	MaxSupply decimal.Decimal
	MintLimit decimal.Decimal
	Decimal   int8
}

type Mint struct {
	Minter string
	Amount decimal.Decimal
	Init   bool
}

type Receive struct {
	Address string
	Amount  decimal.Decimal
	Init    bool
}

type Transfer struct {
	Sender   string
	Receives []*Receive
}

type InscribeTransfer struct {
	Address string
	Amount  decimal.Decimal
}

type TxResult struct {
	MD               *MetaData
	Block            *xycommon.RpcBlock
	Tx               *xycommon.RpcTransaction
	Mint             *Mint
	Deploy           *Deploy
	Transfer         *Transfer
	InscribeTransfer *InscribeTransfer
}
