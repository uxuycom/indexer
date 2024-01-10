package common

import (
	"open-indexer/client/xycommon"
	"open-indexer/dcache"
	"open-indexer/devents"
	"open-indexer/xyerrors"
)

const DataPrefix = "0x646174613a"

type Protocol struct {
	cache *dcache.Manager
}

func NewProtocol(cache *dcache.Manager) *Protocol {
	return &Protocol{
		cache: cache,
	}
}

func (base *Protocol) Parse(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError) {
	switch md.Operate {
	case devents.OperateDeploy:
		return base.Deploy(block, tx, md)
	case devents.OperateMint:
		return base.Mint(block, tx, md)
	case devents.OperateTransfer:
		return base.Transfer(block, tx, md)
	}
	return nil, nil
}
