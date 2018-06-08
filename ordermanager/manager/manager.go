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
	"github.com/Loopring/relay-cluster/usermanager"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/kafka"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	"github.com/Loopring/relay-lib/types"
)

type OrderManager interface {
	Start()
	Stop()
}

type OrderManagerImpl struct {
	options                    *common.OrderManagerOptions
	brokers                    []string
	rds                        *dao.RdsService
	processor                  *ForkProcessor
	cutoffCache                *common.CutoffCache
	mc                         marketcap.MarketCapProvider
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

const kafka_consume_group = "relay_cluster_order_manager"

func NewOrderManager(
	options *common.OrderManagerOptions,
	rds *dao.RdsService,
	market marketcap.MarketCapProvider,
	um usermanager.UserManager,
	brokers []string) *OrderManagerImpl {

	om := &OrderManagerImpl{}
	om.options = options
	om.brokers = brokers
	om.rds = rds
	om.processor = NewForkProcess(om.rds, market)
	om.mc = market
	om.cutoffCache = common.NewCutoffCache(options.CutoffCacheCleanTime)

	// register watchers for kafka
	// om.registryFlexCancelWatcher()

	InitializeWriter(om.rds, um)

	if cache.Invalid() {
		cache.Initialize(om.rds)
	}

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

	om.approveWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleApprove}
	om.depositWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleDeposit}
	om.withdrawalWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleWithdrawal}
	om.transferWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleTransfer}
	om.ethTransferWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleEthTransfer}
	om.unsupportedContractWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleUnsupportedContract}

	om.forkWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleFork}
	om.warningWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleWarning}
	om.submitRingMethodWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleSubmitRingMethod}

	eventemitter.On(eventemitter.NewOrder, om.newOrderWatcher)

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
	eventemitter.On(eventemitter.Miner_SubmitRing_Method, om.submitRingMethodWatcher)
}

func (om *OrderManagerImpl) Stop() {
	eventemitter.Un(eventemitter.NewOrder, om.newOrderWatcher)

	eventemitter.Un(eventemitter.RingMined, om.ringMinedWatcher)
	eventemitter.Un(eventemitter.OrderFilled, om.fillOrderWatcher)
	eventemitter.Un(eventemitter.CancelOrder, om.cancelOrderWatcher)
	eventemitter.Un(eventemitter.CutoffAll, om.cutoffOrderWatcher)

	eventemitter.On(eventemitter.Approve, om.approveWatcher)
	eventemitter.On(eventemitter.WethDeposit, om.depositWatcher)
	eventemitter.On(eventemitter.WethWithdrawal, om.withdrawalWatcher)
	eventemitter.On(eventemitter.Transfer, om.transferWatcher)
	eventemitter.On(eventemitter.EthTransfer, om.ethTransferWatcher)
	eventemitter.On(eventemitter.UnsupportedContract, om.unsupportedContractWatcher)

	eventemitter.Un(eventemitter.ChainForkDetected, om.forkWatcher)
	eventemitter.Un(eventemitter.ExtractorWarning, om.warningWatcher)
	eventemitter.Un(eventemitter.Miner_SubmitRing_Method, om.submitRingMethodWatcher)
}

