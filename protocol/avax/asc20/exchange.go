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

package asc20

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/utils"
	"github.com/uxuycom/indexer/xyerrors"
	"github.com/uxuycom/indexer/xylog"
	"math/big"
	"strings"
)

const (
	// EventTopicHashExchange avascriptions_protocol_TransferASC20TokenForListing (index_topic_1 address from, index_topic_2 address to, bytes32 id)
	EventTopicHashExchange = "0xe2750d6418e3719830794d3db788aa72febcd657bcd18ed8f1facdbf61a69a9a"

	// EventTopicHashExchange2 avascriptions_protocol_TransferASC20Token(index_topic_1 address, index_topic_2 address, index_topic_3 tick_idx string, amount uint256)
	// https://snowtrace.io/tx/0x71e2b9c31608f89b4e191af8817a61b3df4d74d4f829a75000c8a9e4ca67c4f2/eventlog?chainId=43114
	EventTopicHashExchange2 = "0x8cdf9e10a7b20e7a9c4e778fc3eb28f2766e438a9856a62eac39fbd2be98cbc2"
)

type Exchange struct {
	Operate string
	Tick    string
	From    string
	To      string
	Amount  decimal.Decimal
}

// ASC20Order is an auto generated low-level Go binding around an user-defined struct.
type ASC20Order struct {
	Seller  common.Address
	Creator common.Address
	ListId  [32]byte
	Ticker  string
	Amount  *big.Int
	Price   *big.Int
	Operate string
}

func (p *Protocol) Exchange(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, omd *devents.MetaData) (items []*devents.TxResult, err *xyerrors.InsError) {
	// extract valid orders
	exchanges := p.extractValidOrders(tx)
	if len(exchanges) <= 0 {
		return nil, nil
	}

	items = make([]*devents.TxResult, 0, len(exchanges))
	for _, exchange := range exchanges {
		md := omd.Copy()
		md.Operate = exchange.Operate
		md.Tick = strings.ToLower(strings.TrimSpace(exchange.Tick))
		if err1 := p.verifyExchange(md, exchange); err1 != nil {
			xylog.Logger.Infof("exchange verified failed, err:%v, data:%v", err1, exchange)
			continue
		}

		item := &devents.TxResult{
			MD:    md,
			Block: block,
			Tx:    tx,
			Transfer: &devents.Transfer{
				Sender: exchange.From,
				Receives: []*devents.Receive{
					{
						Address: exchange.To,
						Amount:  exchange.Amount,
					},
				},
			},
		}
		items = append(items, item)
	}
	return
}

func (p *Protocol) parseOrderByExchange(e xycommon.RpcLog, orders map[string]*ASC20Order) (*Exchange, *xyerrors.InsError) {
	order, ok := orders[e.Data.String()]
	if !ok {
		return nil, nil
	}

	if order.Amount.String() == "" {
		return nil, xyerrors.NewInsError(-17, fmt.Sprintf("order amount value empty, ticker[%v]", order.Ticker))
	}

	return &Exchange{
		Operate: order.Operate,
		Tick:    order.Ticker,
		From:    e.Address.String(),
		To:      common.BytesToAddress(e.Topics[2].Bytes()).String(),
		Amount:  decimal.NewFromBigInt(order.Amount, 0),
	}, nil
}

func (p *Protocol) extractValidOrders(tx *xycommon.RpcTransaction) []*Exchange {
	mixed := make([]*Exchange, 0, len(tx.Events))
	items := p.extractValidOrdersByExchange(tx)
	if len(items) > 0 {
		mixed = append(mixed, items...)
	}

	items = p.extractValidOrdersByTransfer(tx)
	if len(items) > 0 {
		mixed = append(mixed, items...)
	}
	return mixed
}

type TransferASC20Token struct {
	From   common.Address `json:"from"`
	To     common.Address `json:"to"`
	Ticker common.Hash    `json:"ticker"`
	Amount *big.Int       `json:"amount"`
}

func (p *Protocol) parseOrderByTransfer(transferEvent xycommon.RpcLog) (*Exchange, *xyerrors.InsError) {
	transferASC20TokenResult := &TransferASC20Token{}
	_, err := utils.ParseEventToStruct(ParsedABI, utils.EventLog{
		Address: transferEvent.Address,
		Topics:  transferEvent.Topics,
		Data:    transferEvent.Data,
	}, transferASC20TokenResult)
	if err != nil {
		return nil, xyerrors.NewInsError(-10, fmt.Sprintf("tx execute event parse error[%v], event[%v]", err, transferEvent))
	}

	ok, tick := p.cache.Inscription.GetNameByIdx(transferASC20TokenResult.Ticker.String())
	if !ok {
		return nil, xyerrors.NewInsError(-11, fmt.Sprintf("tx execute event parse failed, tick not found, idx[%s]", transferASC20TokenResult.Ticker))
	}

	if transferASC20TokenResult.Amount.String() == "" {
		return nil, xyerrors.NewInsError(-17, fmt.Sprintf("tx execute event parse failed, amount value empty, tick[%v]", tick))
	}

	return &Exchange{
		Operate: devents.OperateExchange,
		Tick:    tick,
		From:    transferASC20TokenResult.From.String(),
		To:      transferASC20TokenResult.To.String(),
		Amount:  decimal.NewFromBigInt(transferASC20TokenResult.Amount, 0),
	}, nil
}

