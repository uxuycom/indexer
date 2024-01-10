package protocol

import (
	"open-indexer/client/xycommon"
	"open-indexer/config"
	"open-indexer/dcache"
	"open-indexer/devents"
	"open-indexer/model"
	"open-indexer/protocol/avax/asc20"
	btcBrc20 "open-indexer/protocol/btc/brc20"
	"open-indexer/protocol/evm/brc20"
	"open-indexer/protocol/types"
	"open-indexer/storage"
	"open-indexer/xylog"
)

var (
	BTCBrc20Protocol *btcBrc20.Protocol
	EvmAsc20Protocol *asc20.Protocol
	EvmBrc20Protocol *brc20.Protocol
)

func InitProtocols(cache *dcache.Manager) {
	BTCBrc20Protocol = btcBrc20.NewProtocol(cache)
	EvmBrc20Protocol = brc20.NewProtocol(cache)
	EvmAsc20Protocol = asc20.NewProtocol(cache)
}

func GetProtocol(cfg *config.Config, tx *xycommon.RpcTransaction) (types.IProtocol, *devents.MetaData) {
	md, err := ParseMetaData(cfg.Chain.ChainName, tx)
	if md == nil {
		xylog.Logger.Infof("metadata parsed failed, block:%d-tx:%s, err:%v", tx.BlockNumber, tx.Hash, err)
		return nil, nil
	}

	// btc types protocols
	if cfg.Chain.ChainGroup == model.BtcChainGroup {
		switch md.Protocol {
		case types.BRC20Protocol:
			return BTCBrc20Protocol, md
		}
		return nil, nil
	}

	// default protocols: evm
	switch md.Protocol {
	case types.ASC20Protocol:
		return EvmAsc20Protocol, md
	default:
		return EvmBrc20Protocol, md
	}
}

func GetOperateByTxInput(chain, inputData string, db *storage.DBClient) *devents.MetaData {
	md, _ := ParseMetaData(chain, &xycommon.RpcTransaction{Input: inputData})
	return md
}
