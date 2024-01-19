package btc

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/config"
	"github.com/uxuycom/indexer/dcache"
	"math/big"
)

// BClient defines typed wrappers for the Ethereum RPC API.
type BClient struct {
	client  *RawClient
	convert *Convert
}

// NewClient creates a client that uses the given RPC client.
func NewClient(chainCfg *config.ChainConfig, cache *dcache.Manager) (*BClient, error) {
	btcClient, err := NewRawClient(chainCfg)
	if err != nil {
		return nil, err
	}

	convert := NewConvert(8, btcClient, chainCfg.Testnet, chainCfg.OrdRpc, cache)
	return &BClient{
		client:  btcClient,
		convert: convert,
	}, nil
}

// Close closes the underlying RPC connection.
func (c *BClient) Close() {
	c.client.Close()
}

// Blockchain Access

// ChainID retrieves the current chain ID for transaction replay protection.
func (c *BClient) ChainID(ctx context.Context) (*big.Int, error) {
	return c.client.ChainID(ctx)
}

// BlockNumber returns the most recent block number
func (c *BClient) BlockNumber(ctx context.Context) (uint64, error) {
	return c.client.BlockNumber(ctx)
}

func (c *BClient) HeaderByNumber(ctx context.Context, number *big.Int) (*xycommon.RpcHeader, error) {
	return c.convert.convertHeader(c.client.HeaderByNumber(ctx, number))
}

// BlockByHash returns the given full block.
//
// Note that loading full blocks requires two requests. Use HeaderByHash
// if you don't need all transactions or uncle headers.
func (c *BClient) BlockByHash(ctx context.Context, hash string) (*xycommon.RpcBlock, error) {
	return c.convert.convertBlock(c.client.BlockByHash(ctx, hash))
}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (c *BClient) BlockByNumber(ctx context.Context, number *big.Int) (*xycommon.RpcBlock, error) {
	return c.convert.convertBlock(c.client.BlockByNumber(ctx, number))
}

func (c *BClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]xycommon.RpcLog, error) {
	return nil, nil
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (c *BClient) TransactionReceipt(ctx context.Context, txHashStr string) (*xycommon.RpcReceipt, error) {
	return nil, nil
}
