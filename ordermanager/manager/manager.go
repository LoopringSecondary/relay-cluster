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
	"github.com/Loopring/relay-cluster/ordermanager/common"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	"github.com/Loopring/relay-lib/types"
)

type OrderManager interface {
	Start()
	Stop()
}

type OrderManagerImpl struct {
	options                 *common.OrderManagerOptions
	rds                     *dao.RdsService
	processor               *ForkProcessor
	cutoffCache             *common.CutoffCache
	mc                      marketcap.MarketCapProvider
	newOrderWatcher         *eventemitter.Watcher
	ringMinedWatcher        *eventemitter.Watcher
	fillOrderWatcher        *eventemitter.Watcher
	cancelOrderWatcher      *eventemitter.Watcher
	cutoffOrderWatcher      *eventemitter.Watcher
	cutoffPairWatcher       *eventemitter.Watcher
	forkWatcher             *eventemitter.Watcher
	warningWatcher          *eventemitter.Watcher
	submitRingMethodWatcher *eventemitter.Watcher
}

func NewOrderManager(
	options *common.OrderManagerOptions,
	rds *dao.RdsService,
	market marketcap.MarketCapProvider) *OrderManagerImpl {

	om := &OrderManagerImpl{}
	om.options = options
	om.rds = rds
	om.processor = NewForkProcess(om.rds, market)
	om.mc = market
	om.cutoffCache = common.NewCutoffCache(options.CutoffCacheCleanTime)

	return om
}

// Start start orderbook as a service
func (om *OrderManagerImpl) Start() {
	om.newOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleGatewayOrder}
	om.ringMinedWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleRingMined}
	om.fillOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleOrderFilled}
	om.cancelOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleOrderCancelled}
	om.cutoffOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleCutoff}
	om.cutoffPairWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleCutoffPair}
	om.forkWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleFork}
	om.warningWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleWarning}
	om.submitRingMethodWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleSubmitRingMethod}

	eventemitter.On(eventemitter.NewOrder, om.newOrderWatcher)
	eventemitter.On(eventemitter.RingMined, om.ringMinedWatcher)
	eventemitter.On(eventemitter.OrderFilled, om.fillOrderWatcher)
	eventemitter.On(eventemitter.CancelOrder, om.cancelOrderWatcher)
	eventemitter.On(eventemitter.CutoffAll, om.cutoffOrderWatcher)
	eventemitter.On(eventemitter.CutoffPair, om.cutoffPairWatcher)
	eventemitter.On(eventemitter.ChainForkDetected, om.forkWatcher)
	eventemitter.On(eventemitter.ExtractorWarning, om.warningWatcher)
	eventemitter.On(eventemitter.Miner_SubmitRing_Method, om.submitRingMethodWatcher)
}

func (om *OrderManagerImpl) Stop() {
	eventemitter.Un(eventemitter.NewOrder, om.newOrderWatcher)
	eventemitter.Un(eventemitter.RingMined, om.ringMinedWatcher)
	eventemitter.Un(eventemitter.OrderFilled, om.fillOrderWatcher)
	eventemitter.Un(eventemitter.CancelOrder, om.cancelOrderWatcher)
	eventemitter.Un(eventemitter.CutoffAll, om.cutoffOrderWatcher)
	eventemitter.Un(eventemitter.ChainForkDetected, om.forkWatcher)
	eventemitter.Un(eventemitter.ExtractorWarning, om.warningWatcher)
	eventemitter.Un(eventemitter.Miner_SubmitRing_Method, om.submitRingMethodWatcher)
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

// 所有来自gateway的订单都是新订单
func (om *OrderManagerImpl) handleGatewayOrder(input eventemitter.EventData) error {
	handler := &GatewayOrderHandler{
		State:     input.(*types.OrderState),
		Rds:       om.rds,
		MarketCap: om.mc,
	}

	return working(handler)
}

func (om *OrderManagerImpl) handleSubmitRingMethod(input eventemitter.EventData) error {
	handler := &SubmitRingHandler{
		Event: input.(*types.SubmitRingMethodEvent),
		Rds:   om.rds,
	}

	return working(handler)
}

func (om *OrderManagerImpl) handleRingMined(input eventemitter.EventData) error {
	handler := &RingMinedHandler{
		Event: input.(*types.RingMinedEvent),
		Rds:   om.rds,
	}

	return working(handler)
}

func (om *OrderManagerImpl) handleOrderFilled(input eventemitter.EventData) error {
	handler := &FillHandler{
		Event:     input.(*types.OrderFilledEvent),
		Rds:       om.rds,
		MarketCap: om.mc,
	}

	return working(handler)
}

func (om *OrderManagerImpl) handleOrderCancelled(input eventemitter.EventData) error {
	handler := &OrderCancelHandler{
		Event:     input.(*types.OrderCancelledEvent),
		Rds:       om.rds,
		MarketCap: om.mc,
	}

	return working(handler)
}

// 所有cutoff event都应该存起来,但不是所有event都会影响订单
func (om *OrderManagerImpl) handleCutoff(input eventemitter.EventData) error {
	handler := &CutoffHandler{
		Event:       input.(*types.CutoffEvent),
		Rds:         om.rds,
		CutoffCache: om.cutoffCache,
	}

	return working(handler)
}

func (om *OrderManagerImpl) handleCutoffPair(input eventemitter.EventData) error {
	handler := &CutoffPairHandler{
		Event:       input.(*types.CutoffPairEvent),
		Rds:         om.rds,
		CutoffCache: om.cutoffCache,
	}

	return working(handler)
}

func working(handler EventStatusHandler) error {
	handler.HandlePending()
	handler.HandleFailed()
	handler.HandleSuccess()

	return nil
}
