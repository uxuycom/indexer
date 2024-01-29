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
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/model"
	"github.com/uxuycom/indexer/xylog"
	"time"
)

type DBAction string

const (
	DBActionCreate DBAction = "create"
	DBActionUpdate DBAction = "update"
)

type DBModelEvent struct {
	Tx               *model.Transaction
	Inscriptions     map[DBAction]*model.Inscriptions
	InscriptionStats map[DBAction]*model.InscriptionsStats
	Balances         map[DBAction][]*model.Balances
	AddressTxs       []*model.AddressTxs
	BalanceTxs       []*model.BalanceTxn
	UTXOs            map[DBAction]*model.UTXO
}

func (tc *TxResultHandler) BuildModel(r *TxResult) *DBModelEvent {
	dm := &DBModelEvent{}

	dm.Tx = tc.BuildTx(r)
	dm.Inscriptions = tc.BuildInscription(r)
	dm.InscriptionStats = tc.BuildInscriptionStats(r)
	dm.BalanceTxs, dm.Balances = tc.BuildBalance(r)
	dm.AddressTxs = tc.BuildAddressTxs(r)
	dm.UTXOs = tc.BuildUTXOs(r)
	return dm
}

func (tc *TxResultHandler) BuildInscription(e *TxResult) map[DBAction]*model.Inscriptions {
	if e.Deploy == nil {
		return nil
	}

	_, d := tc.cache.Inscription.Get(e.MD.Protocol, e.MD.Tick)
	ret := make(map[DBAction]*model.Inscriptions, 1)
	ret[DBActionCreate] = &model.Inscriptions{
		SID:          d.SID,
		Chain:        e.MD.Chain,
		Protocol:     e.MD.Protocol,
		Tick:         e.MD.Tick,
		Name:         e.Deploy.Name,
		LimitPerMint: e.Deploy.MintLimit,
		TotalSupply:  e.Deploy.MaxSupply,
		DeployBy:     e.Tx.From,
		DeployHash:   e.Tx.Hash,
		DeployTime:   time.Unix(int64(e.Block.Time), 0),
		Decimals:     e.Deploy.Decimal,
	}
	return ret
}

func (tc *TxResultHandler) BuildInscriptionStats(e *TxResult) map[DBAction]*model.InscriptionsStats {
	_, d := tc.cache.InscriptionStats.Get(e.MD.Protocol, e.MD.Tick)

	data := &model.InscriptionsStats{
		SID:      d.SID,
		Chain:    e.MD.Chain,
		Protocol: e.MD.Protocol,
		Tick:     e.MD.Tick,
		Minted:   d.Minted,
		Holders:  uint64(d.Holders),
		TxCnt:    d.TxCnt,
	}

	// update mint stats
	if e.Mint != nil {
		// first mint block record
		if d.Minted.Equal(e.Mint.Amount) {
			data.MintFirstBlock = e.Block.Number.Uint64()
		}

		// final mint block record
		_, inscription := tc.cache.Inscription.Get(e.MD.Protocol, e.MD.Tick)
		if inscription.TotalSupply.LessThanOrEqual(d.Minted) {
			data.MintLastBlock = e.Block.Number.Uint64()

			ts := time.Unix(int64(e.Block.Time), 0)
			data.MintCompletedTime = &ts
		}
	}

	if e.Deploy != nil {
		return map[DBAction]*model.InscriptionsStats{
			DBActionCreate: data,
		}
	} else {
		return map[DBAction]*model.InscriptionsStats{
			DBActionUpdate: data,
		}
	}
}

type AddressTxEvent struct {
	Address string
	Amount  decimal.Decimal
}

