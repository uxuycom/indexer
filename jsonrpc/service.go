package jsonrpc

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/model"
	"github.com/uxuycom/indexer/protocol"
	"github.com/uxuycom/indexer/utils"
	"github.com/uxuycom/indexer/xylog"
	"strings"
	"time"
)

type Service struct {
	rpcServer *RpcServer
}

var service *Service

func NewService(rpcServer *RpcServer) *Service {
	if service != nil {
		return service
	}
	return &Service{
		rpcServer: rpcServer,
	}
}

func (s *Service) GetAddressBalances(limit, offset int, address, chain, protocol, tick string, key string,
	sort int) (interface{}, error) {
	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)
	cacheKey := fmt.Sprintf("addr_balances_%d_%d_%s_%s_%s_%s_%d", limit, offset, address, chain, protocol, tick, sort)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*FindUserBalancesResponse); ok {
			return allIns, nil
		}
	}

	balances, total, err := s.rpcServer.dbc.GetAddressInscriptions(limit, offset, address, chain, protocol, tick, key, sort)
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
	s.rpcServer.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func (s *Service) GetInscriptions(limit, offset int, chain, protocol, tick, deployBy string, sort,
	sortMode int) (interface{}, error) {
	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)
	cacheKey := fmt.Sprintf("all_ins_%d_%d_%s_%s_%s_%s_%d_%d", limit, offset, chain, protocol, tick, deployBy, sort, sortMode)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*IndsGetAllInscriptionsResponse); ok {
			return allIns, nil
		}
	}
	inscriptions, total, err := s.rpcServer.dbc.GetInscriptions(limit, offset, chain, protocol, tick, deployBy, sort, sortMode)
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
			DeployTime:   uint32(ins.DeployTime.Unix()),
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

	resp := &IndsGetAllInscriptionsResponse{
		Inscriptions: result,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}
	s.rpcServer.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func (s *Service) GetInscription(chain, protocol, tick, deployHash string) (interface{}, error) {
	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)

	cacheKey := fmt.Sprintf("tick_%s_%s_%s_%s", chain, protocol, tick, deployHash)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if ticks, ok := ins.(*InscriptionInfo); ok {
			return ticks, nil
		}
	}

	inscription, err := s.rpcServer.dbc.FindInscriptionInfo(chain, protocol, tick, deployHash)
	if err != nil {
		return ErrRPCInternal, err
	}
	if inscription == nil {
		return ErrRPCRecordNotFound, err
	}

	resp := &InscriptionInfo{
		Chain:        inscription.Chain,
		Protocol:     inscription.Protocol,
		Tick:         inscription.Tick,
		Name:         inscription.Name,
		LimitPerMint: inscription.LimitPerMint.String(),
		DeployBy:     inscription.DeployBy,
		TotalSupply:  inscription.TotalSupply.String(),
		DeployHash:   inscription.DeployHash,
		TransferType: inscription.TransferType,
		Decimals:     inscription.Decimals,
		Minted:       inscription.Minted.String(),
		Holders:      inscription.Holders,
		TxCnt:        inscription.TxCnt,
		Progress:     inscription.Progress.String(),
		DeployTime:   uint32(inscription.DeployTime.Unix()),
		CreatedAt:    uint32(inscription.CreatedAt.Unix()),
		UpdatedAt:    uint32(inscription.UpdatedAt.Unix()),
	}

	s.rpcServer.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func (s *Service) GetTickHolders(limit int, offset int, chain, protocol, tick string,
	sortMode int) (interface{},
	error) {
	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)
	cacheKey := fmt.Sprintf("all_ins_%d_%d_%s_%s_%s_%d", limit, offset, chain, protocol, tick, sortMode)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*FindTickHoldersResponse); ok {
			return allIns, nil
		}
	}

	// get inscription info
	inscription, err := s.rpcServer.dbc.FindInscriptionByTick(chain, protocol, tick)
	if err != nil {
		return ErrRPCInternal, err
	}
	if inscription == nil {
		return nil, errors.New("Record not found")
	}

	holders, total, err := s.rpcServer.dbc.GetHoldersByTick(limit, offset, chain, protocol, tick, sortMode)
	if err != nil {
		return ErrRPCInternal, err
	}

	list := make([]*TickHolder, 0, len(holders))
	for _, holder := range holders {
		balance := &TickHolder{
			Chain:       holder.Chain,
			Protocol:    holder.Protocol,
			Tick:        holder.Tick,
			DeployHash:  inscription.DeployHash,
			Address:     holder.Address,
			Balance:     holder.Balance.String(),
			TotalSupply: inscription.TotalSupply.String(),
		}
		list = append(list, balance)
	}

	resp := &FindTickHoldersResponse{
		Holders: list,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}

	s.rpcServer.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func (s *Service) GetTransactions(chain string, address string, tick string, limit int, offset int,
	sortMode int) (interface{},
	error) {

	address = strings.ToLower(address)
	tick = strings.ToLower(tick)
	chain = strings.ToLower(chain)

	cacheKey := fmt.Sprintf("all_transactions_%d_%d_%s_%s_%s_%d", limit, offset, chain, address, tick, sortMode)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if transactions, ok := ins.(*CommonResponse); ok {
			return transactions, nil
		}
	}
	lastMonth := time.Now().AddDate(0, -1, 0).Format("2006-01-02")[:7] + "-01"

	txs, total, err := s.rpcServer.dbc.GetTransactions(lastMonth, chain, address, tick, limit, offset, sortMode)
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
			TxHash:          common.BytesToHash(v.TxHash),
			From:            v.From,
			To:              v.To,
			Op:              v.Op,
			Tick:            v.Tick,
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
	s.rpcServer.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func (s *Service) GetInscriptionsStats(limit int, offset int, sortMode int) (interface{},
	error) {
	txs, total, err := s.rpcServer.dbc.GetInscriptionStatsList(limit, offset, sortMode)
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

func (s *Service) Search(keyword, chain string) (interface{}, error) {

	result := &SearchResult{}
	if len(keyword) <= 10 {
		// Inscription
		inscriptions, _ := s.GetInscriptions(100, 0, chain, "", keyword, "", 2, 0)
		result.Data = inscriptions
		result.Type = "Inscription"
		return result, nil
	}
	if strings.HasPrefix(keyword, "0x") {
		if len(keyword) == 42 {
			// address
			address, _, _ := s.rpcServer.dbc.GetBalancesChainByAddress(100, 0, keyword, chain, "", "")
			result.Data = address
			result.Type = "Address"
		}
		if len(keyword) == 66 {
			// tx hash
			result.Data, _ = s.rpcServer.dbc.FindBalanceByTxHash(keyword)
			result.Type = "TxHash"
		}
	} else {
		if len(keyword) == 64 {
			// tx hash
			result.Data, _ = s.rpcServer.dbc.FindBalanceByTxHash(keyword)
			result.Type = "TxHash"
		} else {
			// address
			address, _, _ := s.rpcServer.dbc.GetBalancesChainByAddress(100, 0, keyword, chain, "", "")
			result.Data = address
			result.Type = "Address"
		}
	}
	return result, nil
}
func (s *Service) GetAllChain() (interface{}, error) {

	blocksMap := make(map[string]model.Block, 0)
	blocks, err := s.rpcServer.dbc.GetAllBlocks()
	if err != nil {
		return ErrRPCInternal, err
	}
	for _, v := range blocks {
		blocksMap[v.Chain] = v
	}

	chains, err := s.rpcServer.dbc.GetAllChainInfo()
	if err != nil {
		return ErrRPCInternal, err
	}

	chainsInfo := make([]ChainInfo, 0)
	for _, v := range chains {
		info := ChainInfo{
			ChainId:    v.ChainId,
			Chain:      v.Chain,
			OuterChain: v.OuterChain,
			Name:       v.Name,
			Logo:       v.Logo,
			NetworkId:  v.NetworkId,
			Ext:        v.Ext,
		}
		if block, ok := blocksMap[v.Chain]; ok {
			info.BlockTime = block.BlockTime
			info.UpdatedAt = block.UpdatedAt
			info.BlockNumber = block.BlockNumber
		}
		chainsInfo = append(chainsInfo, info)
	}
	return chainsInfo, nil
}

func (s *Service) GetInscriptionByTick(protocol string, tick string, chain string) (interface{}, error) {

	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)

	cacheKey := fmt.Sprintf("tick_%s_%s_%s", chain, protocol, tick)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if ticks, ok := ins.(*InscriptionInfo); ok {
			return ticks, nil
		}
	}

	data, err := s.rpcServer.dbc.FindInscriptionByTick(chain, protocol, tick)
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
	s.rpcServer.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func (s *Service) GetAddressTransactions(protocol string, tick string, chain string, limit int,
	offset int,
	address string, event int8) (interface{}, error) {

	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)

	cacheKey := fmt.Sprintf("addr_txs_%d_%d_%s_%s_%s_%s_%d", limit, offset, address, chain, tick,
		tick, event)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*FindUserTransactionsResponse); ok {
			return allIns, nil
		}
	}

	transactions, total, err := s.rpcServer.dbc.GetAddressTxs(limit, offset, address, chain, protocol, tick, event)
	if err != nil {
		return ErrRPCInternal, err
	}

	txsHashes := make(map[string][]common.Hash)
	for _, v := range transactions {
		txsHashes[v.Chain] = append(txsHashes[v.Chain], common.BytesToHash(v.TxHash))
	}

	txMap := make(map[string]*model.Transaction)
	for chain, hashes := range txsHashes {
		txs, err := s.rpcServer.dbc.GetTxsByHashes(chain, hashes)
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
			TxHash:    common.BytesToHash(t.TxHash),
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
	s.rpcServer.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func (s *Service) GetTxByHash(txHash common.Hash, chain string) (interface{}, error) {

	cacheKey := fmt.Sprintf("tx_info_%s_%s", chain, txHash)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*GetTxByHashResponse); ok {
			return allIns, nil
		}
	}
	tx, err := s.rpcServer.dbc.FindTransaction(chain, txHash)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, errors.New("Transaction Record not found")
	}
	resp := &GetTxByHashResponse{}
	inscription, err := s.rpcServer.dbc.FindInscriptionByTick(tx.Chain, tx.Protocol, tx.Tick)
	// get amount from address tx tab
	addressTx, err := s.rpcServer.dbc.FindAddressTxByHash(chain, txHash)
	resp.IsInscription = true

	resp.Inscriptions = inscription
	resp.Address = addressTx

	if tx != nil {
		trs := &TransactionResponse{
			ID:              tx.ID,
			Chain:           tx.Chain,
			Protocol:        tx.Protocol,
			BlockHeight:     tx.BlockHeight,
			PositionInBlock: tx.PositionInBlock,
			BlockTime:       tx.BlockTime,
			TxHash:          common.BytesToHash(tx.TxHash),
			From:            tx.From,
			To:              tx.To,
			Op:              tx.Op,
			Tick:            tx.Tick,
			Amount:          tx.Amount,
			Gas:             tx.Gas,
			GasPrice:        tx.GasPrice,
			Status:          tx.Status,
			CreatedAt:       tx.CreatedAt,
			UpdatedAt:       tx.UpdatedAt,
		}
		resp.Transaction = trs
	}
	inscriptionsData := &InscriptionsData{
		Protocol: tx.Protocol,
		Operate:  tx.Op,
		Tick:     tx.Tick,
		Amount:   tx.Amount,
	}
	resp.InscriptionsData = inscriptionsData
	s.rpcServer.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func (s *Service) GetLastBlockNumber(chains []string) (interface{}, error) {

	var chainsStr string
	if len(chains) > 0 {
		chainsStr = strings.Join(chains, "_")
	} else {
		chainsStr = fmt.Sprintf("%v", len(chains))
	}
	xylog.Logger.Infof("get last block chainsStr:%v, chains len:%v", chainsStr, len(chains))

	cacheKey := fmt.Sprintf("block_number_%s", chainsStr)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.([]*BlockInfo); ok {
			return allIns, nil
		}
	}
	result := make([]*BlockInfo, 0)
	var err error
	chs := chains
	if len(chains) == 0 {
		chs, err = s.rpcServer.dbc.GetAllChainFromBlock()
		if err != nil {
			chs = []string{}
		}
		xylog.Logger.Infof("get last block from db chains:%v", chs)
	}
	for _, chain := range chs {
		block, err := s.rpcServer.dbc.FindLastBlock(chain)
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
	s.rpcServer.cacheStore.Set(cacheKey, result)
	return result, nil
}

