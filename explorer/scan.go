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

package explorer

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/uxuycom/indexer/client/xycommon"
	"github.com/uxuycom/indexer/config"
	"github.com/uxuycom/indexer/dcache"
	"github.com/uxuycom/indexer/devents"
	"github.com/uxuycom/indexer/storage"
	"github.com/uxuycom/indexer/xylog"
	"golang.org/x/sync/errgroup"
	"math/big"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type Explorer struct {
	config          *config.Config
	node            xycommon.IRPCClient
	db              *storage.DBClient
	ctx             context.Context
	cancel          context.CancelFunc
	quit            chan os.Signal
	mu              sync.Mutex
	blocks          chan *xycommon.RpcBlock
	txResultHandler *devents.TxResultHandler
	dCache          *dcache.Manager
	dEvent          *devents.DEvent
	latestBlockNum  atomic.Uint64
	currentBlockNum atomic.Uint64
}

func NewExplorer(rpcClient xycommon.IRPCClient, dbc *storage.DBClient, cfg *config.Config, dCache *dcache.Manager, dEvent *devents.DEvent, quit chan os.Signal) *Explorer {
	ctx, cancel := context.WithCancel(context.Background())

	txResultHandler := devents.NewTxResultHandler(dCache)

	exp := &Explorer{
		ctx:             ctx,
		cancel:          cancel,
		quit:            quit,
		node:            rpcClient,
		db:              dbc,
		config:          cfg,
		mu:              sync.Mutex{},
		dCache:          dCache,
		blocks:          make(chan *xycommon.RpcBlock, 100),
		txResultHandler: txResultHandler,

		dEvent: dEvent,
	}
	return exp
}

func (e *Explorer) Scan() {
	defer func() {
		e.cancel()
		if err := recover(); err != nil {
			xylog.Logger.Panicf("scan error & quit, err[%v]", err)
		}
		xylog.Logger.Infof("scan quit")
		e.quit <- syscall.SIGUSR1
	}()
	xylog.Logger.Infof("start scanning...")

	// Prioritize using data retrieved from the database
	blockNum, err := e.db.QueryLastBlock(e.config.Chain.ChainName)
	if err != nil {
		xylog.Logger.Fatalf("load hisotry block index err:%v", err)
	}

	startBlock := e.config.Scan.StartBlock
	if blockNum.Uint64() > 0 {
		startBlock = blockNum.Uint64() + 1
	}

	// update latest block number
	go e.updateBlockLatestNumberTiming()

	// set start block number
	e.currentBlockNum.Store(startBlock)

	for {
		select {
		case <-e.ctx.Done():
			return
		default:
		}
		startBlock = e.currentBlockNum.Load()
		latestBlockNum := e.latestBlockNum.Load()
		if latestBlockNum < 1 {
			xylog.Logger.Infof("latest block number is zero. chain:%s", e.config.Chain.ChainName)
			<-time.After(time.Second)
			continue
		}

		// wait more blocks for safety
		if startBlock > (latestBlockNum - e.config.Scan.DelayedBlockNum) {
			xylog.Logger.Infof("current block number[%d] is too close to the latest block number[%d]. chain:%s", startBlock, latestBlockNum, e.config.Chain.ChainName)
			<-time.After(time.Second)
			continue
		}

		endBlock := startBlock
		if e.config.Scan.BlockBatchWorkers > 0 {
			endBlock = startBlock + e.config.Scan.BlockBatchWorkers - 1
		}

		if endBlock > latestBlockNum {
			endBlock = latestBlockNum
		}

		err = e.batchScan(startBlock, endBlock)
		if err != nil {
			xylog.Logger.Errorf("batch block scanning failed. blocks[%d-%d] err=%s", startBlock, endBlock, err)
			continue
		}

		// update current block number
		e.currentBlockNum.Store(endBlock + 1)
	}
}

