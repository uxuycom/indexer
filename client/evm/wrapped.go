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
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/uxuycom/indexer/client/xycommon"
	"math/big"
	"time"
)

// EClient defines typed wrappers for the Ethereum RPC API.
type EClient struct {
	rawClient *RawClient
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*EClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client, err := DialContext(ctx, rawurl)

	if err != nil {
		return nil, err
	}
	return &EClient{rawClient: client}, nil
}

// DialContext connects a client to the given URL with context.
func DialContext(ctx context.Context, rawurl string) (*RawClient, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// Close closes the underlying RPC connection.
func (ec *EClient) Close() {
	ec.rawClient.Close()
}

// EClient gets the underlying RPC client.
func (ec *EClient) Client() *rpc.Client {
	return ec.rawClient.Client()
}

// Blockchain Access

// ChainID retrieves the current chain ID for transaction replay protection.
func (ec *EClient) ChainID(ctx context.Context) (*big.Int, error) {
	return ec.rawClient.ChainID(ctx)
}

// BlockNumber returns the most recent block number
func (ec *EClient) BlockNumber(ctx context.Context) (uint64, error) {
	return ec.rawClient.BlockNumber(ctx)
}

func (ec *EClient) convertHeader(head *RpcHeader, err error) (*xycommon.RpcHeader, error) {
	if err != nil {
		if errors.Is(err, rpc.ErrNoResult) {
			return nil, xycommon.ErrNotFound
		}
		return nil, err
	}

	header := &xycommon.RpcHeader{
		ParentHash: head.ParentHash.String(),
		Number:     head.Number.ToInt(),
		Time:       uint64(head.Time),
		TxHash:     head.TxHash.String(),
	}
	return header, nil
}

func (ec *EClient) convertBlock(block *RpcBlock, err error) (*xycommon.RpcBlock, error) {
	if err != nil {
		if errors.Is(err, rpc.ErrNoResult) {
			return nil, xycommon.ErrNotFound
		}
		return nil, err
	}

	cBlock := &xycommon.RpcBlock{
		ParentHash:   block.ParentHash.String(),
		Coinbase:     block.Coinbase.String(),
		Bloom:        block.Bloom,
		Number:       block.Number.ToInt(),
		GasLimit:     block.GasLimit.ToInt(),
		GasUsed:      block.GasUsed.ToInt(),
		Time:         uint64(block.Time),
		TxHash:       block.TxHash.String(),
		Hash:         block.Hash.String(),
		Transactions: make([]*xycommon.RpcTransaction, 0, len(block.Transactions)),
	}

	for _, tx := range block.Transactions {
		cBlock.Transactions = append(cBlock.Transactions, ec.convertTransaction(tx))
	}
	return cBlock, nil
}

func (ec *EClient) convertTransaction(tx *RpcTransaction) *xycommon.RpcTransaction {
	fromAddr := ""
	if tx.From != nil {
		fromAddr = tx.From.String()
	}

	toAddr := ""
	if tx.To != nil {
		toAddr = tx.To.String()
	}
	return &xycommon.RpcTransaction{
		BlockHash:   tx.BlockHash.String(),
		BlockNumber: tx.BlockNumber.ToInt(),
		TxIndex:     tx.TxIndex.ToInt(),
		Type:        tx.Type.ToInt(),
		Hash:        tx.Hash.String(),
		ChainID:     tx.ChainID.ToInt(),
		From:        fromAddr,
		To:          toAddr,
		Input:       tx.Input,
		Value:       tx.Value.ToInt(),
		Gas:         big.NewInt(0).SetUint64(uint64(tx.Gas)),
		GasPrice:    tx.GasPrice.ToInt(),
	}
}

// BlockByHash returns the given full block.
//
// Note that loading full blocks requires two requests. Use HeaderByHash
// if you don't need all transactions or uncle headers.
func (ec *EClient) BlockByHash(ctx context.Context, hash string) (*xycommon.RpcBlock, error) {
	h := common.HexToHash(hash)
	return ec.convertBlock(ec.rawClient.BlockByHash(ctx, h))
}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (ec *EClient) BlockByNumber(ctx context.Context, number *big.Int) (*xycommon.RpcBlock, error) {
	return ec.convertBlock(ec.rawClient.BlockByNumber(ctx, number))
}

func (ec *EClient) HeaderByNumber(ctx context.Context, number *big.Int) (*xycommon.RpcHeader, error) {
	return ec.convertHeader(ec.rawClient.HeaderByNumber(ctx, number))
}

// FilterLogs executes a filter query.
func (ec *EClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]xycommon.RpcLog, error) {
	logs, err := ec.rawClient.FilterLogs(ctx, q)
	if err != nil {
		return nil, err
	}

	rpcLogs := make([]xycommon.RpcLog, 0, len(logs))
	for _, l := range logs {
		rpcLogs = append(rpcLogs, *ec.convertLog(&l))
	}
	return rpcLogs, nil
}

