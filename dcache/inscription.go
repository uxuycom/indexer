package dcache

import (
	"fmt"
	"github.com/shopspring/decimal"
	"open-indexer/utils"
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
		fmt.Println("key:", key)
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
	name, ok := d.tickNames.Load(key)
	if !ok {
		return false, ""
	}
	return true, name.(string)
}
