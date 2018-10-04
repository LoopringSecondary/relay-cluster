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

package manager

import (
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/ordermanager/cache"
	"github.com/Loopring/relay-cluster/ordermanager/common"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	"github.com/Loopring/relay-lib/types"
	"time"
)

type OrderManager interface {
	Start()
	Stop()
}

type OrderManagerImpl struct {
	options                    *common.OrderManagerOptions
	brokers                    []string
	processor                  *ForkProcessor
	newOrderWatcher            *eventemitter.Watcher
	ringMinedWatcher           *eventemitter.Watcher
	fillOrderWatcher           *eventemitter.Watcher
	cancelOrderWatcher         *eventemitter.Watcher
	cutoffOrderWatcher         *eventemitter.Watcher
	cutoffPairWatcher          *eventemitter.Watcher
	approveWatcher             *eventemitter.Watcher
	depositWatcher             *eventemitter.Watcher
	withdrawalWatcher          *eventemitter.Watcher
	transferWatcher            *eventemitter.Watcher
	ethTransferWatcher         *eventemitter.Watcher
	unsupportedContractWatcher *eventemitter.Watcher
	forkWatcher                *eventemitter.Watcher
	warningWatcher             *eventemitter.Watcher
	submitRingMethodWatcher    *eventemitter.Watcher
}

var (
	rds               *dao.RdsService
	marketCapProvider marketcap.MarketCapProvider
	cutoffcache       *common.CutoffCache
)

func NewOrderManager(
	options *common.OrderManagerOptions,
	db *dao.RdsService,
	market marketcap.MarketCapProvider,
	brokers []string) *OrderManagerImpl {

	om := &OrderManagerImpl{}
	om.options = options
	om.brokers = brokers
	om.processor = NewForkProcess()
	cutoffcache = common.NewCutoffCache(options.CutoffCacheCleanTime)

	marketCapProvider = market
	rds = db

	if cache.Invalid() {
		cache.Initialize(rds)
	}

	return om
}

// Start start orderbook as a service
func (om *OrderManagerImpl) Start() {
	// order related
	om.newOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandlerOrderRelatedEvent}
	om.submitRingMethodWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandlerOrderRelatedEvent}
	om.ringMinedWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandlerOrderRelatedEvent}
	om.fillOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandlerOrderRelatedEvent}
	om.cancelOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandlerOrderRelatedEvent}
	om.cutoffOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandlerOrderRelatedEvent}
	om.cutoffPairWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandlerOrderRelatedEvent}

	// order correlated
	om.approveWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandleOrderCorrelatedEvent}
	om.depositWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandleOrderCorrelatedEvent}
	om.withdrawalWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandleOrderCorrelatedEvent}
	om.transferWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandleOrderCorrelatedEvent}
	om.ethTransferWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandleOrderCorrelatedEvent}
	om.unsupportedContractWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.HandleOrderCorrelatedEvent}

	// procedure related
	om.forkWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleFork}
	om.warningWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleWarning}

	eventemitter.On(eventemitter.NewOrder, om.newOrderWatcher)
	eventemitter.On(eventemitter.Miner_SubmitRing_Method, om.submitRingMethodWatcher)
	eventemitter.On(eventemitter.RingMined, om.ringMinedWatcher)
	eventemitter.On(eventemitter.OrderFilled, om.fillOrderWatcher)
	eventemitter.On(eventemitter.CancelOrder, om.cancelOrderWatcher)
	eventemitter.On(eventemitter.CutoffAll, om.cutoffOrderWatcher)
	eventemitter.On(eventemitter.CutoffPair, om.cutoffPairWatcher)

	eventemitter.On(eventemitter.Approve, om.approveWatcher)
	eventemitter.On(eventemitter.WethDeposit, om.depositWatcher)
	eventemitter.On(eventemitter.WethWithdrawal, om.withdrawalWatcher)
	eventemitter.On(eventemitter.Transfer, om.transferWatcher)
	eventemitter.On(eventemitter.EthTransfer, om.ethTransferWatcher)
	eventemitter.On(eventemitter.UnsupportedContract, om.unsupportedContractWatcher)

	eventemitter.On(eventemitter.ChainForkDetected, om.forkWatcher)
	eventemitter.On(eventemitter.ExtractorWarning, om.warningWatcher)
}

