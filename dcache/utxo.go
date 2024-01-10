package dcache

import (
	"github.com/shopspring/decimal"
	"strings"
	"sync"
)

// UTXO
/*****************************************************
 * Build cache for all utxo records
 * Mainly used for mint & transfer data checking
 ****************************************************/
type UTXO struct {
	hashes *sync.Map //record mint hash items
}

type UTXOItem struct {
	Protocol string
	Tick     string
	Amount   decimal.Decimal
	Owner    string
	SN       string
}

func NewUTXO() *UTXO {
	return &UTXO{
		hashes: &sync.Map{},
	}
}

/***************************************
 * idx define utxo unique id
 ***************************************/
func (d *UTXO) idx(txHash string) string {
	return strings.ToLower(txHash)
}

// Add
/***************************************
 * Add new utxo record
 ***************************************/
func (d *UTXO) Add(protocol, tick, txHash, address string, amount decimal.Decimal, sn string) {
	idx := d.idx(txHash)
	d.hashes.Store(idx, &UTXOItem{
		Protocol: protocol,
		Tick:     tick,
		Amount:   amount,
		Owner:    address,
		SN:       sn,
	})
}

// Get
/***************************************
 * get utxo record by mint tx hash
 ***************************************/
func (d *UTXO) Get(txHash string) (bool, *UTXOItem) {
	idx := d.idx(txHash)
	item, ok := d.hashes.Load(idx)
	if !ok {
		return false, nil
	}
	return true, item.(*UTXOItem)
}