func (p *Protocol) extractValidOrdersByTransfer(tx *xycommon.RpcTransaction) []*Exchange {
	items := make([]*Exchange, 0, len(tx.Events))
	for _, e := range tx.Events {
		if len(e.Topics) != 4 || e.Topics[0].String() != EventTopicHashExchange2 {
			continue
		}

		xylog.Logger.Infof("hit avax-transfer-types, tx:%s", tx.Hash)
		item, err := p.parseOrderByTransfer(e)
		if err != nil {
			xylog.Logger.Infof("tx[%s] - transfer decode err:%v", tx.Hash, err)
			continue
		}
		items = append(items, item)
	}
	return items
}

func (p *Protocol) extractInputOrders(hash, input string) map[string]*ASC20Order {
	callData := common.FromHex(input)
	if len(callData) < 4 {
		return nil
	}

	sigData, argData := callData[:4], callData[4:]
	method, err := ParsedABI.MethodById(sigData)
	if err != nil {
		xylog.Logger.Errorf("tx[%s] - exchange input abi method query err:%v", hash, err)
		return nil
	}

	unpacked, err := method.Inputs.UnpackValues(argData)
	if err != nil {
		xylog.Logger.Errorf("tx[%s] - exchange input abi method UnpackValues err:%v", hash, err)
		return nil
	}

	orderOperate := devents.OperateExchange
	if method.Name == "cancelOrder" || method.Name == "cancelOrders" {
		orderOperate = devents.OperateDelist
	}

	orders := make(map[string]*ASC20Order, 10)
	for k, v := range unpacked {
		encodeBytes, _ := json.Marshal(v)
		if method.Inputs[k].Name == "order" {
			item := &ASC20Order{}
			if err = json.Unmarshal(encodeBytes, item); err != nil {
				xylog.Logger.Errorf("tx[%s] - exchange input abi method Unmarshal order err:%v", hash, err)
				return nil
			}

			item.Operate = orderOperate
			orders[common.BytesToHash(item.ListId[:]).String()] = item
			return orders
		}

		if method.Inputs[k].Name == "orders" {
			items := make([]*ASC20Order, 0, 10)
			if err = json.Unmarshal(encodeBytes, &items); err != nil {
				xylog.Logger.Errorf("tx[%s] - exchange input abi method Unmarshal orders err:%v", hash, err)
				return nil
			}

			if len(items) == 0 {
				xylog.Logger.Errorf("tx[%s] - exchange input abi method orders is nil", hash)
				return nil
			}

			for _, item := range items {
				item.Operate = orderOperate
				orders[common.BytesToHash(item.ListId[:]).String()] = item
			}
			return orders
		}
	}
	return nil
}

func (p *Protocol) extractValidOrdersByExchange(tx *xycommon.RpcTransaction) []*Exchange {
	orders := p.extractInputOrders(tx.Hash, tx.Input)
	if len(orders) == 0 {
		return nil
	}

	items := make([]*Exchange, 0, len(tx.Events))
	for _, e := range tx.Events {
		if len(e.Topics) != 3 || e.Topics[0].String() != EventTopicHashExchange {
			continue
		}

		item, err := p.parseOrderByExchange(e, orders)
		if err != nil {
			xylog.Logger.Infof("tx[%v] - exchange decode err:%v", tx, err)
			continue
		}
		items = append(items, item)
	}
	return items
}

func (p *Protocol) verifyExchange(md *devents.MetaData, e *Exchange) *xyerrors.InsError {
	var (
		protocol = md.Protocol
		tick     = md.Tick
	)
	ok, inscription := p.cache.Inscription.Get(protocol, tick)
	if !ok || inscription == nil {
		return xyerrors.NewInsError(-15, fmt.Sprintf("inscription not exist, protocol[%s]-tick[%s]", protocol, tick))
	}

	// sender balance checking
	ok, balance := p.cache.Balance.Get(protocol, tick, e.From)
	if !ok {
		return xyerrors.NewInsError(-16, fmt.Sprintf("sender balance record not exist, tick[%s-%s], address[%s]", protocol, tick, e.From))
	}

	// balance available checking
	if balance.Overall.LessThan(e.Amount) {
		return xyerrors.NewInsError(-17, fmt.Sprintf("sender total balance[%v] < transfer amount[%v]", balance.Overall, e.Amount))
	}
	return nil
}
