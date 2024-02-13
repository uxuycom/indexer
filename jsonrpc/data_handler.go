package jsonrpc

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/model"
	"github.com/uxuycom/indexer/protocol"
	"github.com/uxuycom/indexer/xylog"
	"strings"
)

func getAddressBalances(s *RpcServer, limit, offset int, address, chain, protocol, tick string, key string,
	sort int) (interface{}, error) {
	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)
	cacheKey := fmt.Sprintf("addr_balances_%d_%d_%s_%s_%s_%s_%d", limit, offset, address, chain, protocol, tick, sort)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*FindUserBalancesResponse); ok {
			return allIns, nil
		}
	}

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
			TxHash:          common.Bytes2Hex(v.TxHash),
			From:            v.From,
			To:              v.To,
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

func search(s *RpcServer, keyword, chain string) (interface{}, error) {

	result := &SearchResult{}
	if len(keyword) <= 10 {
		// Inscription
		inscriptions, _ := findInscriptions(s, 10, 0, chain, "", keyword, "", 2, 0)
		result.Data = inscriptions
		result.Type = "Inscription"
	}
	if strings.HasPrefix(keyword, "0x") {
		if len(keyword) == 42 {
			// address
			address, _, _ := s.dbc.GetBalancesChainByAddress(10, 0, keyword, chain, "", "")
			result.Data = address
			result.Type = "Address"
		}
		if len(keyword) == 66 {
			// tx hash
			result.Data, _ = s.dbc.FindBalanceByTxHash(keyword)
			result.Type = "TxHash"
		}
	} else {
		if len(keyword) == 64 {
			// tx hash
			result.Data, _ = s.dbc.FindBalanceByTxHash(keyword)
			result.Type = "tx"
		} else {
			// address
			address, _, _ := s.dbc.GetBalancesChainByAddress(10, 0, keyword, chain, "", "")
			result.Data = address
			result.Type = "Address"
		}
	}
	return result, nil
}
func getAllChain(s *RpcServer) (interface{}, error) {
	chains, err := s.dbc.GetAllChainFromBlock()
	if err != nil {
		return ErrRPCInternal, err
	}
	return chains, nil
}

func getInscriptionByTick(s *RpcServer, protocol string, tick string, chain string) (interface{}, error) {

	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)

	cacheKey := fmt.Sprintf("tick_%s_%s_%s", chain, protocol, tick)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if ticks, ok := ins.(*InscriptionInfo); ok {
			return ticks, nil
		}
	}

	data, err := s.dbc.FindInscriptionByTick(chain, protocol, tick)
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

func getAddressTransactions(s *RpcServer, protocol string, tick string, chain string, limit int, offset int,
	address string, event int8) (interface{}, error) {

	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)

	cacheKey := fmt.Sprintf("addr_txs_%d_%d_%s_%s_%s_%s_%d", limit, offset, address, chain, tick,
		tick, event)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*FindUserTransactionsResponse); ok {
			return allIns, nil
		}
	}

	transactions, total, err := s.dbc.GetAddressTxs(limit, offset, address, chain, protocol, tick, event)
	if err != nil {
		return ErrRPCInternal, err
	}

	txsHashes := make(map[string][]string)
	for _, v := range transactions {
		txsHashes[v.Chain] = append(txsHashes[v.Chain], common.Bytes2Hex(v.TxHash))
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
			TxHash:    common.Bytes2Hex(t.TxHash),
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
		Limit:        limit,
		Offset:       offset,
	}
	s.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func getTxByHash(s *RpcServer, txHash string, chain string) (interface{}, error) {

	txHash = strings.ToLower(txHash)
	cacheKey := fmt.Sprintf("tx_info_%s_%s", chain, txHash)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*GetTxByHashResponse); ok {
			return allIns, nil
		}
	}
	tx, err := s.dbc.FindTransaction(chain, txHash)
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
	transInfo.DeployHash = inscription.DeployHash

	// get amount from address tx tab
	addressTx, err := s.dbc.FindAddressTxByHash(chain, txHash)
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

func getLastBlockNumber(s *RpcServer, chains []string) (interface{}, error) {

	var chainsStr string
	if len(chains) > 0 {
		chainsStr = strings.Join(chains, "_")
	} else {
		chainsStr = fmt.Sprintf("%v", len(chains))
	}
	xylog.Logger.Infof("get last block chainsStr:%v, chains len:%v", chainsStr, len(chains))

	cacheKey := fmt.Sprintf("block_number_%s", chainsStr)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.([]*BlockInfo); ok {
			return allIns, nil
		}
	}
	result := make([]*BlockInfo, 0)
	var err error
	chs := chains
	if len(chains) == 0 {
		chs, err = s.dbc.GetAllChainFromBlock()
		if err != nil {
			chs = []string{}
		}
		xylog.Logger.Infof("get last block from db chains:%v", chs)
	}
	for _, chain := range chs {
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

func getTxOperate(s *RpcServer, chain string, inputData string) (interface{}, error) {
	cacheKey := fmt.Sprintf("tx_operate_%s_%s", chain, inputData)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*TxOperateResponse); ok {
			return allIns, nil
		}
	}
	operate := protocol.GetOperateByTxInput(chain, inputData, s.dbc)
	xylog.Logger.Infof("handleGetTxOperate operate =%v, inputdata=%v, chain=%v", operate, inputData, chain)
	if operate == nil {
		return nil, errors.New("Record not found")
	}
	var deployHash string
	if operate.Protocol != "" && operate.Tick != "" {
		inscription, err := s.dbc.FindInscriptionByTick(strings.ToLower(chain),
			strings.ToLower(string(operate.Protocol)), strings.ToLower(operate.Tick))
		if err != nil {
			xylog.Logger.Errorf("the query for the inscription failed. chain:%s protocol:%s tick:%s err=%s", chain,
				string(operate.Protocol), operate.Tick, err)
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

func getAddressBalance(s *RpcServer, protocol string, chain string, tick string, address string) (interface{}, error) {

	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)
	cacheKey := fmt.Sprintf("addr_balance_%s_%s_%s", chain, protocol, tick)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*BalanceBrief); ok {
			return allIns, nil
		}
	}
	inscription, err := s.dbc.FindInscriptionByTick(chain, protocol, tick)
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
	balance, err := s.dbc.FindUserBalanceByTick(chain, protocol, tick, address)
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
		result, err := s.dbc.GetUtxosByAddress(address, chain, protocol, tick)
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

func getTickBriefs(s *RpcServer, addresses []*TickAddress) (interface{}, error) {

	deployHashGroups := make(map[string][]string)
	key := ""
	for _, address := range addresses {
		deployHashGroups[address.Chain] = append(deployHashGroups[address.Chain], address.DeployHash)
		key += fmt.Sprintf("%s_%s", address.Chain, address.DeployHash)
	}

	cacheKey := fmt.Sprintf("tick_briefs_%s", key)
	if ins, ok := s.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*GetTickBriefsResp); ok {
			return allIns, nil
		}
	}

	result := make([]*model.InscriptionOverView, 0, len(addresses))
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
