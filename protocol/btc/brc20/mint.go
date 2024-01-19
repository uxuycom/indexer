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
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/xyerrors"
)

func (p *Protocol) Mint(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError) {
	amount, err := p.verifyMint(md)
	if err != nil {
		return nil, xyerrors.ErrDataVerifiedFailed.WrapCause(err)
	}
	result := &devents.TxResult{
		MD:    md,
		Block: block,
		Tx:    tx,
		Mint: &devents.Mint{
			Minter: tx.To,
			Amount: amount,
		},
	}
	return []*devents.TxResult{result}, nil
}

func (p *Protocol) verifyMint(md *devents.MetaData) (decimal.Decimal, *xyerrors.InsError) {
	amount, precision, err := ParseAmountParam(md.Data)
	if err != nil {
		return decimal.Zero, xyerrors.NewInsError(-11, fmt.Sprintf("amount format error, data[%s]", md.Data))
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, xyerrors.NewInsError(-14, "mint amount <= 0")
	}

	var (
		protocol = md.Protocol
		tick     = md.Tick
	)
	ok, inscription := p.cache.Inscription.Get(protocol, tick)
	if !ok || inscription == nil {
		return decimal.Zero, xyerrors.NewInsError(-15, fmt.Sprintf("inscription not exist, protocol[%s], tick[%s]", protocol, tick))
	}

	if precision > int(inscription.Decimals) {
		return decimal.Zero, xyerrors.NewInsError(-16, fmt.Sprintf("mint amount precision > inscription decimals, precision[%d], decimals[%d]", precision, inscription.Decimals))
	}

	// mint amount maximum checking
	if amount.GreaterThan(inscription.LimitPerMint) {
		return decimal.Zero, xyerrors.NewInsError(-17, "mint amount exceeds limit per mint")
	}

	// mint finished checking
	ok, stats := p.cache.InscriptionStats.Get(protocol, tick)
	if !ok {
		return decimal.Zero, xyerrors.ErrInternal.WrapCause(xyerrors.NewInsError(-19, fmt.Sprintf("the inscription stats does not exist, tick[%s-%s]", protocol, tick)))
	}

	if stats.Minted.GreaterThanOrEqual(inscription.TotalSupply) {
		return decimal.Zero, xyerrors.NewInsError(-20, "mint completed")
	}

	// final mint = math.Min(Total Supply - Minted)
	mintLeft := inscription.TotalSupply.Sub(stats.Minted)
	if amount.GreaterThan(mintLeft) {
		amount = mintLeft
	}
	return amount, nil
}
