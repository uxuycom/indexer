package evm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type RpcHeader struct {
	ParentHash common.Hash    `json:"parentHash"       gencodec:"required"`
	Coinbase   common.Address `json:"miner"`
	Bloom      types.Bloom    `json:"logsBloom"        gencodec:"required"`
	Number     *hexutil.Big   `json:"number"           gencodec:"required"`
	GasLimit   *hexutil.Big   `json:"gasLimit"         gencodec:"required"`
	GasUsed    *hexutil.Big   `json:"gasUsed"          gencodec:"required"`
	Time       hexutil.Uint64 `json:"timestamp"`
	TxHash     common.Hash    `json:"transactionsRoot" gencodec:"required"`
	Hash       common.Hash    `json:"hash"`
}

type RpcBlock struct {
	ParentHash common.Hash    `json:"parentHash"       gencodec:"required"`
	Coinbase   common.Address `json:"miner"`
	Bloom      types.Bloom    `json:"logsBloom"        gencodec:"required"`
	Number     *hexutil.Big   `json:"number"           gencodec:"required"`
	GasLimit   *hexutil.Big   `json:"gasLimit"         gencodec:"required"`
	GasUsed    *hexutil.Big   `json:"gasUsed"          gencodec:"required"`
	Time       hexutil.Uint64 `json:"timestamp"`

	TxHash       common.Hash       `json:"transactionsRoot" gencodec:"required"`
	Hash         common.Hash       `json:"hash"`
	Transactions []*RpcTransaction `json:"transactions"`
	TraceId      string            `json:"-"`
}

type RpcTransaction struct {
	BlockHash   common.Hash     `json:"blockHash"`
	BlockNumber *hexutil.Big    `json:"blockNumber"`
	TxIndex     *hexutil.Big    `json:"transactionIndex"`
	Type        *hexutil.Big    `json:"type"`
	Hash        common.Hash     `json:"hash"`
	ChainID     *hexutil.Big    `json:"chainId,omitempty"`
	From        *common.Address `json:"from"`
	To          *common.Address `json:"to"`
	Input       string          `json:"input"`
	Value       *hexutil.Big    `json:"value"`

	Gas                  hexutil.Uint64 `json:"gas"`
	GasPrice             *hexutil.Big   `json:"gasPrice"`
	MaxPriorityFeePerGas *hexutil.Big   `json:"maxPriorityFeePerGas"`
	MaxFeePerGas         *hexutil.Big   `json:"maxFeePerGas"`
	MaxFeePerDataGas     *hexutil.Big   `json:"maxFeePerDataGas,omitempty"`
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
	Type              *hexutil.Big  `json:"type,omitempty"`
	PostState         hexutil.Bytes `json:"root"`
	Status            *hexutil.Big  `json:"status"`
	CumulativeGasUsed *hexutil.Big  `json:"cumulativeGasUsed" gencodec:"required"`
	Logs              []*RpcLog     `json:"logs"              gencodec:"required"`

	// Implementation fields: These fields are added by geth when processing a transaction.
	TxHash            common.Hash     `json:"transactionHash" gencodec:"required"`
	ContractAddress   *common.Address `json:"contractAddress"`
	GasUsed           *hexutil.Big    `json:"gasUsed" gencodec:"required"`
	EffectiveGasPrice *hexutil.Big    `json:"effectiveGasPrice"` // required, but tag omitted for backwards compatibility

	// Inclusion information: These fields provide information about the inclusion of the
	// transaction corresponding to this receipt.
	BlockHash        common.Hash  `json:"blockHash,omitempty"`
	BlockNumber      *hexutil.Big `json:"blockNumber,omitempty"`
	TransactionIndex *hexutil.Big `json:"transactionIndex"`
}