// SubscribeFilterLogs subscribes to the results of a streaming filter query.
func (ec *EClient) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	return ec.rawClient.SubscribeFilterLogs(ctx, q, ch)
}

// SubscribeNewHead subscribes to notifications about the current blockchain head
// on the given channel.
func (ec *EClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header, ct chan<- *xycommon.RpcBlock) (ethereum.Subscription, error) {
	return ec.rawClient.SubscribeNewHead(ctx, ch)
}

// TransactionSender returns the sender address of the given transaction. The transaction
// must be known to the remote node and included in the blockchain at the given block and
// index. The sender is the one derived by the protocol at the time of inclusion.
//
// There is a fast-path for transactions retrieved by TransactionByHash and
// TransactionInBlock. Getting their sender address can be done without an RPC interaction.
func (ec *EClient) TransactionSender(ctx context.Context, txHashStr, blockHashStr string, txIndex uint) (string, error) {
	txHash := common.HexToHash(txHashStr)
	blockHash := common.HexToHash(blockHashStr)
	addr, err := ec.rawClient.TransactionSender(ctx, blockHash, txHash, txIndex)
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}

func (ec *EClient) convertReceipt(r *RpcReceipt) *xycommon.RpcReceipt {
	c := &xycommon.RpcReceipt{
		Type:              r.Type.ToInt(),
		PostState:         r.PostState,
		Status:            r.Status.ToInt(),
		CumulativeGasUsed: r.CumulativeGasUsed.ToInt(),
		Logs:              make([]*xycommon.RpcLog, 0, len(r.Logs)),
		TxHash:            r.TxHash,
		ContractAddress:   r.ContractAddress,
		GasUsed:           r.GasUsed.ToInt(),
		EffectiveGasPrice: r.EffectiveGasPrice.ToInt(),
		BlockHash:         r.BlockHash,
		BlockNumber:       r.BlockNumber.ToInt(),
		TransactionIndex:  r.TransactionIndex.ToInt(),
	}

	for _, l := range r.Logs {
		c.Logs = append(c.Logs, ec.convertLog(l))
	}
	return c
}

func (ec *EClient) convertLog(l *RpcLog) *xycommon.RpcLog {
	return &xycommon.RpcLog{
		Address:     l.Address,
		Topics:      l.Topics,
		Data:        l.Data,
		BlockNumber: l.BlockNumber,
		TxHash:      l.TxHash,
		TxIndex:     l.TxIndex,
		BlockHash:   l.BlockHash,
		Index:       l.Index,
		Removed:     l.Removed,
	}
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (ec *EClient) TransactionReceipt(ctx context.Context, txHashStr string) (*xycommon.RpcReceipt, error) {
	txHash := common.HexToHash(txHashStr)
	r, err := ec.rawClient.TransactionReceipt(ctx, txHash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) || errors.Is(err, rpc.ErrNoResult) {
			return nil, xycommon.ErrNotFound
		}
		return nil, err
	}
	return ec.convertReceipt(r), nil
}
