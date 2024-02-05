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

package jsonrpc

import (
	"errors"
	"fmt"
	"github.com/uxuycom/indexer/model"
	"github.com/uxuycom/indexer/protocol"
	"github.com/uxuycom/indexer/storage"
	"github.com/uxuycom/indexer/xylog"
	"strings"
)

var rpcHandlersBeforeInit = map[string]commandHandler{
	"inscription.All":           handleFindAllInscriptions,
	"inscription.Tick":          handleFindInscriptionTick,
	"address.Transactions":      handleFindAddressTransactions,
	"address.Balances":          handleFindAddressBalances,
	"address.Balance":           handleFindAddressBalance,
	"tick.Holders":              handleFindTickHolders,
	"block.LastNumber":          handleGetLastBlockNumber,
	"tool.InscriptionTxOperate": handleGetTxOperate,
	"transaction.Info":          handleGetTxByHash,
	"tick.GetBriefs":            handleGetTickBriefs,
}

func handleFindAllInscriptions(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindAllInscriptionsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all Inscriptions cmd params:%v", req)
	return findInsciptions(s, req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.DeployBy, req.Sort, storage.OrderByModeDesc)
}

func handleFindInscriptionTick(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindInscriptionTickCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find inscriptions tick cmd params:%v", req)

	req.Protocol = strings.ToLower(req.Protocol)
	req.Tick = strings.ToLower(req.Tick)

	cacheKey := fmt.Sprintf("tick_%s_%s_%s", req.Chain, req.Protocol, req.Tick)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if ticks, ok := ins.(*InscriptionInfo); ok {
			return ticks, nil
		}
	}

	data, err := s.dbc.FindInscriptionByTick(req.Chain, req.Protocol, req.Tick)
	if err != nil {
		return ErrRPCInternal, err
	}
	if data == nil {
		return ErrRPCRecordNotFound, err
	}

	resp := &InscriptionInfo{
		Chain:        data.Chain,
		Protocol:     data.Protocol,
		Tick:         data.Tick,
		Name:         data.Name,
		LimitPerMint: data.LimitPerMint.String(),
		DeployBy:     data.DeployBy,
		TotalSupply:  data.TotalSupply.String(),
		DeployHash:   data.DeployHash,
		DeployTime:   uint32(data.DeployTime.Unix()),
		TransferType: data.TransferType,
		CreatedAt:    uint32(data.CreatedAt.Unix()),
		UpdatedAt:    uint32(data.UpdatedAt.Unix()),
		Decimals:     data.Decimals,
	}
	s.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func handleFindAddressTransactions(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindUserTransactionsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user transactions cmd params:%v", req)

	req.Protocol = strings.ToLower(req.Protocol)
	req.Tick = strings.ToLower(req.Tick)

	cacheKey := fmt.Sprintf("addr_txs_%d_%d_%s_%s_%s_%s_%d", req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick, req.Event)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*FindUserTransactionsResponse); ok {
			return allIns, nil
		}
	}

	transactions, total, err := s.dbc.GetAddressTxs(req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick, req.Event)
	if err != nil {
		return ErrRPCInternal, err
	}

	txsHashes := make(map[string][]string)
	for _, v := range transactions {
		txsHashes[v.Chain] = append(txsHashes[v.Chain], v.TxHash)
	}

	txMap := make(map[string]*model.Transaction)
	for chain, hashes := range txsHashes {
		txs, err := s.dbc.GetTxsByHashes(chain, hashes)
		if err != nil {
			xylog.Logger.Error(err)
			continue
		}
		if len(txs) > 0 {
			for _, v := range txs {
				key := fmt.Sprintf("%s_%s", v.Chain, v.TxHash)
				txMap[key] = v
			}
		}
	}

	list := make([]*AddressTransaction, 0, len(transactions))
	for _, t := range transactions {
		key := fmt.Sprintf("%s_%s", t.Chain, t.TxHash)
		from := ""
		to := ""
		if tx, ok := txMap[key]; ok {
			from = tx.From
			to = tx.To
		}

		trans := &AddressTransaction{
			Event:     t.Event,
			TxHash:    t.TxHash,
			Address:   t.Address,
			From:      from,
			To:        to,
			Amount:    t.Amount.String(),
			Tick:      t.Tick,
			Protocol:  t.Protocol,
			Operate:   t.Operate,
			Chain:     t.Chain,
			Status:    t.Status,
			CreatedAt: uint32(t.CreatedAt.Unix()),
			UpdatedAt: uint32(t.UpdatedAt.Unix()),
		}
		list = append(list, trans)
	}

	resp := &FindUserTransactionsResponse{
		Transactions: list,
		Total:        total,
		Limit:        req.Limit,
		Offset:       req.Offset,
	}
	s.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func handleFindAddressBalances(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindUserBalancesCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balances cmd params:%v", req)

	return findAddressBalances(s, req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick, storage.OrderByModeDesc)
}

