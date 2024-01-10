package devents

import (
	"github.com/shopspring/decimal"
	"open-indexer/client/xycommon"
)

const (
	OperateDeploy   string = "deploy"
	OperateMint     string = "mint"
	OperateTransfer string = "transfer"
	OperateList     string = "list"
	OperateExchange string = "exchange"
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

type TxResult struct {
	MD       *MetaData
	Block    *xycommon.RpcBlock
	Tx       *xycommon.RpcTransaction
	Mint     *Mint
	Deploy   *Deploy
	Transfer *Transfer
}
