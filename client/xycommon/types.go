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

package xycommon

import (
	"context"
	"errors"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

var ErrNotFound = errors.New("not found")

type IRPCClient interface {
	BlockNumber(ctx context.Context) (uint64, error)

	BlockByNumber(ctx context.Context, number *big.Int) (*RpcBlock, error)

	HeaderByNumber(ctx context.Context, number *big.Int) (*RpcHeader, error)

	TransactionSender(ctx context.Context, txHash, blockHash string, txIndex uint) (string, error)

	TransactionReceipt(ctx context.Context, txHash string) (*RpcReceipt, error)

	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]RpcLog, error)
}

type RpcHeader struct {
	ParentHash string   `json:"parentHash"       gencodec:"required"`
	Number     *big.Int `json:"number"           gencodec:"required"`
	Time       uint64   `json:"timestamp"`
	TxHash     string   `json:"transactionsRoot" gencodec:"required"`
}

type RpcBlock struct {
	ParentHash   string            `json:"parentHash"       gencodec:"required"`
	Coinbase     string            `json:"miner"`
	Bloom        types.Bloom       `json:"logsBloom"        gencodec:"required"`
	Number       *big.Int          `json:"number"           gencodec:"required"`
	GasLimit     *big.Int          `json:"gasLimit"         gencodec:"required"`
	GasUsed      *big.Int          `json:"gasUsed"          gencodec:"required"`
	Time         uint64            `json:"timestamp"`
	TxHash       string            `json:"transactionsRoot" gencodec:"required"`
	Hash         string            `json:"hash"`
	Transactions []*RpcTransaction `json:"transactions"`
}

type Inscription struct {
}

type RpcTransaction struct {
	BlockHash   string               `json:"blockHash"`
	BlockNumber *big.Int             `json:"blockNumber"`
	TxIndex     *big.Int             `json:"transactionIndex"`
	Type        *big.Int             `json:"type"`
	Hash        string               `json:"hash"`
	ChainID     *big.Int             `json:"chainId,omitempty"`
	From        string               `json:"from"`
	To          string               `json:"to"`
	Input       string               `json:"input"`
	Value       *big.Int             `json:"value"`
	Gas         *big.Int             `json:"gas"`
	GasPrice    *big.Int             `json:"gasPrice"`
	Vin         []btcjson.VinPrevOut `json:"vin"`
	Vout        []btcjson.Vout       `json:"vout"`
	Events      []RpcLog             `json:"events"`
	Status      int64                `json:"status"`
}

type RpcLog struct {
	// Consensus fields:
	// address of the contract that generated the event
	Address common.Address `json:"address" gencodec:"required"`
	// list of topics provided by the contract.
	Topics []common.Hash `json:"topics" gencodec:"required"`
	// supplied by the contract, usually ABI-encoded
	Data hexutil.Bytes `json:"data" gencodec:"required"`

	// Derived fields. These fields are filled in by the node
	// but not secured by consensus.
	// block in which the transaction was included
	BlockNumber *hexutil.Big `json:"blockNumber"`
	// hash of the transaction
	TxHash common.Hash `json:"transactionHash" gencodec:"required"`
	// index of the transaction in the block
	TxIndex *hexutil.Big `json:"transactionIndex"`
	// hash of the block in which the transaction was included
	BlockHash common.Hash `json:"blockHash"`
	// index of the xylog in the block
	Index *hexutil.Big `json:"logIndex"`

	// The Removed field is true if this xylog was reverted due to a chain reorganisation.
	// You must pay attention to this field if you receive logs through a filter query.
	Removed bool `json:"removed"`
}

// RpcReceipt represents the results of a transaction.
type RpcReceipt struct {
	// Consensus fields: These fields are defined by the Yellow Paper
	Type              *big.Int      `json:"type,omitempty"`
	PostState         hexutil.Bytes `json:"root"`
	Status            *big.Int      `json:"status"`
	CumulativeGasUsed *big.Int      `json:"cumulativeGasUsed" gencodec:"required"`
	Logs              []*RpcLog     `json:"logs"              gencodec:"required"`

	// Implementation fields: These fields are added by geth when processing a transaction.
	TxHash            common.Hash     `json:"transactionHash" gencodec:"required"`
	ContractAddress   *common.Address `json:"contractAddress"`
	GasUsed           *big.Int        `json:"gasUsed" gencodec:"required"`
	EffectiveGasPrice *big.Int        `json:"effectiveGasPrice"` // required, but tag omitted for backwards compatibility

	// Inclusion information: These fields provide information about the inclusion of the
	// transaction corresponding to this receipt.
	BlockHash        common.Hash `json:"blockHash,omitempty"`
	BlockNumber      *big.Int    `json:"blockNumber,omitempty"`
	TransactionIndex *big.Int    `json:"transactionIndex"`
}
