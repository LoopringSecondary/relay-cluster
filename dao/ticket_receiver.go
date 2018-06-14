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
	ID      int    `gorm:"column:id;primary_key;" json:"id"`
	Name    string `gorm:"column:name;type:varchar(42)" json:"name"`
	Email   string `gorm:"column:email;type:varchar(128)" json:"email"`
	Phone   string `gorm:"column:phone;type:varchar(128)" json:"phone"`
	Address string `gorm:"column:address;type:varchar(128);unique_index" json:"address"`
}

func (s *RdsService) QueryTicketByAddress(address string) (ticket TicketReceiver, err error) {
	ticket = TicketReceiver{}
	err = s.Db.Model(&TicketReceiver{}).Where("address = ? ", address).Find(&ticket).Error
	return ticket, err
}

func (s *RdsService) TicketCount() (count int, err error) {
	err = s.Db.Model(&TicketReceiver{}).Count(&count).Error
	return count, err
}
