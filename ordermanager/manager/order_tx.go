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
	"fmt"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/ordermanager/cache"
	omtyp "github.com/Loopring/relay-cluster/ordermanager/types"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

// orderTx中同一个order 最多有三条记录 分别属于order owner&miner
// 1、当订单处于pending状态时允许用户cancel/cutoff
// 2、当订单处于cancel/cutoff时不允许miner pending
// 第一种情况,

type OrderTxHandler struct {
	Event *omtyp.OrderTx
	BaseHandler
}

func NewOrderTxHandler(basehandler BaseHandler) *OrderTxHandler {
	event := &omtyp.OrderTx{
		Owner:  basehandler.TxInfo.From,
		TxHash: basehandler.TxInfo.TxHash,
		Nonce:  basehandler.TxInfo.Nonce.Int64(),
	}
	handler := &OrderTxHandler{BaseHandler: basehandler, Event: event}
	return handler
}

func (handler *OrderTxHandler) HandleOrderRelatedTxPending() error {
	if handler.TxInfo.Status != types.TX_STATUS_PENDING {
		return nil
	}
	// 写入orderTx table
	if err := handler.addOrderPendingTx(); err != nil {
		return err
	}
	// 获取当前orderTx中跟order相关记录
	list, err := handler.getOrderPendingTx()
	if err != nil {
		return err
	}

	// 重新计算订单状态,并更新order表状态记录
	return handler.setOrderStatus(list)
}

func (handler *OrderTxHandler) HandleOrderRelatedTxFailed() error {
	if handler.TxInfo.Status != types.TX_STATUS_FAILED {
		return nil
	}
	if err := handler.validate(); err != nil {
		return err
	}
	return handler.processSingleOrder()
}

func (handler *OrderTxHandler) HandleOrderRelatedTxSuccess() error {
	if handler.TxInfo.Status != types.TX_STATUS_SUCCESS {
		return nil
	}
	if err := handler.validate(); err != nil {
		return err
	}
	return handler.processSingleOrder()
}

func (handler *OrderTxHandler) HandleOrderCorrelatedTxPending() error {
	return nil
}

func (handler *OrderTxHandler) HandleOrderCorrelatedTxFailed() error {
	if handler.TxInfo.Status != types.TX_STATUS_FAILED {
		return nil
	}
	orderHashList := cache.GetPendingOrders(handler.Event.Owner)
	if len(orderHashList) == 0 {
		return nil
	}

	for _, orderhash := range orderHashList {
		handler.fullFilled(orderhash)
		if err := handler.processSingleOrder(); err != nil {
			continue
		}
	}

	return nil
}

func (handler *OrderTxHandler) HandleOrderCorrelatedTxSuccess() error {
	if handler.TxInfo.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	orderHashList := cache.GetPendingOrders(handler.Event.Owner)
	if len(orderHashList) == 0 {
		return nil
	}

	for _, orderhash := range orderHashList {
		handler.fullFilled(orderhash)
		if err := handler.processSingleOrder(); err != nil {
			continue
		}
	}

	return nil
}

func (handler *OrderTxHandler) processSingleOrder() error {
	list, err := handler.getOrderPendingTx()
	if err != nil {
		return err
	}
	if err := handler.setOrderStatus(list); err != nil {
		return fmt.Errorf(handler.format("err:%s"), handler.value(err.Error())...)
	}
	return handler.delOrderPendingTx(list)
}

// todo
// 从数据库中获取订单status
// 根据当前的orderTx以及当前订单状态生成最终状态
// 更新order表订单最终状态
func (handler *OrderTxHandler) setOrderStatus(list []omtyp.OrderTx) error {
	return nil
}

// todo add cache
// 如果orderTx的nonce都大于当前nonce则不用管
func (handler *OrderTxHandler) getOrderPendingTx() ([]omtyp.OrderTx, error) {
	var list []omtyp.OrderTx

	event := handler.Event
	rds := handler.Rds

	if err := handler.validate(); err != nil {
		return list, err
	}

	if !cache.ExistPendingOrder(event.Owner, event.OrderHash) {
		return list, fmt.Errorf(handler.format("can not find owner:%s's pending order:%s in cache"), handler.value(event.Owner.Hex(), event.OrderHash.Hex())...)
	}
	models, err := rds.GetPendingOrderTx(event.Owner, event.OrderHash)
	if err != nil {
		return list, err
	}

	for _, model := range models {
		var tx omtyp.OrderTx
		model.ConvertUp(&tx)
		list = append(list, tx)
	}

	return list, nil
}

func (handler *OrderTxHandler) addOrderPendingTx() error {
	if err := handler.validate(); err != nil {
		return err
	}

	var (
		model = &dao.OrderPendingTransaction{}
		err   error
	)

	rds := handler.Rds
	event := handler.Event

	if model, err = rds.FindPendingOrderTx(event.TxHash, event.OrderHash); err == nil {
		return fmt.Errorf(handler.format("err:order %s already exist"), handler.value(event.OrderHash.Hex())...)
	}

	model.ConvertDown(event)
	rds.Add(model)

	if !cache.ExistPendingOrder(event.Owner, event.OrderHash) {
		cache.SetPendingOrder(event.Owner, event.OrderHash)
	}

	return nil
}

// 删除某个订单下txhash相同,以及txhash不同但是nonce<=当前nonce对应的tx
// 如果在orderTx表里的数据全被删除 则应在cache里删除order
func (handler *OrderTxHandler) delOrderPendingTx(list []omtyp.OrderTx) error {
	if err := handler.validate(); err != nil {
		return err
	}

	var validTxHashList []common.Hash

	event := handler.Event
	rds := handler.Rds

	for _, v := range list {
		if v.TxHash == event.TxHash {
			validTxHashList = append(validTxHashList, v.TxHash)
		} else if v.Nonce <= event.Nonce {
			validTxHashList = append(validTxHashList, v.TxHash)
		}
	}

	affectedRows := rds.DelPendingOrderTx(event.Owner, event.OrderHash, validTxHashList)
	if int(affectedRows) == len(list) {
		cache.DelPendingOrder(event.Owner, event.OrderHash)
	}

	return nil
}

func (handler *OrderTxHandler) fullFilled(orderhash common.Hash) {
	handler.Event.OrderHash = orderhash
}

func (handler *OrderTxHandler) validate() error {
	event := handler.Event
	if event.OrderHash == types.NilHash {
		return fmt.Errorf(handler.format("err:orderhash should not be nil"), handler.value()...)
	}
	return nil
}

func (handler *OrderTxHandler) format(fields ...string) string {
	baseformat := "order manager, orderTxHandler, tx:%s, owner:%s, txstatus:%s, nonce:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *OrderTxHandler) value(values ...interface{}) []interface{} {
	basevalues := []interface{}{handler.TxInfo.TxHash.Hex(), handler.TxInfo.From.Hex(), types.StatusStr(handler.TxInfo.Status), handler.TxInfo.Nonce.String()}
	basevalues = append(basevalues, values...)
	return basevalues
}
