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

package cache

import (
	"github.com/Loopring/relay-cluster/dao"
	omtyp "github.com/Loopring/relay-cluster/ordermanager/types"
	"github.com/Loopring/relay-lib/cache"
	"github.com/ethereum/go-ethereum/common"
)

// miner & order owner
func HasOrderPermission(rds *dao.RdsService, owner common.Address) bool {
	ttl := int64(86400 * 10)

	key := "om_order_permission_" + owner.Hex()
	if ok, _ := cache.Exists(key); ok {
		return true
	}

	if !rds.IsOrderOwner(owner) && !rds.IsMiner(owner) {
		return false
	}

	cache.Set(key, []byte(""), ttl)
	return true
}

// todo
func SetPendingOrders(owner common.Address, orderhash common.Hash) error {
	return nil
}

// todo
func DelPendingOrders(owner common.Address, orderhash common.Hash) error {
	return nil
}

// todo
func GetPendingOrders(owner common.Address) []common.Hash {
	var list []common.Hash

	return list
}

// todo
func GetOrderPendingTx(orderhash common.Hash) []omtyp.OrderTx {
	var list []omtyp.OrderTx
	return list
}
