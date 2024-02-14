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

package evm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/uxuycom/indexer/xylog"
	"math/big"
	"time"
)

// RawClient defines typed wrappers for the Ethereum RPC API.
type RawClient struct {
	c *rpc.Client
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *RawClient {
	return &RawClient{c}
}

// Close closes the underlying RPC connection.
func (ec *RawClient) Close() {
	ec.c.Close()
}

// Client gets the underlying RPC client.
func (ec *RawClient) Client() *rpc.Client {
	return ec.c
}

func (ec *RawClient) doCallContext(retry int, result interface{}, method string, args ...interface{}) (err error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	t1 := time.Now()
	err = ec.c.CallContext(timeCtx, result, method, args...)

	//build logs
	if method == "eth_getLogs" {
		args = []interface{}{"<hidden>"}
	}
	msg := fmt.Sprintf("JSONRPC-CALL, method:%s, args[%v], cost[%v]", method, args, time.Since(t1))
	if retry > 0 {
		msg += fmt.Sprintf(", retry[%d]", retry)
	}

	if err != nil {
		msg += fmt.Sprintf(", err[%v]", err)
	}
	xylog.Logger.Debug(msg)
	return
}

func (ec *RawClient) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) (err error) {
	retry := 10
	for i := 0; i < retry; i++ {
		//call
		err = ec.doCallContext(i, result, method, args...)
		if err == nil {
			if result == nil {
				return rpc.ErrNoResult
			}
			d, _ := json.Marshal(result)
			xylog.Logger.Debugf("CallContext result=%v", string(d))
			return nil
		}

		if errors.Is(err, rpc.ErrNoResult) {
			return rpc.ErrNoResult
		}

		if err.Error() == "cannot query unfinalized data" {
			return rpc.ErrNoResult
		}

		select {
		case <-time.After(time.Millisecond * 100):
			//do nothing
		case <-ctx.Done():
			return errors.New("ctx done quit")
		}
	}
	return err
}

// Blockchain Access

// ChainID retrieves the current chain ID for transaction replay protection.
func (ec *RawClient) ChainID(ctx context.Context) (*big.Int, error) {
	var result hexutil.Big
	err := ec.CallContext(ctx, &result, "eth_chainId")
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), err
}

// BlockNumber returns the most recent block number
func (ec *RawClient) BlockNumber(ctx context.Context) (uint64, error) {
	var result hexutil.Uint64
	err := ec.CallContext(ctx, &result, "eth_blockNumber")
	return uint64(result), err
}

func (ec *RawClient) HeaderByNumber(ctx context.Context, number *big.Int) (*RpcHeader, error) {
	var result RpcHeader
	err := ec.CallContext(ctx, &result, "eth_getBlockByNumber", toBlockNumArg(number), true)
	if err == nil && result.Number == nil {
		return nil, rpc.ErrNoResult
	}
	return &result, err
}

// BlockByHash returns the given full block.
//
// Note that loading full blocks requires two requests. Use HeaderByHash
// if you don't need all transactions or uncle headers.
func (ec *RawClient) BlockByHash(ctx context.Context, hash common.Hash) (*RpcBlock, error) {
	return ec.getBlock(ctx, "eth_getBlockByHash", hash, true)
}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (ec *RawClient) BlockByNumber(ctx context.Context, number *big.Int) (*RpcBlock, error) {
	return ec.getBlock(ctx, "eth_getBlockByNumber", toBlockNumArg(number), true)
}

func (ec *RawClient) getBlock(ctx context.Context, method string, args ...interface{}) (*RpcBlock, error) {
	var result RpcBlock
	err := ec.CallContext(ctx, &result, method, args...)
	if err != nil {
		return nil, err
	}

	if result.Number == nil {
		return nil, rpc.ErrNoResult
	}

	if result.TxHash != types.EmptyTxsHash && result.TxHash.String() != "0x0000000000000000000000000000000000000000000000000000000000000000" && len(result.Transactions) == 0 {
		return nil, fmt.Errorf("block[%v] tx header checked failed, b[%+v]", args, result)
	}
	return &result, nil
}

