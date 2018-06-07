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
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"math/big"
)

type FillHandler struct {
	Event *types.OrderFilledEvent
	BaseHandler
}

func (handler *FillHandler) HandlePending() error {
	return nil
}

func (handler *FillHandler) HandleFailed() error {
	return nil
}

func (handler *FillHandler) HandleSuccess() error {
	if handler.Event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	event := handler.Event
	rds := handler.Rds
	mc := handler.MarketCap

	// save fill event
	_, err := rds.FindFillEvent(event.TxHash.Hex(), event.FillIndex.Int64())
	if err == nil {
		return fmt.Errorf(handler.format("err:fill already exist"), handler.value()...)
	}

	// get rds.Order and types.OrderState
	state := &types.OrderState{UpdatedBlock: event.BlockNumber}
	model, err := rds.GetOrderByHash(event.OrderHash)
	if err != nil {
		return fmt.Errorf(handler.format("err:%s"), handler.value(err.Error())...)
	}
	if err := model.ConvertUp(state); err != nil {
		return fmt.Errorf(handler.format("err:%s"), handler.value(err.Error())...)
	}

	newFillModel := &dao.FillEvent{}
	newFillModel.ConvertDown(event)
	newFillModel.Fork = false
	newFillModel.OrderType = state.RawOrder.OrderType
	newFillModel.Side = util.GetSide(util.AddressToAlias(event.TokenS.Hex()), util.AddressToAlias(event.TokenB.Hex()))
	newFillModel.Market, _ = util.WrapMarketByAddress(event.TokenB.Hex(), event.TokenS.Hex())

	if err := rds.Add(newFillModel); err != nil {
		return fmt.Errorf(handler.format("err:%s"), handler.value(err.Error())...)
	}

	// judge order status
	if state.Status == types.ORDER_CUTOFF || state.Status == types.ORDER_FINISHED || state.Status == types.ORDER_UNKNOWN {
		return fmt.Errorf(handler.format("err:order status:%d invalid"), handler.value(state.Status)...)
	}

	// calculate dealt amount
	state.UpdatedBlock = event.BlockNumber
	state.DealtAmountS = new(big.Int).Add(state.DealtAmountS, event.AmountS)
	state.DealtAmountB = new(big.Int).Add(state.DealtAmountB, event.AmountB)
	state.SplitAmountS = new(big.Int).Add(state.SplitAmountS, event.SplitS)
	state.SplitAmountB = new(big.Int).Add(state.SplitAmountB, event.SplitB)

	// update order status
	SettleOrderStatus(state, mc, false)

	// update rds.Order
	if err := model.ConvertDown(state); err != nil {
		return fmt.Errorf(handler.format("err:%s"), handler.value(err.Error())...)
	}
	if err := rds.UpdateOrderWhileFill(state.RawOrder.Hash, state.Status, state.DealtAmountS, state.DealtAmountB, state.SplitAmountS, state.SplitAmountB, state.UpdatedBlock); err != nil {
		return fmt.Errorf(handler.format("err:%s"), handler.value(err.Error())...)
	}

	log.Debugf(handler.format("dealAmountS:%s, dealtAmountB:%s"), handler.value(state.DealtAmountS.String(), state.DealtAmountB.String())...)

	notify.NotifyOrderFilled(newFillModel)

	return nil
}

func (handler *FillHandler) format(fields ...string) string {
	baseformat := "order manager fillHandler, tx:%s, fillIndex:%s, orderhash:%s, txstatus:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *FillHandler) value(values ...interface{}) []interface{} {
	basevalues := []interface{}{handler.Event.TxHash.Hex(), handler.Event.FillIndex.String(), handler.Event.OrderHash.Hex(), types.StatusStr(handler.Event.Status)}
	basevalues = append(basevalues, values...)
	return basevalues
}
