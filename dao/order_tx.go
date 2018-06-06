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
	OrderStatus uint8  `gorm:"column:order_status;type:tinyint(4)"`
	TxHash      string `gorm:"column:tx_hash;type:varchar(82)"`
	TxStatus    uint8  `gorm:"column:tx_status;type:tinyint(4)"`
	Nonce       int64  `gorm:"column:nonce;type:bigint"`
	BlockNumber int64  `gorm:"column:block_number"`
	Fork        bool   `gorm:"column:fork"`
}

// convert types/orderTxRecord to dao/ordertx
func (tx *OrderTransaction) ConvertDown(src *omtyp.OrderTx) error {
	tx.OrderHash = src.OrderHash.Hex()
	tx.OrderStatus = uint8(src.OrderStatus)
	tx.TxHash = src.TxHash.Hex()
	tx.Owner = src.Owner.Hex()
	tx.Nonce = src.Nonce
	tx.TxStatus = uint8(src.TxStatus)
	tx.BlockNumber = src.BlockNumber

	return nil
}

func (tx *OrderTransaction) ConvertUp(dst *omtyp.OrderTx) error {
	dst.OrderHash = common.HexToHash(tx.OrderHash)
	dst.TxHash = common.HexToHash(tx.TxHash)
	dst.Owner = common.HexToAddress(tx.Owner)
	dst.OrderStatus = types.OrderStatus(tx.OrderStatus)
	dst.Nonce = tx.Nonce
	dst.TxStatus = types.TxStatus(tx.TxStatus)
	dst.BlockNumber = tx.BlockNumber

	return nil
}

func (s *RdsService) FindOrderTx(txhash, orderhash common.Hash) (*OrderTransaction, error) {
	var (
		tx  OrderTransaction
		err error
	)

	err = s.Db.Where("tx_hash=?", txhash.Hex()).
		Where("order_hash=?", orderhash.Hex()).
		Where("fork=?", false).
		First(&tx).Error
	return &tx, err
}

func (s *RdsService) GetPendingOrderTxByOwner(owner common.Address) ([]OrderTransaction, error) {
	var (
		list []OrderTransaction
		err  error
	)

	err = s.Db.Where("owner=?", owner.Hex()).
		Where("tx_status=?", types.TX_STATUS_PENDING).
		Where("fork=?", false).
		Find(&list).Error

	return list, err
}

func (s *RdsService) SetPendingOrderTxFailed(owner common.Address, txhash common.Hash, nonce int64) int64 {
	num := s.Db.Model(&OrderTransaction{}).
		Where("owner=?", owner.Hex()).
		Where("tx_hash<>?", txhash.Hex()).
		Where("nonce<=", nonce).
		Where("fork=?", false).
		Update("tx_status", types.TX_STATUS_FAILED).RowsAffected
	return num
}

func (s *RdsService) UpdateOrderTxStatus(txhash common.Hash, blockNumber int64, status types.TxStatus) error {
	return s.Db.Model(&OrderTransaction{}).
		Where("tx_hash=?", txhash.Hex()).
		Where("fork=?", false).
		Update("tx_status", status).
		Update("block_number", blockNumber).Error
}

// 分叉不管之前的pending记录
func (s *RdsService) RollBackOrderTx(from, to int64) error {
	return s.Db.Model(&OrderTransaction{}).
		Where("block_number > ? and block_number <= ?", from, to).
		Update("fork", true).Error
}
