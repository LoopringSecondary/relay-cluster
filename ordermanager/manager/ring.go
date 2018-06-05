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
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

type SubmitRingHandler struct {
	Event *types.SubmitRingMethodEvent
	BaseHandler
}

func (handler *SubmitRingHandler) HandlePending() error {
	if handler.Event.Status != types.TX_STATUS_PENDING {
		return nil
	}

	// save pending tx
	if model, err := handler.Rds.FindRingMined(handler.Event.TxHash.Hex()); err == nil {
		return fmt.Errorf(handler.format("err:tx already exist"), handler.value())
	} else {
		log.Debugf(handler.format(), handler.value())
		model.FromSubmitRingMethod(handler.Event)
		return handler.Rds.Add(model)
	}

	// return NewOrderTxHandlerAndSaving(handler.BaseHandler, handler.orderHashList(), types.ORDER_PENDING)
}

func (handler *SubmitRingHandler) HandleFailed() error {
	if handler.Event.Status != types.TX_STATUS_FAILED {
		return nil
	}

	// save failed tx
	if model, err := handler.Rds.FindRingMined(handler.Event.TxHash.Hex()); err != nil {
		return fmt.Errorf(handler.format("err:tx already exist"), handler.value())
	} else {
		log.Debugf(handler.format(), handler.value())
		model.FromSubmitRingMethod(handler.Event)
		return handler.Rds.Save(model)
	}
}

func (handler *SubmitRingHandler) orderHashList() []common.Hash {
	var list []common.Hash
	for _, v := range handler.Event.OrderList {
		list = append(list, v.Hash)
	}
	return list
}

func (handler *SubmitRingHandler) format(fields ...string) string {
	baseformat := "order manager ringMinedHandler, tx:%s, txstatus:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *SubmitRingHandler) value(values ...string) []string {
	basevalues := []string{handler.Event.TxHash.Hex(), types.StatusStr(handler.Event.Status)}
	basevalues = append(basevalues, values...)
	return basevalues
}

func (handler *SubmitRingHandler) HandleSuccess() error {
	return nil
}

type RingMinedHandler struct {
	Event *types.RingMinedEvent
	BaseHandler
}

func (handler *RingMinedHandler) HandlePending() error {
	return nil
}

func (handler *RingMinedHandler) HandleFailed() error {
	return nil
}

func (handler *RingMinedHandler) HandleSuccess() error {
	if handler.Event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	event := handler.Event
	rds := handler.Rds

	if model, err := rds.FindRingMined(event.TxHash.Hex()); err != nil {
		return fmt.Errorf(handler.format("err:tx already exist"), handler.value())
	} else {
		log.Debugf(handler.format(), handler.value())
		model.ConvertDown(event)
		return rds.Save(model)
	}

	// todo
	//return NewOrderTxHandlerAndSaving(handler.BaseHandler, handler.orderHashList(), types.ORDER_PENDING)
}

func (handler *RingMinedHandler) format(fields ...string) string {
	baseformat := "order manager ringMinedHandler, tx:%s, ringhash:%s, txstatus:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *RingMinedHandler) value(values ...string) []string {
	basevalues := []string{handler.Event.TxHash.Hex(), handler.Event.Ringhash.Hex(), types.StatusStr(handler.Event.Status)}
	basevalues = append(basevalues, values...)
	return basevalues
}
