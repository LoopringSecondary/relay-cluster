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

package ordermanager

import (
	"fmt"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/marketcap"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"math/big"
)

func newOrderEntity(state *types.OrderState, mc marketcap.MarketCapProvider, blockNumber *big.Int) (*dao.Order, error) {
	state.DealtAmountS = big.NewInt(0)
	state.DealtAmountB = big.NewInt(0)
	state.SplitAmountS = big.NewInt(0)
	state.SplitAmountB = big.NewInt(0)
	state.CancelledAmountB = big.NewInt(0)
	state.CancelledAmountS = big.NewInt(0)
	state.RawOrder.Side = util.GetSide(state.RawOrder.TokenS.Hex(), state.RawOrder.TokenB.Hex())

	if blockNumber == nil {
		state.UpdatedBlock = big.NewInt(0)
	} else {
		state.UpdatedBlock = blockNumber
	}

	// calculate order amount and settled
	cancelAmount, dealtAmount, getAmountErr := getCancelledAndDealtAmount(state, blockNumber)
	if getAmountErr != nil {
		return nil, getAmountErr
	}
	settleOrderCancelAndFilled(state, cancelAmount, dealtAmount)

	// check order finished status
	settleOrderStatus(state, mc, false)

	// convert order
	model := &dao.Order{}
	model.ConvertDown(state)

	return model, nil
}

func settleOrderStatus(state *types.OrderState, mc marketcap.MarketCapProvider, isCancel bool) {
	zero := big.NewInt(0)
	finishAmountS := big.NewInt(0).Add(state.CancelledAmountS, state.DealtAmountS)
	totalAmountS := big.NewInt(0).Add(finishAmountS, state.SplitAmountS)
	finishAmountB := big.NewInt(0).Add(state.CancelledAmountB, state.DealtAmountB)
	totalAmountB := big.NewInt(0).Add(finishAmountB, state.SplitAmountB)
	totalAmount := big.NewInt(0).Add(totalAmountS, totalAmountB)

	if totalAmount.Cmp(zero) <= 0 {
		state.Status = types.ORDER_NEW
	} else if !mc.IsOrderValueDust(state) {
		state.Status = types.ORDER_PARTIAL
	} else if isCancel {
		state.Status = types.ORDER_CANCEL
	} else {
		state.Status = types.ORDER_FINISHED
	}
}

func settleOrderCancelAndFilled(state *types.OrderState, cancelAmount, dealtAmount *big.Int) {
	if state.RawOrder.BuyNoMoreThanAmountB {
		state.DealtAmountB = dealtAmount
		state.CancelledAmountB = cancelAmount
	} else {
		state.DealtAmountS = dealtAmount
		state.CancelledAmountS = cancelAmount
	}
}

func blockNumberToString(blockNumber *big.Int) string {
	var blockNumberStr string
	if blockNumber == nil {
		blockNumberStr = "latest"
	} else {
		blockNumberStr = types.BigintToHex(blockNumber)
	}

	return blockNumberStr
}

func getCancelledAndDealtAmount(state *types.OrderState, blockNumber *big.Int) (*big.Int, *big.Int, error) {
	// TODO(fuk): 系统暂时只会从gateway接收新订单,而不会有部分成交的订单
	return big.NewInt(0), big.NewInt(0), nil

	var (
		cancelled, cancelOrFilled, dealt *big.Int
		err                              error
	)

	blockNumberStr := blockNumberToString(blockNumber)
	protocol := state.RawOrder.DelegateAddress
	orderhash := state.RawOrder.Hash

	// get order cancelled amount from chain
	if cancelled, err = loopringaccessor.GetCancelled(protocol, orderhash, blockNumberStr); err != nil {
		return nil, nil, fmt.Errorf("order manager,handle gateway order,order %s getCancelled error:%s", orderhash.Hex(), err.Error())
	}

	// get order cancelledOrFilled amount from chain
	if cancelOrFilled, err = loopringaccessor.GetCancelledOrFilled(protocol, orderhash, blockNumberStr); err != nil {
		return nil, nil, fmt.Errorf("order manager,handle gateway order,order %s getCancelledOrFilled error:%s", orderhash.Hex(), err.Error())
	}

	if cancelOrFilled.Cmp(cancelled) < 0 {
		return nil, nil, fmt.Errorf("order manager,handle gateway order,order %s cancelOrFilledAmount:%s < cancelledAmount:%s", orderhash.Hex(), cancelOrFilled.String(), cancelled.String())
	}

	dealt = big.NewInt(0).Sub(cancelOrFilled, cancelled)

	return cancelled, dealt, nil
}
