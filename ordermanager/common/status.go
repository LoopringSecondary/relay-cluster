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

package common

import "github.com/Loopring/relay-lib/types"

var PendingStatus = []types.OrderStatus{
	types.ORDER_PENDING,
	types.ORDER_CANCELLING,
	types.ORDER_CUTOFFING,
}

func IsPendingStatus(status types.OrderStatus) bool {
	for _, v := range PendingStatus {
		if status == v {
			return true
		}
	}

	return false
}

var ValidFlexCancelStatus = []types.OrderStatus{
	types.ORDER_NEW,
	types.ORDER_PARTIAL,
	types.ORDER_PENDING,
}

var ValidMinerStatus = []types.OrderStatus{
	types.ORDER_NEW,
	types.ORDER_PARTIAL,
	types.ORDER_PENDING,
}

// 同一个订单必须允许多次cancel&cutoff,有的cancel/cutoff可能会不成功,后续的动作可以跟进
var ValidCutoffStatus = []types.OrderStatus{
	types.ORDER_NEW,
	types.ORDER_PARTIAL,
	types.ORDER_PENDING,
	types.ORDER_CANCELLING,
	types.ORDER_CUTOFFING,
}
