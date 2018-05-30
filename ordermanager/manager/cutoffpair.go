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
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

type CutoffPairHandler struct {
	Event *types.CutoffPairEvent
	BaseHandler
}

func (handler *CutoffPairHandler) HandleFailed() error {
	return nil
}

func (handler *CutoffPairHandler) HandlePending() error {
	return nil
}

func (handler *CutoffPairHandler) HandleSuccess() error {
	event := handler.Event
	rds := handler.Rds
	cutoffCache := handler.CutoffCache

	if event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	// check tx exist
	_, err := rds.GetCutoffPairEvent(event.TxHash)
	if err == nil {
		log.Debugf("order manager,handle order cutoffPair event tx:%s error:transaction have already exist", event.TxHash.Hex())
		return nil
	}

	lastCutoffPair := cutoffCache.GetCutoffPair(event.Protocol, event.Owner, event.Token1, event.Token2)

	var orderHashList []common.Hash
	// 首次存储到缓存，lastCutoffPair == currentCutoffPair
	if event.Cutoff.Cmp(lastCutoffPair) < 0 {
		log.Debugf("order manager,handle cutoffPair event tx:%s, protocol:%s - owner:%s lastCutoffPairtime:%s > currentCutoffPairTime:%s", event.TxHash.Hex(), event.Protocol.Hex(), event.Owner.Hex(), lastCutoffPair.String(), event.Cutoff.String())
	} else {
		cutoffCache.UpdateCutoffPair(event.Protocol, event.Owner, event.Token1, event.Token2, event.Cutoff)
		if orders, _ := rds.GetCutoffPairOrders(event.Owner, event.Token1, event.Token2, event.Cutoff); len(orders) > 0 {
			for _, v := range orders {
				var state types.OrderState
				v.ConvertUp(&state)
				orderHashList = append(orderHashList, state.RawOrder.Hash)
			}
			rds.SetCutOffOrders(orderHashList, event.BlockNumber)
		}
		log.Debugf("order manager,handle cutoffPair event tx:%s, owner:%s, token1:%s, token2:%s, cutoffTimestamp:%s", event.TxHash.Hex(), event.Owner.Hex(), event.Token1.Hex(), event.Token2.Hex(), event.Cutoff.String())
	}

	// save transaction
	event.OrderHashList = orderHashList
	newCutoffPairEventModel := &dao.CutOffPairEvent{}
	newCutoffPairEventModel.ConvertDown(event)
	newCutoffPairEventModel.Fork = false

	if err := rds.Add(newCutoffPairEventModel); err != nil {
		return err
	}

	notify.NotifyCutoffPair(event)
	return nil
}
