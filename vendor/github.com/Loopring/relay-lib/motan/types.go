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

package motan

import (
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type AccountBalanceAndAllowanceReq struct {
	Owner   common.Address
	Token   common.Address
	Spender common.Address
}

type AccountBalanceAndAllowanceRes struct {
	Balance, Allowance *big.Int
	Err                string
}

type MinerOrdersReq struct {
	Delegate             common.Address
	TokenS               common.Address
	TokenB               common.Address
	Length               int
	ReservedTime         int64
	StartBlockNumber     int64
	EndBlockNumber       int64
	FilterOrderHashLists []*types.OrderDelayList
}

type MinerOrdersRes struct {
	List []*types.OrderState
}
