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

package brc20

import (
	"encoding/json"
	"fmt"
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/dcache"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/protocol/common"
	"github.com/uxuycom/indexer/protocol/types"
	"github.com/uxuycom/indexer/xyerrors"
	"github.com/uxuycom/indexer/xylog"
	"strings"
)

type Protocol struct {
	*common.Protocol
	cache *dcache.Manager
}

func NewProtocol(cache *dcache.Manager) *Protocol {
	return &Protocol{
		Protocol: common.NewProtocol(cache),
		cache:    cache,
	}
}

func (p *Protocol) Parse(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError) {
	switch md.Operate {
	case devents.OperateDeploy:
		return p.Deploy(block, tx, md)

	case devents.OperateMint:
		return p.Mint(block, tx, md)

	case devents.OperateTransfer:
		if tx.To != "" {
			return p.Transfer(block, tx, md)
		}
		md.Operate = devents.OperateInscribeTransfer
		return p.InscribeTransfer(block, tx, md)
	}
	return nil, nil
}

func ParseMetaData(chain string, tx *xycommon.RpcTransaction) (*devents.MetaData, error) {
	proto := &devents.MetaData{}
	if err := json.Unmarshal([]byte(tx.Input), proto); err != nil {
		return nil, fmt.Errorf("tx input data parsed failed, data[%s], err[%v]", tx.Input, err)
	}

	if proto.Protocol != types.BRC20Protocol {
		xylog.Logger.Infof("protocol <> brc-20 & ignored, protocol[%s]", proto.Protocol)
		return nil, nil
	}

	if len(proto.Tick) != 4 {
		xylog.Logger.Infof("protocol tick length <> 4 & ignored, tick[%s]", proto.Tick)
		return nil, nil
	}

	switch proto.Operate {
	case devents.OperateDeploy, devents.OperateMint, devents.OperateTransfer:
		return &devents.MetaData{
			Chain:    chain,
			Protocol: proto.Protocol,
			Tick:     strings.ToLower(proto.Tick),
			Operate:  proto.Operate,
			Data:     tx.Input,
		}, nil
	}
	xylog.Logger.Infof("protocol operate invalid & ignored, operation[%s]", proto.Operate)
	return nil, nil
}
