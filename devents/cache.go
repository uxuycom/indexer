package devents

import (
	"github.com/shopspring/decimal"
	"open-indexer/dcache"
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
		} else {
			receiveAmount := receiveBalance.Overall.Add(item.Amount)
			tc.cache.Balance.Update(r.MD.Protocol, r.MD.Tick, item.Address, &dcache.BalanceItem{
				Overall: receiveAmount,
			})
		}
	}
}
