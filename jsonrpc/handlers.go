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
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/model"
	"github.com/uxuycom/indexer/protocol"
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
}

func handleFindAllInscriptions(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindAllInscriptionsCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find all Inscriptions cmd params:%v", req)

	req.Protocol = strings.ToLower(req.Protocol)
	req.Tick = strings.ToLower(req.Tick)

	inscriptions, total, err := s.dbc.GetInscriptions(req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick, req.DeployBy, req.Sort)
	if err != nil {
		return ErrRPCInternal, err
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
	return resp, nil
}

func handleFindInscriptionTick(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*FindInscriptionTickCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("find inscriptions tick cmd params:%v", req)

	req.Protocol = strings.ToLower(req.Protocol)
	req.Tick = strings.ToLower(req.Tick)

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
		DeployBy:     data.DeployedBy,
		TotalSupply:  data.TotalSupply.String(),
		DeployHash:   data.DeployHash,
		DeployTime:   uint32(data.DeployTime.Unix()),
		TransferType: data.TransferType,
		CreatedAt:    uint32(data.CreatedAt.Unix()),
		UpdatedAt:    uint32(data.UpdatedAt.Unix()),
		Decimals:     data.Decimals,
	}

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

	transactions, total, err := s.dbc.GetTransactionsByAddress(req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick, req.Event)
	if err != nil {
		return ErrRPCInternal, err
	}

	list := make([]*AddressTransaction, 0, len(transactions))
	for _, t := range transactions {
		trans := &AddressTransaction{
			Event:     int8(t.Event),
			TxHash:    t.TxHash,
			Address:   t.Address,
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

	req.Protocol = strings.ToLower(req.Protocol)
	req.Tick = strings.ToLower(req.Tick)

	balances, total, err := s.dbc.GetAddressInscriptions(req.Limit, req.Offset, req.Address, req.Chain, req.Protocol, req.Tick)
	if err != nil {
		return ErrRPCInternal, err
	}

	list := make([]*BalanceInfo, 0, len(balances))
	for _, b := range balances {
		balance := &BalanceInfo{
			Chain:        b.Chain,
			Protocol:     b.Protocol,
			Tick:         b.Tick,
			Address:      b.Address,
			Balance:      b.Balance.String(),
			DeployHash:   b.DeployHash,
			TransferType: b.TransferType,
		}
		list = append(list, balance)
	}

	resp := &FindUserBalancesResponse{
		Inscriptions: list,
		Total:        total,
		Limit:        req.Limit,
		Offset:       req.Offset,
	}
	return resp, nil
}

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

	holders, total, err := s.dbc.GetHoldersByTick(req.Limit, req.Offset, req.Chain, req.Protocol, req.Tick)
	if err != nil {
		return ErrRPCInternal, err
	}

	list := make([]*BalanceInfo, 0, len(holders))
	for _, b := range holders {
		balance := &BalanceInfo{
			Chain:    b.Chain,
			Protocol: b.Protocol,
			Tick:     b.Tick,
			Address:  b.Address,
			Balance:  b.Balance.String(),
		}
		list = append(list, balance)
	}

	resp := &FindTickHoldersResponse{
		Holders: list,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}
	return resp, nil
}

func handleGetLastBlockNumber(s *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	req, ok := cmd.(*LastBlockNumberCmd)
	if !ok {
		return ErrRPCInvalidParams, errors.New("invalid params")
	}
	xylog.Logger.Infof("get last block number cmd params:%v", req)

	blockNumber, err := s.dbc.QueryLastBlock(req.Chain)
	if err != nil {
		return ErrRPCInternal, err
	}

	resp := &LastBlockNumberResponse{
		BlockNumber: blockNumber,
	}
	return resp, nil
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

	// todo, get inscription events from address_txs
	transInfo := &TransactionInfo{
		Protocol: "",
		Tick:     "",
		From:     tx.From,
		To:       tx.To,
	}

	inscription, err := s.dbc.FindInscriptionByTick(tx.Chain, "", "")
	if err != nil {
		return ErrRPCInternal, err
	}
	if inscription == nil {
		return nil, errors.New("Record not found")
	}
	transInfo.DeployBy = inscription.DeployedBy

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
