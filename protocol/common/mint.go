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

type Mint struct {
	Amount decimal.Decimal `json:"amt"`
}

func (base *Protocol) Mint(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError) {
	m, err := base.verifyMint(tx, md)
	if err != nil {
		return nil, xyerrors.ErrDataVerifiedFailed.WrapCause(err)
	}
	result := &devents.TxResult{
		MD:    md,
		Block: block,
		Tx:    tx,
		Mint: &devents.Mint{
			Minter: tx.To,
			Amount: m.Amount,
		},
	}
	return []*devents.TxResult{result}, nil
}

func (base *Protocol) verifyMint(tx *xycommon.RpcTransaction, md *devents.MetaData) (*Mint, *xyerrors.InsError) {
	mint := &Mint{}
	err := json.Unmarshal([]byte(md.Data), mint)
	if err != nil {
		return nil, xyerrors.NewInsError(-13, fmt.Sprintf("data json deocde err:%v, data[%s]", err, md.Data))
	}

	if mint.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, xyerrors.NewInsError(-14, "mint amount <= 0")
	}

	var (
		protocol = md.Protocol
		tick     = md.Tick
	)
	ok, inscription := base.cache.Inscription.Get(protocol, tick)
	if !ok || inscription == nil {
		return nil, xyerrors.NewInsError(-15, fmt.Sprintf("inscription not exist, protocol[%s], tick[%s]", protocol, tick))
	}

	// mint amount maximum checking
	if mint.Amount.GreaterThan(inscription.LimitPerMint) {
		return nil, xyerrors.NewInsError(-17, "mint amount exceeds limit per mint")
	}

	// mint finished checking
	ok, stats := base.cache.InscriptionStats.Get(protocol, tick)
	if !ok {
		return nil, xyerrors.ErrInternal.WrapCause(xyerrors.NewInsError(-19, fmt.Sprintf("the inscription stats does not exist, tick[%s-%s]", protocol, tick)))
	}

	if stats.Minted.GreaterThanOrEqual(inscription.TotalSupply) {
		return nil, xyerrors.NewInsError(-20, "mint completed")
	}

	// final mint = math.Min(Total Supply - Minted)
	mintLeft := inscription.TotalSupply.Sub(stats.Minted)
	if mint.Amount.GreaterThan(mintLeft) {
		mint.Amount = mintLeft
	}
	return mint, nil
}
