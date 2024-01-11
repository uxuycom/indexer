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

package devents

import (
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/dcache"
)

type TxResultHandler struct {
	cache *dcache.Manager
}

func NewTxResultHandler(cache *dcache.Manager) *TxResultHandler {
	return &TxResultHandler{
		cache: cache,
	}
}

func (tc *TxResultHandler) UpdateCache(r *TxResult) {
	if r.Deploy != nil {
		tc.updateDeployCache(r)
	}

	if r.Mint != nil {
		tc.updateMintCache(r)
	}

	if r.Transfer != nil {
		tc.updateTransferCache(r)
	}
}

func (tc *TxResultHandler) updateDeployCache(r *TxResult) {
	//Add new tick
	t := &dcache.Tick{
		LimitPerMint: r.Deploy.MintLimit,
		TotalSupply:  r.Deploy.MaxSupply,
		Decimals:     r.Deploy.Decimal,
	}
	tc.cache.Inscription.Create(r.MD.Protocol, r.MD.Tick, t)

	//Add new tick stats
	ts := &dcache.InsStats{
		TxCnt: 1,
	}
	tc.cache.InscriptionStats.Create(r.MD.Protocol, r.MD.Tick, ts)
}

func (tc *TxResultHandler) updateMintCache(r *TxResult) {
	//Update mint stats
	tc.cache.InscriptionStats.Mint(r.MD.Protocol, r.MD.Tick, r.Mint.Amount)
	tc.cache.InscriptionStats.TxCnt(r.MD.Protocol, r.MD.Tick, 1)

	//Update minter balances
	ok, balance := tc.cache.Balance.Get(r.MD.Protocol, r.MD.Tick, r.Mint.Minter)
	if !ok {
		tc.cache.Balance.Create(r.MD.Protocol, r.MD.Tick, r.Mint.Minter, &dcache.BalanceItem{
			Overall: r.Mint.Amount,
		})
		tc.cache.InscriptionStats.Holders(r.MD.Protocol, r.MD.Tick, 1)

		//mark minter init
		r.Mint.Init = true
	} else {
		amount := balance.Overall.Add(r.Mint.Amount)
		tc.cache.Balance.Update(r.MD.Protocol, r.MD.Tick, r.Mint.Minter, &dcache.BalanceItem{
			Overall: amount,
		})
	}
}

func (tc *TxResultHandler) updateTransferCache(r *TxResult) {
	//Update transfer stats
	tc.cache.InscriptionStats.TxCnt(r.MD.Protocol, r.MD.Tick, 1)

	//Update sender balances
	sendTotalAmount := decimal.Zero
	for _, item := range r.Transfer.Receives {
		sendTotalAmount = sendTotalAmount.Add(item.Amount)
	}

	holders := int64(0)
	_, senderBalance := tc.cache.Balance.Get(r.MD.Protocol, r.MD.Tick, r.Transfer.Sender)
	senderAmount := senderBalance.Overall.Sub(sendTotalAmount)
	if senderAmount.LessThanOrEqual(decimal.Zero) {
		holders--
	}
	tc.cache.Balance.Update(r.MD.Protocol, r.MD.Tick, r.Transfer.Sender, &dcache.BalanceItem{
		Overall: senderAmount,
	})

	for _, item := range r.Transfer.Receives {
		ok, receiveBalance := tc.cache.Balance.Get(r.MD.Protocol, r.MD.Tick, item.Address)
		if !ok {
			holders++
			receiveAmount := item.Amount
			tc.cache.Balance.Create(r.MD.Protocol, r.MD.Tick, item.Address, &dcache.BalanceItem{
				Overall: receiveAmount,
			})

			//mark minter init
			item.Init = true
		} else {
			receiveAmount := receiveBalance.Overall.Add(item.Amount)
			tc.cache.Balance.Update(r.MD.Protocol, r.MD.Tick, item.Address, &dcache.BalanceItem{
				Overall: receiveAmount,
			})
		}
	}
	tc.cache.InscriptionStats.Holders(r.MD.Protocol, r.MD.Tick, holders)
}
