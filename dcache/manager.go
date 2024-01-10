package dcache

import (
	"open-indexer/storage"
	"open-indexer/xylog"
	"time"
)

// Manager
/*****************************************************
 * Do all cache operate
 ****************************************************/
type Manager struct {
	chain            string
	db               *storage.DBClient
	Balance          *Balance
	UTXO             *UTXO
	Inscription      *Inscription
	InscriptionStats *InscriptionStats
}

func NewManager(db *storage.DBClient, chain string) *Manager {
	e := &Manager{
		db:    db,
		chain: chain,
	}

	if db == nil {
		return e
	}

	e.initInscriptionCache()
	e.initInscriptionStatsCache()
	e.initBalanceCache()
	e.initUtxoCache()
	return e
}

func (h *Manager) initInscriptionCache() {
	h.Inscription = NewInscription()

	startTs := time.Now()
	idx := 0
	start := uint32(0)
	limit := 10000
	xylog.Logger.Infof("load inscriptions data start...")
	for {
		items, err := h.db.GetInscriptionsByIdLimit(uint64(start), limit)
		if err != nil {
			xylog.Logger.Fatalf("failed to initialize inscription cache data. err:%v", err)
		}
		idx++
		xylog.Logger.Infof("load inscriptions ret, items[%d], idx:%d", len(items), idx)

		if len(items) <= 0 {
			break
		}

		for _, v := range items {
			h.Inscription.Create(v.Protocol, v.Tick, &Tick{
				SID:          v.SID,
				TransferType: v.TransferType,
				LimitPerMint: v.LimitPerMint,
				TotalSupply:  v.TotalSupply,
				Decimals:     v.Decimals,
			})
		}
		//update sid
		h.Inscription.SetSid(items[len(items)-1].SID)

		//update id index
		start = items[len(items)-1].ID
	}
	xylog.Logger.Infof("load inscriptions data finished, cost ts:%v", time.Since(startTs))
}

func (h *Manager) initInscriptionStatsCache() {
	h.InscriptionStats = NewInscriptionStats()

	startTs := time.Now()
	idx := 0
	start := uint32(0)
	limit := 10000
	xylog.Logger.Infof("load inscription-stats data start...")
	for {
		items, err := h.db.GetInscriptionStatsByIdLimit(uint64(start), limit)
		if err != nil {
			xylog.Logger.Fatalf("failed to initialize inscription-stats cache data. err:%v", err)
		}
		idx++
		xylog.Logger.Infof("load inscription-stats ret, items[%d], idx:%d", len(items), idx)

		if len(items) <= 0 {
			break
		}

		for _, v := range items {
			h.InscriptionStats.Create(v.Protocol, v.Tick, &InsStats{
				SID:     v.SID,
				Minted:  v.Minted,
				Holders: int64(v.Holders),
				TxCnt:   v.TxCnt,
			})
		}
		//update sid
		h.InscriptionStats.SetSid(items[len(items)-1].SID)

		//update id index
		start = items[len(items)-1].ID
	}
	xylog.Logger.Infof("load inscription-stats data finished, cost ts:%v", time.Since(startTs))
}

func (h *Manager) initBalanceCache() {
	h.Balance = NewBalance()

	startTs := time.Now()
	idx := 0
	start := uint64(0)
	limit := 10000
	xylog.Logger.Infof("load balances data start...")
	for {
		balances, err := h.db.GetBalancesByIdLimit(start, limit)
		if err != nil {
			xylog.Logger.Fatalf("failed to initialize balance cache data. err:%v", err)
		}
		idx++
		xylog.Logger.Infof("load balances ret, items[%d], idx:%d", len(balances), idx)

		if len(balances) <= 0 {
			break
		}

		for _, v := range balances {
			h.Balance.Create(v.Protocol, v.Tick, v.Address, &BalanceItem{
				SID:       v.SID,
				Available: v.Available,
				Overall:   v.Balance,
			})
		}

		//update sid
		h.Balance.SetSid(balances[len(balances)-1].SID)

		//update id index
		start = balances[len(balances)-1].ID
	}
	xylog.Logger.Infof("load balances data finished, cost ts:%v", time.Since(startTs))
}

func (h *Manager) initUtxoCache() {
	h.UTXO = NewUTXO()

	startTs := time.Now()
	idx := 0
	start := uint64(0)
	limit := 1000
	xylog.Logger.Infof("load utxos data start...")
	for {
		utxos, err := h.db.GetUTXOsByIdLimit(start, limit)
		if err != nil {
			xylog.Logger.Fatalf("failed to initialize utxos cache data. err:%v", err)
		}
		idx++
		xylog.Logger.Infof("load utxos ret, items[%d], idx:%d", len(utxos), idx)

		if len(utxos) <= 0 {
			break
		}

		for _, v := range utxos {
			h.UTXO.Add(v.Protocol, v.Tick, v.RootHash, v.Address, v.Amount, v.Sn)
		}

		//update id index
		start = utxos[len(utxos)-1].ID
	}
	xylog.Logger.Infof("load utxos data finished, cost ts:%v", time.Since(startTs))
}
