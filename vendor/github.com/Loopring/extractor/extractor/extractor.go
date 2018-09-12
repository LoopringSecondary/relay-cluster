/*

  Copyright 2017 Loopring Project Ltd (Loopring Foundation).

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/

package extractor

import (
	"fmt"
	"github.com/Loopring/extractor/dao"
	"github.com/Loopring/extractor/watch"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/eth/contract"
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"math/big"
	"sort"
	"sync"
	"time"
)

/**
区块链的listener, 得到order以及ring的事件，
*/

const (
	defaultEndBlockNumber  = 1000000000
	defaultForkWaitingTime = 10
)

type ExtractorService interface {
	Start()
	Stop()
	ForkProcess(block *types.Block) error
	WatchingPendingTransaction(input interface{}) error
}

type ExtractorServiceImpl struct {
	options          ExtractorOptions
	detector         *forkDetector
	processor        *AbiProcessor
	dao              dao.RdsService
	stop             chan bool
	lock             sync.RWMutex
	startBlockNumber *big.Int
	endBlockNumber   *big.Int
	iterator         *accessor.BlockIterator
	syncComplete     bool
	forkComplete     bool
}

type ExtractorOptions struct {
	StartBlockNumber   *big.Int
	EndBlockNumber     *big.Int
	ConfirmBlockNumber uint64
	ForkWaitingTime    int64
	Debug              bool
}

func NewExtractorService(options ExtractorOptions, db dao.RdsService) *ExtractorServiceImpl {
	var l ExtractorServiceImpl

	if options.ForkWaitingTime <= 0 {
		options.ForkWaitingTime = defaultForkWaitingTime
	}

	l.options = options
	l.dao = db
	l.processor = newAbiProcessor()
	l.detector = newForkDetector(db, l.options.StartBlockNumber)
	l.stop = make(chan bool, 1)
	l.setBlockNumberRange()

	return &l
}

func (l *ExtractorServiceImpl) Start() {
	log.Infof("extractor started! scanning block:%s......", l.startBlockNumber.String())
	l.syncComplete = false

	l.iterator = accessor.NewBlockIterator(l.startBlockNumber, l.endBlockNumber, true, l.options.ConfirmBlockNumber)
	go func() {
		for {
			select {
			case <-l.stop:
				return
			default:
				if err := l.ProcessBlock(); nil != err {
					log.Error(err.Error())
					time.Sleep(1 * time.Second)
				}
			}
		}
	}()
}

func (l *ExtractorServiceImpl) Stop() {
	l.stop <- true
}

// 重启(分叉)时先关停subscribeEvents，然后关
func (l *ExtractorServiceImpl) ForkProcess(currentBlock *types.Block) error {
	forkEvent, err := l.detector.Detect(currentBlock)
	if err != nil {
		l.Warning(err)
		return err
	}

	if forkEvent == nil {
		return nil
	}

	log.Debugf("extractor,detected chain fork, from :%d to %d", forkEvent.ForkBlock.Int64(), forkEvent.DetectedBlock.Int64())

	l.Stop()

	// emit event
	Produce(forkEvent)

	// reset start blockNumber
	l.startBlockNumber = new(big.Int).Add(forkEvent.ForkBlock, big.NewInt(1))

	// waiting for the eth node catch up
	time.Sleep(time.Duration(l.options.ForkWaitingTime) * time.Second)

	l.Start()

	return fmt.Errorf("extractor,detected chain fork")
}

func (l *ExtractorServiceImpl) Sync(blockNumber *big.Int) {
	var syncBlock types.Big
	if err := accessor.BlockNumber(&syncBlock); err != nil {
		l.Warning(fmt.Errorf("extractor,Sync chain block,get ethereum node current block number error:%s", err.Error()))
	}
	currentBlockNumber := new(big.Int).Add(blockNumber, big.NewInt(int64(l.options.ConfirmBlockNumber)))
	if syncBlock.BigInt().Cmp(currentBlockNumber) <= 0 {
		Produce(syncBlock)
		l.syncComplete = true
		log.Info("extractor,Sync chain block complete!")
	} else {
		log.Debugf("extractor,chain block syncing... ")
	}
}

// Warning 当发生严重错误时需关停extractor，并通知其他模块
func (l *ExtractorServiceImpl) Warning(err error) {
	l.Stop()
	log.Warnf("extractor, warning:%s", err.Error())
}

func (l *ExtractorServiceImpl) WatchingPendingTransaction(input interface{}) error {
	tx := input.(*ethtyp.Transaction)
	if err := l.ProcessPendingTransaction(tx); err != nil {
		log.Errorf("extractor, watching pending transaction error:%s", err.Error())
	}
	return nil
}

