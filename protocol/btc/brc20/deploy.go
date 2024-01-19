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
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/xyerrors"
	"math"
	"math/big"
	"strconv"
)

type Deploy struct {
	Name      string
	MaxSupply decimal.Decimal
	MintLimit decimal.Decimal
	Decimal   decimal.Decimal
}

func (p *Protocol) Deploy(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError) {
	d, err := p.verifyDeploy(md)
	if err != nil {
		return nil, xyerrors.ErrDataVerifiedFailed.WrapCause(err)
	}

	result := &devents.TxResult{
		MD:    md,
		Block: block,
		Tx:    tx,
		Deploy: &devents.Deploy{
			Name:      d.Name,
			MaxSupply: d.MaxSupply,
			MintLimit: d.MintLimit,
			Decimal:   int8(d.Decimal.IntPart()),
		},
	}
	return []*devents.TxResult{result}, nil
}

func (p *Protocol) verifyDeploy(md *devents.MetaData) (*Deploy, *xyerrors.InsError) {
	// metadata protocol / tick checking
	if md.Protocol == "" || md.Tick == "" {
		return nil, xyerrors.NewInsError(-10, fmt.Sprintf("protocol[%s] / tick[%s] nil", md.Protocol, md.Tick))
	}

	if len(md.Tick) != 4 {
		return nil, xyerrors.NewInsError(-11, fmt.Sprintf("tick[%s] len != 4", md.Tick))
	}

	// exists checking
	if ok, _ := p.cache.Inscription.Get(md.Protocol, md.Tick); ok {
		return nil, xyerrors.NewInsError(-12, fmt.Sprintf("inscription deployed & abort, protocol[%s], tick[%s]", md.Protocol, md.Tick))
	}

	// deploy params parse
	var deployParams map[string]interface{}
	err := json.Unmarshal([]byte(md.Data), &deployParams)
	if err != nil {
		return nil, xyerrors.NewInsError(-13, fmt.Sprintf("json decode err:%v, data[%s]", err, md.Data))
	}

	// decimal param parse
	decimals := 18
	if dec, exist := deployParams["dec"]; exist {
		decString, ok := dec.(string)
		if !ok {
			return nil, xyerrors.NewInsError(-14, fmt.Sprintf("decimal type err:%v", dec))
		}

		decimals, err = strconv.Atoi(decString)
		if err != nil {
			return nil, xyerrors.NewInsError(-14, fmt.Sprintf("decimal parse err:%v", err))
		}

		if decimals > MaxPrecision {
			return nil, xyerrors.NewInsError(-15, fmt.Sprintf("decimal[%d] > max_precision[%d]", decimals, MaxPrecision))
		}
	}

	// max param parse
	maxVal, exist := deployParams["max"]
	if !exist {
		return nil, xyerrors.NewInsError(-16, "max not exist")
	}
	maxString, ok := maxVal.(string)
	if !ok {
		return nil, xyerrors.NewInsError(-16, fmt.Sprintf("max type err:%v", maxVal))
	}
	maxSupply, precision, err := NewDecimalFromString(maxString)
	if err != nil {
		return nil, xyerrors.NewInsError(-16, fmt.Sprintf("max supply decimal parse err:%v", err))
	}
	if precision > decimals {
		return nil, xyerrors.NewInsError(-17, fmt.Sprintf("max supply precision[%d] > decimal[%d]", precision, decimals))
	}
	if maxSupply.LessThanOrEqual(decimal.Zero) {
		return nil, xyerrors.NewInsError(-15, "max <= 0")
	}
	// MaxSupply must <= uint64
	maxUint64Decimal := decimal.NewFromBigInt(new(big.Int).SetUint64(math.MaxUint64), 0)
	if maxSupply.GreaterThan(maxUint64Decimal) {
		return nil, xyerrors.NewInsError(-18, fmt.Sprintf("max[%s] > max_uint64", maxSupply.String()))
	}

	// limit param parse
	mintLimit := maxSupply
	limitVal, exist := deployParams["lim"]
	if exist {
		limitString, ok1 := limitVal.(string)
		if !ok1 {
			return nil, xyerrors.NewInsError(-19, fmt.Sprintf("limit type err:%v", limitVal))
		}

		mintLimit, precision, err = NewDecimalFromString(limitString)
		if err != nil {
			return nil, xyerrors.NewInsError(-20, fmt.Sprintf("max supply decimal parse err:%v", err))
		}

		if precision > decimals {
			return nil, xyerrors.NewInsError(-21, fmt.Sprintf("max supply precision[%d] > decimal[%d]", precision, decimals))
		}

		// limit > 0
		if mintLimit.LessThanOrEqual(decimal.Zero) {
			return nil, xyerrors.NewInsError(-22, "limit <= 0")
		}

		// max >= limit
		if maxSupply.LessThan(mintLimit) {
			return nil, xyerrors.NewInsError(-33, "max < limit")
		}
	}

	name := deployParams["tick"].(string)
	return &Deploy{
		Name:      name,
		MaxSupply: maxSupply,
		MintLimit: mintLimit,
		Decimal:   decimal.NewFromInt(int64(decimals)),
	}, nil
}
