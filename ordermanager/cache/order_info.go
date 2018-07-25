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
	"encoding/json"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	OrderPrefix = "om_order_"
	OrderTtl    = 120 // todo modify it to 86400
)

func BaseInfo(orderhash common.Hash) (*types.OrderState, error) {
	key := OrderPrefix + orderhash.Hex()
	state := &types.OrderState{}

	// todo delete it
	model, err := rds.GetOrderByHash(orderhash)
	if err != nil {
		return nil, err
	}
	model.ConvertUp(state)
	return state, nil

	if bs, err := cache.Get(key); err != nil {
		model, err := rds.GetOrderByHash(orderhash)
		if err != nil {
			return nil, err
		}
		model.ConvertUp(state)
		bs, _ := json.Marshal(state)
		cache.Set(key, bs, OrderTtl)
	} else {
		json.Unmarshal(bs, state)
	}

	return state, nil
}
