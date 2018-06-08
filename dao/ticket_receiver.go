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

type TicketReceiver struct {
	ID           int    `gorm:"column:id;primary_key;"`
	Name      string `gorm:"column:name;type:varchar(42) json:"name"`
	Email   string  `gorm:"column:email;type:varchar(128)" json:"email"`
	Phone   string  `gorm:"column:phone;type:varchar(128)" json:"phone"`
	Address   string  `gorm:"column:address;type:varchar(128)" json:"address"`
	V                     uint8   `gorm:"column:v;type:tinyint(4)" json:"v"'`
	R                     string  `gorm:"column:r;type:varchar(66)" json:"r"`
	S                     string  `gorm:"column:s;type:varchar(66)" json:"s"`
}

