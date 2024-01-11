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
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/xylog"
	"strings"
	"sync"
)

// InscriptionStats
/*****************************************************
 * Build cache for all inscriptions stats data
 * Mainly used for statics data query
 ****************************************************/
type InscriptionStats struct {
	sid   uint32
	ticks *sync.Map
}

type InsStats struct {
	SID     uint32
	Minted  decimal.Decimal
	Holders int64
	TxCnt   uint64
}

func NewInscriptionStats() *InscriptionStats {
	return &InscriptionStats{
		ticks: &sync.Map{},
	}
}

/***************************************
 * idx define protocol tick unique id
 ***************************************/
func (d *InscriptionStats) idx(protocol, tick string) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(protocol), strings.ToLower(tick))
}

// Update
/***************************************
 * update ticks
 ***************************************/
func (d *InscriptionStats) Update(protocol, tick string, stats *InsStats) *InsStats {
	ok, insStats := d.Get(protocol, tick)
	if !ok {
		return nil
	}

	if stats.Minted.GreaterThan(decimal.Zero) {
		insStats.Minted = stats.Minted
	}

	if stats.Holders > 0 {
		insStats.Holders = stats.Holders
	}

	if stats.TxCnt > 0 {
		insStats.TxCnt = stats.TxCnt
	}
	return insStats
}

// Create
/***************************************
 * init tick's id
 ***************************************/
func (d *InscriptionStats) Create(protocol, tick string, stats *InsStats) *InsStats {
	// Add auto_increment ID
	if stats.SID <= 0 {
		d.sid++
		stats.SID = d.sid
	}

	idx := d.idx(protocol, tick)
	d.ticks.Store(idx, stats)
	return stats
}

func (d *InscriptionStats) Mint(protocol, tick string, amount decimal.Decimal) *InsStats {
	ok, insStats := d.Get(protocol, tick)
	if !ok {
		return nil
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return insStats
	}

	insStats.Minted = insStats.Minted.Add(amount)
	return insStats
}

func (d *InscriptionStats) Holders(protocol, tick string, incr int64) *InsStats {
	ok, insStats := d.Get(protocol, tick)
	if !ok {
		return nil
	}

	insStats.Holders = insStats.Holders + incr

	if insStats.Holders < 0 {
		xylog.Logger.Fatalf("protocol:%s, tick:%s holders < 0", protocol, tick)
	}
	return insStats
}

func (d *InscriptionStats) TxCnt(protocol, tick string, incr uint64) *InsStats {
	ok, insStats := d.Get(protocol, tick)
	if !ok {
		return nil
	}

	insStats.TxCnt = insStats.TxCnt + incr
	return insStats
}

// SetSid set auto_increment id
func (d *InscriptionStats) SetSid(sid uint32) {
	if sid > d.sid {
		d.sid = sid
	}
}

// Get
/***************************************
 * get tick meta data contains filed (id, transfer_type)
 ***************************************/
func (d *InscriptionStats) Get(protocol, tick string) (bool, *InsStats) {
	idx := d.idx(protocol, tick)
	t, ok := d.ticks.Load(idx)
	if !ok {
		return false, nil
	}
	return true, t.(*InsStats)
}
