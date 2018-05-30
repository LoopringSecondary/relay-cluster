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
	omcm "github.com/Loopring/relay-cluster/ordermanager/common"
	notify "github.com/Loopring/relay-cluster/util"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

type CutoffHandler struct {
	Event       *types.CutoffEvent
	Rds         *dao.RdsService
	CutoffCache *omcm.CutoffCache
}

func (handler *CutoffHandler) HandleFailed() error {
	return nil
}

func (handler *CutoffHandler) HandlePending() error {
	return nil
}

func (handler *CutoffHandler) HandleSuccess() error {
	event := handler.Event
	rds := handler.Rds
	cutoffCache := handler.CutoffCache

	if event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	// check tx exist
	_, err := rds.GetCutoffEvent(event.TxHash)
	if err == nil {
		log.Debugf("order manager,handle order cutoff event tx:%s error:transaction have already exist", event.TxHash.Hex())
		return nil
	}

	lastCutoff := cutoffCache.GetCutoff(event.Protocol, event.Owner)

	var orderHashList []common.Hash

	// 首次存储到缓存，lastCutoff == currentCutoff
	if event.Cutoff.Cmp(lastCutoff) < 0 {
		log.Debugf("order manager,handle cutoff event tx:%s, protocol:%s - owner:%s lastCutofftime:%s > currentCutoffTime:%s", event.TxHash.Hex(), event.Protocol.Hex(), event.Owner.Hex(), lastCutoff.String(), event.Cutoff.String())
	} else {
		cutoffCache.UpdateCutoff(event.Protocol, event.Owner, event.Cutoff)
		if orders, _ := rds.GetCutoffOrders(event.Owner, event.Cutoff); len(orders) > 0 {
			for _, v := range orders {
				var state types.OrderState
				v.ConvertUp(&state)
				orderHashList = append(orderHashList, state.RawOrder.Hash)
			}
			rds.SetCutOffOrders(orderHashList, event.BlockNumber)
		}
		log.Debugf("order manager,handle cutoff event tx:%s, owner:%s, cutoffTimestamp:%s", event.TxHash.Hex(), event.Owner.Hex(), event.Cutoff.String())
	}

	// save cutoff event
	event.OrderHashList = orderHashList
	newCutoffEventModel := &dao.CutOffEvent{}
	newCutoffEventModel.ConvertDown(event)
	newCutoffEventModel.Fork = false

	if err := rds.Add(newCutoffEventModel); err != nil {
		return err
	}

	notify.NotifyCutoff(event)

	return nil
}