func (e *Explorer) updateBlockLatestNumberTiming() {
	defer func() {
		e.cancel()
	}()

	_ = e.syncLatestBlockNumber()

	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if err := e.syncLatestBlockNumber(); err != nil {
				xylog.Logger.Errorf("failed to obtain the current block height. chain:%s err=%s", e.config.Chain.ChainName, err)
			}
		case <-e.ctx.Done():
			return
		}
	}
}

func (e *Explorer) syncLatestBlockNumber() error {
	// Add latency updating strategy for history data sync
	if e.latestBlockNum.Load()-e.currentBlockNum.Load() > 100 {
		return nil
	}

	num, err := e.node.BlockNumber(e.ctx)
	if err != nil {
		return err
	}

	if num <= 0 {
		return errors.New("block number is zero")
	}

	e.latestBlockNum.Store(num)
	xylog.Logger.Info("latestBlockNum:", num)
	return nil
}

func (e *Explorer) scanLogs(startBlock, endBlock uint64, result chan map[string][]xycommon.RpcLog) {
	if e.config.Filters == nil || len(e.config.Filters.EventTopics) <= 0 {
		result <- nil
		return
	}

	// filter Logs
	topics := [][]common.Hash{{}}
	topics[0] = make([]common.Hash, 0, len(e.config.Filters.EventTopics))
	for _, ts := range e.config.Filters.EventTopics {
		topics[0] = append(topics[0], common.HexToHash(ts))
	}

	retry := 0
DoFilter:
	query := ethereum.FilterQuery{
		Topics:    topics,
		FromBlock: new(big.Int).SetUint64(startBlock),
		ToBlock:   new(big.Int).SetUint64(endBlock),
	}
	logs, err := e.node.FilterLogs(e.ctx, query)
	if err != nil {
		xylog.Logger.Errorf("rpc FilterLogs call err:%v, retry[%d]", err, retry)
		retry++
		if retry > 10 {
			result <- nil
			return
		}
		goto DoFilter
	}

	groupLogs := make(map[string][]xycommon.RpcLog, 200)
	for _, log := range logs {
		txIdx := log.TxHash.String()
		if _, ok := groupLogs[txIdx]; !ok {
			groupLogs[txIdx] = make([]xycommon.RpcLog, 0, 2)
		}
		groupLogs[txIdx] = append(groupLogs[txIdx], log)
	}
	result <- groupLogs
}

func (e *Explorer) batchScan(startBlock, endBlock uint64) error {
	startTs := time.Now()
	defer func() {
		xylog.Logger.Infof("batchScan blocks cost[%v], blocks[%d-%d], num[%d], delayed[%d]", time.Since(startTs), startBlock, endBlock, endBlock-startBlock+1, e.latestBlockNum.Load()-e.currentBlockNum.Load())
	}()

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	blockLogsChan := make(chan map[string][]xycommon.RpcLog)
	go e.scanLogs(startBlock, endBlock, blockLogsChan)

	blockMap := &sync.Map{}
	g, ctx := errgroup.WithContext(ctx)
	for i := startBlock; i <= endBlock; i++ {
		blockNum := i
		g.Go(func() error {
			block, err := e.node.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
			if err != nil {
				xylog.Logger.Errorf("scan call rpc BlockByNumber[%d], err=%s", blockNum, err)
				return err
			}
			blockMap.Store(blockNum, block)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("concurrent block scanning failed. err=%s", err)
	}

	// wait rpc logs result
	blockLogs := <-blockLogsChan

	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		blockVal, ok := blockMap.Load(blockNum)
		if !ok {
			return fmt.Errorf("failed to obtain block[%d] data", blockNum)
		}

		block := blockVal.(*xycommon.RpcBlock)
		// add logs data
		for _, tx := range block.Transactions {
			if logs, ok1 := blockLogs[tx.Hash]; ok1 {
				tx.Events = logs
			}
		}
		e.blocks <- block
	}
	return nil
}

func (e *Explorer) Stop() {
	e.cancel()
}
