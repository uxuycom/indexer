package jsonrpc

import (
	"errors"
	"github.com/uxuycom/indexer/storage"
	"github.com/uxuycom/indexer/xylog"
)

// v1 version
var rpcHandlersBeforeInit = map[string]commandHandler{
	"inscription.All":           indsGetInscriptions,
	"inscription.Tick":          indsGetInscriptionByTick,
	"address.Transactions":      indsGetAddressTransactions,
	"address.Balances":          indsGetBalancesByAddress,
	"address.Balance":           indsGetAddressBalance,
	"tick.Holders":              indsGetHoldersByTick,
	"block.LastNumber":          indsGetLastBlockNumber,
	"tool.InscriptionTxOperate": indsGetTxOperate,
	"transaction.Info":          indsGetTxByHash,
	"tick.GetBriefs":            indsGetTickBriefs,
}

// v2 version
var rpcHandlersBeforeInitV2 = map[string]commandHandler{
	"inds_getInscriptions":           indsGetInscriptions,
	"index_getInscriptionByTick":     indsGetInscriptionByTick,
	"inds_search":                    indsSearch,
	"inds_getAllChain":               indsGetAllChain,
	"inds_getTicks":                  indsGetTicks,
	"inds_getTransactions":           indsGetTransactions,
	"inds_getTransactionByAddress":   indsGetAddressTransactions,
	"inds_getTransactionByHash":      indsGetTxByHash,
	"inds_getBalancesByAddress":      indsGetBalancesByAddress,
	"inds_getHoldersByTick":          indsGetHoldersByTick,
	"inds_getLastBlockNumberIndexed": indsGetLastBlockNumber,
	"inds_getTickByCallData":         indsGetTxOperate,
	"inds_getAddressBalance":         indsGetAddressBalance,
	"inds_getTickBriefs":             indsGetTickBriefs,
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
	return getInscriptions(s, req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.DeployBy, req.Sort,
		storage.OrderByModeDesc)
}

func indsGetTransactions(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	req, ok := cmd.(*IndsGetTransactionCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all txs cmd params:%v", req)
	return getTransactions(s, req.Address, req.Tick, req.Limit, req.Offset, req.SortMode)

}

func indsGetTicks(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetTicksCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}

	xylog.Logger.Infof("find all Inscriptions cmd params:%v", req)
	return getInscriptions(s, req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.DeployBy, req.Sort,
		req.SortMode)
}

func indsGetBalancesByAddress(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetBalanceByAddressCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balances cmd params:%v", req)

	return getAddressBalances(s, req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick, req.Key,
		req.Sort)
}

func indsGetHoldersByTick(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetHoldersByTickCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balances cmd params:%v", req)

	return getTickHolders(s, req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.SortMode)
}

func indsGetInscriptionByTick(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindInscriptionTickCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find inscriptions tick cmd params:%v", req)

	return getInscriptionByTick(s, req.Protocol, req.Tick, req.Chain)
}

func indsGetAddressTransactions(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindUserTransactionsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user transactions cmd params:%v", req)

	return getAddressTransactions(s, req.Protocol, req.Tick, req.Chain, req.Limit, req.Offset, req.Address, req.Event)
}

func indsGetTxByHash(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*GetTxByHashCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get tx by hash cmd params:%v", req)
	return getTxByHash(s, req.TxHash, req.Chain)
}

func indsGetLastBlockNumber(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*LastBlockNumberCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get last block number cmd params:%v", req)
	return getLastBlockNumber(s, req.Chains)
}

func indsGetTxOperate(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*TxOperateCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	return getTxOperate(s, req.Chain, req.InputData)
}

func indsGetAddressBalance(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindUserBalanceCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balance cmd params:%v", req)
	return getAddressBalance(s, req.Protocol, req.Chain, req.Tick, req.Address)
}

func indsGetTickBriefs(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*GetTickBriefsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get tick briefs cmd params:%v", req)
	return getTickBriefs(s, req.Addresses)
}
