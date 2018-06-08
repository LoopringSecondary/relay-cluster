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

type OrderPendingTransaction struct {
	ID          int    `gorm:"column:id;primary_key;"`
	Owner       string `gorm:"column:owner;type:varchar(42)"`
	OrderHash   string `gorm:"column:order_hash;type:varchar(82)"`
	OrderStatus uint8  `gorm:"column:order_status;type:tinyint(4)"`
	TxHash      string `gorm:"column:tx_hash;type:varchar(82)"`
	Nonce       int64  `gorm:"column:nonce;type:bigint"`
}

// convert types/orderTxRecord to dao/ordertx
func (tx *OrderPendingTransaction) ConvertDown(src *omtyp.OrderTx) error {
	tx.OrderHash = src.OrderHash.Hex()
	tx.OrderStatus = uint8(src.OrderStatus)
	tx.TxHash = src.TxHash.Hex()
	tx.Owner = src.Owner.Hex()
	tx.Nonce = src.Nonce

	return nil
}

func (tx *OrderPendingTransaction) ConvertUp(dst *omtyp.OrderTx) error {
	dst.OrderHash = common.HexToHash(tx.OrderHash)
	dst.TxHash = common.HexToHash(tx.TxHash)
	dst.Owner = common.HexToAddress(tx.Owner)
	dst.OrderStatus = types.OrderStatus(tx.OrderStatus)
	dst.Nonce = tx.Nonce

	return nil
}

func (s *RdsService) FindPendingOrderTx(txhash, orderhash common.Hash) (*OrderPendingTransaction, error) {
	var tx OrderPendingTransaction
	err := s.Db.Where("tx_hash=?", txhash.Hex()).Where("order_hash=?", orderhash.Hex()).First(&tx).Error
	return &tx, err
}

func (s *RdsService) GetPendingOrderTxs(owner common.Address) ([]OrderPendingTransaction, error) {
	var list []OrderPendingTransaction
	err := s.Db.Where("owner=?", owner.Hex()).Find(&list).Error
	return list, err
}

func (s *RdsService) GetPendingOrderTx(owner common.Address, orderhash common.Hash) ([]OrderPendingTransaction, error) {
	var list []OrderPendingTransaction
	err := s.Db.Where("owner=?", owner.Hex()).Where("order_hash=?", orderhash.Hex()).Find(&list).Error
	return list, err
}

func (s *RdsService) DelPendingOrderTx(owner common.Address, orderhash common.Hash, txhashlist []common.Hash) int64 {
	if len(txhashlist) == 0 {
		return 0
	}

	var list []string

	for _, v := range txhashlist {
		list = append(list, v.Hex())
	}

	return s.Db.Model(&OrderPendingTransaction{}).
		Where("owner=?", owner.Hash()).
		Where("order_hash=?", orderhash.Hex()).
		Where("tx_hash in (?)", list).
		RowsAffected
}
