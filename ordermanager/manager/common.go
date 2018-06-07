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
	"github.com/Loopring/relay-lib/marketcap"
	"github.com/Loopring/relay-lib/types"
	"math/big"
)

func SettleOrderStatus(state *types.OrderState, mc marketcap.MarketCapProvider, isCancel bool) {
	zero := big.NewInt(0)
	finishAmountS := big.NewInt(0).Add(state.CancelledAmountS, state.DealtAmountS)
	totalAmountS := big.NewInt(0).Add(finishAmountS, state.SplitAmountS)
	finishAmountB := big.NewInt(0).Add(state.CancelledAmountB, state.DealtAmountB)
	totalAmountB := big.NewInt(0).Add(finishAmountB, state.SplitAmountB)
	totalAmount := big.NewInt(0).Add(totalAmountS, totalAmountB)

	if totalAmount.Cmp(zero) <= 0 {
		state.Status = types.ORDER_NEW
	} else if !mc.IsOrderValueDust(state) {
		state.Status = types.ORDER_PARTIAL
	} else if isCancel {
		state.Status = types.ORDER_CANCEL
	} else {
		state.Status = types.ORDER_FINISHED
	}
}

func EventRecordDuplicated(eventStatus types.TxStatus, modelStatus uint8, noRecord bool) bool {
	if uint8(eventStatus) == modelStatus && !noRecord {
		return true
	} else {
		return false
	}
}
