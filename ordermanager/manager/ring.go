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
	model, err := handler.Rds.FindRingMined(handler.Event.TxHash.Hex())
	if err == nil {
		return fmt.Errorf(handler.format("err:tx already exist"), handler.value()...)
	}

	log.Debugf(handler.format(), handler.value()...)
	model.FromSubmitRingMethod(handler.Event)
	handler.Rds.Add(model)

	for _, v := range handler.Event.OrderList {
		txhandler := FullOrderTxHandler(handler.BaseHandler, v.Hash, types.ORDER_CANCELLING)
		txhandler.HandleOrderRelatedTxPending()
	}

	return nil
}

func (handler *SubmitRingHandler) HandleFailed() error {
	if handler.Event.Status != types.TX_STATUS_FAILED {
		return nil
	}

	// save failed tx
	model, err := handler.Rds.FindRingMined(handler.Event.TxHash.Hex())
	if err != nil {
		return fmt.Errorf(handler.format("err:tx already exist"), handler.value()...)
	}

	log.Debugf(handler.format(), handler.value()...)
	model.FromSubmitRingMethod(handler.Event)
	handler.Rds.Save(model)

	for _, v := range handler.Event.OrderList {
		txhandler := FullOrderTxHandler(handler.BaseHandler, v.Hash, types.ORDER_CANCELLING)
		txhandler.HandleOrderRelatedTxFailed()
	}

	return nil
}

func (handler *SubmitRingHandler) orderHashList() []common.Hash {
	var list []common.Hash
	for _, v := range handler.Event.OrderList {
		list = append(list, v.Hash)
	}
	return list
}

func (handler *SubmitRingHandler) format(fields ...string) string {
	baseformat := "order manager, ringMinedHandler, tx:%s, txstatus:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *SubmitRingHandler) value(values ...interface{}) []interface{} {
	basevalues := []interface{}{handler.Event.TxHash.Hex(), types.StatusStr(handler.Event.Status)}
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

	model, err := rds.FindRingMined(event.TxHash.Hex())
	if err != nil {
		return fmt.Errorf(handler.format("err:tx already exist"), handler.value()...)
	}

	log.Debugf(handler.format(), handler.value()...)
	model.ConvertDown(event)
	rds.Save(model)

	orderhashlist := dao.UnmarshalStrToHashList(model.OrderHashList)
	for _, v := range orderhashlist {
		txhandler := FullOrderTxHandler(handler.BaseHandler, v, types.ORDER_CANCELLING)
		txhandler.HandleOrderRelatedTxSuccess()
	}

	return nil
}

func (handler *RingMinedHandler) format(fields ...string) string {
	baseformat := "order manager, ringMinedHandler, tx:%s, ringhash:%s, txstatus:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *RingMinedHandler) value(values ...interface{}) []interface{} {
	basevalues := []interface{}{handler.Event.TxHash.Hex(), handler.Event.Ringhash.Hex(), types.StatusStr(handler.Event.Status)}
	basevalues = append(basevalues, values...)
	return basevalues
}