// FilterLogs executes a filter query.
func (ec *RawClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]RpcLog, error) {
	var result []RpcLog
	arg, err := toFilterArg(q)
	if err != nil {
		return nil, err
	}
	err = ec.CallContext(ctx, &result, "eth_getLogs", arg)
	return result, err
}

// SubscribeFilterLogs subscribes to the results of a streaming filter query.
func (ec *RawClient) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	arg, err := toFilterArg(q)
	if err != nil {
		return nil, err
	}
	sub, err := ec.c.EthSubscribe(ctx, ch, "logs", arg)
	if err != nil {
		// Defensively prefer returning nil interface explicitly on error-path, instead
		// of letting default golang behavior wrap it with non-nil interface that stores
		// nil concrete type value.
		return nil, err
	}
	return sub, nil
}

// SubscribeNewHead subscribes to notifications about the current blockchain head
// on the given channel.
func (ec *RawClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	sub, err := ec.c.EthSubscribe(ctx, ch, "newHeads")
	if err != nil {
		// Defensively prefer returning nil interface explicitly on error-path, instead
		// of letting default golang behavior wrap it with non-nil interface that stores
		// nil concrete type value.
		return nil, err
	}
	return sub, nil
}

// TransactionSender returns the sender address of the given transaction. The transaction
// must be known to the remote node and included in the blockchain at the given block and
// index. The sender is the one derived by the protocol at the time of inclusion.
//
// There is a fast-path for transactions retrieved by TransactionByHash and
// TransactionInBlock. Getting their sender address can be done without an RPC interaction.
func (ec *RawClient) TransactionSender(ctx context.Context, txHash common.Hash, blockHash common.Hash, txIndex uint) (common.Address, error) {
	var meta struct {
		Hash common.Hash
		From common.Address
	}
	if err := ec.CallContext(ctx, &meta, "eth_getTransactionByBlockHashAndIndex", blockHash, hexutil.Uint64(txIndex)); err != nil {
		return common.Address{}, err
	}

	if meta.Hash == (common.Hash{}) || meta.Hash != txHash {
		return common.Address{}, errors.New("wrong inclusion block/index")
	}
	return meta.From, nil
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (ec *RawClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*RpcReceipt, error) {
	var r *RpcReceipt
	err := ec.CallContext(ctx, &r, "eth_getTransactionReceipt", txHash)
	if err == nil {
		if r == nil {
			return nil, ethereum.NotFound
		}
	}
	return r, err
}

func (ec *RawClient) TransactionByHash(ctx context.Context, hash common.Hash) (tx *RpcTransaction, isPending bool, err error) {
	tx = &RpcTransaction{}
	err = ec.CallContext(ctx, tx, "eth_getTransactionByHash", hash)
	if err != nil {
		return nil, false, err
	} else if tx.Hash == common.HexToHash("0x00") {
		return nil, false, ethereum.NotFound
	}
	return tx, tx.BlockNumber == nil, nil
}

func (ec *RawClient) IsContract(ctx context.Context, address common.Address) (ok bool, err error) {
	var result hexutil.Bytes
	err = ec.CallContext(ctx, &result, "eth_getCode", address, "latest")
	if err == nil && result == nil {
		return false, nil
	}
	return len(result) > 0, err
}

func toFilterArg(q ethereum.FilterQuery) (interface{}, error) {
	arg := map[string]interface{}{
		"address": q.Addresses,
		"topics":  q.Topics,
	}

	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.FromBlock != nil || q.ToBlock != nil {
			return nil, errors.New("cannot specify both BlockHash and FromBlock/ToBlock")
		}
	} else {
		if q.FromBlock == nil {
			arg["fromBlock"] = "0x0"
		} else {
			arg["fromBlock"] = toBlockNumArg(q.FromBlock)
		}
		arg["toBlock"] = toBlockNumArg(q.ToBlock)
	}
	return arg, nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	if number.Sign() >= 0 {
		return hexutil.EncodeBig(number)
	}
	// It's negative.
	if number.IsInt64() {
		return rpc.BlockNumber(number.Int64()).String()
	}
	// It's negative and large, which is invalid.
	return fmt.Sprintf("<invalid %d>", number)
}
