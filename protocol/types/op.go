package types

import (
	"open-indexer/client/xycommon"
	"open-indexer/devents"
	"open-indexer/xyerrors"
)

type IProtocol interface {
	Parse(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError)
}

const (
	BRC20Protocol = "brc-20"
	ASC20Protocol = "asc-20"
	BSC20Protocol = "bsc-20"
	PRC20Protocol = "prc-20"
)
