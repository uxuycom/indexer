package btc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
	"open-indexer/client/xycommon"
	"open-indexer/xylog"
)

type Convert struct {
	decimals    int
	chainParams *chaincfg.Params
	client      *RawClient
}

func NewConvert(decimals int, client *RawClient, testnet bool) *Convert {
	chainParams := &chaincfg.MainNetParams
	if testnet {
		chainParams = &chaincfg.TestNet3Params
	}
	return &Convert{
		decimals:    decimals,
		client:      client,
		chainParams: chainParams,
	}
}

func (c *Convert) convertHeader(block *btcjson.GetBlockVerboseResult, be error) (header *xycommon.RpcHeader, err error) {
	if be != nil {
		return nil, be
	}
	header = &xycommon.RpcHeader{
		ParentHash: block.PreviousHash,
		Number:     big.NewInt(block.Height),
		Time:       uint64(block.Time),
		TxHash:     block.MerkleRoot,
	}
	return header, nil
}

func (c *Convert) prefetchRawTransactions(block *btcjson.GetBlockVerboseTxResult) (txs map[string]*btcjson.TxRawResult, err error) {
	hashes := make([]string, 0, len(block.Tx)*2)
	for _, tx := range block.Tx {
		for _, vin := range tx.Vin {
			// Retrieve the previous transaction output
			if vin.IsCoinBase() {
				continue
			}
			hashes = append(hashes, vin.Txid)
		}
	}
	return c.client.GetMultiRawTransactionVerbose(context.Background(), hashes)
}

func (c *Convert) convertBlock(block *btcjson.GetBlockVerboseTxResult, be error) (cBlock *xycommon.RpcBlock, err error) {
	if be != nil {
		return nil, be
	}

	cBlock = &xycommon.RpcBlock{
		ParentHash:   block.PreviousHash,
		Number:       big.NewInt(block.Height),
		Time:         uint64(block.Time),
		TxHash:       block.MerkleRoot,
		Hash:         block.Hash,
		Transactions: make([]*xycommon.RpcTransaction, 0, len(block.Tx)),
	}

	rawTxs, err := c.prefetchRawTransactions(block)
	if err != nil {
		return nil, fmt.Errorf("prefetchRawTransactions err[%v], block[%d]", err, block.Height)
	}

	for idx, tx := range block.Tx {
		fTxs, err1 := c.convertTransaction(block.Height, idx, tx, rawTxs)
		if err1 != nil {
			return nil, fmt.Errorf("convert transaction tx[%+v] idx[%d] data error[%v]", tx, idx, err1)
		}
		cBlock.Transactions = append(cBlock.Transactions, fTxs...)
	}
	return cBlock, nil
}

func (c *Convert) getAddressFromScriptPubKey(scriptPubKey btcjson.ScriptPubKeyResult) (btcutil.Address, error) {
	if scriptPubKey.Address != "" {
		addr, err := btcutil.DecodeAddress(scriptPubKey.Address, c.chainParams)
		if err == nil {
			return addr, nil
		}
	}

	scriptPubKeyBytes, err := hex.DecodeString(scriptPubKey.Hex)
	if err != nil {
		return nil, fmt.Errorf("decode scriptPubKey hex[%s] err[%v]", scriptPubKey.Hex, err)
	}

	_, addresses, _, err := txscript.ExtractPkScriptAddrs(scriptPubKeyBytes, c.chainParams)
	if err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, nil
	}
	return addresses[0], nil
}

func (c *Convert) convertBitcoinSats(value decimal.Decimal) decimal.Decimal {
	if value.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}
	return value.Mul(decimal.NewFromFloat(math.Pow10(c.decimals))).Round(int32(c.decimals))
}

