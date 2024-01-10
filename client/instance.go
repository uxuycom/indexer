package client

import (
	"open-indexer/client/btc"
	"open-indexer/client/evm"
	"open-indexer/client/xycommon"
	"open-indexer/model"
)

func NewRPCClient(rpc string, proto model.ChainGroup) (xycommon.IRPCClient, error) {
	if proto == model.BtcChainGroup {
		return btc.NewClient(rpc)
	}
	return evm.Dial(rpc)
}
