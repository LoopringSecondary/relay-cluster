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
	switcher := handler.FullSwitcher(types.NilHash, types.ORDER_CANCELLING)

	if err := switcher.FlexibleCancellationPendingProcedure(); err != nil {
		log.Errorf(err.Error())
	}

	return nil
}

func (handler *OrderCancelHandler) HandleFailed() error {
	switcher := handler.FullSwitcher(types.NilHash, types.ORDER_PENDING)

	if err := switcher.FlexibleCancellationFailedProcedure(); err != nil {
		log.Errorf(err.Error())
	}

	return nil
}

func (handler *OrderCancelHandler) HandleSuccess() error {
	if handler.Event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	event := handler.Event
	rds := handler.Rds
	mc := handler.MarketCap

	// save cancel event
	_, err := rds.GetCancelEvent(event.TxHash)
	if err == nil {
		log.Debugf("order manager,handle order cancelled event tx:%s, orderhash:%s error:order have already exist", event.TxHash.Hex(), event.OrderHash.Hex())
		return nil
	}
	newCancelEventModel := &dao.CancelEvent{}
	newCancelEventModel.ConvertDown(event)
	newCancelEventModel.Fork = false
	if err := rds.Add(newCancelEventModel); err != nil {
		return err
	}

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
		log.Debugf("order manager,handle order cancelled event tx:%s, order:%s cancelled amountb:%s", event.TxHash.Hex(), state.RawOrder.Hash.Hex(), state.CancelledAmountB.String())
	} else {
		state.CancelledAmountS = new(big.Int).Add(state.CancelledAmountS, event.AmountCancelled)
		log.Debugf("order manager,handle order cancelled event tx:%s, order:%s cancelled amounts:%s", event.TxHash.Hex(), state.RawOrder.Hash.Hex(), state.CancelledAmountS.String())
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

	return nil
}
