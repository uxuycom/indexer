package btc

import (
	"context"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/ethereum/go-ethereum"
	"math/big"
	"open-indexer/client/xycommon"
)

// BClient defines typed wrappers for the Ethereum RPC API.
type BClient struct {
	testnet bool
	client  *RawClient
	convert *Convert
}

// NewClient creates a client that uses the given RPC client.
func NewClient(rpc string) (*BClient, error) {
	client, err := NewRawClient(rpc)
	if err != nil {
		return nil, err
	}

	convert := NewConvert(8, client, false)
	return &BClient{
		testnet: false,
		client:  client,
		convert: convert,
	}, nil
}

// Close closes the underlying RPC connection.
func (c *BClient) Close() {
	c.client.Close()
}

// BlockNumber returns the most recent block number
func (c *BClient) BlockNumber(ctx context.Context) (uint64, error) {
	return c.client.BlockNumber(ctx)
}

func (c *BClient) HeaderByNumber(ctx context.Context, number *big.Int) (*xycommon.RpcHeader, error) {
	return c.convert.convertHeader(c.client.HeaderByNumber(ctx, number))
}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (c *BClient) BlockByNumber(ctx context.Context, number *big.Int) (*xycommon.RpcBlock, error) {
	return c.convert.convertBlock(c.client.BlockByNumber(ctx, number))
}

// TransactionSender returns the sender address of the given transaction. The transaction
// must be known to the remote node and included in the blockchain at the given block and
// index. The sender is the one derived by the protocol at the time of inclusion.
//
// There is a fast-path for transactions retrieved by TransactionByHash and
// TransactionInBlock. Getting their sender address can be done without an RPC interaction.
func (c *BClient) TransactionSender(ctx context.Context, txHashStr, blockHashStr string, txIndex uint) (string, error) {
	txs, _, err := c.getTransactionsByHash(ctx, txHashStr)
	if err != nil {
		return "", fmt.Errorf("getTransactionsByHash[%s] error[%v]", txHashStr, err)
	}

	sendersMap := make(map[string]struct{}, len(txs))
	senders := make([]string, 0, len(txs))
	for _, tx := range txs {
		if tx.From != "" {
			sendersMap[tx.From] = struct{}{}
			senders = append(senders, tx.From)
		}
	}

	if len(sendersMap) == 1 {
		return senders[0], nil
	}
	return "", nil
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (c *BClient) TransactionReceipt(ctx context.Context, txHashStr string) (*xycommon.RpcReceipt, error) {
	return nil, nil
}

func (c *BClient) getTransactionsByHash(ctx context.Context, txHashStr string) (txs []*xycommon.RpcTransaction, isPending bool, err error) {
	rpctx, err := c.client.GetRawTransactionVerbose(ctx, txHashStr)
	if err != nil {
		return nil, false, fmt.Errorf("GetRawTransactionVerbose tx[%s] error[%v]", txHashStr, err)
	}

	var block *btcjson.GetBlockVerboseResult
	if rpctx.BlockHash != "" {
		block, err = c.client.GetBlockVerbose(ctx, rpctx.BlockHash)
		if err != nil {
			return nil, false, fmt.Errorf("GetBlockVerbose block hash[%s] error[%v]", rpctx.BlockHash, err)
		}
	} else {
		block = &btcjson.GetBlockVerboseResult{
			Height: 0,
			Tx:     []string{txHashStr},
		}
	}

	idx := 0
	for i, txStr := range block.Tx {
		if txStr == txHashStr {
			idx = i
			break
		}
	}
	isPending = block.Confirmations <= 0

	hashes := make([]string, 0, len(block.Tx[idx])*2)
	for _, vin := range rpctx.Vin {
		// Retrieve the previous transaction output
		if vin.IsCoinBase() {
			continue
		}
		hashes = append(hashes, vin.Txid)
	}
	rawTxs, err := c.client.GetMultiRawTransactionVerbose(ctx, hashes)
	if err != nil {
		return nil, false, fmt.Errorf("GetMultiRawTransactionVerbose err[%v], hash[%v], tx[%s]", err, hashes, txHashStr)
	}

	rtxs, err := c.convert.convertTransaction(block.Height, idx, *rpctx, rawTxs)
	if err != nil {
		return nil, false, fmt.Errorf("convertTransaction error[%v]", err)
	}

	//filter tx from = to
	txs = make([]*xycommon.RpcTransaction, 0, len(txs))
	for _, tx := range rtxs {
		if tx.From != tx.To {
			txs = append(txs, tx)
		}
	}
	return txs, isPending, nil
}

func (c *BClient) TransactionByHash(ctx context.Context, hashStr string) (tx *xycommon.RpcTransaction, isPending bool, err error) {
	txs, isPending, err := c.getTransactionsByHash(ctx, hashStr)
	if err != nil {
		return nil, false, fmt.Errorf("getTransactionsByHash error[%v]", err)
	}

	if len(txs) > 1 {
		return nil, false, fmt.Errorf("multi input addrs -> multi output addrs not support now. txs[%v]", txs)
	}
	return txs[0], isPending, nil
}

// FilterLogs executes a filter query.
func (ec *BClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]xycommon.RpcLog, error) {
	return nil, nil
}
