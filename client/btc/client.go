package btc

import (
	"context"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btclog"
	"github.com/uxuycom/indexer/config"
	"math/big"
	"net/url"
	"os"
)

// RawClient defines typed wrappers for the Ethereum RPC API.
type RawClient struct {
	client      *rpcclient.Client
	batchClient *rpcclient.Client
	chainID     int
}

// NewRawClient creates a client that uses the given RPC client.
func NewRawClient(chainCfg *config.ChainConfig) (*RawClient, error) {
	ul, err := url.Parse(chainCfg.ChainRPC)
	if err != nil {
		return nil, fmt.Errorf("invalid rpc[%s] url error[%v]", chainCfg.ChainRPC, err)
	}

	connConfig := &rpcclient.ConnConfig{
		Host:         ul.Host + ul.Path,
		User:         chainCfg.UserName,
		Pass:         chainCfg.PassWord,
		HTTPPostMode: ul.Scheme == "http" || ul.Scheme == "https",
		DisableTLS:   ul.Scheme != "https" && ul.Scheme != "wss",
	}

	backendLogger := btclog.NewBackend(os.Stdout).Logger("MAIN")
	backendLogger.SetLevel(btclog.LevelTrace)

	rpcclient.UseLogger(backendLogger)
	client, err := rpcclient.New(connConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("rpc[%s] client init error[%v]", chainCfg.ChainRPC, err)
	}

	batchClient, err := rpcclient.NewBatch(connConfig)
	if err != nil {
		return nil, fmt.Errorf("rpc[%s] client init error[%v]", chainCfg.ChainRPC, err)
	}

	return &RawClient{
		chainID:     chainCfg.ChainId,
		client:      client,
		batchClient: batchClient,
	}, nil
}

// Close closes the underlying RPC connection.
func (c *RawClient) Close() {
	c.client.Shutdown()
}

// Blockchain Access

// ChainID retrieves the current chain ID for transaction replay protection.
func (c *RawClient) ChainID(ctx context.Context) (*big.Int, error) {
	return big.NewInt(int64(c.chainID)), nil
}

// BlockNumber returns the most recent block number
func (c *RawClient) BlockNumber(ctx context.Context) (uint64, error) {
	count, err := c.client.GetBlockCount()
	if err != nil {
		return 0, fmt.Errorf("get status data error[%v]", err)
	}
	return uint64(count), nil
}

func (c *RawClient) HeaderByNumber(ctx context.Context, number *big.Int) (*btcjson.GetBlockVerboseResult, error) {
	if number == nil {
		if num, err := c.BlockNumber(ctx); err != nil {
			return nil, err
		} else {
			number = big.NewInt(0).SetUint64(num)
		}
	}
	blockHash, err := c.client.GetBlockHash(number.Int64())
	if err != nil {
		return nil, fmt.Errorf("get block hash err[%v]", err)
	}
	return c.GetBlockVerbose(ctx, blockHash.String())
}

// BlockByHash returns the given full block.
//
// Note that loading full blocks requires two requests. Use HeaderByHash
// if you don't need all transactions or uncle headers.
func (c *RawClient) BlockByHash(ctx context.Context, hash string) (*btcjson.GetBlockVerboseTxResult, error) {
	return c.GetBlockVerboseTx(ctx, hash)
}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (c *RawClient) BlockByNumber(ctx context.Context, number *big.Int) (*btcjson.GetBlockVerboseTxResult, error) {
	blockHash, err := c.client.GetBlockHash(number.Int64())
	if err != nil {
		return nil, fmt.Errorf("get block hash err[%v]", err)
	}
	return c.GetBlockVerboseTx(ctx, blockHash.String())
}

func (c *RawClient) GetBlock(ctx context.Context, hash string) (b *wire.MsgBlock, err error) {
	hs, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, fmt.Errorf("decode hash[%s] err[%v]", hash, err)
	}
	return c.client.GetBlock(hs)
}

func (c *RawClient) GetBlockVerbose(ctx context.Context, hash string) (b *btcjson.GetBlockVerboseResult, err error) {
	hs, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, fmt.Errorf("decode hash[%s] err[%v]", hash, err)
	}
	return c.client.GetBlockVerbose(hs)
}

func (c *RawClient) GetBlockVerboseTx(ctx context.Context, hash string) (b *btcjson.GetBlockVerboseTxResult, err error) {
	hs, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, fmt.Errorf("decode hash[%s] err[%v]", hash, err)
	}
	return c.client.GetBlockVerboseTx(hs)
}

func (c *RawClient) GetRawTransaction(ctx context.Context, hash string) (tx *btcutil.Tx, err error) {
	hs, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, fmt.Errorf("decode hash[%s] err[%v]", hash, err)
	}
	return c.client.GetRawTransaction(hs)
}

func (c *RawClient) GetRawTransactionVerbose(ctx context.Context, hash string) (tx *btcjson.TxRawResult, err error) {
	hs, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, fmt.Errorf("decode hash[%s] err[%v]", hash, err)
	}
	return c.client.GetRawTransactionVerbose(hs)
}

func (c *RawClient) GetMultiRawTransactionVerbose(ctx context.Context, reqHashes []string) (txs map[string]*btcjson.TxRawResult, err error) {
	if len(reqHashes) <= 0 {
		return nil, nil
	}

	//unique hash list
	hashesMap := make(map[string]struct{}, len(reqHashes))
	for _, h := range reqHashes {
		hashesMap[h] = struct{}{}
	}

	hashes := make([]string, 0, len(hashesMap))
	for h := range hashesMap {
		hashes = append(hashes, h)
	}

	i := 0
	batch := 100
	maxIdx := len(hashes)
	ended := false
	txs = make(map[string]*btcjson.TxRawResult, len(hashes))
	for {
		start := i * 100
		end := (i + 1) * 100
		if end >= maxIdx {
			end = maxIdx
			ended = true
		}
		i++
		sliceHashes := hashes[start:end]
		futures := make(map[string]rpcclient.FutureGetRawTransactionVerboseResult, len(sliceHashes))
		for _, hash := range sliceHashes {
			hs, err1 := chainhash.NewHashFromStr(hash)
			if err1 != nil {
				return nil, fmt.Errorf("decode hash[%s] err[%v]", hash, err1)
			}
			futures[hash] = c.batchClient.GetRawTransactionVerboseAsync(hs)
		}

		if err = c.batchClient.Send(); err != nil {
			return nil, fmt.Errorf("batch request error[%v], hashes[%v]", err, batch)
		}

		for hash, future := range futures {
			tx, err1 := future.Receive()
			if err1 != nil {
				return nil, fmt.Errorf("receive tx hash[%s] error[%v]", hash, err1)
			}
			txs[hash] = tx
		}

		if ended {
			break
		}
	}
	return txs, nil
}
