package jsonrpc

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/model"
	"github.com/uxuycom/indexer/xylog"
	"strings"
)

func findAddressBalances(s *RpcServer, limit, offset int, address, chain, protocol, tick, key string, sort int) (interface{}, error) {
	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)
	balances, total, err := s.dbc.GetAddressInscriptions(limit, offset, address, chain, protocol, tick, key, sort)
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
	return resp, nil
}

func findInsciptions(s *RpcServer, limit, offset int, chain, protocol, tick, deployBy string, sort, sortMode int) (interface{}, error) {
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

	xylog.Logger.Info(resp)

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
