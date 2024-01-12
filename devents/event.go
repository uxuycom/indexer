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

package devents

import (
	"context"
	"github.com/uxuycom/indexer/storage"
	"github.com/uxuycom/indexer/xylog"
	"gorm.io/gorm"
	"time"
)

type Event struct {
	Chain     string
	BlockNum  uint64
	BlockTime uint64
	BlockHash string
	Items     []*DBModelEvent
}

type DEvent struct {
	ctx    context.Context
	events chan *Event
	db     *storage.DBClient
}

func NewDEvents(ctx context.Context, db *storage.DBClient) *DEvent {
	return &DEvent{
		ctx:    ctx,
		db:     db,
		events: make(chan *Event, 1024),
	}
}

func (h *DEvent) WriteDBAsync(e *Event) {
	h.events <- e
}

func (h *DEvent) Read(num int) (items []*Event) {
	items = make([]*Event, 0, num)
	for i := 0; i < num; i++ {
		select {
		case item := <-h.events:
			items = append(items, item)
		default:
			return items
		}
	}
	return
}

func (h *DEvent) Flush() {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	xylog.Logger.Infof("start flushing...")
	for {
		select {
		case <-t.C:
			if !h.Sink(h.db) {
				return
			}
		case <-h.ctx.Done():
			return
		}
	}
}

func (h *DEvent) Sink(db *storage.DBClient) bool {
	//get events from channel
	events := h.Read(100)

	// merge events data
	if len(events) < 1 {
		return true
	}

	dm := BuildDBUpdateModel(events)
	chain := dm.BlockStatus.Chain

	startTs := time.Now()
	err := db.SqlDB.Transaction(func(tx *gorm.DB) error {
		// insert inscriptions
		if items := dm.Inscriptions[DBActionCreate]; len(items) > 0 {
			if err := db.BatchAddInscription(tx, items); err != nil {
				xylog.Logger.Errorf("failed to save the inscription. err=%s", err)
				return err
			}
		}

		// update inscriptions
		if items := dm.Inscriptions[DBActionUpdate]; len(items) > 0 {
			err := db.BatchUpdateInscription(tx, chain, items)
			if err != nil {
				xylog.Logger.Errorf("failed to update inscription. err=%s", err)
				return err
			}
		}

		// insert inscriptions stats
		if items := dm.InscriptionStats[DBActionCreate]; len(items) > 0 {
			err := db.BatchAddInscriptionStats(tx, items)
			if err != nil {
				xylog.Logger.Errorf("failed to update inscription. err=%s", err)
				return err
			}
		}

		if items := dm.InscriptionStats[DBActionUpdate]; len(items) > 0 {
			// batch updates，minted / holders / tx_cnt
			err := db.BatchUpdateInscriptionStats(tx, chain, items)
			if err != nil {
				xylog.Logger.Errorf("failed to update inscription. err=%s", err)
				return err
			}

			//update mint ext data
			for _, item := range items {
				if item.MintFirstBlock == 0 && item.MintLastBlock == 0 && item.MintCompletedTime == nil {
					continue
				}

				updates := make(map[string]interface{})
				if item.MintFirstBlock > 0 {
					updates["mint_first_block"] = item.MintFirstBlock
				}

				if item.MintLastBlock > 0 {
					updates["mint_last_block"] = item.MintLastBlock
				}

				if item.MintCompletedTime != nil {
					updates["mint_completed_time"] = item.MintCompletedTime
				}

				err = db.UpdateInscriptionsStatsBySID(tx, chain, item.SID, updates)
				if err != nil {
					xylog.Logger.Errorf("failed to update inscription stats. err=%s", err)
					return err
				}
			}
		}

		// insert transactions
		if len(dm.Txs) > 0 {
			if err := db.BatchAddTransaction(tx, dm.Txs); err != nil {
				xylog.Logger.Errorf("failed to create transactions. err=%s", err)
				return err
			}
		}

		// insert address transactions
		if len(dm.AddressTxs) > 0 {
			if err := db.BatchAddAddressTx(tx, dm.AddressTxs); err != nil {
				xylog.Logger.Errorf("failed insert address transaction records. err=%s", err)
				return err
			}
		}

		// insert balance related transactions
		if len(dm.BalanceTxs) > 0 {
			if err := db.BatchAddBalanceTx(tx, dm.BalanceTxs); err != nil {
				xylog.Logger.Errorf("failed insert balances related tx records. err=%s", err)
				return err
			}
		}

		// update balances
		if items := dm.Balances[DBActionCreate]; len(items) > 0 {
			if err := db.BatchAddBalances(tx, items); err != nil {
				xylog.Logger.Errorf("failed insert balances records. err=%s", err)
				return err
			}
		}

		// update inscriptions
		if items := dm.Balances[DBActionUpdate]; len(items) > 0 {
			err := db.BatchUpdateBalances(tx, chain, items)
			if err != nil {
				xylog.Logger.Errorf("failed update balances records. err=%s", err)
				return err
			}
		}

		// record block status
		if err := db.SaveLastBlock(tx, dm.BlockStatus); err != nil {
			xylog.Logger.Errorf("failed to save block information. err=%s", err)
			return err
		}
		return nil
	})

	if err != nil {
		xylog.Logger.Errorf("flush db error. err=%s, cost:%v", err, time.Since(startTs))
		return false
	}
	xylog.Logger.Infof("flush db success, cost:%v", time.Since(startTs))
	return true
}
