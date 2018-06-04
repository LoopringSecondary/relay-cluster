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
	"github.com/Loopring/relay-cluster/ordermanager/common"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
)

type FlexCancelOrderHandler struct {
	Event *types.FlexCancelOrderEvent
	BaseHandler
}

func (handler *FlexCancelOrderHandler) HandlePending() error {
	return nil
}

func (handler *FlexCancelOrderHandler) HandleFailed() error {
	return nil
}

func (handler *FlexCancelOrderHandler) HandleSuccess() error {
	if err := handler.Validate(); err != nil {
		log.Debugf(err.Error())
		return err
	}

	validStatus := common.ValidFlexCancelStatus
	status := types.ORDER_FLEX_CANCEL

	var nums int64 = 0
	switch handler.Event.Type {
	case types.FLEX_CANCEL_BY_HASH:
		log.Debugf("order manager, FlexCancelOrderHandler, FLEX_CANCEL_BY_HASH orderhash:%s", handler.Event.OrderHash.Hex())
		nums = handler.Rds.FlexCancelOrderByHash(handler.Event.Owner, handler.Event.OrderHash, validStatus, status)

	case types.FLEX_CANCEL_BY_OWNER:
		log.Debugf("order manager, FlexCancelOrderHandler, FLEX_CANCEL_BY_OWNER owner:%s", handler.Event.Owner.Hex())
		nums = handler.Rds.FlexCancelOrderByOwner(handler.Event.Owner, validStatus, status)

	case types.FLEX_CANCEL_BY_TIME:
		log.Debugf("order manager, FlexCancelOrderHandler, FLEX_CANCEL_BY_TIME cutofftime:%d", handler.Event.CutoffTime)
		nums = handler.Rds.FlexCancelOrderByTime(handler.Event.Owner, handler.Event.CutoffTime, validStatus, status)

	case types.FLEX_CANCEL_BY_MARKET:
		market, _ := util.WrapMarketByAddress(handler.Event.TokenS.Hex(), handler.Event.TokenB.Hex())
		log.Debugf("order manager, FlexCancelOrderHandler, FLEX_CANCEL_BY_MARKET market:%s", market)
		nums = handler.Rds.FlexCancelOrderByMarket(handler.Event.Owner, handler.Event.CutoffTime, market, validStatus, status)
	}

	if nums == 0 {
		log.Debugf("order manager, flex cancel order, no invalid orders")
	}

	return nil
}

func (handler *FlexCancelOrderHandler) Validate() error {
	if types.IsZeroAddress(handler.Event.Owner) {
		return fmt.Errorf("order manager, FlexCancelOrderHandler validate, owner invalid")
	}

	switch handler.Event.Type {
	case types.FLEX_CANCEL_BY_HASH:
		if types.IsZeroHash(handler.Event.OrderHash) {
			return fmt.Errorf("order manager, FlexCancelOrderHandler validate, orderhash invalid")
		}

	case types.FLEX_CANCEL_BY_OWNER:
		if len(handler.Event.Owner.Hex()) < 10 {
			return fmt.Errorf("order manager, FlexCancelOrderHandler validate, owner length invalid")
		}

	case types.FLEX_CANCEL_BY_TIME:
		if handler.Event.CutoffTime <= 0 {
			return fmt.Errorf("order manager, FlexCancelOrderHandler validate, cutoffTimeStamp invalid")
		}

	case types.FLEX_CANCEL_BY_MARKET:
		if _, err := util.WrapMarketByAddress(handler.Event.TokenS.Hex(), handler.Event.TokenB.Hex()); err != nil {
			return fmt.Errorf("order manager, FlexCancelOrderHandler validate, market invalid")
		}

	default:
		return fmt.Errorf("order manager, FlexCancelOrderHandler validate, event type invalid")
	}

	return nil
}
