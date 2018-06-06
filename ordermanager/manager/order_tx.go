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
	"github.com/Loopring/relay-cluster/ordermanager/cache"
	omtyp "github.com/Loopring/relay-cluster/ordermanager/types"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

// orderTx中同一个order 最多有三条记录 分别属于order owner&miner
// 1、当订单处于pending状态时允许用户cancel/cutoff
// 2、当订单处于cancel/cutoff时不允许miner pending
// 第一种情况,

type OrderTxHandler struct {
	Event *omtyp.OrderTx
	BaseHandler
}

func NewOrderTxHandler(basehandler BaseHandler) *OrderTxHandler {
	event := &omtyp.OrderTx{
		Owner:  basehandler.TxInfo.From,
		TxHash: basehandler.TxInfo.TxHash,
		Nonce:  basehandler.TxInfo.Nonce.Int64(),
	}
	handler := &OrderTxHandler{BaseHandler: basehandler, Event: event}
	return handler
}

func NewOrderTxHandlerAndSaving(basehandler BaseHandler, orderhashlist []common.Hash, orderStatus types.OrderStatus) error {
	handler := NewOrderTxHandler(basehandler)
	for _, v := range orderhashlist {
		handler.Event.OrderHash = v
		handler.Event.OrderStatus = orderStatus
		handler.saveOrderPendingTx()
	}
	return nil
}

func (handler *OrderTxHandler) HandlePending() error {
	//if handler.TxInfo.Status != types.TX_STATUS_PENDING {
	//	return nil
	//}
	//if !handler.requirePermission() {
	//	return nil
	//}
	//if err := handler.saveOrderPendingTx(); err != nil {
	//	log.Debugf(handler.format(), handler.value())
	//}
	return nil
}

func (handler *OrderTxHandler) HandleFailed() error {
	if handler.TxInfo.Status != types.TX_STATUS_FAILED {
		return nil
	}
	if !handler.requirePermission() {
		return nil
	}
	return handler.processPendingTx()
}

func (handler *OrderTxHandler) HandleSuccess() error {
	if handler.TxInfo.Status != types.TX_STATUS_SUCCESS {
		return nil
	}
	if !handler.requirePermission() {
		return nil
	}
	return handler.processPendingTx()
}

// 查询用户是否拥有修改订单状态的权限
func (handler *OrderTxHandler) requirePermission() bool {
	owner := handler.TxInfo.From
	return cache.HasOrderPermission(handler.Rds, owner)
}

func (handler *OrderTxHandler) processPendingTx() error {
	//todo 1.删除用户无效pending tx
	//todo 2.获取用户其他pending tx
	models, _ := handler.Rds.GetOrderRelatedPendingTxList(handler.TxInfo.From)
	if len(models) == 0 {
		return nil
	}

	for _, model := range models {
		var tx omtyp.OrderTx
		model.ConvertUp(&tx)
	}

	return nil
}

func (handler *OrderTxHandler) saveOrderPendingTx() error {
	var (
		model = &dao.OrderTransaction{}
		err   error
	)

	rds := handler.Rds
	event := handler.Event

	model, err = rds.GetOrderRelatedPendingTx(event.OrderHash, handler.TxInfo.TxHash)
	if EventRecordDuplicated(handler.TxInfo.Status, uint8(types.TX_STATUS_PENDING), err != nil) {
		return fmt.Errorf(handler.format("err:order %s already exist"), handler.value(event.OrderHash.Hex())...)
	}

	model.ConvertDown(event)

	if handler.TxInfo.Status == types.TX_STATUS_PENDING {
		return rds.Add(model)
	} else {
		return rds.Del(model)
	}
}

func (handler *OrderTxHandler) format(fields ...string) string {
	baseformat := "order manager, orderTxHandler, tx:%s, owner:%s, txstatus:%s, nonce:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *OrderTxHandler) value(values ...interface{}) []interface{} {
	basevalues := []interface{}{handler.TxInfo.TxHash.Hex(), handler.TxInfo.From.Hex(), types.StatusStr(handler.TxInfo.Status), handler.TxInfo.Nonce.String()}
	basevalues = append(basevalues, values...)
	return basevalues
}
