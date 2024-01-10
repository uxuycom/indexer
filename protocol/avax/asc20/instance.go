package asc20

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"open-indexer/client/xycommon"
	"open-indexer/dcache"
	"open-indexer/devents"
	"open-indexer/protocol/common"
	"open-indexer/protocol/types"
	"open-indexer/xyerrors"
	"open-indexer/xylog"
	"strings"
	"sync"
)

type Protocol struct {
	common *common.Protocol
	cache  *dcache.Manager
	ticks  *sync.Map // record deploy tx -> tick code
}

var ParsedABI abi.ABI

func NewProtocol(cache *dcache.Manager) *Protocol {
	return &Protocol{
		common: common.NewProtocol(cache),
		cache:  cache,
		ticks:  &sync.Map{},
	}
}

func (p *Protocol) Parse(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, md *devents.MetaData) ([]*devents.TxResult, *xyerrors.InsError) {
	switch md.Operate {
	case devents.OperateList:
		return p.List(block, tx, md)

	case devents.OperateExchange:
		return p.Exchange(block, tx, md)
	}
	return p.common.Parse(block, tx, md)
}

func ParseMetaDataByEventLogs(chain string, tx *xycommon.RpcTransaction) (*devents.MetaData, error) {
	for _, event := range tx.Events {
		if len(event.Topics) < 1 {
			continue
		}

		topic := event.Topics[0].String()
		if topic == EventTopicHashExchange || topic == EventTopicHashExchange2 {
			return &devents.MetaData{
				Chain:    chain,
				Protocol: types.ASC20Protocol,
				Operate:  devents.OperateExchange,
			}, nil
		}
	}
	return nil, nil
}

func init() {
	var err error
	ParsedABI, err = abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		xylog.Logger.Fatalf("asc20 abi decode err:%v", err)
	}
}