func handleFindAddressBalance(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindUserBalanceCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balance cmd params:%v", req)

	req.Protocol = strings.ToLower(req.Protocol)
	req.Tick = strings.ToLower(req.Tick)
	cacheKey := fmt.Sprintf("addr_balance_%s_%s_%s", req.Chain, req.Protocol, req.Tick)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*BalanceBrief); ok {
			return allIns, nil
		}
	}
	inscription, err := s.dbc.FindInscriptionByTick(req.Chain, req.Protocol, req.Tick)
	if err != nil {
		return ErrRPCInternal, err
	}
	if inscription == nil {
		return nil, errors.New("Record not found")
	}

	resp := &BalanceBrief{
		Tick:         inscription.Tick,
		TransferType: inscription.TransferType,
		DeployHash:   inscription.DeployHash,
	}

	// balance
	balance, err := s.dbc.FindUserBalanceByTick(req.Chain, req.Protocol, req.Tick, req.Address)
	if err != nil {
		return ErrRPCInternal, err
	}
	if balance == nil {
		return nil, errors.New("Record not found")
	}
	resp.Balance = balance.Balance.String()

	switch inscription.TransferType {
	case model.TransferTypeHash:
		// transfer with hash
		result, err := s.dbc.GetUtxosByAddress(req.Address, req.Chain, req.Protocol, req.Tick)
		if err != nil {
			return ErrRPCInternal, err
		}
		utxos := make([]*UTXOBrief, 0, len(result))
		for _, u := range result {
			utxos = append(utxos, &UTXOBrief{
				Tick:     u.Tick,
				Amount:   u.Amount.String(),
				RootHash: u.RootHash,
			})
		}
		resp.Utxos = utxos
	}
	s.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func handleFindTickHolders(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindTickHoldersCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find tick holders cmd params:%v", req)
	return findTickHolders(s, req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, storage.OrderByModeDesc)
}

func handleGetLastBlockNumber(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*LastBlockNumberCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get last block number cmd params:%v", req)

	var chainsStr string
	if len(req.Chains) > 0 {
		chainsStr = strings.Join(req.Chains, "_")
	} else {
		chainsStr = fmt.Sprintf("%v", len(req.Chains))
	}
	xylog.Logger.Infof("get last block chainsStr:%v, chains len:%v", chainsStr, len(req.Chains))
	cacheKey := fmt.Sprintf("block_number_%s", chainsStr)

	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.([]*BlockInfo); ok {
			return allIns, nil
		}
	}
	result := make([]*BlockInfo, 0)
	var err error
	var chains []string
	if len(req.Chains) == 0 {
		chains, err = s.dbc.GetAllChainFromBlock()
		if err != nil {
			chains = []string{}
		}
		xylog.Logger.Infof("get last block from db chains:%v", chains)
	}
	for _, chain := range chains {
		block, err := s.dbc.FindLastBlock(chain)
		if err != nil {
			return ErrRPCInternal, err
		}
		blockInfo := &BlockInfo{
			Chain:       chain,
			BlockNumber: block.BlockNumber,
			TimeStamp:   uint32(block.BlockTime.Unix()),
			BlockTime:   block.BlockTime.String(),
		}
		result = append(result, blockInfo)
	}
	s.cacheStore.Set(cacheKey, result)
	return result, nil
}

