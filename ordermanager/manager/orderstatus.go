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
	"github.com/Loopring/relay-lib/marketcap"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
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

type OrderTxSwitcher struct {
	Rds         *dao.RdsService
	OrderHash   common.Hash
	OrderStatus types.OrderStatus
	TxInfo      types.TxInfo
	MarketCap   marketcap.MarketCapProvider
}

// todo
func (s *OrderTxSwitcher) FlexibleCancellationPendingProcedure() error {
	return nil

	state, err := cache.BaseInfo(s.Rds, s.OrderHash)
	if err != nil {
		return err
	}

	if !IsValidPendingStatus(state.Status) {
		return fmt.Errorf("status invalid")
	}

	//list, _ := rds.GetOrderTxList(state.RawOrder.Hash)

	//// 1.去重
	//maxnonce, err := rds.MaxNonce(txinfo.From, state.RawOrder.Hash)
	//if err != nil {
	//	var tx dao.OrderTransaction
	//	tx.ConvertUp(state.RawOrder.Hash, newstatus, txinfo)
	//	return rds.Add(&tx)
	//}
	//
	//if maxnonce > txinfo.Nonce.Int64() {
	//	return fmt.Errorf("maxnonce:%d > currentNonce:%d", maxnonce, txinfo.Nonce.Int64())
	//}
	//
	//for _, v := range list {
	//	state.Status = types.OrderStatus(v.Status)
	//
	//	// 以用户状态为准
	//	if common.HexToAddress(v.Owner) == state.RawOrder.Owner {
	//		break
	//	}
	//}
	//
	//var tx dao.OrderTransaction
	//tx.ConvertUp(state.RawOrder.Hash, newstatus, txinfo)
	//rds.Add(&tx)

	return s.Rds.UpdateOrderStatus(s.OrderHash, s.OrderStatus)
}

func (s *OrderTxSwitcher) FlexibleCancellationFailedProcedure() error {
	return nil
}

// cancelling/cutoffing与pending冲突的问题,以用户行为cancelling/cutoffing为准

func IsValidPendingStatus(status types.OrderStatus) bool {
	list := []types.OrderStatus{
		types.ORDER_NEW,
		types.ORDER_PARTIAL,
	}

	for _, v := range list {
		if v == status {
			return true
		}
	}

	return false
}

var IngStatus = []types.OrderStatus{
	types.ORDER_PENDING,
	types.ORDER_CANCELLING,
	types.ORDER_CUTOFFING,
}