func (s *Service) GetTxOperate(chain string, inputData string) (interface{}, error) {
	cacheKey := fmt.Sprintf("tx_operate_%s_%s", chain, inputData)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*TxOperateResponse); ok {
			return allIns, nil
		}
	}
	operate := protocol.GetOperateByTxInput(chain, inputData, s.rpcServer.dbc)
	xylog.Logger.Infof("handleGetTxOperate operate =%v, inputdata=%v, chain=%v", operate, inputData, chain)
	if operate == nil {
		return nil, errors.New("Record not found")
	}
	var deployHash string
	if operate.Protocol != "" && operate.Tick != "" {
		inscription, err := s.rpcServer.dbc.FindInscriptionByTick(strings.ToLower(chain),
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
	s.rpcServer.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func (s *Service) GetAddressBalance(protocol string, chain string, tick string,
	address string) (interface{}, error) {

	protocol = strings.ToLower(protocol)
	tick = strings.ToLower(tick)
	cacheKey := fmt.Sprintf("addr_balance_%s_%s_%s", chain, protocol, tick)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*BalanceBrief); ok {
			return allIns, nil
		}
	}
	inscription, err := s.rpcServer.dbc.FindInscriptionByTick(chain, protocol, tick)
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
	balance, err := s.rpcServer.dbc.FindUserBalanceByTick(chain, protocol, tick, address)
	if err != nil {
		return ErrRPCInternal, err
	}
	if balance == nil {
		return nil, errors.New("Record not found")
	}
	resp.Balance = balance.Balance.String()
	resp.Available = balance.Available.String()

	switch inscription.TransferType {
	case model.TransferTypeHash:
		// transfer with hash
		result, err := s.rpcServer.dbc.GetUtxosByAddress(address, chain, protocol, tick)
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
	s.rpcServer.cacheStore.Set(cacheKey, resp)
	return resp, nil
}

func (s *Service) GetTickBriefs(addresses []*TickAddress) (interface{}, error) {

	deployHashGroups := make(map[string][]string)
	key := ""
	for _, address := range addresses {
		deployHashGroups[address.Chain] = append(deployHashGroups[address.Chain], address.DeployHash)
		key += fmt.Sprintf("%s_%s", address.Chain, address.DeployHash)
	}

	cacheKey := fmt.Sprintf("tick_briefs_%s", key)
	if ins, ok := s.rpcServer.cacheStore.Get(cacheKey); ok {
		if allIns, ok := ins.(*GetTickBriefsResp); ok {
			return allIns, nil
		}
	}

	result := make([]*model.InscriptionOverView, 0, len(addresses))
	for chain, groups := range deployHashGroups {
		dbTicks, err := s.rpcServer.dbc.GetInscriptionsByChain(chain, groups)
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
			stat, _ := s.rpcServer.dbc.FindInscriptionsStatsByTick(dbTick.Chain, dbTick.Protocol, dbTick.Tick)
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
	s.rpcServer.cacheStore.Set(cacheKey, resp)

	return resp, nil
}
func (s *Service) GetChainStat(chain []string) (interface{}, error) {
	// get 24H chain stat from chain_stats_hour
	nowUint := utils.TimeHourInt(time.Now())
	yesterdayUint := utils.TimeHourInt(utils.YesterdayHour())
	dayBeforeYesterdayUint := utils.TimeHourInt(utils.BeforeYesterdayHour())
	xylog.Logger.Infof("chain stat nowUint:%v yesterdayUint:%v dayBeforeYesterdayUint:%v", nowUint, yesterdayUint, dayBeforeYesterdayUint)

	todayStat, err := s.rpcServer.dbc.GroupChainStatHour(24, 0, chain)
	if err != nil {
		return ErrRPCInternal, err
	}
	yesterdayStat, err := s.rpcServer.dbc.GroupChainStatHour(24, 24, chain)
	result := make([]*model.Chain24HourStat, 0)
	xylog.Logger.Infof("chain stat today:%v yesterdayStat:%v", todayStat, yesterdayStat)
	for _, a := range todayStat {
		xylog.Logger.Infof("chain stat hour:%v", a)
		chainInfo, _ := s.rpcServer.dbc.GetChainInfoByChain(a.Chain)
		if chainInfo != nil {
			chain24HourStat := &model.Chain24HourStat{
				ChainId:    chainInfo.ChainId,
				Chain:      a.Chain,
				Name:       chainInfo.Name,
				Logo:       chainInfo.Logo,
				Address24h: a.AddressCount,
				Balance24h: a.BalanceSum,
				Tick24h:    a.InscriptionsCount,
			}
			for _, b := range yesterdayStat {
				if b.Chain == a.Chain {
					if a.AddressCount == 0 || b.AddressCount == 0 {
						chain24HourStat.Address24hPercent = 0
					} else {
						chain24HourStat.Address24hPercent = a.AddressCount / b.AddressCount
					}
					partA := a.BalanceSum.IntPart()
					partB := b.BalanceSum.IntPart()
					if partA == 0 || partB == 0 {
						chain24HourStat.Balance24hPercent = 0
					} else {
						chain24HourStat.Balance24hPercent = uint32(partA / partB)
					}
					if a.InscriptionsCount == 0 || b.InscriptionsCount == 0 {
						chain24HourStat.Tick24hPercent = 0
					} else {
						chain24HourStat.Tick24hPercent = a.InscriptionsCount / b.InscriptionsCount
					}

				}
			}
			result = append(result, chain24HourStat)
		}
	}
	return result, nil
}

func (s *Service) GetChainBlockStat(chain string) (interface{}, error) {

	block, err := s.rpcServer.dbc.FindLastBlock(chain)
	endTime := time.Now()
	if block != nil {
		endTime = block.BlockTime
	}
	startTime := endTime.Add(-72 * time.Hour)
	stat, err := s.rpcServer.dbc.GroupChainBlockStat(startTime, endTime, 0, chain)
	if err != nil {
		return ErrRPCInternal, err
	}
	return stat, nil
}
func (s *Service) GetChainInfo(chain string) (interface{}, error) {
	chainInfo, err := s.rpcServer.dbc.GetChainInfoByChain(chain)
	if err != nil {
		return ErrRPCInternal, err
	}
	chifo := model.ChainInfo{
		ID:         chainInfo.ID,
		ChainId:    chainInfo.ChainId,
		Chain:      chainInfo.Chain,
		OuterChain: chainInfo.OuterChain,
		Name:       chainInfo.Name,
		Logo:       chainInfo.Logo,
		NetworkId:  chainInfo.NetworkId,
		Ext:        chainInfo.Ext,
		UpdatedAt:  chainInfo.UpdatedAt,
		CreatedAt:  chainInfo.CreatedAt,
	}

	chainInfoExt := &model.ChainInfoExt{
		ChainInfo:    chifo,
		TickCount:    s.rpcServer.dbc.CountTickByChain(chain), // TODO
		AddressCount: 0,                                       // TODO
		MintCount:    0,                                       // TODO

	}
	return chainInfoExt, nil
}
