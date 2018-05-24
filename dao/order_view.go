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

// 用于前端查询

type OrderView struct {
	ID              int     `gorm:"column:id;primary_key;"`
	OrderId         int     `gorm:"column:order_id"`
	Protocol        string  `gorm:"column:protocol;type:varchar(42)"`
	DelegateAddress string  `gorm:"column:delegate_address;type:varchar(42)"`
	Owner           string  `gorm:"column:owner;type:varchar(42)"`
	OrderHash       string  `gorm:"column:order_hash;type:varchar(82)"`
	TokenS          string  `gorm:"column:token_s;type:varchar(42)"`
	TokenB          string  `gorm:"column:token_b;type:varchar(42)"`
	CreateTime      int64   `gorm:"column:create_time;type:bigint"`
	Price           float64 `gorm:"column:price;type:decimal(28,16);"`
	Status          uint8   `gorm:"column:status;type:tinyint(4)"`
	BroadcastTime   int     `gorm:"column:broadcast_time;type:bigint"`
	Market          string  `gorm:"column:market;type:varchar(40)"`
	Side            string  `gorm:"column:side;type:varchar(40)"`
	OrderType       string  `gorm:"column:order_type;type:varchar(40)"`
}
