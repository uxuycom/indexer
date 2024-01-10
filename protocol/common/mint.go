package common

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"open-indexer/client/xycommon"
	"open-indexer/devents"
	"open-indexer/xyerrors"
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
