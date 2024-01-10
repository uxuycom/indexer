package asc20

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"math/big"
	"open-indexer/client/xycommon"
	"open-indexer/devents"
	"open-indexer/utils"
	"open-indexer/xyerrors"
	"open-indexer/xylog"
)

const ExchangeMethodID = "0xd9b3d6d0"

const (
	// EventTopicHashExchange avascriptions_protocol_TransferASC20TokenForListing (index_topic_1 address from, index_topic_2 address to, bytes32 id)
	EventTopicHashExchange = "0xe2750d6418e3719830794d3db788aa72febcd657bcd18ed8f1facdbf61a69a9a"

	// EventTopicHashOrderExecuted ASC20OrderExecuted (address seller, address taker, bytes32 listId, string ticker, uint256 amount, uint256 price, uint16 feeRate, uint64 timestamp)
	EventTopicHashOrderExecuted = "0x3efe873bf4d1c1061b9980e7aed9b564e024844522ec8c80aec160809948ef77"

	// EventTopicHashExchange2 avascriptions_protocol_TransferASC20Token(index_topic_1 address, index_topic_2 address, index_topic_3 tick_idx string, amount uint256)
	EventTopicHashExchange2 = "0x8cdf9e10a7b20e7a9c4e778fc3eb28f2766e438a9856a62eac39fbd2be98cbc2"
)

type Exchange struct {
	Tick   string
	From   string
	To     string
	Amount decimal.Decimal
}

func (p *Protocol) Exchange(block *xycommon.RpcBlock, tx *xycommon.RpcTransaction, omd *devents.MetaData) (items []*devents.TxResult, err *xyerrors.InsError) {
	exchanges := p.extractValidOrders(tx)
	if len(exchanges) <= 0 {
		return nil, nil
	}

	items = make([]*devents.TxResult, 0, len(exchanges))
	for _, exchange := range exchanges {
		md := omd.Copy()
		md.Operate = devents.OperateExchange
		md.Tick = exchange.Tick
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

type OrderExecuted struct {
	Seller    common.Address `json:"seller"`
	Taker     common.Address `json:"taker"`
	ListId    [32]byte       `json:"listId"`
	Ticker    string         `json:"ticker"`
	Amount    *big.Int       `json:"amount"`
	Price     *big.Int       `json:"price"`
	FeeRate   uint16         `json:"feeRate"`
	Timestamp uint64         `json:"timestamp"`
}

func (p *Protocol) parseOrderByExchange(transferEvent, orderExecuteEvent xycommon.RpcLog) (*Exchange, *xyerrors.InsError) {
	orderExecuteResult := &OrderExecuted{}
	_, err := utils.ParseEventToStruct(ParsedABI, utils.EventLog{
		Address: orderExecuteEvent.Address,
		Topics:  orderExecuteEvent.Topics,
		Data:    orderExecuteEvent.Data,
	}, orderExecuteResult)
	if err != nil {
		return nil, xyerrors.NewInsError(-10, fmt.Sprintf("tx execute event parse error[%v], event[%v]", err, orderExecuteEvent))
	}

	if orderExecuteResult.Amount.String() == "" {
		return nil, xyerrors.NewInsError(-17, fmt.Sprintf("tx execute event parse failed, amount value empty, ticker[%v]", orderExecuteResult.Ticker))
	}

	if hex.EncodeToString(orderExecuteResult.ListId[:]) != transferEvent.Data.String()[2:] {
		return nil, xyerrors.NewInsError(-17, fmt.Sprintf("tx execute event parse failed, listId not match, ticker[%v]", orderExecuteResult.Ticker))
	}

	return &Exchange{
		Tick:   orderExecuteResult.Ticker,
		From:   transferEvent.Address.String(),
		To:     common.BytesToAddress(transferEvent.Topics[2].Bytes()).String(),
		Amount: decimal.NewFromBigInt(orderExecuteResult.Amount, 0),
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
	Ticker string         `json:"ticker"`
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

	ok, tick := p.cache.Inscription.GetNameByIdx(transferASC20TokenResult.Ticker)
	if !ok {
		return nil, xyerrors.NewInsError(-11, fmt.Sprintf("tx execute event parse failed, tick not found, idx[%s]", transferASC20TokenResult.Ticker))
	}

	if transferASC20TokenResult.Amount.String() == "" {
		return nil, xyerrors.NewInsError(-17, fmt.Sprintf("tx execute event parse failed, amount value empty, tick[%v]", tick))
	}

	return &Exchange{
		Tick:   tick,
		From:   transferASC20TokenResult.From.String(),
		To:     transferASC20TokenResult.To.String(),
		Amount: decimal.NewFromBigInt(transferASC20TokenResult.Amount, 0),
	}, nil
}

func (p *Protocol) extractValidOrdersByTransfer(tx *xycommon.RpcTransaction) []*Exchange {
	items := make([]*Exchange, 0, len(tx.Events))
	for _, e := range tx.Events {
		if len(e.Topics) != 3 || e.Topics[0].String() != EventTopicHashExchange2 {
			continue
		}

		item, err := p.parseOrderByTransfer(e)
		if err != nil {
			xylog.Logger.Infof("tx[%s] - transfer decode err:%v", tx.Hash, err)
			continue
		}
		items = append(items, item)
	}
	return items
}

func (p *Protocol) extractValidOrdersByExchange(tx *xycommon.RpcTransaction) []*Exchange {
	items := make([]*Exchange, 0, len(tx.Events))
	for i, e := range tx.Events {
		if i >= len(tx.Events)-1 {
			break
		}

		if len(e.Topics) != 3 || e.Topics[0].String() != EventTopicHashExchange {
			continue
		}

		orderExecuteEvent := tx.Events[i+1]
		if len(orderExecuteEvent.Topics) > 0 && orderExecuteEvent.Topics[0].String() == EventTopicHashOrderExecuted {
			item, err := p.parseOrderByExchange(e, orderExecuteEvent)
			if err != nil {
				xylog.Logger.Infof("tx[%v] - exchange decode err:%v", tx, err)
				continue
			}
			items = append(items, item)
		}
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
