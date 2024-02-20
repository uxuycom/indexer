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
	"time"
)

type ChainStatsTask struct {
	Task
}

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
			chains, _ := t.dbc.FindAllChain()
			limit := 5000
			for i := range chains {
				chainStatHour := &model.ChainStatHour{}
				chainStatHour.AddressCount = 0
				chainStatHour.InscriptionsCount = 0
				chainStatHour.BalanceSum = decimal.NewFromInt(0)

				chainStatHour.Chain = chains[i].Chain

				// get current hour and next hour
				now := time.Now()
				chainStatHour.CreatedAt = now
				chainStatHour.UpdatedAt = now

				nowHour := now.Truncate(time.Hour)
				lastHour := now.Add(-1 * time.Hour).Truncate(time.Hour)
				format := lastHour.Format("2006010215")
				parseUint, _ := strconv.ParseUint(format, 10, 32)
				chainStatHour.DateHour = uint32(parseUint)

				u, _ := strconv.ParseUint(lastHour.Add(-1*time.Hour).Truncate(time.Hour).Format("2006010215"), 10, 32)
				chainStat, _ := t.dbc.FindLastChainStatHourByChainAndDateHour(chains[i].Chain, uint32(u))
				if chainStat == nil {
					// first stat
					chainStat.AddressLastId = t.cfg.Stat.AddressStartId
					chainStat.BalanceLastId = t.cfg.Stat.BalanceStartId
					chainStat.Chain = chains[i].Chain
				}
				// address
				addressIndex := chainStat.AddressLastId
				address, _ := t.dbc.FindAddressTxByIdAndChainAndLimit(chainStat.Chain, addressIndex, limit)
				for {
					if len(address) > 0 {
						for j := range address {
							if address[j].CreatedAt.After(nowHour) {
								break
							}
							if address[j].CreatedAt.Before(nowHour) && address[j].CreatedAt.After(lastHour) {
								chainStatHour.AddressCount++
								chainStatHour.AddressLastId = address[j].ID
							}
							addressIndex = address[j].ID
						}
						address, _ = t.dbc.FindAddressTxByIdAndChainAndLimit(chainStat.Chain, addressIndex, limit)
					} else {
						break
					}
				}
				// inscriptions
				inscriptions, _ := t.dbc.FindInscriptionsTxByIdAndChainAndLimit(chainStat.Chain, nowHour, lastHour)
				chainStatHour.InscriptionsCount = uint32(len(inscriptions))
				// balance
				balanceIndex := chainStat.BalanceLastId
				balance, _ := t.dbc.FindBalanceTxByIdAndChainAndLimit(chainStat.Chain, balanceIndex, limit)
				for {
					if len(balance) > 0 {
						for j := range balance {
							if balance[j].CreatedAt.After(nowHour) {
								break
							}
							if balance[j].CreatedAt.Before(nowHour) && balance[j].CreatedAt.After(lastHour) {
								chainStatHour.BalanceSum.Add(balance[j].Amount)
								chainStatHour.BalanceLastId = balance[j].ID
							}
							balanceIndex = balance[j].ID
						}
						balance, _ = t.dbc.FindBalanceTxByIdAndChainAndLimit(chainStat.Chain, balanceIndex, limit)
					} else {
						break
					}
				}
				// add stat
				err := t.dbc.AddChainStatHour(chainStatHour)
				if err != nil {
					xylog.Logger.Errorf("AddChainStatHour error: %v chainStatHour: %v", err, chainStatHour)
					return
				}
			}

		}
	}
}
