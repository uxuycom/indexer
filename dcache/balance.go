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
	"strings"
	"sync"
)

// Balance
/*****************************************************
 * Build cache for all ticks all address balance
 * Mainly used for real-time verification of data
 ****************************************************/
type Balance struct {
	sid   uint64
	ticks *sync.Map
}

type BalanceItem struct {
	SID       uint64
	Available decimal.Decimal
	Overall   decimal.Decimal
}

func NewBalance() *Balance {
	return &Balance{
		ticks: &sync.Map{},
	}
}

/***************************************
 * idx define protocol tick unique id
 ***************************************/
func (d *Balance) idx(protocol, tick, address string) string {
	return fmt.Sprintf("%s_%s_%s", strings.ToLower(protocol), strings.ToLower(tick), strings.ToLower(address))
}

// Update
/***************************************
 * update addr tick's balance
 ***************************************/
func (d *Balance) Update(protocol, tick string, addr string, b *BalanceItem) *BalanceItem {
	//idx := d.idx(protocol, tick, addr)
	ok, balanceItem := d.Get(protocol, tick, addr)
	if !ok {
		return nil
	}

	balanceItem.Available = b.Available
	balanceItem.Overall = b.Overall
	return balanceItem
}

// Create
/***************************************
 * create addr tick's balance
 ***************************************/
func (d *Balance) Create(protocol, tick string, addr string, b *BalanceItem) *BalanceItem {
	if b.SID <= 0 {
		d.sid++
		b.SID = d.sid
	}

	balanceItem := &BalanceItem{
		SID:       b.SID,
		Available: b.Available,
		Overall:   b.Overall,
	}

	idx := d.idx(protocol, tick, addr)
	d.ticks.Store(idx, balanceItem)
	return balanceItem
}

// SetSid set auto_increment id
func (d *Balance) SetSid(sid uint64) {
	if sid > d.sid {
		d.sid = sid
	}
}

// Get
/***************************************
 * get addr tick's balance
 ***************************************/
func (d *Balance) Get(protocol, tick string, addr string) (ok bool, val *BalanceItem) {
	idx := d.idx(protocol, tick, addr)
	balances, ok := d.ticks.Load(idx)
	if !ok {
		return false, nil
	}
	//addr = strings.ToLower(addr)
	return true, balances.(*BalanceItem)
}
