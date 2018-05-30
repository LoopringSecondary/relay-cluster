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

package dao

// order amountS 上限1e30

type OrderTransaction struct {
	ID        int    `gorm:"column:id;primary_key;"`
	Owner     string `gorm:"column:owner;type:varchar(42)"`
	OrderHash string `gorm:"column:order_hash;type:varchar(82)"`
	TxHash    string `gorm:"column:tx_hash;type:varchar(82)"`
	Status    uint8  `gorm:"column:status;type:tinyint(4)"`
	Nonce     int64  `gorm:"column:nonce;type:bigint"`
}