func (om *OrderManagerImpl) registryFlexCancelWatcher() error {
	register := &kafka.ConsumerRegister{}
	register.Initialize(om.brokers)

	topic := kafka.Kafka_Topic_OrderManager_FlexCancelOrder
	group := kafka_consume_group

	err := register.RegisterTopicAndHandler(topic, group, types.FlexCancelOrderEvent{}, om.handleFlexOrderCancellation)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return nil
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

func (om *OrderManagerImpl) basehandler() BaseHandler {
	var base BaseHandler
	base.Rds = om.rds
	base.CutoffCache = om.cutoffCache
	base.MarketCap = om.mc

	return base
}

// 所有来自gateway的订单都是新订单
func (om *OrderManagerImpl) handleGatewayOrder(input eventemitter.EventData) error {
	handler := &GatewayOrderHandler{
		State:     input.(*types.OrderState),
		Rds:       om.rds,
		MarketCap: om.mc,
	}

	return om.orderRelatedWorking(handler)
}

func (om *OrderManagerImpl) handleSubmitRingMethod(input eventemitter.EventData) error {
	handler := &SubmitRingHandler{
		Event:       input.(*types.SubmitRingMethodEvent),
		BaseHandler: om.basehandler(),
	}

	return om.orderRelatedWorking(handler)
}

func (om *OrderManagerImpl) handleRingMined(input eventemitter.EventData) error {
	handler := &RingMinedHandler{
		Event:       input.(*types.RingMinedEvent),
		BaseHandler: om.basehandler(),
	}

	return om.orderRelatedWorking(handler)
}

func (om *OrderManagerImpl) handleOrderFilled(input eventemitter.EventData) error {
	handler := &FillHandler{
		Event:       input.(*types.OrderFilledEvent),
		BaseHandler: om.basehandler(),
	}

	return om.orderRelatedWorking(handler)
}

func (om *OrderManagerImpl) handleOrderCancelled(input eventemitter.EventData) error {
	handler := &OrderCancelHandler{
		Event:       input.(*types.OrderCancelledEvent),
		BaseHandler: om.basehandler(),
	}

	return om.orderRelatedWorking(handler)
}

func (om *OrderManagerImpl) handleFlexOrderCancellation(input interface{}) error {
	handler := &FlexCancelOrderHandler{
		Event:       input.(*types.FlexCancelOrderEvent),
		BaseHandler: om.basehandler(),
	}

	return om.orderRelatedWorking(handler)
}

func (om *OrderManagerImpl) handleCutoff(input eventemitter.EventData) error {
	handler := &CutoffHandler{
		Event:       input.(*types.CutoffEvent),
		BaseHandler: om.basehandler(),
	}

	return om.orderRelatedWorking(handler)
}

func (om *OrderManagerImpl) handleCutoffPair(input eventemitter.EventData) error {
	handler := &CutoffPairHandler{
		Event:       input.(*types.CutoffPairEvent),
		BaseHandler: om.basehandler(),
	}

	return om.orderRelatedWorking(handler)
}

func (om *OrderManagerImpl) handleApprove(input eventemitter.EventData) error {
	src := input.(*types.ApprovalEvent)
	return om.orderCorrelatedWorking(src.TxInfo)
}

func (om *OrderManagerImpl) handleDeposit(input eventemitter.EventData) error {
	src := input.(*types.WethDepositEvent)
	return om.orderCorrelatedWorking(src.TxInfo)
}

func (om *OrderManagerImpl) handleWithdrawal(input eventemitter.EventData) error {
	src := input.(*types.WethWithdrawalEvent)
	return om.orderCorrelatedWorking(src.TxInfo)
}

func (om *OrderManagerImpl) handleTransfer(input eventemitter.EventData) error {
	src := input.(*types.TransferEvent)
	return om.orderCorrelatedWorking(src.TxInfo)
}

func (om *OrderManagerImpl) handleEthTransfer(input eventemitter.EventData) error {
	src := input.(*types.EthTransferEvent)
	return om.orderCorrelatedWorking(src.TxInfo)
}

func (om *OrderManagerImpl) handleUnsupportedContract(input eventemitter.EventData) error {
	src := input.(*types.UnsupportedContractEvent)
	return om.orderCorrelatedWorking(src.TxInfo)
}

func (om *OrderManagerImpl) orderRelatedWorking(handler EventStatusHandler) error {
	if err := handler.HandlePending(); err != nil {
		log.Debugf(err.Error())
		return nil
	}
	if err := handler.HandleFailed(); err != nil {
		log.Debugf(err.Error())
		return nil
	}
	if err := handler.HandleSuccess(); err != nil {
		log.Debugf(err.Error())
		return nil
	}

	return nil
}

func (om *OrderManagerImpl) orderCorrelatedWorking(txinfo types.TxInfo) error {
	basehandler := om.basehandler()
	basehandler.TxInfo = txinfo
	handler := NewOrderTxHandler(basehandler)

	handler.HandleOrderCorrelatedTxFailed()
	handler.HandleOrderCorrelatedTxSuccess()

	return nil
}
