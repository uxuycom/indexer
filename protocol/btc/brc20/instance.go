package brc20

import (
	"open-indexer/dcache"
	"open-indexer/protocol/common"
)

type Protocol struct {
	*common.Protocol
}

func NewProtocol(cache *dcache.Manager) *Protocol {
	return &Protocol{
		Protocol: common.NewProtocol(cache),
	}
}
