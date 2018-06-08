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
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

func BaseInfo(rds *dao.RdsService, orderhash common.Hash) (*types.OrderState, error) {
	state := &types.OrderState{}
	model, err := rds.GetOrderByHash(orderhash)
	if err != nil {
		return nil, err
	}

	model.ConvertUp(state)

	// todo(fuk):set in redis
	return state, nil
}
