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
	if _, err := handler.saveEvent(); err != nil {
		log.Debugf(err.Error())
	} else {
		log.Debugf(handler.format(), handler.value())
	}
	return nil
}

func (handler *CutoffHandler) HandleFailed() error {
	if _, err := handler.saveEvent(); err != nil {
		log.Debugf(err.Error())
	} else {
		log.Debugf(handler.format(), handler.value())
	}
	return nil
}

func (handler *CutoffHandler) HandleSuccess() error {
	event := handler.Event
	rds := handler.Rds
	cutoffCache := handler.CutoffCache

	if event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	orderhashList, err := handler.saveEvent()
	if err != nil {
		return err
	}

	// 首次存储到缓存，lastCutoff == currentCutoff
	lastCutoff := cutoffCache.GetCutoff(event.Protocol, event.Owner)
	if event.Cutoff.Cmp(lastCutoff) < 0 {
		log.Debugf(handler.format("lastCutofftime:%s > currentCutoffTime:%s"), handler.value(lastCutoff.String(), event.Cutoff.String()))
	} else {
		cutoffCache.UpdateCutoff(event.Protocol, event.Owner, event.Cutoff)
		log.Debugf(handler.format("cutoffTimestamp:%s"), handler.value(event.Cutoff.String()))
	}

	rds.SetCutOffOrders(orderhashList, event.BlockNumber)

	notify.NotifyCutoff(event)

	return nil
}

// return orderhash list
func (handler *CutoffHandler) saveEvent() ([]common.Hash, error) {
	rds := handler.Rds
	event := handler.Event

	var (
		model dao.CutOffEvent
		list  []common.Hash
		err   error
	)

	// save cancel event
	model, err = rds.GetCutoffEvent(event.TxHash)
	if EventRecordDuplicated(event.Status, model.Status, err != nil) {
		return list, fmt.Errorf(handler.format("err:tx already exist"), handler.value())
	}

	if orders, _ := rds.GetCutoffOrders(event.Owner, event.Cutoff); len(orders) > 0 {
		for _, v := range orders {
			var state types.OrderState
			v.ConvertUp(&state)
			list = append(list, state.RawOrder.Hash)
		}
	}

	event.OrderHashList = list
	model.ConvertDown(event)
	model.Fork = false

	if event.Status == types.TX_STATUS_PENDING {
		err = rds.Add(model)
	} else {
		err = rds.Save(model)
	}

	if err != nil {
		return list, fmt.Errorf(handler.format("err:%s"), handler.value(err.Error()))
	} else {
		return list, nil
	}
}

func (handler *CutoffHandler) format(fields ...string) string {
	baseformat := "order manager CutoffHandler, tx:%s, txstatus:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *CutoffHandler) value(values ...string) []string {
	basevalues := []string{handler.Event.TxHash.Hex(), types.StatusStr(handler.Event.Status)}
	basevalues = append(basevalues, values...)
	return basevalues
}