func (tc *TxResultHandler) BuildAddressTxEvents(e *TxResult) []*AddressTxEvent {
	items := make([]*AddressTxEvent, 0, 10)
	if e.Deploy != nil {
		items = append(items, &AddressTxEvent{
			Address: e.Deploy.Address,
			Amount:  decimal.Zero,
		})
	}

	if e.Mint != nil {
		items = append(items, &AddressTxEvent{
			Address: e.Tx.To,
			Amount:  e.Mint.Amount,
		})
	}

	if e.Transfer != nil {
		sendTotalAmount := decimal.Zero
		for _, item := range e.Transfer.Receives {
			sendTotalAmount = sendTotalAmount.Add(item.Amount)
		}

		items = append(items, &AddressTxEvent{
			Address: e.Transfer.Sender,
			Amount:  sendTotalAmount,
		})

		for _, item := range e.Transfer.Receives {
			items = append(items, &AddressTxEvent{
				Address: item.Address,
				Amount:  item.Amount,
			})
		}
	}

	if e.InscribeTransfer != nil {
		items = append(items, &AddressTxEvent{
			Address: e.InscribeTransfer.Address,
			Amount:  e.InscribeTransfer.Amount,
		})
	}
	return items
}

func (tc *TxResultHandler) BuildAddressTxs(e *TxResult) (txs []*model.AddressTxs) {
	addressTxEvents := tc.BuildAddressTxEvents(e)
	txs = make([]*model.AddressTxs, 0, len(addressTxEvents))
	for _, item := range addressTxEvents {
		txs = append(txs, &model.AddressTxs{
			Event:     tc.getEventByOperate(e.MD.Operate),
			Address:   item.Address,
			Amount:    item.Amount,
			TxHash:    e.Tx.Hash,
			Tick:      e.MD.Tick,
			Protocol:  e.MD.Protocol,
			Operate:   e.MD.Operate,
			Chain:     e.MD.Chain,
			CreatedAt: time.Unix(int64(e.Block.Time), 0),
		})
	}
	return txs
}

func (tc *TxResultHandler) BuildUTXOs(e *TxResult) (items map[DBAction]*model.UTXO) {
	if e.Tx.InscriptionID == "" {
		return
	}

	if e.InscribeTransfer != nil {
		return map[DBAction]*model.UTXO{
			DBActionCreate: {
				Chain:    e.MD.Chain,
				Protocol: e.MD.Protocol,
				Tick:     e.MD.Tick,
				SN:       e.Tx.InscriptionID,
				Address:  e.InscribeTransfer.Address,
				Amount:   e.InscribeTransfer.Amount,
			},
		}
	}

	if e.Transfer != nil {
		address := ""
		amount := decimal.NewFromInt(0)
		if len(e.Transfer.Receives) > 0 {
			address = e.Transfer.Receives[0].Address
			amount = e.Transfer.Receives[0].Amount
		}

		xylog.Logger.Infof("BuildUTXOs txid:%s address:[%s]", e.Tx.Hash, address)
		return map[DBAction]*model.UTXO{
			DBActionUpdate: {
				Chain:   e.MD.Chain,
				SN:      e.Tx.InscriptionID,
				Address: address,
				Amount:  amount,
			},
		}
	}
	return nil
}

func (tc *TxResultHandler) getEventByOperate(operate string) model.TxEvent {
	switch operate {
	case OperateDeploy:
		return model.TransactionEventDeploy
	case OperateMint:
		return model.TransactionEventMint
	case OperateTransfer:
		return model.TransactionEventTransfer
	case OperateList:
		return model.TransactionEventList
	case OperateDelist:
		return model.TransactionEventDelist
	case OperateExchange:
		return model.TransactionEventExchange
	case OperateInscribeTransfer:
		return model.TransactionEventInscribeTransfer
	}
	return model.TxEvent(0)
}

type BalanceTxEvent struct {
	Action           DBAction
	SID              uint64
	Address          string
	Amount           decimal.Decimal
	AvailableBalance decimal.Decimal
	OverallBalance   decimal.Decimal
}

