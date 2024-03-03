package jsonrpc

import (
	"errors"
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
	"inds_getAllChains":              indsGetAllChains,
	"inds_getTicks":                  indsGetTicks,
	"inds_getTick":                   indsGetTick,
	"inds_getTransactions":           indsGetTransactions,
	"inds_getTransactionByAddress":   indsGetAddressTransactions,
	"inds_getTransactionByHash":      indsGetTxByHash,
	"inds_getBalancesByAddress":      indsGetBalancesByAddress,
	"inds_getHoldersByTick":          indsGetHoldersByTick,
	"inds_getLastBlockNumberIndexed": indsGetLastBlockNumber,
	"inds_getTickByCallData":         indsGetTxOperate,
	"inds_getInscriptionTxOperate":   indsGetTxOperate,
	"inds_getAddressBalance":         indsGetAddressBalance,
	"inds_getTickBriefs":             indsGetTickBriefs,
	"inds_chainStat":                 indsChainStat,
	"inds_chainBlockStat":            indsChainBlockStat,
}

func indsGetAllChains(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*GetAllChainCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all chain cmd params:%v", req)
	svr := NewService(s)
	return svr.GetAllChain()
}

func indsSearch(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsSearchCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all txs cmd params:%v", req)
	svr := NewService(s)
	return svr.Search(req.Keyword, req.Chain)

}

func indsGetInscriptions(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	req, ok := cmd.(*IndsGetInscriptionsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all txs cmd params:%v", req)
	svr := NewService(s)
	return svr.GetInscriptions(req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.DeployBy, req.Sort,
		req.SortMode)
}

func indsGetTransactions(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {

	req, ok := cmd.(*IndsGetTransactionCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all txs cmd params:%v", req)
	svr := NewService(s)
	return svr.GetTransactions(req.Chain, req.Address, req.Tick, req.Limit, req.Offset, req.SortMode)

}

func indsGetTicks(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetTicksCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}

	xylog.Logger.Infof("find all Inscriptions cmd params:%v", req)
	svr := NewService(s)
	return svr.GetInscriptions(req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.DeployBy, req.Sort,
		req.SortMode)
}

func indsGetTick(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetTickCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find Inscription cmd params:%v", req)

	svr := NewService(s)

	return svr.GetInscription(req.Chain, req.Protocol, req.Tick, req.DeployHash)
}

func indsGetBalancesByAddress(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetBalanceByAddressCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balances cmd params:%v", req)
	svr := NewService(s)
	return svr.GetAddressBalances(req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick, req.Key,
		req.Sort)
}

func indsGetHoldersByTick(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetHoldersByTickCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balances cmd params:%v", req)
	svr := NewService(s)
	return svr.GetTickHolders(req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.SortMode)
}

func indsGetInscriptionByTick(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetInscriptionTickCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find inscriptions tick cmd params:%v", req)
	svr := NewService(s)
	return svr.GetInscriptionByTick(req.Protocol, req.Tick, req.Chain)
}

func indsGetAddressTransactions(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*IndsGetUserTransactionsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user transactions cmd params:%v", req)
	svr := NewService(s)
	return svr.GetAddressTransactions(req.Protocol, req.Tick, req.Chain, req.Limit, req.Offset, req.Address, req.Event)
}

func indsGetTxByHash(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*GetTxByHashCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get tx by hash cmd params:%v", req)
	svr := NewService(s)
	return svr.GetTxByHash(req.TxHash, req.Chain)
}

func indsGetLastBlockNumber(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*LastBlockNumberCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get last block number cmd params:%v", req)
	svr := NewService(s)
	return svr.GetLastBlockNumber(req.Chains)
}

func indsGetTxOperate(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*TxOperateCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	svr := NewService(s)
	return svr.GetTxOperate(req.Chain, req.InputData)
}

func indsGetAddressBalance(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindUserBalanceCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balance cmd params:%v", req)
	svr := NewService(s)
	return svr.GetAddressBalance(req.Protocol, req.Chain, req.Tick, req.Address)
}

func indsGetTickBriefs(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*GetTickBriefsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get tick briefs cmd params:%v", req)
	svr := NewService(s)
	return svr.GetTickBriefs(req.Addresses)
}
func indsChainStat(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*ChainStatCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("chain stat cmd params:%v", req)
	svr := NewService(s)
	return svr.GetChainStat(req.Chains)
}
func indsChainBlockStat(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*ChainBlockStatCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("chain block stat cmd params:%v", req)
	svr := NewService(s)
	return svr.GetChainBlockStat(req.Chain)
}
