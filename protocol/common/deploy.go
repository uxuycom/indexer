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

package common

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/xyerrors"
)

type Deploy struct {
	Tick      string          `json:"tick"`
	MaxSupply decimal.Decimal `json:"max"`
	MintLimit decimal.Decimal `json:"lim"`
	Decimal   decimal.Decimal `json:"dec"`
}

func (base *Protocol) Deploy(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError) {
	d, err := base.verifyDeploy(tx, md)
	if err != nil {
		return nil, xyerrors.ErrDataVerifiedFailed.WrapCause(err)
	}

	result := &devents.TxResult{
		MD:    md,
		Block: block,
		Tx:    tx,
		Deploy: &devents.Deploy{
			Name:      d.Tick,
			MaxSupply: d.MaxSupply,
			MintLimit: d.MintLimit,
			Decimal:   int8(d.Decimal.IntPart()),
		},
	}
	return []*devents.TxResult{result}, nil
}

func (base *Protocol) verifyDeploy(tx *xycommon.RpcTransaction, md *devents.MetaData) (*Deploy, *xyerrors.InsError) {
	// metadata protocol / tick checking
	if md.Protocol == "" || md.Tick == "" {
		return nil, xyerrors.NewInsError(-12, fmt.Sprintf("protocol[%s] / tick[%s] nil", md.Protocol, md.Tick))
	}

	// exists checking
	if ok, _ := base.cache.Inscription.Get(md.Protocol, md.Tick); ok {
		return nil, xyerrors.NewInsError(-15, fmt.Sprintf("inscription deployed & abort, protocol[%s], tick[%s]", md.Protocol, md.Tick))
	}

	deploy := &Deploy{}
	err := json.Unmarshal([]byte(md.Data), deploy)
	if err != nil {
		return nil, xyerrors.NewInsError(-13, fmt.Sprintf("json decode err:%v", err))
	}

	// max > 0
	if deploy.MaxSupply.LessThanOrEqual(decimal.Zero) {
		return nil, xyerrors.NewInsError(-14, "max <= 0")
	}

	// limit > 0
	if deploy.MintLimit.LessThanOrEqual(decimal.Zero) {
		return nil, xyerrors.NewInsError(-15, "limit <= 0")
	}

	// max >= limit
	if deploy.MaxSupply.LessThan(deploy.MintLimit) {
		return nil, xyerrors.NewInsError(-16, "max < limit")
	}

	// decimal value only int type is valid
	if !deploy.Decimal.IsInteger() {
		return nil, xyerrors.NewInsError(-17, fmt.Sprintf("invalid decimal:%s", deploy.Decimal.String()))
	}

	// maximum decimals is 18
	if deploy.Decimal.IntPart() > 18 {
		return nil, xyerrors.NewInsError(-18, fmt.Sprintf("decimal[%d] > 18", deploy.Decimal.IntPart()))
	}
	return deploy, nil
}
