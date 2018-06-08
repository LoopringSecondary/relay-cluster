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
	"github.com/ethereum/go-ethereum/common"
)

type CutoffHandler struct {
	Event *types.CutoffEvent
	BaseHandler
}

func (handler *CutoffHandler) HandlePending() error {
	if handler.Event.Status != types.TX_STATUS_PENDING {
		return nil
	}
	orderhashList, err := handler.getOrdersAndSaveEvent()
	if err != nil {
		return err
	}

	log.Debugf(handler.format(), handler.value()...)

	for _, orderhash := range orderhashList {
		txhandler := FullOrderTxHandler(handler.BaseHandler, orderhash, types.ORDER_CANCELLING)
		txhandler.HandleOrderRelatedTxPending()
	}

	return nil
}

func (handler *CutoffHandler) HandleFailed() error {
	if handler.Event.Status != types.TX_STATUS_FAILED {
		return nil
	}
	orderhashList, err := handler.getOrdersAndSaveEvent()
	if err != nil {
		return err
	}

	log.Debugf(handler.format(), handler.value()...)

	for _, orderhash := range orderhashList {
		txhandler := FullOrderTxHandler(handler.BaseHandler, orderhash, types.ORDER_CANCELLING)
		txhandler.HandleOrderRelatedTxFailed()
	}

	return nil
}

func (handler *CutoffHandler) HandleSuccess() error {
	if handler.Event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	event := handler.Event
	rds := handler.Rds
	cutoffCache := handler.CutoffCache

	orderhashList, err := handler.getOrdersAndSaveEvent()
	if err != nil {
		return err
	}

	log.Debugf(handler.format(), handler.value()...)

	// 首次存储到缓存，lastCutoff == currentCutoff
	lastCutoff := cutoffCache.GetCutoff(event.Protocol, event.Owner)
	if event.Cutoff.Cmp(lastCutoff) < 0 {
		return fmt.Errorf(handler.format("lastCutofftime:%s > currentCutoffTime:%s"), handler.value(lastCutoff.String(), event.Cutoff.String())...)
	}

	cutoffCache.UpdateCutoff(event.Protocol, event.Owner, event.Cutoff)
	rds.SetCutOffOrders(orderhashList, event.BlockNumber)

	notify.NotifyCutoff(event)

	for _, orderhash := range orderhashList {
		txhandler := FullOrderTxHandler(handler.BaseHandler, orderhash, types.ORDER_CANCELLING)
		txhandler.HandleOrderRelatedTxSuccess()
	}

	return nil
}

// return orderhash list
func (handler *CutoffHandler) getOrdersAndSaveEvent() ([]common.Hash, error) {
	rds := handler.Rds
	event := handler.Event

	var (
		model dao.CutOffEvent
		list  []common.Hash
		err   error
	)

	model, err = rds.GetCutoffEvent(event.TxHash)
	if EventRecordDuplicated(event.Status, model.Status, err != nil) {
		return list, fmt.Errorf(handler.format("err:tx already exist"), handler.value()...)
	}

	if handler.Event.Status == types.TX_STATUS_PENDING {
		orders, _ := rds.GetCutoffOrders(event.Owner, event.Cutoff)
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

func (handler *CutoffHandler) format(fields ...string) string {
	baseformat := "order manager, CutoffHandler, tx:%s, owner:%s, cutofftime:%s, txstatus:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *CutoffHandler) value(values ...interface{}) []interface{} {
	basevalues := []interface{}{handler.Event.TxHash.Hex(), handler.Event.Owner.Hex(), handler.Event.Cutoff.String(), types.StatusStr(handler.Event.Status)}
	basevalues = append(basevalues, values...)
	return basevalues
}
