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
	notify "github.com/Loopring/relay-cluster/util"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"math/big"
)

type OrderCancelHandler struct {
	Event *types.OrderCancelledEvent
	BaseHandler
}

func (handler *OrderCancelHandler) HandlePending() error {
	if handler.Event.Status != types.TX_STATUS_PENDING {
		return nil
	}
	if err := handler.saveEvent(); err != nil {
		return err
	}

	log.Debugf(handler.format(), handler.value()...)

	txhandler := FullOrderTxHandler(handler.BaseHandler, handler.Event.OrderHash, types.ORDER_CANCELLING)
	return txhandler.HandlerOrderRelatedTx()
}

func (handler *OrderCancelHandler) HandleFailed() error {
	if handler.Event.Status != types.TX_STATUS_FAILED {
		return nil
	}
	if err := handler.saveEvent(); err != nil {
		return err
	}

	log.Debugf(handler.format(), handler.value()...)

	txhandler := FullOrderTxHandler(handler.BaseHandler, handler.Event.OrderHash, types.ORDER_CANCELLING)
	return txhandler.HandlerOrderRelatedTx()
}

func (handler *OrderCancelHandler) HandleSuccess() error {
	if handler.Event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	event := handler.Event
	rds := handler.Rds
	mc := handler.MarketCap

	// save cancel event
	if err := handler.saveEvent(); err != nil {
		return err
	}

	log.Debugf(handler.format(), handler.value()...)

	// get rds.Order and types.OrderState
	state := &types.OrderState{}
	model, err := rds.GetOrderByHash(event.OrderHash)
	if err != nil {
		return err
	}
	if err := model.ConvertUp(state); err != nil {
		return err
	}

	// calculate remainAmount and cancelled amount should be saved whether order is finished or not
	if state.RawOrder.BuyNoMoreThanAmountB {
		state.CancelledAmountB = new(big.Int).Add(state.CancelledAmountB, event.AmountCancelled)
		log.Debugf(handler.format("cancelledAmountB"), handler.value(state.CancelledAmountB.String())...)
	} else {
		state.CancelledAmountS = new(big.Int).Add(state.CancelledAmountS, event.AmountCancelled)
		log.Debugf(handler.format("cancelledAmountS"), handler.value(state.CancelledAmountS.String())...)
	}

	// update order status
	SettleOrderStatus(state, mc, true)
	state.UpdatedBlock = event.BlockNumber

	// update rds.Order
	if err := model.ConvertDown(state); err != nil {
		return err
	}
	if err := rds.UpdateOrderWhileCancel(state.RawOrder.Hash, state.Status, state.CancelledAmountS, state.CancelledAmountB, state.UpdatedBlock); err != nil {
		return err
	}

	notify.NotifyOrderUpdate(state)

	// 原则上不允许订单状态干扰到其他动作
	txhandler := FullOrderTxHandler(handler.BaseHandler, handler.Event.OrderHash, types.ORDER_CANCELLING)
	return txhandler.HandlerOrderRelatedTx()
}

func (handler *OrderCancelHandler) saveEvent() error {
	rds := handler.Rds
	event := handler.Event

	var (
		model dao.CancelEvent
		err   error
	)

	// save cancel event
	model, err = rds.GetCancelEvent(event.TxHash)
	if err := ValidateDuplicateEvent(event.Status, model.Status, err); err != nil {
		return fmt.Errorf(handler.format("err:%s"), handler.value(err.Error())...)
	}

	model.ConvertDown(event)
	model.Fork = false

	if err != nil {
		return rds.Add(&model)
	} else {
		return rds.Save(&model)
	}
}

func (handler *OrderCancelHandler) format(fields ...string) string {
	baseformat := "order manager orderCancelHandler, tx:%s, orderhash:%s, txstatus:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *OrderCancelHandler) value(values ...interface{}) []interface{} {
	basevalues := []interface{}{handler.Event.TxHash.Hex(), handler.Event.OrderHash.Hex(), types.StatusStr(handler.Event.Status)}
	basevalues = append(basevalues, values...)
	return basevalues
}
