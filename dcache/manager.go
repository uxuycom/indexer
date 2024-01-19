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

package dcache

import (
	"github.com/uxuycom/indexer/storage"
	"github.com/uxuycom/indexer/xylog"
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

	e.initInscriptionCache(chain)
	e.initInscriptionStatsCache(chain)
	e.initBalanceCache(chain)
	e.initUtxoCache()
	return e
}

func (h *Manager) initInscriptionCache(chain string) {
	h.Inscription = NewInscription()

	startTs := time.Now()
	idx := 0
	start := uint32(0)
	limit := 10000
	maxSid := uint32(0)
	xylog.Logger.Infof("load inscriptions data start...")
	for {
		items, err := h.db.GetInscriptionsByIdLimit(chain, uint64(start), limit)
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

			if v.SID > maxSid {
				maxSid = v.SID
			}
		}

		//update id index
		start = items[len(items)-1].ID
	}

	//update sid
	h.Inscription.SetSid(maxSid)

	xylog.Logger.Infof("load inscriptions data finished, cost ts:%v", time.Since(startTs))
}

func (h *Manager) initInscriptionStatsCache(chain string) {
	h.InscriptionStats = NewInscriptionStats()

	startTs := time.Now()
	idx := 0
	start := uint32(0)
	limit := 10000
	maxSid := uint32(0)
	xylog.Logger.Infof("load inscription-stats data start...")
	for {
		items, err := h.db.GetInscriptionStatsByIdLimit(chain, uint64(start), limit)
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

			if v.SID > maxSid {
				maxSid = v.SID
			}
		}

		//update id index
		start = items[len(items)-1].ID
	}

	//update sid
	h.InscriptionStats.SetSid(maxSid)

	xylog.Logger.Infof("load inscription-stats data finished, cost ts:%v", time.Since(startTs))
}

func (h *Manager) initBalanceCache(chain string) {
	h.Balance = NewBalance()

	startTs := time.Now()
	idx := 0
	start := uint64(0)
	limit := 10000
	maxSid := uint64(0)
	xylog.Logger.Infof("load balances data start...")
	for {
		balances, err := h.db.GetBalancesByIdLimit(chain, start, limit)
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

			if v.SID > maxSid {
				maxSid = v.SID
			}
		}

		//update id index
		start = balances[len(balances)-1].ID
	}

	//update sid
	h.Balance.SetSid(maxSid)

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
			h.UTXO.Add(v.Protocol, v.Tick, v.TxHash, v.Address, v.Amount)
		}

		//update id index
		start = utxos[len(utxos)-1].ID
	}
	xylog.Logger.Infof("load utxos data finished, cost ts:%v", time.Since(startTs))
}
