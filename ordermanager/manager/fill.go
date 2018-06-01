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
		log.Debugf("order manager,handle order filled event tx:%s fillIndex:%d fill already exist", event.TxHash.String(), event.FillIndex)
		return nil
	}

	// get rds.Order and types.OrderState
	state := &types.OrderState{UpdatedBlock: event.BlockNumber}
	model, err := rds.GetOrderByHash(event.OrderHash)
	if err != nil {
		return err
	}
	if err := model.ConvertUp(state); err != nil {
		return err
	}

	newFillModel := &dao.FillEvent{}
	newFillModel.ConvertDown(event)
	newFillModel.Fork = false
	newFillModel.OrderType = state.RawOrder.OrderType
	newFillModel.Side = util.GetSide(util.AddressToAlias(event.TokenS.Hex()), util.AddressToAlias(event.TokenB.Hex()))
	newFillModel.Market, _ = util.WrapMarketByAddress(event.TokenB.Hex(), event.TokenS.Hex())

	if err := rds.Add(newFillModel); err != nil {
		log.Debugf("order manager,handle order filled event tx:%s fillIndex:%s orderhash:%s error:insert failed",
			event.TxHash.Hex(), event.FillIndex.String(), event.OrderHash.Hex())
		return err
	}

	// judge order status
	if state.Status == types.ORDER_CUTOFF || state.Status == types.ORDER_FINISHED || state.Status == types.ORDER_UNKNOWN {
		log.Debugf("order manager,handle order filled event tx:%s fillIndex:%s orderhash:%s status:%d invalid",
			event.TxHash.Hex(), event.FillIndex.String(), event.OrderHash.Hex(), state.Status)
		return nil
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
		log.Errorf(err.Error())
		return err
	}
	if err := rds.UpdateOrderWhileFill(state.RawOrder.Hash, state.Status, state.DealtAmountS, state.DealtAmountB, state.SplitAmountS, state.SplitAmountB, state.UpdatedBlock); err != nil {
		return err
	}

	log.Debugf("order manager,handle order filled event tx:%s, fillIndex:%s, orderhash:%s, dealAmountS:%s, dealtAmountB:%s",
		event.TxHash.Hex(), event.FillIndex.String(), state.RawOrder.Hash.Hex(), state.DealtAmountS.String(), state.DealtAmountB.String())

	notify.NotifyOrderFilled(newFillModel)

	return nil
}
