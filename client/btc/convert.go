package btc

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/shopspring/decimal"
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/dcache"
	"github.com/uxuycom/indexer/xylog"
	"math"
	"math/big"
	"time"
)

type Convert struct {
	decimals    int
	chainParams *chaincfg.Params
	btcClient   *RawClient
	ordClient   *OrdClient
	cache       *dcache.Manager
}

func NewConvert(decimals int, client *RawClient, testnet bool, ordEndpoint string, cache *dcache.Manager) *Convert {
	chainParams := &chaincfg.MainNetParams
	if testnet {
		chainParams = &chaincfg.TestNet3Params
	}
	return &Convert{
		cache:       cache,
		decimals:    decimals,
		btcClient:   client,
		chainParams: chainParams,
		ordClient:   NewOrdClient(ordEndpoint),
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

	brc20Inscriptions, err := c.ordClient.BlockBRC20Inscriptions(context.Background(), block.Height)
	if err != nil {
		return nil, fmt.Errorf("prefetchRawTransactions err[%v], block[%d]", err, block.Height)
	}

	for idx, tx := range block.Tx {
		if inscription, ok := brc20Inscriptions[tx.Txid]; ok {
			insTx, err1 := c.convertInscriptionTx(block.Height, idx, len(block.Tx), tx, inscription)
			if err1 != nil {
				return nil, fmt.Errorf("convert inscription tx[%+v] idx[%d] data error[%v]", tx, idx, err1)
			}
			cBlock.Transactions = append(cBlock.Transactions, insTx)
		} else {
			ttx, err1 := c.convertTransferTx(block.Height, idx, len(block.Tx), tx, brc20Inscriptions)
			if err1 != nil {
				return nil, fmt.Errorf("convert transaction tx[%+v] idx[%d] data error[%v]", tx, idx, err1)
			}
			if ttx == nil {
				continue
			}
			cBlock.Transactions = append(cBlock.Transactions, ttx)
		}
	}
	xylog.Logger.Infof("convert block[%d] txs[%d] success", block.Height, len(cBlock.Transactions))
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

func (c *Convert) convertInscriptionTx(blockHeight int64, idx int, num int, tx btcjson.TxRawResult, inscription Inscription) (*xycommon.RpcTransaction, error) {
	startTs := time.Now()
	defer func() {
		xylog.Logger.Debugf("[%d/%d]convertInscriptionTx cost[%v], block[%d], tx[%s]", idx+1, num, time.Since(startTs), blockHeight, tx.Txid)
	}()

	blockHash, _ := chainhash.NewHashFromStr(tx.BlockHash)
	txHash, _ := chainhash.NewHashFromStr(tx.Txid)
	rtx := &xycommon.RpcTransaction{
		BlockHash:     blockHash.String(),
		BlockNumber:   big.NewInt(blockHeight),
		TxIndex:       big.NewInt(int64(idx)),
		Hash:          txHash.String(),
		InscriptionID: inscription.ID,
	}
	rtx.Input = inscription.Content
	addr, err := c.getOutputReceivers(tx.Vout, inscription.Meta, true)
	if err != nil {
		return nil, fmt.Errorf("getOutputReceivers error[%v]", err)
	}

	rtx.To = addr
	rtx.Gas = big.NewInt(int64(inscription.Meta.GenesisFee))
	if tx.Vsize > 0 {
		rtx.GasPrice = big.NewInt(1).Div(rtx.Gas, big.NewInt(int64(tx.Vsize)))
	}
	return rtx, nil
}

func (c *Convert) extractFirstOutputReceivers(vouts []btcjson.Vout) (to string, err error) {
	for _, vout := range vouts {
		addr, err1 := c.getAddressFromScriptPubKey(vout.ScriptPubKey)
		if err1 != nil {
			return "", fmt.Errorf("getAddressFromScriptPubKey, ScriptPubKey[%+v] err[%v]", vout.ScriptPubKey, err1)
		}
		if addr == nil {
			continue
		}
		return addr.EncodeAddress(), nil
	}
	return "", fmt.Errorf("can't find output address")
}

func (c *Convert) getOutputReceivers(vouts []btcjson.Vout, metadata InscriptionMeta, failover bool) (to string, err error) {
	addrs := make([]string, 0, len(vouts))
	for _, vout := range vouts {
		addr, err1 := c.getAddressFromScriptPubKey(vout.ScriptPubKey)
		if err1 != nil {
			return "", fmt.Errorf("getAddressFromScriptPubKey, ScriptPubKey[%+v] err[%v]", vout.ScriptPubKey, err1)
		}
		if addr == nil {
			continue
		}

		if c.convertBitcoinSats(decimal.NewFromFloat(vout.Value)).Equal(decimal.NewFromInt(metadata.OutputValue)) {
			return addr.EncodeAddress(), nil
		}
		addrs = append(addrs, addr.EncodeAddress())
	}

	if failover {
		return addrs[0], nil
	}
	return "", nil
}

func (c *Convert) tryFindInscriptionByTxID(txId string, brc20Inscriptions map[string]Inscription) bool {
	// query from cache
	if ok, _ := c.cache.UTXO.Get(txId); ok {
		return true
	}

	if _, ok := brc20Inscriptions[txId]; ok {
		return true
	}
	return false
}

func (c *Convert) findInscriptionFromVins(vins []btcjson.Vin, brc20Inscriptions map[string]Inscription) (bool, string, error) {
	for _, vin := range vins {
		if vin.IsCoinBase() {
			continue
		}

		// try find from cache & current block inscriptions
		if !c.tryFindInscriptionByTxID(vin.Txid, brc20Inscriptions) {
			continue
		}

		// valid inscription input
		output, err := c.ordClient.InscriptionOutput(context.Background(), vin.Txid, vin.Vout)
		if err != nil {
			return false, "", fmt.Errorf("get inscription output error[%v], txid[%s:%d]", err, vin.Txid, vin.Vout)
		}

		// if output inscriptions nil, it means that the input is not inscription
		if len(output.Inscriptions) == 0 {
			continue
		}
		return true, output.Inscriptions[0], nil
	}
	return false, "", nil
}

func (c *Convert) convertTransferTx(blockHeight int64, idx, num int, tx btcjson.TxRawResult, brc20Inscriptions map[string]Inscription) (*xycommon.RpcTransaction, error) {
	startTs := time.Now()
	defer func() {
		xylog.Logger.Debugf("[%d/%d]convertTransferTx cost[%v], block[%d], tx[%s]", idx+1, num, time.Since(startTs), blockHeight, tx.Txid)
	}()

	blockHash, _ := chainhash.NewHashFromStr(tx.BlockHash)
	txHash, _ := chainhash.NewHashFromStr(tx.Txid)

	ttx := &xycommon.RpcTransaction{
		BlockHash:   blockHash.String(),
		BlockNumber: big.NewInt(blockHeight),
		TxIndex:     big.NewInt(int64(idx)),
		Type:        big.NewInt(0),
		Hash:        txHash.String(),
	}

	// query inscription from vins
	ok, inscriptionID, err := c.findInscriptionFromVins(tx.Vin, brc20Inscriptions)
	if err != nil {
		return nil, fmt.Errorf("findInscriptionFromVins error[%v]", err)
	}
	if !ok {
		return nil, nil
	}
	xylog.Logger.Infof("find tx inscription[%s] from vins, tx[%s]", inscriptionID, tx.Txid)
	ttx.InscriptionID = inscriptionID

	// query inscription metadata
	metadata, err := c.ordClient.InscriptionMetaByID(context.Background(), inscriptionID)
	if err != nil {
		return nil, fmt.Errorf("get inscription metadata error[%v], id[%s]", err, inscriptionID)
	}

	// query inscription content
	ttx.Input, err = c.ordClient.InscriptionContentByID(context.Background(), inscriptionID)
	if err != nil {
		return nil, fmt.Errorf("get inscription content error[%v], id[%s]", err, inscriptionID)
	}

	// extract inscription output address
	ttx.From = "00"
	ttx.To, err = c.getOutputReceivers(tx.Vout, metadata, false)
	if err != nil {
		return nil, fmt.Errorf("getOutputReceivers error[%v]", err)
	}

	// Iterate over each transaction input
	totalInAmount, err := c.getTxVinTotalAmount(tx.Vin)
	if err != nil {
		return nil, fmt.Errorf("getTxVinTotalAmount error[%v]", err)
	}

	totalOutAmount := decimal.Zero
	for _, output := range tx.Vout {
		totalOutAmount = totalOutAmount.Add(decimal.NewFromFloatWithExponent(output.Value, int32(-c.decimals)))
	}

	feeAmount := totalInAmount.Sub(totalOutAmount)
	feeAmountSats := c.convertBitcoinSats(feeAmount)
	gasPrice := decimal.Zero
	if tx.Vsize > 0 {
		gasPrice = feeAmountSats.Div(decimal.NewFromInt(int64(tx.Vsize)))
	}
	ttx.Gas = feeAmountSats.BigInt()
	ttx.GasPrice = c.convertBitcoinSats(gasPrice).BigInt()
	return ttx, nil
}

func (c *Convert) getTxVinTotalAmount(vins []btcjson.Vin) (amount decimal.Decimal, err error) {
	hashes := make([]string, 0, len(vins))
	for _, vin := range vins {
		if vin.IsCoinBase() {
			continue
		}
		hashes = append(hashes, vin.Txid)
	}

	prevOuts, err := c.btcClient.GetMultiRawTransactionVerbose(context.Background(), hashes)
	if err != nil {
		return decimal.Zero, fmt.Errorf("getMultiRawTransactionVerbose error[%v]", err)
	}

	for _, vin := range vins {
		if vin.IsCoinBase() {
			continue
		}
		amount = amount.Add(decimal.NewFromFloat(prevOuts[vin.Txid].Vout[vin.Vout].Value))
	}
	return
}