func (om *OrderManagerImpl) Stop() {
	eventemitter.Un(eventemitter.NewOrder, om.newOrderWatcher)
	eventemitter.Un(eventemitter.Miner_SubmitRing_Method, om.submitRingMethodWatcher)
	eventemitter.Un(eventemitter.RingMined, om.ringMinedWatcher)
	eventemitter.Un(eventemitter.OrderFilled, om.fillOrderWatcher)
	eventemitter.Un(eventemitter.CancelOrder, om.cancelOrderWatcher)
	eventemitter.Un(eventemitter.CutoffAll, om.cutoffOrderWatcher)
	eventemitter.Un(eventemitter.CutoffPair, om.cutoffPairWatcher)

	eventemitter.Un(eventemitter.Approve, om.approveWatcher)
	eventemitter.Un(eventemitter.WethDeposit, om.depositWatcher)
	eventemitter.Un(eventemitter.WethWithdrawal, om.withdrawalWatcher)
	eventemitter.Un(eventemitter.Transfer, om.transferWatcher)
	eventemitter.Un(eventemitter.EthTransfer, om.ethTransferWatcher)
	eventemitter.Un(eventemitter.UnsupportedContract, om.unsupportedContractWatcher)

	eventemitter.Un(eventemitter.ChainForkDetected, om.forkWatcher)
	eventemitter.Un(eventemitter.ExtractorWarning, om.warningWatcher)
}

func (om *OrderManagerImpl) handleFork(input eventemitter.EventData) error {
	log.Debugf("order manager processing chain fork......")

	om.Stop()
	if err := om.processor.Fork(input.(*types.ForkedEvent)); err != nil {
		log.Fatalf("order manager,handle fork error:%s", err.Error())
	}
	om.Start()

	return nil
}

func (om *OrderManagerImpl) handleWarning(input eventemitter.EventData) error {
	log.Debugf("order manager processing extractor warning")
	om.Stop()
	return nil
}

func (om *OrderManagerImpl) HandlerOrderRelatedEvent(input eventemitter.EventData) error {
	var err error

	switch event := input.(type) {
	case *types.OrderState:
		err = HandleGatewayOrder(event)
	case *types.SubmitRingMethodEvent:
		err = HandleSubmitRingMethodEvent(event)
		err = HandleP2PSubmitRing(event)
	case *types.RingMinedEvent:
		err = HandleRingMinedEvent(event)
	case *types.OrderFilledEvent:
		err = HandleOrderFilledEvent(event)
		err = HandleP2POrderFilled(event)
	case *types.OrderCancelledEvent:
		err = HandleOrderCancelledEvent(event)
	case *types.CutoffEvent:
		err = HandleCutoffEvent(event)
	case *types.CutoffPairEvent:
		err = HandleCutoffPair(event)
	default:
		return nil
	}

	if err != nil {
		log.Errorf(err.Error())
	}

	return nil
}

func (om *OrderManagerImpl) HandleOrderCorrelatedEvent(input eventemitter.EventData) error {
	var txinfo types.TxInfo

	start := time.Now().UnixNano()
	switch event := input.(type) {
	case *types.ApprovalEvent:
		txinfo = event.TxInfo
	case *types.WethDepositEvent:
		txinfo = event.TxInfo
	case *types.WethWithdrawalEvent:
		txinfo = event.TxInfo
	case *types.TransferEvent:
		txinfo = event.TxInfo
	case *types.EthTransferEvent:
		txinfo = event.TxInfo
	case *types.UnsupportedContractEvent:
		txinfo = event.TxInfo
	default:
		return nil
	}

	handler := BaseOrderTxHandler(txinfo)
	if err := handler.HandlerOrderCorrelatedTx(); err != nil {
		log.Errorf(err.Error())
	}

	stop := time.Now().UnixNano()
	execmsec := (stop - start) / 1e6
	if execmsec > 0 {
		log.Debugf("order manager handle order correlated event time:%d(msec)", execmsec)
	}
	return nil
}
