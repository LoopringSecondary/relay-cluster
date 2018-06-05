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
	omtyp "github.com/Loopring/relay-cluster/ordermanager/types"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

// txhash唯一索引
type OrderTransaction struct {
	ID          int    `gorm:"column:id;primary_key;"`
	Owner       string `gorm:"column:owner;type:varchar(42)"`
	OrderHash   string `gorm:"column:order_hash;type:varchar(82)"`
	TxHash      string `gorm:"column:tx_hash;type:varchar(82)"`
	OrderStatus uint8  `gorm:"column:order_status;type:tinyint(4)"`
	Nonce       int64  `gorm:"column:nonce;type:bigint"`
}

// convert types/orderTxRecord to dao/ordertx
func (tx *OrderTransaction) ConvertDown(src *omtyp.OrderRelatedPendingTx) error {
	tx.OrderHash = src.OrderHash.Hex()
	tx.OrderStatus = uint8(src.OrderStatus)
	tx.TxHash = src.TxHash.Hex()
	tx.Owner = src.Owner.Hex()
	tx.Nonce = src.Nonce

	return nil
}

func (tx *OrderTransaction) ConvertUp(dst *omtyp.OrderRelatedPendingTx) error {
	dst.OrderHash = common.HexToHash(tx.OrderHash)
	dst.TxHash = common.HexToHash(tx.TxHash)
	dst.Owner = common.HexToAddress(tx.Owner)
	dst.OrderStatus = types.OrderStatus(tx.OrderStatus)
	dst.Nonce = tx.Nonce

	return nil
}

func (s *RdsService) GetOrderTx(orderhash, txhash common.Hash) (*OrderTransaction, error) {
	var (
		tx  OrderTransaction
		err error
	)

	err = s.Db.Where("order_hash=?", orderhash.Hex()).Where("tx_hash=?", txhash.Hex()).First(&tx).Error
	return &tx, err
}

func (s *RdsService) GetPendingOrderTxs(owner common.Address, pendingstatus []types.OrderStatus) ([]OrderTransaction, error) {
	var list []OrderTransaction
	err := s.Db.Where("owner=?", owner.Hex()).Where("order_status in (?)", pendingstatus).Find(&list).Error
	return list, err
}
