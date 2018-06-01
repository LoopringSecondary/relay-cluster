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
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
)

type SubmitRingHandler struct {
	Event *types.SubmitRingMethodEvent
	BaseHandler
}

func (handler *SubmitRingHandler) HandlePending() error {
	if handler.Event.Status != types.TX_STATUS_PENDING {
		return nil
	}

	switcher := handler.FullSwitcher(types.NilHash, types.ORDER_PENDING)

	for _, v := range handler.Event.OrderList {
		switcher.OrderHash = v.Hash
		if err := switcher.FlexibleCancellationPendingProcedure(); err != nil {
			log.Errorf(err.Error())
		}
	}

	// save pending tx
	if model, err := handler.Rds.FindRingMined(handler.Event.TxHash.Hex()); err == nil {
		log.Errorf("order manager, submitRingHandler, pending tx:%s already exist", handler.Event.TxHash.Hex())
		return nil
	} else {
		log.Debugf("order manager, submitRingHandler, pending tx:%s insert", handler.Event.TxHash.Hex())
		model.FromSubmitRingMethod(handler.Event)
		return handler.Rds.Add(model)
	}
}

func (handler *SubmitRingHandler) HandleFailed() error {
	if handler.Event.Status != types.TX_STATUS_FAILED {
		return nil
	}

	switcher := handler.FullSwitcher(types.NilHash, types.ORDER_PENDING)

	for _, v := range handler.Event.OrderList {
		switcher.OrderHash = v.Hash
		if err := switcher.FlexibleCancellationFailedProcedure(); err != nil {
			log.Errorf(err.Error())
		}
	}

	// save failed tx
	if model, err := handler.Rds.FindRingMined(handler.Event.TxHash.Hex()); err != nil {
		log.Errorf("order manager, submitRingHandler, failed tx:%s not exist", handler.Event.TxHash.Hex())
		return nil
	} else {
		log.Debugf("order manager, submitRingHandler, failed tx:%s updated", handler.Event.TxHash.Hex())
		model.FromSubmitRingMethod(handler.Event)
		return handler.Rds.Save(model)
	}
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
		log.Errorf("order manager,handle ringmined event,tx:%s ringhash:%s not exist", event.TxHash.Hex(), event.Ringhash.Hex())
		return nil
	} else {
		log.Debugf("order manager,handle ringmined event,tx:%s, ringhash:%s inserted", event.TxHash.Hex(), event.Ringhash.Hex())
		model.ConvertDown(event)
		return rds.Save(model)
	}
}
