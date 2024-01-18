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

type Transfer struct {
	Amount decimal.Decimal `json:"amt"`
}

func (p *Protocol) Transfer(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError) {
	amount, err := p.verifyTransfer(tx, md)
	if err != nil {
		return nil, xyerrors.ErrDataVerifiedFailed.WrapCause(err)
	}
	result := &devents.TxResult{
		MD:    md,
		Block: block,
		Tx:    tx,
		Transfer: &devents.Transfer{
			Sender: tx.From,
			Receives: []*devents.Receive{
				{
					Address: tx.To,
					Amount:  amount,
				},
			},
		},
	}
	return []*devents.TxResult{result}, nil
}

func (p *Protocol) verifyTransfer(tx *xycommon.RpcTransaction, md *devents.MetaData) (decimal.Decimal, *xyerrors.InsError) {
	amount, precision, err := ParseAmountParam(md.Data)
	if err != nil {
		return decimal.Zero, xyerrors.NewInsError(-11, fmt.Sprintf("amount format error, data[%s]", md.Data))
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, xyerrors.NewInsError(-14, "transfer amount <= 0")
	}
	var (
		protocol = md.Protocol
		tick     = md.Tick
	)
	ok, inscription := p.cache.Inscription.Get(protocol, tick)
	if !ok || inscription == nil {
		return decimal.Zero, xyerrors.NewInsError(-15, fmt.Sprintf("inscription not exist, protocol[%s]-tick[%s]", protocol, tick))
	}

	if precision > int(inscription.Decimals) {
		return decimal.Zero, xyerrors.NewInsError(-16, fmt.Sprintf("inscribe transfer amount precision > inscription decimals, precision[%d], decimals[%d]", precision, inscription.Decimals))
	}

	// sender balance checking
	ok, balance := p.cache.Balance.Get(protocol, tick, tx.From)
	if !ok {
		return decimal.Zero, xyerrors.NewInsError(-16, fmt.Sprintf("sender balance record not exist, tick[%s-%s], address[%s]", protocol, tick, tx.From))
	}

	// balance available checking
	if balance.Overall.Sub(balance.Available).LessThan(amount) {
		return decimal.Zero, xyerrors.NewInsError(-17, fmt.Sprintf("sender tranferable balance[%v] = overall[%v] - available[%v] < transfer amount[%v]", balance.Overall.Sub(balance.Available), balance.Overall, balance.Available, amount))
	}
	return amount, nil
}
