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
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

type OrderTxHandler struct {
	OrderHash   common.Hash
	OrderStatus types.OrderStatus
	BaseHandler
}

func (handler *OrderTxHandler) HandlePending() error {
	if handler.TxInfo.Status != types.TX_STATUS_PENDING {
		return nil
	}
	if !handler.requirePermission() {
		return nil
	}
	if err := handler.saveOrderPendingTx(); err != nil {
		log.Debugf(handler.format(), handler.value())
	}
	return nil
}

func (handler *OrderTxHandler) HandleFailed() error {
	if handler.TxInfo.Status != types.TX_STATUS_FAILED {
		return nil
	}
	if !handler.requirePermission() {
		return nil
	}
	if err := handler.saveOrderPendingTx(); err != nil {
		log.Debugf(handler.format(), handler.value())
	}
	return nil
}

func (handler *OrderTxHandler) HandleSuccess() error {
	if handler.TxInfo.Status != types.TX_STATUS_SUCCESS {
		return nil
	}
	if !handler.requirePermission() {
		return nil
	}
	if err := handler.saveOrderPendingTx(); err != nil {
		log.Debugf(handler.format(), handler.value())
	}
	return nil
}

func (handler *OrderTxHandler) HandlerCorrelatedTx() error {
	return nil
}

// 查询用户是否拥有修改订单状态的权限
func (handler *OrderTxHandler) requirePermission() bool {
	owner := handler.TxInfo.From
	return cache.HasOrderPermission(handler.Rds, owner)
}

// todo:查询orderTx表里是否有pending的tx
func (handler *OrderTxHandler) getPendingTx(orderhash common.Hash) []omtyp.OrderRelatedPendingTx {
	var list []omtyp.OrderRelatedPendingTx
	return list
}

// func (handler *OrderTxHandler)

func (handler *OrderTxHandler) saveOrderPendingTx() error {
	var (
		model = &dao.OrderTransaction{}
		err   error
	)

	rds := handler.Rds

	if model, err = rds.GetOrderTx(handler.OrderHash, handler.TxInfo.TxHash); err == nil && model.OrderStatus == uint8(handler.TxInfo) {
		return fmt.Errorf(handler.format("err:order %s already exist"), handler.value(handler.OrderHash.Hex()))
	}

	var record omtyp.OrderRelatedPendingTx
	record.Owner = handler.TxInfo.From
	record.TxHash = handler.TxInfo.TxHash
	record.Nonce = handler.TxInfo.Nonce.Int64()
	record.OrderHash = handler.OrderHash
	record.OrderStatus = handler.OrderStatus
	model.ConvertDown(&record)

	if handler.TxInfo.Status == types.TX_STATUS_PENDING {
		err = rds.Add(model)
	} else {
		err = rds.Del(model)
	}

	if err != nil {
		return fmt.Errorf(handler.format("err"), handler.value(err.Error()))
	} else {
		return nil
	}
}

func (handler *OrderTxHandler) format(fields ...string) string {
	baseformat := "order manager orderTxHandler, tx:%s, owner:%s, txstatus:%s, nonce:%s"
	for _, v := range fields {
		baseformat += ", " + v
	}
	return baseformat
}

func (handler *OrderTxHandler) value(values ...string) []string {
	basevalues := []string{handler.TxInfo.TxHash.Hex(), handler.TxInfo.From.Hex(), types.StatusStr(handler.TxInfo.Status), handler.TxInfo.Nonce.String()}
	basevalues = append(basevalues, values...)
	return basevalues
}
