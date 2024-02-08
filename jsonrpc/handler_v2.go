package jsonrpc

import (
	"errors"
	"github.com/uxuycom/indexer/storage"
	"github.com/uxuycom/indexer/xylog"
)

var rpcHandlersBeforeInitV2 = map[string]commandHandler{
	"inds_getInscriptions":           indsGetInscriptions,
	"index_getInscriptionByTick":     indsGetInscriptionByTick,
	"inds_search":                    indsSearch,
	"inds_getAllChain":               indsGetAllChain,
	"inds_getTicks":                  indsGetTicks, //handleFindAllInscriptions,
	"inds_getTransactions":           indsGetTransactions,
	"inds_getTransactionByAddress":   handleFindAddressTransactions,
	"inds_getTransactionByHash":      handleGetTxByHash,
	"inds_getBalanceByAddress":       indsGetBalanceByAddress,
	"inds_getHoldersByTick":          indsGetHoldersByTick,
	"inds_getLastBlockNumberIndexed": handleGetLastBlockNumber,
	"inds_getTickByCallData":         handleGetTxOperate,
}

func indsGetAllChain(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*GetAllChainCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all chain cmd params:%v", req)
	return getAllChain(s)
}

func indsSearch(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsSearchCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all txs cmd params:%v", req)
	return search(s, req.Keyword, req.Chain)

}

func indsGetInscriptions(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	req, ok := cmd.(*IndsGetInscriptionsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all txs cmd params:%v", req)
	return findInscriptions(s, req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.DeployBy, req.Sort, storage.OrderByModeDesc)
}

func indsGetTransactions(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	req, ok := cmd.(*IndsGetTransactionCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all txs cmd params:%v", req)
	return findTransactions(s, req.Address, req.Tick, req.Limit, req.Offset, req.SortMode)

}

func indsGetTicks(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetTicksCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}

	xylog.Logger.Infof("find all Inscriptions cmd params:%v", req)
	return findInscriptions(s, req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.DeployBy, req.Sort,
		req.SortMode)
}

func indsGetBalanceByAddress(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetBalanceByAddressCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balances cmd params:%v", req)

	return findAddressBalances(s, req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick, req.Key,
		req.Sort)
}

func indsGetHoldersByTick(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetHoldersByTickCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balances cmd params:%v", req)

	return findTickHolders(s, req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.SortMode)
}
