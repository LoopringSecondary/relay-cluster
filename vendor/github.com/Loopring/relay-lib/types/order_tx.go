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

//go:generate gencodec -type OrderTxRecord -out gen_orderTxRecord_json.go
type OrderTxRecord struct {
	Owner       common.Address `json:"owner"`
	TxHash      common.Hash    `json:"tx_hash"`
	OrderHash   common.Hash    `json:"order_hash"`
	OrderStatus OrderStatus    `json:"order_status"`
	Nonce       int64          `json:"nonce"`
}
