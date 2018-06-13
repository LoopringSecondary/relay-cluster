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
	omcm "github.com/Loopring/relay-cluster/ordermanager/common"
	notify "github.com/Loopring/relay-cluster/util"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

type CutoffPairHandler struct {
	Event *types.CutoffPairEvent
	BaseHandler
}

func (handler *CutoffPairHandler) HandlePending() error {
	if handler.Event.Status != types.TX_STATUS_PENDING {
		return nil
	}

	orderhashList, err := handler.getOrdersAndSaveEvent()
	if err != nil {
		return err
	}

	log.Debugf(handler.format(), handler.value()...)

	for _, orderhash := range orderhashList {
		txhandler := FullOrderTxHandler(handler.BaseHandler, orderhash, types.ORDER_CUTOFFING)
		txhandler.HandleOrderRelatedTxPending()
	}

	return nil
}

func (handler *CutoffPairHandler) HandleFailed() error {
	if handler.Event.Status != types.TX_STATUS_FAILED {
		return nil
	}

	orderhashList, err := handler.getOrdersAndSaveEvent()
	if err != nil {
		return err
	}

	log.Debugf(handler.format(), handler.value()...)

	for _, orderhash := range orderhashList {
		txhandler := FullOrderTxHandler(handler.BaseHandler, orderhash, types.ORDER_CUTOFFING)
		txhandler.HandleOrderRelatedTxFailed()
	}

	return nil
}

func (handler *CutoffPairHandler) HandleSuccess() error {
	if handler.Event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	event := handler.Event
	rds := handler.Rds
	cutoffCache := handler.CutoffCache

	orderhashlist, err := handler.getOrdersAndSaveEvent()
	if err != nil {
		return err
	}

	log.Debugf(handler.format(), handler.value()...)

	// 首次存储到缓存，lastCutoffPair == currentCutoffPair
	lastCutoffPair := cutoffCache.GetCutoffPair(event.Protocol, event.Owner, event.Token1, event.Token2)
	if event.Cutoff.Cmp(lastCutoffPair) < 0 {
		return fmt.Errorf(handler.format("lastCutoffPairTime:%s > currentCutoffPairTime:%s"), handler.value(lastCutoffPair.String(), event.Cutoff.String())...)
	}

	cutoffCache.UpdateCutoffPair(event.Protocol, event.Owner, event.Token1, event.Token2, event.Cutoff)
	rds.SetCutOffOrders(orderhashlist, event.BlockNumber)

	notify.NotifyCutoffPair(event)

	for _, orderhash := range orderhashlist {
		txhandler := FullOrderTxHandler(handler.BaseHandler, orderhash, types.ORDER_CUTOFFING)
		txhandler.HandleOrderRelatedTxSuccess()
	}

	return nil
}

func (handler *CutoffPairHandler) getOrdersAndSaveEvent() ([]common.Hash, error) {
	rds := handler.Rds
	event := handler.Event

	var (
		model dao.CutOffPairEvent
		list  []common.Hash
		err   error
	)

	model, err = rds.GetCutoffPairEvent(event.TxHash)
	if err := ValidateExistEvent(event.Status, model.Status, err); err != nil {
		return list, fmt.Errorf(handler.format("err:%s"), handler.value(err.Error())...)
	}

	if handler.Event.Status == types.TX_STATUS_PENDING {
		orders, _ := rds.GetCutoffPairOrders(event.Owner, event.Token1, event.Token2, event.Cutoff, omcm.ValidCutoffStatus)
		for _, v := range orders {
			var state types.OrderState
			v.ConvertUp(&state)
			list = append(list, state.RawOrder.Hash)
		}
		model.Fork = false
	} else {
		list = dao.UnmarshalStrToHashList(model.OrderHashList)
	}

	event.OrderHashList = list
	model.ConvertDown(event)

	if handler.Event.Status == types.TX_STATUS_PENDING {
		return list, rds.Add(&model)
	} else {
		return list, rds.Save(&model)
	}
}

func (handler *CutoffPairHandler) format(fields ...string) string {
	baseformat := "order manager cutoffPairHandler, tx:%s, owner:%s, token1:%s, token2:%s, cutoffTimestamp:%s, txstatus:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *CutoffPairHandler) value(values ...interface{}) []interface{} {
	basevalues := []interface{}{handler.Event.TxHash.Hex(), handler.Event.Owner.Hex(), handler.Event.Token1.Hex(), handler.Event.Token2.Hex(), handler.Event.Cutoff.String(), types.StatusStr(handler.Event.Status)}
	basevalues = append(basevalues, values...)
	return basevalues
}