func handleGetTxOperate(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*TxOperateCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	cacheKey := fmt.Sprintf("tx_operate_%s_%s", req.Chain, req.InputData)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*TxOperateResponse); ok {
			return allIns, nil
		}
	}
	operate := protocol.GetOperateByTxInput(req.Chain, req.InputData, s.dbc)
	if operate == nil {
		return nil, errors.New("Record not found")
	}
	var deployHash string
	if operate.Protocol != "" && operate.Tick != "" {
		inscription, err := s.dbc.FindInscriptionByTick(strings.ToLower(req.Chain), strings.ToLower(string(operate.Protocol)), strings.ToLower(operate.Tick))
		if err != nil {
			xylog.Logger.Errorf("the query for the inscription failed. chain:%s protocol:%s tick:%s err=%s", req.Chain, string(operate.Protocol), operate.Tick, err)
		}
		if inscription != nil {
			deployHash = inscription.DeployHash
		}
	}

	resp := &TxOperateResponse{
		Protocol:   operate.Protocol,
		Operate:    operate.Operate,
		Tick:       operate.Tick,
		DeployHash: deployHash,
	}
	s.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func handleGetTxByHash(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*GetTxByHashCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get tx by hash cmd params:%v", req)

	req.TxHash = strings.ToLower(req.TxHash)
	cacheKey := fmt.Sprintf("tx_info_%s_%s", req.Chain, req.TxHash)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*GetTxByHashResponse); ok {
			return allIns, nil
		}
	}
	tx, err := s.dbc.FindTransaction(req.Chain, req.TxHash)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, errors.New("Record not found")
	}

	resp := &GetTxByHashResponse{}

	// not inscription transaction
	if tx == nil {
		resp.IsInscription = false
		return resp, nil
	}

	transInfo := &TransactionInfo{
		Protocol: tx.Protocol,
		Tick:     tx.Tick,
		From:     tx.From,
		To:       tx.To,
	}

	inscription, err := s.dbc.FindInscriptionByTick(tx.Chain, tx.Protocol, tx.Tick)
	if err != nil {
		return ErrRPCInternal, err
	}
	if inscription == nil {
		return nil, errors.New("Record not found")
	}
	transInfo.DeployHash = inscription.DeployHash

	// get amount from address tx tab
	addressTx, err := s.dbc.FindAddressTxByHash(req.Chain, req.TxHash)
	if err != nil {
		return ErrRPCInternal, err
	}
	if addressTx == nil {
		return nil, errors.New("Record not found")
	}
	transInfo.Amount = addressTx.Amount.String()

	resp.IsInscription = true
	resp.Transaction = transInfo
	s.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func handleGetTickBriefs(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*GetTickBriefsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get tick briefs cmd params:%v", req)

	deployHashGroups := make(map[string][]string)
	key := ""
	for _, address := range req.Addresses {
		deployHashGroups[address.Chain] = append(deployHashGroups[address.Chain], address.DeployHash)
		key += fmt.Sprintf("%s_%s", address.Chain, address.DeployHash)
	}

	cacheKey := fmt.Sprintf("tick_briefs_%s", key)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*GetTickBriefsResp); ok {
			return allIns, nil
		}
	}

	result := make([]*model.InscriptionOverView, 0, len(req.Addresses))
	for chain, groups := range deployHashGroups {
		dbTicks, err := s.dbc.GetInscriptionsByChain(chain, groups)
		if err != nil {
			continue
		}
		for _, dbTick := range dbTicks {
			overview := &model.InscriptionOverView{
				Chain:        dbTick.Chain,
				Protocol:     dbTick.Protocol,
				Tick:         dbTick.Tick,
				Name:         dbTick.Name,
				LimitPerMint: dbTick.LimitPerMint,
				TotalSupply:  dbTick.TotalSupply,
				DeployBy:     dbTick.DeployBy,
				DeployHash:   dbTick.DeployHash,
				DeployTime:   dbTick.DeployTime,
				TransferType: dbTick.TransferType,
				Decimals:     dbTick.Decimals,
				CreatedAt:    dbTick.CreatedAt,
			}
			stat, _ := s.dbc.FindInscriptionsStatsByTick(dbTick.Chain, dbTick.Protocol, dbTick.Tick)
			if stat != nil {
				overview.Holders = stat.Holders
				overview.Minted = stat.Minted
				overview.TxCnt = stat.TxCnt
			}
			result = append(result, overview)
		}
	}

	resp := &GetTickBriefsResp{}
	resp.Inscriptions = result
	s.cacheStore.Set(cacheKey, resp)

	return resp, nil
}
