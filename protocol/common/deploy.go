package common

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"open-indexer/client/xycommon"
	"open-indexer/devents"
	"open-indexer/xyerrors"
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
