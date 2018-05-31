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
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	"github.com/Loopring/relay-lib/types"
	"math/big"
)

type GatewayOrderHandler struct {
	State     *types.OrderState
	Rds       *dao.RdsService
	MarketCap marketcap.MarketCapProvider
}

func (handler *GatewayOrderHandler) HandleFailed() error {
	return nil
}

func (handler *GatewayOrderHandler) HandlePending() error {
	return nil
}

func (handler *GatewayOrderHandler) HandleSuccess() error {
	state := handler.State
	rds := handler.Rds
	mc := handler.MarketCap

	model, err := NewOrderEntity(state, mc, nil)
	if err != nil {
		log.Errorf("order manager,handle gateway order:%s error", state.RawOrder.Hash.Hex())
		return err
	}

	if err = rds.Add(model); err != nil {
		return err
	}

	log.Debugf("order manager,handle gateway order,order.hash:%s amountS:%s", state.RawOrder.Hash.Hex(), state.RawOrder.AmountS.String())

	notify.NotifyOrderUpdate(state)

	return nil
}

func NewOrderEntity(state *types.OrderState, mc marketcap.MarketCapProvider, blockNumber *big.Int) (*dao.Order, error) {
	state.DealtAmountS = big.NewInt(0)
	state.DealtAmountB = big.NewInt(0)
	state.SplitAmountS = big.NewInt(0)
	state.SplitAmountB = big.NewInt(0)
	state.CancelledAmountB = big.NewInt(0)
	state.CancelledAmountS = big.NewInt(0)

	if blockNumber == nil {
		state.UpdatedBlock = big.NewInt(0)
	} else {
		state.UpdatedBlock = blockNumber
	}

	// calculate order amount and settled
	SettleOrderAmountOnChain(state)

	// check order finished status
	SettleOrderStatus(state, mc, false)

	// convert order
	model := &dao.Order{}
	model.ConvertDown(state)

	return model, nil
}

func SettleOrderAmountOnChain(state *types.OrderState) error {
	// TODO(fuk): 系统暂时只会从gateway接收新订单,而不会有部分成交的订单
	return nil

	var (
		cancelled, cancelOrFilled, dealt *big.Int
		err                              error
	)

	protocol := state.RawOrder.DelegateAddress
	orderhash := state.RawOrder.Hash

	// get order cancelled amount from chain
	if cancelled, err = loopringaccessor.GetCancelled(protocol, orderhash, "latest"); err != nil {
		return fmt.Errorf("order manager,handle gateway order,order %s getCancelled error:%s", orderhash.Hex(), err.Error())
	}

	// get order cancelledOrFilled amount from chain
	if cancelOrFilled, err = loopringaccessor.GetCancelledOrFilled(protocol, orderhash, "latest"); err != nil {
		return fmt.Errorf("order manager,handle gateway order,order %s getCancelledOrFilled error:%s", orderhash.Hex(), err.Error())
	}

	if cancelOrFilled.Cmp(cancelled) < 0 {
		return fmt.Errorf("order manager,handle gateway order,order %s cancelOrFilledAmount:%s < cancelledAmount:%s", orderhash.Hex(), cancelOrFilled.String(), cancelled.String())
	}

	dealt = big.NewInt(0).Sub(cancelOrFilled, cancelled)

	if state.RawOrder.BuyNoMoreThanAmountB {
		state.DealtAmountB = dealt
		state.CancelledAmountB = cancelled
	} else {
		state.DealtAmountS = dealt
		state.CancelledAmountS = cancelled
	}

	return nil
}