func (c *Convert) convertTransaction(blockHeight int64, idx int, tx btcjson.TxRawResult, prevRawTxs map[string]*btcjson.TxRawResult) ([]*xycommon.RpcTransaction, error) {
	blockHash, _ := chainhash.NewHashFromStr(tx.BlockHash)
	txHash, _ := chainhash.NewHashFromStr(tx.Txid)

	totalOutAmount := decimal.Zero
	for _, output := range tx.Vout {
		totalOutAmount = totalOutAmount.Add(decimal.NewFromFloatWithExponent(output.Value, int32(-c.decimals)))
	}

	// Iterate over each transaction input
	totalInAmount := decimal.Zero
	senders := make(map[int]btcutil.Address, len(tx.Vin))
	sendersMap := make(map[string]struct{}, len(tx.Vin))
	prevTxs := make(map[int]*btcjson.TxRawResult, len(tx.Vin))
	for j, vin := range tx.Vin {
		// Retrieve the previous transaction output
		if vin.IsCoinBase() {
			continue
		}

		prevTx, ok := prevRawTxs[vin.Txid]
		if !ok {
			xylog.Logger.Infof("GetRawTransactionVerbose rpc tx[%s] nil", vin.Txid)
			continue
		}
		prevTxs[j] = prevTx
		if len(prevTx.Vout) == 0 || int(vin.Vout) >= len(prevTx.Vout) {
			prefTxBytes, _ := json.Marshal(prevTx)
			xylog.Logger.Infof("get prev tx out data empty, tx[%s-%d], data[%s]", vin.Txid, vin.Vout, prefTxBytes)
			continue
		}

		senderAddress, err := c.getAddressFromScriptPubKey(prevTx.Vout[vin.Vout].ScriptPubKey)
		if err != nil {
			return nil, fmt.Errorf("decode vin public[%+v] addr error[%v]", prevTx.Vout[vin.Vout].ScriptPubKey, err)
		}

		if senderAddress == nil {
			continue
		}
		senders[j] = senderAddress
		totalInAmount = totalInAmount.Add(decimal.NewFromFloatWithExponent(prevTx.Vout[vin.Vout].Value, int32(-c.decimals)))
		sendersMap[senderAddress.EncodeAddress()] = struct{}{}
	}
	feeAmount := totalInAmount.Sub(totalOutAmount)
	feeAmountSats := c.convertBitcoinSats(feeAmount)
	gasPrice := decimal.Zero
	if tx.Vsize > 0 {
		gasPrice = feeAmountSats.Div(decimal.NewFromInt(int64(tx.Vsize)))
	}

	// Iterate over each transaction output
	receivers := make(map[int]btcutil.Address, len(tx.Vout))
	receiversMap := make(map[string]struct{}, len(tx.Vout))
	for k, vout := range tx.Vout {
		// Retrieve the receiver address from the transaction output
		receiverAddress, err := c.getAddressFromScriptPubKey(vout.ScriptPubKey)
		if err != nil {
			return nil, fmt.Errorf("getAddressFromScriptPubKey, ScriptPubKey[%+v] err[%v]", vout.ScriptPubKey, err)
		}

		if receiverAddress == nil {
			continue
		}
		receivers[k] = receiverAddress
		receiversMap[receiverAddress.EncodeAddress()] = struct{}{}
	}

	defaultTx := &xycommon.RpcTransaction{
		BlockHash:   blockHash.String(),
		BlockNumber: big.NewInt(blockHeight),
		TxIndex:     big.NewInt(int64(idx)),
		Type:        big.NewInt(0),
		Hash:        txHash.String(),
		Gas:         big.NewInt(feeAmountSats.IntPart()),
		GasPrice:    big.NewInt(gasPrice.IntPart()),
	}

	txs := make([]*xycommon.RpcTransaction, 0, len(sendersMap)+len(receiversMap))
	if len(sendersMap) <= 1 {
		var from btcutil.Address
		if len(senders) > 0 {
			from = senders[0]
		}
		for j, vout := range tx.Vout {
			to := receivers[j]
			rtx := c.formatTx(defaultTx, from, to, vout.Value)
			txs = append(txs, rtx)
		}
	} else if len(receiversMap) == 1 {
		to := receivers[0]
		for k, vin := range tx.Vin {
			if vin.IsCoinBase() {
				continue
			}

			from := senders[k]
			if int(vin.Vout) >= len(prevTxs[k].Vout) {
				xylog.Logger.Infof("get vin prev tx empty, tx[%s-%d]", vin.Txid, vin.Vout)
				continue
			}

			rtx := c.formatTx(defaultTx, from, to, prevTxs[k].Vout[vin.Vout].Value)
			txs = append(txs, rtx)
		}
	} else {
		for j, vin := range tx.Vin {
			if vin.IsCoinBase() {
				continue
			}

			if int(vin.Vout) >= len(prevTxs[j].Vout) {
				xylog.Logger.Infof("get vin prev tx empty, tx[%s-%d]", vin.Txid, vin.Vout)
				continue
			}

			from := senders[j]
			rtx := c.formatTx(defaultTx, from, nil, prevTxs[j].Vout[vin.Vout].Value)
			txs = append(txs, rtx)
		}

		for k, vout := range tx.Vout {
			to := receivers[k]
			rtx := c.formatTx(defaultTx, nil, to, vout.Value)
			txs = append(txs, rtx)
		}
	}
	return txs, nil
}

func (c *Convert) formatTx(defaultTx *xycommon.RpcTransaction, from, to btcutil.Address, vout float64) *xycommon.RpcTransaction {
	fromAddr := ""
	if from != nil {
		fromAddr = from.String()
	}

	toAddr := ""
	if to != nil {
		toAddr = to.String()
	}

	return &xycommon.RpcTransaction{
		BlockHash:   defaultTx.BlockHash,
		BlockNumber: defaultTx.BlockNumber,
		TxIndex:     defaultTx.TxIndex,
		Type:        defaultTx.Type,
		Hash:        defaultTx.Hash,
		ChainID:     defaultTx.ChainID,
		From:        fromAddr,
		To:          toAddr,
		Value:       big.NewInt(c.convertBitcoinSats(decimal.NewFromFloat(vout)).IntPart()),
		Gas:         defaultTx.Gas,
		GasPrice:    defaultTx.GasPrice,
	}
}
