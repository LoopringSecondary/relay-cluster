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
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

type OrderCorrelatedTxEvent struct {
	From   common.Address
	Nonce  int64
	Status types.TxStatus
}

func (e *OrderCorrelatedTxEvent) FromTxInfo(src types.TxInfo) {
	e.From = src.From
	e.Nonce = src.Nonce.Int64()
	e.Status = src.Status
}

type OrderCorrelatedTxHandler struct {
	Event *OrderCorrelatedTxEvent
	BaseHandler
}

func (handler *OrderCorrelatedTxHandler) HandlePending() error {
	return nil
}

func (handler *OrderCorrelatedTxHandler) HandleFailed() error {
	return nil
}

func (handler *OrderCorrelatedTxHandler) HandleSuccess() error {
	return nil
}

// 获取跟订单相关的pending tx,判断当前pending tx在新的nonce到来后的状态