func (l *ExtractorServiceImpl) ProcessBlock() error {
	inter, err := l.iterator.Next()
	if err != nil {
		return fmt.Errorf("extractor,iterator next error:%s", err.Error())
	}

	// get current block
	block := inter.(*ethtyp.BlockWithTxAndReceipt)
	log.Infof("extractor,get block:%s->%s, transaction number:%d", block.Number.BigInt().String(), block.Hash.Hex(), len(block.Transactions))

	currentBlock := &types.Block{}
	currentBlock.BlockNumber = block.Number.BigInt()
	currentBlock.ParentHash = block.ParentHash
	currentBlock.BlockHash = block.Hash
	currentBlock.CreateTime = block.Timestamp.Int64()

	// convert and save block
	var entity dao.Block
	entity.ConvertDown(currentBlock)
	l.dao.SaveBlock(&entity)

	// sync block on chain
	if l.syncComplete == false {
		l.Sync(block.Number.BigInt())
	}

	// detect chain fork
	if err := l.ForkProcess(currentBlock); err != nil {
		return err
	}

	// emit new block
	blockEvent := &types.BlockEvent{}
	blockEvent.BlockNumber = block.Number.BigInt()
	blockEvent.BlockHash = block.Hash
	blockEvent.BlockTime = block.Timestamp.Int64()
	blockEvent.IsFinished = false
	Produce(blockEvent)

	// add custom tokens
	l.processor.AddCustomTokens()

	if len(block.Transactions) > 0 {
		for idx, transaction := range block.Transactions {
			receipt := block.Receipts[idx]
			l.debug("extractor,tx:%s", transaction.Hash)
			if err := l.ProcessMinedTransaction(&transaction, &receipt, block.Timestamp.BigInt()); err != nil {
				log.Errorf("extractor, process mined transaction error:%s", err.Error())
			}
		}
	}

	blockEvent.IsFinished = true
	Produce(blockEvent)

	watch.ReportHeartBeat()

	return nil
}

func (l *ExtractorServiceImpl) ProcessPendingTransaction(tx *ethtyp.Transaction) error {
	log.Debugf("extractor,process pending transaction:%s, input:%s", tx.Hash, tx.Input)

	blockTime := big.NewInt(time.Now().Unix())

	if l.processor.IsSupportedMethod(tx) {
		return l.ProcessMethod(tx, nil, blockTime)
	}

	return handleOtherTransaction(tx, nil, blockTime)
}

func (l *ExtractorServiceImpl) ProcessMinedTransaction(tx *ethtyp.Transaction, receipt *ethtyp.TransactionReceipt, blockTime *big.Int) error {
	l.debug("extractor,process mined transaction,tx:%s status :%s,logs:%d", tx.Hash, receipt.Status.BigInt().String(), len(receipt.Logs))

	if l.processor.HaveSupportedEvents(receipt) {
		return l.ProcessEvent(tx, receipt, blockTime)
	}

	if l.processor.IsSupportedMethod(tx) {
		return l.ProcessMethod(tx, receipt, blockTime)
	}

	return handleOtherTransaction(tx, receipt, blockTime)
}

func (l *ExtractorServiceImpl) ProcessMethod(tx *ethtyp.Transaction, receipt *ethtyp.TransactionReceipt, blockTime *big.Int) error {
	method, ok := l.processor.GetMethod(tx)
	if !ok {
		l.debug("extractor,process method,tx:%s,unsupported contract method", tx.Hash)
		return nil
	}

	gasUsed := getGasUsed(receipt)
	status := getStatus(tx, receipt)
	return method.handleMethod(tx, gasUsed, blockTime, status, method.Name)
}

func (l *ExtractorServiceImpl) ProcessEvent(tx *ethtyp.Transaction, receipt *ethtyp.TransactionReceipt, blockTime *big.Int) error {
	methodName := l.processor.GetMethodName(tx)

	// 如果是submitRing的相关事件，必须保证fill在前，transfer在后
	if contract.TxIsSubmitRing(methodName) && len(receipt.Logs) > 1 {
		sort.SliceStable(receipt.Logs, func(i, j int) bool {
			cmpEventName := contract.EVENT_RING_MINED

			evti, _ := l.processor.GetEvent(receipt.Logs[i])

			if evti.Name == cmpEventName {
				return true
			}
			return false
		})
	}

	gasUsed := getGasUsed(receipt)
	for _, evtLog := range receipt.Logs {
		event, ok := l.processor.GetEvent(evtLog)
		if !ok {
			l.debug("extractor,process event,tx:%s,unsupported contract event", tx.Hash)
			continue
		}

		if !l.processor.IsSupportedEvent(&evtLog) {
			continue
		}

		if err := event.handleEvent(tx, &evtLog, gasUsed, blockTime, methodName); err != nil {
			log.Errorf("extractor, process event, tx:%s, logIndex:%s error:%s", tx.Hash, evtLog.LogIndex.BigInt().String(), err.Error())
		}
	}

	return nil
}

func (l *ExtractorServiceImpl) setBlockNumberRange() {
	l.startBlockNumber = l.options.StartBlockNumber
	l.endBlockNumber = l.options.EndBlockNumber
	if l.endBlockNumber.Cmp(big.NewInt(0)) == 0 {
		l.endBlockNumber = big.NewInt(defaultEndBlockNumber)
	}

	// 寻找最新块
	var ret types.Block
	latestBlock, err := l.dao.FindLatestBlock()
	if err != nil {
		log.Debugf("extractor,get latest block number error:%s", err.Error())
		return
	}
	latestBlock.ConvertUp(&ret)
	l.startBlockNumber = ret.BlockNumber

	log.Debugf("extractor,configStartBlockNumber:%s latestBlockNumber:%s", l.options.StartBlockNumber.String(), l.startBlockNumber.String())
}

func (l *ExtractorServiceImpl) debug(template string, args ...interface{}) {
	if l.options.Debug {
		log.Debugf(template, args...)
	}
}
