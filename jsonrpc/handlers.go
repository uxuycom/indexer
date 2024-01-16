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
	/*
		req.Protocol = strings.ToLower(req.Protocol)
		req.Tick = strings.ToLower(req.Tick)

		inscriptions, total, err := s.dbc.GetInscriptions(req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.DeployBy, req.Sort, req.SortMode)
		if err != nil {
			return ErrRPCInternal, err
		}

		cacheKey := fmt.Sprintf("all_ins_%d_%d_%s_%s_%s_%s_%d_%d", req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.DeployBy, req.Sort, req.SortMode)
		if ins, ok := s.cacheStore.Get(cacheKey); ok {
			if allIns, ok := ins.(FindAllInscriptionsResponse); ok {
				return allIns, nil
			}
		}

		result := make([]*model.InscriptionBrief, 0, len(inscriptions))

		for _, ins := range inscriptions {
			brief := &model.InscriptionBrief{
				Chain:        ins.Chain,
				Protocol:     ins.Protocol,
				Tick:         ins.Name,
				DeployBy:     ins.DeployBy,
				DeployHash:   ins.DeployHash,
				TotalSupply:  ins.TotalSupply.String(),
				Holders:      ins.Holders,
				Minted:       ins.Minted.String(),
				LimitPerMint: ins.LimitPerMint.String(),
				TransferType: ins.TransferType,
				Status:       model.MintStatusProcessing,
				TxCnt:        ins.TxCnt,
				CreatedAt:    uint32(ins.CreatedAt.Unix()),
			}

			minted := ins.Minted
			totalSupply := ins.TotalSupply

			if totalSupply != decimal.Zero && minted != decimal.Zero {
				percentage, _ := minted.Div(totalSupply).Float64()
				if percentage >= 1 {
					percentage = 1
				}
				brief.MintedPercent = fmt.Sprintf("%.4f", percentage)
			}

			if ins.Minted.Cmp(ins.TotalSupply) >= 0 {
				brief.Status = model.MintStatusAllMinted
			}

			result = append(result, brief)
		}

		resp := &FindAllInscriptionsResponse{
			Inscriptions: result,
			Total:        total,
			Limit:        req.Limit,
			Offset:       req.Offset,
		}

		s.cacheStore.Set(cacheKey, resp)

		return resp, nil
	*/
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
		if ticks, ok := ins.(InscriptionInfo); ok {
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

	transactions, total, err := s.dbc.GetTransactionsByAddress(req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick, req.Key, req.Event)
	if err != nil {
		return ErrRPCInternal, err
	}

	list := make([]*AddressTransaction, 0, len(transactions))
	for _, t := range transactions {
		trans := &AddressTransaction{
			Event:     t.Event,
			TxHash:    t.TxHash,
			Address:   t.Address,
			From:      t.From,
			To:        t.To,
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
	return resp, nil
}

func handleFindAddressBalances(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindUserBalancesCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balances cmd params:%v", req)

	return findAddressBalances(s, req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick, req.Key, storage.OrderByModeDesc)
}

//func handleFindAddressBalancesWithSort(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
//	req, ok := cmd.(*FindUserBalancesWithSortCmd)
//	if !ok {
//		return ErrRPCInvalidParams, errors.New("invalid params")
//	}
//	xylog.Logger.Infof("find user balances cmd params:%v", req)
//
//	req.Protocol = strings.ToLower(req.Protocol)
//	req.Tick = strings.ToLower(req.Tick)
//
//	return findAddressBalances(s, req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick, req.Key, req.Sort)
//}

func handleFindAddressBalance(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindUserBalanceCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find user balance cmd params:%v", req)

	req.Protocol = strings.ToLower(req.Protocol)
	req.Tick = strings.ToLower(req.Tick)

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

	return resp, nil
}

func handleFindTickHolders(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindTickHoldersCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find tick holders cmd params:%v", req)

	req.Protocol = strings.ToLower(req.Protocol)
	req.Tick = strings.ToLower(req.Tick)

	return findTickHolders(s, req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, storage.OrderByModeDesc)
}

func handleGetLastBlockNumber(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*LastBlockNumberCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get last block number cmd params:%v", req)

	result := make([]*BlockInfo, 0)

	for _, chain := range req.Chains {
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

	return result, nil
}

func handleGetTxOperate(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*TxOperateCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
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

	return TxOperateResponse{
		Protocol:   string(operate.Protocol),
		Operate:    string(operate.Operate),
		Tick:       operate.Tick,
		DeployHash: deployHash,
	}, nil
}

func handleGetTxByHash(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*GetTxByHashCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get tx by hash cmd params:%v", req)

	req.TxHash = strings.ToLower(req.TxHash)

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

	return resp, nil
}

func handleGetTickBriefs(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*GetTickBriefsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get tick briefs cmd params:%v", req)

	deployHashGroups := make(map[string][]string)
	for _, address := range req.Addresses {
		deployHashGroups[address.Chain] = append(deployHashGroups[address.Chain], address.DeployHash)
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

	return resp, nil
}
