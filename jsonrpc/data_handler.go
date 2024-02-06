package jsonrpc

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/model"
	"strings"
)

func findAddressBalances(s *RpcServer, limit, offset int, address, chain, protocol, tick string, sort int) (interface{}, error) {
	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)
	cacheKey := fmt.Sprintf("addr_balances_%d_%d_%s_%s_%s_%s_%d", limit, offset, address, chain, protocol, tick, sort)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*FindUserBalancesResponse); ok {
			return allIns, nil
		}
	}

	balances, total, err := s.dbc.GetAddressInscriptions(limit, offset, address, chain, protocol, tick, sort)
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
		Limit:        limit,
		Offset:       offset,
	}
	s.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func findInscriptions(s *RpcServer, limit, offset int, chain, protocol, tick, deployBy string, sort, 
	sortMode int) (interface{}, error) {
	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)
	cacheKey := fmt.Sprintf("all_ins_%d_%d_%s_%s_%s_%s_%d_%d", limit, offset, chain, protocol, tick, deployBy, sort, sortMode)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*FindAllInscriptionsResponse); ok {
			return allIns, nil
		}
	}
	inscriptions, total, err := s.dbc.GetInscriptions(limit, offset, chain, protocol, tick, deployBy, sort, sortMode)
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
		Limit:        limit,
		Offset:       offset,
	}
	s.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func findTickHolders(s *RpcServer, limit int, offset int, chain, protocol, tick string, sortMode int) (interface{}, error) {
	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)
	cacheKey := fmt.Sprintf("all_ins_%d_%d_%s_%s_%s_%d", limit, offset, chain, protocol, tick, sortMode)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*FindTickHoldersResponse); ok {
			return allIns, nil
		}
	}

	holders, total, err := s.dbc.GetHoldersByTick(limit, offset, chain, protocol, tick, sortMode)
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
		Limit:   limit,
		Offset:  offset,
	}

	s.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func findTransactions(s *RpcServer, address string, tick string, limit int, offset int, sortMode int) (interface{},
	error) {

	address = strings.ToLower(address)
	tick = strings.ToLower(tick)

	cacheKey := fmt.Sprintf("all_transactions_%d_%d_%s_%s_%d", limit, offset, address, tick, sortMode)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if transactions, ok := ins.(*CommonResponse); ok {
			return transactions, nil
		}
	}
	txs, total, err := s.dbc.GetTransactions(address, tick, limit, offset, sortMode)
	if err != nil {
		return ErrRPCInternal, err
	}
	transactions := make([]*TransactionResponse, 0)
	for _, v := range txs {

		trs := &TransactionResponse{
			ID:              v.ID,
			Chain:           v.Chain,
			Protocol:        v.Protocol,
			BlockHeight:     v.BlockHeight,
			PositionInBlock: v.PositionInBlock,
			BlockTime:       v.BlockTime,
			TxHash:          v.TxHash,
			From:            v.From,
			To:              v.To,
			Tick:            v.Tick,
			Amount:          v.Amount,
			Gas:             v.Gas,
			GasPrice:        v.GasPrice,
			Status:          v.Status,
			CreatedAt:       v.CreatedAt,
			UpdatedAt:       v.UpdatedAt,
		}
		transactions = append(transactions, trs)
	}

	resp := &CommonResponse{
		Data:   transactions,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	s.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func findInscriptionsStats(s *RpcServer, limit int, offset int, sortMode int) (interface{},
	error) {
	txs, total, err := s.dbc.GetInscriptionStatsList(limit, offset, sortMode)
	if err != nil {
		return ErrRPCInternal, err
	}

	resp := &CommonResponse{
		Data:   txs,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	return resp, nil
}

func search(s *RpcServer, keyword string, chain string) (interface{},
	error) {
	resp := &CommonResponse{}

	result := &SearchResult{}
	// todo do some validate
	if strings.HasPrefix(keyword, "0x") {
		if len(keyword) == 42 {
			// address
			result.Data, _ = findAddressBalances(s, 1, 0, keyword, chain, "", "", 0)
			result.Type = "address"
		}
		if len(keyword) == 66 {
			// tx hash
		}
	} else {
		if len(keyword) == 64 {
			// tx hash
			result.Data, _ = findTransactions(s, keyword, "", 1, 0, 0)
			result.Type = "tx"
		} else {
			// address
		}
	}
	resp.Data = result

	return resp, nil
}
func getAllChain(s *RpcServer) (interface{}, error) {
	chains, err := s.dbc.GetAllChainFromBlock()
	if err != nil {
		return ErrRPCInternal, err
	}
	return chains, nil
}
