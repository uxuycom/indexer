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

package asc20

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/dcache"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/protocol/common"
	"github.com/uxuycom/indexer/protocol/types"
	"github.com/uxuycom/indexer/xyerrors"
	"github.com/uxuycom/indexer/xylog"
	"strings"
	"sync"
)

type Protocol struct {
	common *common.Protocol
	cache  *dcache.Manager
	ticks  *sync.Map // record deploy tx -> tick code
}

var ParsedABI abi.ABI

func NewProtocol(cache *dcache.Manager) *Protocol {
	return &Protocol{
		common: common.NewProtocol(cache),
		cache:  cache,
		ticks:  &sync.Map{},
	}
}

func (p *Protocol) Parse(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError) {
	switch md.Operate {
	case devents.OperateList:
		return p.List(block, tx, md)

	case devents.OperateExchange:
		return p.Exchange(block, tx, md)
	}
	return p.common.Parse(block, tx, md)
}
func (p *Protocol) MaxDateLength() int {
	return types.MaxDataLengthASC20Protocol
}

func ParseMetaDataByEventLogs(chain string, tx *xycommon.RpcTransaction) (*devents.MetaData, error) {
	for _, event := range tx.Events {
		if len(event.Topics) < 1 {
			continue
		}

		topic := event.Topics[0].String()
		if topic == EventTopicHashExchange || topic == EventTopicHashExchange2 {
			return &devents.MetaData{
				Chain:    chain,
				Protocol: types.ASC20Protocol,
				Operate:  devents.OperateExchange,
			}, nil
		}
	}
	return nil, nil
}

func init() {
	var err error
	ParsedABI, err = abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		xylog.Logger.Fatalf("asc20 abi decode err:%v", err)
	}
}