func (tc *TxResultHandler) BuildBalanceTxEvents(e *TxResult) []BalanceTxEvent {
	items := make([]BalanceTxEvent, 0, 10)
	if e.Mint != nil {
		_, balance := tc.cache.Balance.Get(e.MD.Protocol, e.MD.Tick, e.Mint.Minter)
		action := DBActionUpdate
		if e.Mint.Init {
			action = DBActionCreate
		}

		items = append(items, BalanceTxEvent{
			Action:           action,
			SID:              balance.SID,
			Address:          e.Mint.Minter,
			Amount:           e.Mint.Amount,
			AvailableBalance: balance.Available,
			OverallBalance:   balance.Overall,
		})
	}

	if e.Transfer != nil {
		sendTotalAmount := decimal.Zero
		for _, item := range e.Transfer.Receives {
			sendTotalAmount = sendTotalAmount.Add(item.Amount)
		}

		_, senderBalance := tc.cache.Balance.Get(e.MD.Protocol, e.MD.Tick, e.Transfer.Sender)
		items = append(items, BalanceTxEvent{
			Action:           DBActionUpdate,
			SID:              senderBalance.SID,
			Address:          e.Transfer.Sender,
			Amount:           sendTotalAmount.Neg(),
			AvailableBalance: senderBalance.Available,
			OverallBalance:   senderBalance.Overall,
		})

		for _, item := range e.Transfer.Receives {
			_, receiveBalance := tc.cache.Balance.Get(e.MD.Protocol, e.MD.Tick, item.Address)
			action := DBActionUpdate
			if item.Init {
				action = DBActionCreate
			}

			items = append(items, BalanceTxEvent{
				Action:           action,
				SID:              receiveBalance.SID,
				Address:          item.Address,
				Amount:           item.Amount,
				AvailableBalance: receiveBalance.Available,
				OverallBalance:   receiveBalance.Overall,
			})
		}
	}

	if e.InscribeTransfer != nil {
		_, senderBalance := tc.cache.Balance.Get(e.MD.Protocol, e.MD.Tick, e.InscribeTransfer.Address)
		items = append(items, BalanceTxEvent{
			Action:           DBActionUpdate,
			SID:              senderBalance.SID,
			Address:          e.InscribeTransfer.Address,
			Amount:           e.InscribeTransfer.Amount.Neg(),
			AvailableBalance: senderBalance.Available,
			OverallBalance:   senderBalance.Overall,
		})
	}
	return items
}

func (tc *TxResultHandler) BuildBalance(e *TxResult) (txns []*model.BalanceTxn, balances map[DBAction][]*model.Balances) {
	balanceTxEvents := tc.BuildBalanceTxEvents(e)
	txns = make([]*model.BalanceTxn, 0, len(balanceTxEvents))
	balances = make(map[DBAction][]*model.Balances, 2)
	for _, event := range balanceTxEvents {
		txns = append(txns, &model.BalanceTxn{
			Chain:     e.MD.Chain,
			Protocol:  e.MD.Protocol,
			Event:     tc.getEventByOperate(e.MD.Operate),
			Address:   event.Address,
			Tick:      e.MD.Tick,
			Amount:    event.Amount,
			Balance:   event.OverallBalance,
			Available: event.AvailableBalance,
			TxHash:    e.Tx.Hash,
			CreatedAt: time.Unix(int64(e.Block.Time), 0),
		})

		if _, ok := balances[event.Action]; !ok {
			balances[event.Action] = make([]*model.Balances, 0, len(balanceTxEvents))
		}
		balances[event.Action] = append(balances[event.Action], &model.Balances{
			SID:       event.SID,
			Chain:     e.MD.Chain,
			Protocol:  e.MD.Protocol,
			Address:   event.Address,
			Tick:      e.MD.Tick,
			Balance:   event.OverallBalance,
			Available: event.AvailableBalance,
		})
	}
	return txns, balances
}

func (tc *TxResultHandler) BuildTx(e *TxResult) *model.Transaction {
	return &model.Transaction{
		Chain:           e.MD.Chain,
		Protocol:        e.MD.Protocol,
		BlockHeight:     e.Tx.BlockNumber.Uint64(),
		PositionInBlock: e.Tx.TxIndex.Uint64(),
		BlockTime:       time.Unix(int64(e.Block.Time), 0),
		TxHash:          e.Tx.Hash,
		From:            e.Tx.From,
		To:              e.Tx.To,
		Op:              e.MD.Operate,
		Tick:            e.MD.Tick,
		Gas:             e.Tx.Gas.Int64(),
		GasPrice:        e.Tx.GasPrice.Int64(),
	}
}

