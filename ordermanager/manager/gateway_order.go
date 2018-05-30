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
	"github.com/Loopring/relay-lib/marketcap"
	"github.com/Loopring/relay-lib/types"
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
