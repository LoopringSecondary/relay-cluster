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

package types

import "github.com/ethereum/go-ethereum/common"

type FlexCancelType uint8

const (
	FLEX_CANCEL_BY_HASH   FlexCancelType = 1
	FLEX_CANCEL_BY_OWNER  FlexCancelType = 2
	FLEX_CANCEL_BY_TIME   FlexCancelType = 3
	FLEX_CANCEL_BY_MARKET FlexCancelType = 4
)

type FlexCancelOrderEvent struct {
	Owner      common.Address `json:"owner"`
	OrderHash  common.Hash    `json:"order_hash"`
	CutoffTime int64          `json:"cutoff_time"`
	TokenS     common.Address `json:"token_s"`
	TokenB     common.Address `json:"token_b"`
	Type       FlexCancelType `json:"type"`
}