type DBModelsFattened struct {
	Inscriptions     map[DBAction][]*model.Inscriptions
	InscriptionStats map[DBAction][]*model.InscriptionsStats
	Balances         map[DBAction][]*model.Balances
	UTXOs            map[DBAction][]*model.UTXO
	Txs              []*model.Transaction
	AddressTxs       []*model.AddressTxs
	BalanceTxs       []*model.BalanceTxn
	BlockStatus      *model.BlockStatus
}

type DBModels struct {
	Inscriptions     map[DBAction]map[uint32]*model.Inscriptions
	InscriptionStats map[DBAction]map[uint32]*model.InscriptionsStats
	Balances         map[DBAction]map[uint64]*model.Balances
	UTXOs            map[DBAction]map[string]*model.UTXO
	Txs              map[string]*model.Transaction
	AddressTxs       []*model.AddressTxs
	BalanceTxs       []*model.BalanceTxn
}

func BuildDBUpdateModel(blocksEvents []*Event) (dmf *DBModelsFattened) {
	dm := &DBModels{
		Inscriptions: map[DBAction]map[uint32]*model.Inscriptions{
			DBActionCreate: make(map[uint32]*model.Inscriptions, 100),
			DBActionUpdate: make(map[uint32]*model.Inscriptions, 100),
		},
		InscriptionStats: map[DBAction]map[uint32]*model.InscriptionsStats{
			DBActionCreate: make(map[uint32]*model.InscriptionsStats, 100),
			DBActionUpdate: make(map[uint32]*model.InscriptionsStats, 100),
		},
		Balances: map[DBAction]map[uint64]*model.Balances{
			DBActionCreate: make(map[uint64]*model.Balances, 100),
			DBActionUpdate: make(map[uint64]*model.Balances, 100),
		},
		UTXOs: map[DBAction]map[string]*model.UTXO{
			DBActionCreate: make(map[string]*model.UTXO, 100),
			DBActionUpdate: make(map[string]*model.UTXO, 100),
		},
		Txs:        make(map[string]*model.Transaction, len(blocksEvents)*2),
		AddressTxs: make([]*model.AddressTxs, 0, len(blocksEvents)*2),
		BalanceTxs: make([]*model.BalanceTxn, 0, len(blocksEvents)*2),
	}
	for _, blockEvent := range blocksEvents {
		for _, event := range blockEvent.Items {
			for action, item := range event.Inscriptions {
				if _, ok := dm.Inscriptions[action][item.SID]; ok {
					xylog.Logger.Debugf("ins sid[%d] exist & force update, tick[%s]", item.SID, item.Tick)
				}
				dm.Inscriptions[action][item.SID] = item
			}

			for action, item := range event.InscriptionStats {
				if lastItem, ok := dm.InscriptionStats[action][item.SID]; ok {
					xylog.Logger.Debugf("ins stats sid[%d] exist & force update, tick[%s]", item.SID, item.Tick)

					// reserve history mint stats data if exist
					if lastItem.MintFirstBlock > 0 {
						item.MintFirstBlock = lastItem.MintFirstBlock
					}
					if lastItem.MintLastBlock > 0 {
						item.MintLastBlock = lastItem.MintLastBlock
					}
					if lastItem.MintCompletedTime != nil {
						item.MintCompletedTime = lastItem.MintCompletedTime
					}
				}
				dm.InscriptionStats[action][item.SID] = item
			}

			txIdx := event.Tx.TxHash
			if _, ok := dm.Txs[txIdx]; ok {
				xylog.Logger.Debugf("tx[%s] exist & force update", txIdx)
			}
			dm.Txs[txIdx] = event.Tx

			if len(event.AddressTxs) > 0 {
				dm.AddressTxs = append(dm.AddressTxs, event.AddressTxs...)
			}

			if len(event.BalanceTxs) > 0 {
				dm.BalanceTxs = append(dm.BalanceTxs, event.BalanceTxs...)
			}

			for action, item := range event.UTXOs {
				if _, ok := dm.UTXOs[action][item.SN]; ok {
					xylog.Logger.Debugf("utxo sn[%s] exist & force update, tick[%s]", item.SN, item.Tick)
				}
				dm.UTXOs[action][item.SN] = item
			}

			for action, items := range event.Balances {
				for _, item := range items {
					if _, ok := dm.Balances[action][item.SID]; ok {
						xylog.Logger.Debugf("balance sid[%d] exist & force update, address[%s]-tick[%s]", item.SID, item.Address, item.Tick)
					}
					dm.Balances[action][item.SID] = item
				}
			}
		}
	}

	lastBlockEvent := blocksEvents[len(blocksEvents)-1]
	bs := &model.BlockStatus{
		Chain:       lastBlockEvent.Chain,
		BlockHash:   lastBlockEvent.BlockHash,
		BlockNumber: lastBlockEvent.BlockNum,
		BlockTime:   time.Unix(int64(lastBlockEvent.BlockTime), 0),
	}

	dmf = &DBModelsFattened{
		Inscriptions: map[DBAction][]*model.Inscriptions{
			DBActionCreate: make([]*model.Inscriptions, 0, 100),
			DBActionUpdate: make([]*model.Inscriptions, 0, 100),
		},
		InscriptionStats: map[DBAction][]*model.InscriptionsStats{
			DBActionCreate: make([]*model.InscriptionsStats, 0, 100),
			DBActionUpdate: make([]*model.InscriptionsStats, 0, 100),
		},
		Balances: map[DBAction][]*model.Balances{
			DBActionCreate: make([]*model.Balances, 0, 100),
			DBActionUpdate: make([]*model.Balances, 0, 100),
		},
		UTXOs: map[DBAction][]*model.UTXO{
			DBActionCreate: make([]*model.UTXO, 0, 100),
			DBActionUpdate: make([]*model.UTXO, 0, 100),
		},
		Txs:         make([]*model.Transaction, 0, len(dm.Txs)),
		AddressTxs:  dm.AddressTxs,
		BalanceTxs:  dm.BalanceTxs,
		BlockStatus: bs,
	}

	// flatten tx
	for _, tx := range dm.Txs {
		dmf.Txs = append(dmf.Txs, tx)
	}

	// flatten inscription records
	for _, item := range dm.Inscriptions[DBActionCreate] {
		dmf.Inscriptions[DBActionCreate] = append(dmf.Inscriptions[DBActionCreate], item)
	}
	for _, item := range dm.Inscriptions[DBActionUpdate] {
		dmf.Inscriptions[DBActionUpdate] = append(dmf.Inscriptions[DBActionUpdate], item)
	}

	// flatten inscription stats records
	for _, item := range dm.InscriptionStats[DBActionCreate] {
		dmf.InscriptionStats[DBActionCreate] = append(dmf.InscriptionStats[DBActionCreate], item)
	}
	for _, item := range dm.InscriptionStats[DBActionUpdate] {
		dmf.InscriptionStats[DBActionUpdate] = append(dmf.InscriptionStats[DBActionUpdate], item)
	}

	// flatten balances records
	for _, item := range dm.Balances[DBActionCreate] {
		dmf.Balances[DBActionCreate] = append(dmf.Balances[DBActionCreate], item)
	}
	for _, item := range dm.Balances[DBActionUpdate] {
		dmf.Balances[DBActionUpdate] = append(dmf.Balances[DBActionUpdate], item)
	}

	// flatten balances records
	for _, item := range dm.UTXOs[DBActionCreate] {
		dmf.UTXOs[DBActionCreate] = append(dmf.UTXOs[DBActionCreate], item)
	}
	for _, item := range dm.UTXOs[DBActionUpdate] {
		dmf.UTXOs[DBActionUpdate] = append(dmf.UTXOs[DBActionUpdate], item)
	}
	return dmf
}
