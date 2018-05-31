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

import (
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

// txhash唯一索引
type OrderTransaction struct {
	ID        int    `gorm:"column:id;primary_key;"`
	Owner     string `gorm:"column:owner;type:varchar(42)"`
	OrderHash string `gorm:"column:order_hash;type:varchar(82)"`
	TxHash    string `gorm:"column:tx_hash;type:varchar(82)"`
	Status    uint8  `gorm:"column:status;type:tinyint(4)"`
	Nonce     int64  `gorm:"column:nonce;type:bigint"`
}

// convert dao/fill to types/fill
func (f *OrderTransaction) ConvertUp(orderhash common.Hash, orderstatus types.OrderStatus, txinfo types.TxInfo) error {
	f.OrderHash = orderhash.Hex()
	f.Status = uint8(orderstatus)
	f.TxHash = txinfo.TxHash.Hex()
	f.Owner = txinfo.From.Hex()
	f.Nonce = txinfo.Nonce.Int64()

	return nil
}

func (s *RdsService) MaxNonce(owner common.Address, orderhash common.Hash) (int64, error) {
	var nonce int64
	err := s.Db.Where("owner=?", owner.Hex()).Where("order_hash=?", orderhash.Hex()).Pluck("max(nonce)", &nonce).Error
	return nonce, err
}

func (s *RdsService) GetOrderTxList(orderhash common.Hash) ([]OrderTransaction, error) {
	var list []OrderTransaction
	err := s.Db.Where("order_hash=?", orderhash.Hex()).Find(&list).Order("nonce desc").Error
	return list, err
}
