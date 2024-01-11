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
	"github.com/uxuycom/indexer/utils"
	"strings"
	"sync"
)

// Inscription
/*****************************************************
 * Build cache for all inscriptions
 * Mainly used for basic meta data query
 *********************Remo*******************************/
type Inscription struct {
	sid       uint32
	ticks     *sync.Map
	tickNames *sync.Map // used for asc20
}

type Tick struct {
	SID          uint32
	TransferType int8
	LimitPerMint decimal.Decimal
	TotalSupply  decimal.Decimal
	Decimals     int8
}

func NewInscription() *Inscription {
	return &Inscription{
		ticks:     &sync.Map{},
		tickNames: &sync.Map{},
	}
}

/***************************************
 * idx define protocol tick unique id
 ***************************************/
func (d *Inscription) idx(protocol, tick string) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(protocol), strings.ToLower(tick))
}

// Create
/***************************************
 * init tick's metadata
 ***************************************/
func (d *Inscription) Create(protocol, tick string, nt *Tick) {
	// Add auto_increment ID
	if nt.SID <= 0 {
		d.sid++
		nt.SID = d.sid
	}
	idx := d.idx(protocol, tick)
	d.ticks.Store(idx, nt)

	// asc20 Add cache names
	if protocol == "asc-20" {
		key := utils.Keccak256(strings.ToLower(tick))
		d.tickNames.Store(key, tick)
	}
}

// SetSid set auto_increment id
func (d *Inscription) SetSid(sid uint32) {
	if sid > d.sid {
		d.sid = sid
	}
}

// Update
/***************************************
 * update tick's data
 ***************************************/
func (d *Inscription) Update(protocol, tick string, nt *Tick) *Tick {
	ok, t := d.Get(protocol, tick)
	if !ok {
		return nil
	}

	if nt.TransferType > 0 {
		t.TransferType = nt.TransferType
	}
	return t
}

// Get
/***************************************
 * get tick meta data contains filed (id, transfer_type)
 ***************************************/
func (d *Inscription) Get(protocol, tick string) (bool, *Tick) {
	idx := d.idx(protocol, tick)
	t, ok := d.ticks.Load(idx)
	if !ok {
		return false, nil
	}
	return true, t.(*Tick)
}

// GetNameByIdx
/***************************************
 * get tick name by idx
 ***************************************/
func (d *Inscription) GetNameByIdx(key string) (bool, string) {
	key = strings.TrimPrefix(key, "0x")
	name, ok := d.tickNames.Load(key)
	if !ok {
		return false, ""
	}
	return true, name.(string)
}
