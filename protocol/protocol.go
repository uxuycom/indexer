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

package protocol

import (
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/config"
	"github.com/uxuycom/indexer/dcache"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/model"
	"github.com/uxuycom/indexer/protocol/avax/asc20"
	btcBrc20 "github.com/uxuycom/indexer/protocol/btc/brc20"
	"github.com/uxuycom/indexer/protocol/evm/brc20"
	"github.com/uxuycom/indexer/protocol/evm/erc20"
	"github.com/uxuycom/indexer/protocol/types"
	"github.com/uxuycom/indexer/storage"
	"github.com/uxuycom/indexer/xylog"
)

var (
	BTCBrc20Protocol *btcBrc20.Protocol
	EvmAsc20Protocol *asc20.Protocol
	EvmBrc20Protocol *brc20.Protocol
	EvmErc20Protocol *erc20.Protocol
)

func InitProtocols(cache *dcache.Manager) {
	BTCBrc20Protocol = btcBrc20.NewProtocol(cache)
	EvmBrc20Protocol = brc20.NewProtocol(cache)
	EvmAsc20Protocol = asc20.NewProtocol(cache)
	EvmErc20Protocol = erc20.NewProtocol(cache)
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
	case types.ERC20Protocol:
		return EvmErc20Protocol, md
	default:
		return EvmBrc20Protocol, md
	}
}

func GetOperateByTxInput(chain, inputData string, db *storage.DBClient) *devents.MetaData {
	md, _ := ParseMetaData(chain, &xycommon.RpcTransaction{Input: inputData})
	return md
}
