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
	"github.com/Loopring/relay-lib/cache"
	"github.com/ethereum/go-ethereum/common"
)

const UserPendingOrderKeyPrefix = "om_user_pending_order_"

func SetPendingOrder(owner common.Address, orderhash common.Hash) error {
	key := getKey(owner)
	member := getMember(orderhash)
	cache.SAdd(key, 0, member)
	return nil
}

func DelPendingOrder(owner common.Address, orderhash common.Hash) error {
	key := getKey(owner)
	member := getMember(orderhash)
	cache.SRem(key, member)
	return nil
}

func ExistPendingOrder(owner common.Address, orderhash common.Hash) bool {
	key := getKey(owner)
	member := getMember(orderhash)
	ok, _ := cache.SIsMember(key, member)
	return ok
}

func getKey(owner common.Address) string {
	return UserPendingOrderKeyPrefix + owner.Hex()
}

func getMember(orderhash common.Hash) []byte {
	return []byte(orderhash.Hex())
}

// todo
func GetPendingOrders(owner common.Address) []common.Hash {
	var list []common.Hash

	return list
}
