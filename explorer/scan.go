package explorer

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"
	"math/big"
	"open-indexer/client/xycommon"
	"open-indexer/config"
	"open-indexer/dcache"
	"open-indexer/devents"
	"open-indexer/storage"
	"open-indexer/xylog"
	"os"
	"sync"
	"syscall"
	"time"
)

const (
	scanLimit uint64 = 2
)

type Explorer struct {
	config          *config.Config
	node            xycommon.IRPCClient
	db              *storage.DBClient
	fromBlock       uint64
	ctx             context.Context
	cancel          context.CancelFunc
	quit            chan os.Signal
	isStop          bool
	isPause         bool
	isStatusMu      sync.Mutex
	mu              sync.Mutex
	scanChn         chan uint64
	blocks          chan *xycommon.RpcBlock
	txResultHandler *devents.TxResultHandler
	dCache          *dcache.Manager
	dEvent          *devents.DEvent
	currentBlock    uint64
	updatedBlock    uint64
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
		fromBlock:       cfg.Server.FromBlock,
		isStop:          false,
		isPause:         false,
		isStatusMu:      sync.Mutex{},
		mu:              sync.Mutex{},
		dCache:          dCache,
		blocks:          make(chan *xycommon.RpcBlock, 1024),
		txResultHandler: txResultHandler,

		scanChn: make(chan uint64, 4),
		dEvent:  dEvent,
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
	blockNum, err := e.db.LastBlock(e.config.Chain.ChainName)
	if err != nil {
		xylog.Logger.Fatalf("load hisotry block index err:%v", err)
	}
	if blockNum.Uint64() > 0 {
		e.fromBlock = blockNum.Uint64() + 1
	}

	// update latest block number
	if err := e.syncLatestBlockNumber(); err != nil {
		xylog.Logger.Errorf("failed to obtain the current block height. chain:%s err=%s", e.config.Chain.ChainName, err)
	}
	go e.updateBlockLatestNumberTiming()

	//go e.runSave()
	e.scanChn <- e.fromBlock
	e.mu.Lock()
	e.updatedBlock = e.fromBlock
	e.mu.Unlock()

	for {
		select {
		case <-e.ctx.Done():
			return
		default:
		}
		blockNumber := <-e.scanChn

		//fmt.Println("scanChn len", len(e.scanChn))
		if e.currentBlock < 1 || e.currentBlock < e.updatedBlock {
			xylog.Logger.Infof("the updated height is greater than the current height. chain: %s currentBlock:%v updatedBlock:%v", e.config.Chain.ChainName, e.currentBlock, e.updatedBlock)
			e.scanChn <- blockNumber
			continue
		}

		// wait more blocks for safety
		if blockNumber > (e.currentBlock - e.config.Server.DelayedScanNumber) {
			time.Sleep(1 * time.Second)
			e.scanChn <- blockNumber
			continue
		}

		if e.currentBlock-blockNumber > 500 {
			endBlock := blockNumber + scanLimit
			if e.config.Server.ScanLimit > 0 {
				endBlock = blockNumber + e.config.Server.ScanLimit
			}

			// setting endBlock num
			if endBlock > e.currentBlock {
				endBlock = e.currentBlock
			}

			updateBlock, err := e.batchScan(blockNumber, endBlock)
			xylog.Logger.Infof("startBlock:%v endBlock:%v, scanLimit:%v updatedBlock:%v", blockNumber, endBlock, e.config.Server.ScanLimit, e.updatedBlock)
			if err != nil {
				e.scanChn <- updateBlock
				xylog.Logger.Errorf("batch block scanning failed. startBlock:%d offsetBlock:%d err=%s", blockNumber, endBlock, err)
			} else {
				e.mu.Lock()
				e.updatedBlock = updateBlock
				e.mu.Unlock()
				e.scanChn <- updateBlock + 1
			}
		} else {
			updateBlock, err := e.scan(blockNumber)
			xylog.Logger.Infof("block:%v, scanLimit:%v updatedBlock:%v", blockNumber, e.config.Server.ScanLimit, e.updatedBlock)
			if err != nil {
				e.scanChn <- updateBlock
				xylog.Logger.Errorf("batch block scanning failed. startBlock:%d err=%s", blockNumber, err)
			} else {
				e.mu.Lock()
				e.updatedBlock = updateBlock
				e.mu.Unlock()
				e.scanChn <- updateBlock + 1
			}
		}
	}
}

func (e *Explorer) updateBlockLatestNumberTiming() {
	defer func() {
		e.cancel()
	}()

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
	if e.currentBlock-e.updatedBlock > 100 {
		return nil
	}

	blockNumber, err := e.node.BlockNumber(e.ctx)
	if err != nil {
		return err
	}

	if blockNumber <= 0 {
		return errors.New("block number is zero")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.currentBlock = blockNumber
	xylog.Logger.Info("currentBlock:", e.currentBlock)
	if e.updatedBlock == 0 {
		e.updatedBlock = e.currentBlock
	}
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

func (e *Explorer) batchScan(startBlock, endBlock uint64) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	blockLogsChan := make(chan map[string][]xycommon.RpcLog)
	go e.scanLogs(startBlock, endBlock, blockLogsChan)

	blockMap := &sync.Map{}
	g, ctx := errgroup.WithContext(ctx)
	xylog.Logger.Info("batch scan startBlock:", startBlock, " endBlock:", endBlock)
	for i := startBlock; i <= endBlock; i++ {
		blockNum := i
		g.Go(func() error {
			block, err := e.node.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
			if err != nil {
				xylog.Logger.Errorf("[%v]call rpc BlockByNumber err:%v", blockNum, err)
				return err
			}
			blockMap.Store(blockNum, block)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		xylog.Logger.Errorf("concurrent block scanning failed. err=%s", err)
		return startBlock, err
	}

	// wait rpc logs result
	blockLogs := <-blockLogsChan

	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		blockVal, ok := blockMap.Load(blockNum)
		if !ok {
			return startBlock, errors.New("failed to obtain block information")
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
	return endBlock, nil
}

func (e *Explorer) scan(blockNum uint64) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	xylog.Logger.Info("batch scan block:", blockNum)
	block, err := e.node.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
	if err != nil {
		xylog.Logger.Errorf("[%v]call rpc BlockByNumber err:%v", blockNum, err)
		return blockNum, err
	}

	e.blocks <- block
	return blockNum, nil
}

func (e *Explorer) Stop() {
	e.cancel()
}
