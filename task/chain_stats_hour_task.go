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

package task

import (
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/config"
	"github.com/uxuycom/indexer/model"
	"github.com/uxuycom/indexer/storage"
	"github.com/uxuycom/indexer/xylog"
	"strconv"
	"sync"
	"time"
)

type ChainStatsTask struct {
	Task
}

const (
	limit = 5000
)

func NewChainStatsTask(dbc *storage.DBClient, cfg *config.Config) *ChainStatsTask {
	task := &ChainStatsTask{
		Task{
			dbc: dbc,
			cfg: cfg,
		},
	}
	return task
}

func (t *ChainStatsTask) Exec() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	xylog.Logger.Infof("task starting...")
	for {
		select {
		case <-ticker.C:
			xylog.Logger.Infof("Exec ChainStatsTask  task!")
			chain := t.cfg.Chain.ChainName
			// get current hour and last hour
			//now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-02-21 20:00:00", time.Local)
			now := time.Now()
			nowHour := now.Truncate(time.Hour)
			lastHour := now.Add(-1 * time.Hour).Truncate(time.Hour)
			format := lastHour.Format("2006010215")
			parseUint, _ := strconv.ParseUint(format, 10, 32)
			chainStatHour := &model.ChainStatHour{
				AddressCount:      0,
				InscriptionsCount: 0,
				BalanceSum:        decimal.NewFromInt(0),
				Chain:             chain,
				CreatedAt:         now,
				UpdatedAt:         now,
				DateHour:          uint32(parseUint),
			}
			u, _ := strconv.ParseUint(lastHour.Add(-1*time.Hour).Truncate(time.Hour).Format("2006010215"), 10, 32)
			chainStat, _ := t.dbc.FindLastChainStatHourByChainAndDateHour(chain, uint32(u))
			if chainStat == nil {
				// first stat
				chainStat = &model.ChainStatHour{
					AddressLastId: t.cfg.Stat.AddressStartId,
					BalanceLastId: t.cfg.Stat.BalanceStartId,
					Chain:         chain,
				}
			}

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				// address
				HandleAddress(t, chainStat, chainStatHour, nowHour, lastHour)
				defer wg.Done()
			}()

			go func() {
				// balance
				HandleBalance(t, chainStat, chainStatHour, nowHour, lastHour)
				defer wg.Done()
			}()
			wg.Wait()

			// inscriptions
			inscriptions, _ := t.dbc.FindInscriptionsTxByIdAndChainAndLimit(chainStat.Chain, nowHour, lastHour)
			chainStatHour.InscriptionsCount = uint32(len(inscriptions))
			// add stat
			err := t.dbc.AddChainStatHour(chainStatHour)
			if err != nil {
				xylog.Logger.Errorf("AddChainStatHour error: %v chainStatHour: %v", err, chainStatHour)
			}
		}

	}
}
func HandleAddress(t *ChainStatsTask, chainStat, chainStatHour *model.ChainStatHour,
	nowHour, lastHour time.Time) *model.ChainStatHour {

	addresses, _ := t.dbc.FindAddressTxByIdAndChainAndLimit(chainStat.Chain, chainStat.AddressLastId, limit)
	if len(addresses) > 0 {
		for _, a := range addresses {
			if a.CreatedAt.After(nowHour) {
				return chainStatHour
			}
			if a.CreatedAt.Before(nowHour) && a.CreatedAt.After(lastHour) {
				chainStatHour.AddressCount++
				chainStatHour.AddressLastId = a.ID
			}
			chainStat.AddressLastId = a.ID
		}
		return HandleAddress(t, chainStat, chainStatHour, nowHour, lastHour)
	} else {
		return chainStatHour
	}
}

func HandleBalance(t *ChainStatsTask, chainStat, chainStatHour *model.ChainStatHour,
	nowHour, lastHour time.Time) *model.ChainStatHour {

	balances, _ := t.dbc.FindBalanceTxByIdAndChainAndLimit(chainStat.Chain, chainStat.BalanceLastId, limit)
	if len(balances) > 0 {
		for _, b := range balances {
			if b.CreatedAt.After(nowHour) {
				return chainStatHour
			}
			if b.CreatedAt.Before(nowHour) && b.CreatedAt.After(lastHour) {
				chainStatHour.BalanceSum = chainStatHour.BalanceSum.Add(b.Amount)
				chainStatHour.BalanceLastId = b.ID
			}
			chainStat.BalanceLastId = b.ID
		}
		return HandleBalance(t, chainStat, chainStatHour, nowHour, lastHour)
	} else {
		return chainStatHour
	}
}
