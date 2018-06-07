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
	"math/big"
)

// todo(fuk): rename table
type CutOffEvent struct {
	ID              int    `gorm:"column:id;primary_key;"`
	Protocol        string `gorm:"column:contract_address;type:varchar(42)"`
	DelegateAddress string `gorm:"column:delegate_address;type:varchar(42)"`
	Owner           string `gorm:"column:owner;type:varchar(42)"`
	TxHash          string `gorm:"column:tx_hash;type:varchar(82)"`
	OrderHashList   string `gorm:"column:order_hash_list;type:text"`
	BlockNumber     int64  `gorm:"column:block_number"`
	Cutoff          int64  `gorm:"column:cutoff"`
	LogIndex        int64  `gorm:"column:log_index"`
	Status          uint8  `gorm:"column:status"`
	Fork            bool   `gorm:"column:fork"`
	CreateTime      int64  `gorm:"column:create_time"`
}

// convert types/cutoffEvent to dao/CancelEvent
func (e *CutOffEvent) ConvertDown(src *types.CutoffEvent) error {
	e.Owner = src.Owner.Hex()
	e.Protocol = src.Protocol.Hex()
	e.DelegateAddress = src.DelegateAddress.Hex()
	e.TxHash = src.TxHash.Hex()
	e.Cutoff = src.Cutoff.Int64()
	e.LogIndex = src.TxLogIndex
	e.BlockNumber = src.BlockNumber.Int64()
	e.CreateTime = src.BlockTime
	e.Status = uint8(src.Status)
	e.OrderHashList = MarshalHashListToStr(src.OrderHashList)

	return nil
}

// convert dao/cutoffEvent to types/cutoffEvent
func (e *CutOffEvent) ConvertUp(dst *types.CutoffEvent) error {
	dst.Owner = common.HexToAddress(e.Owner)
	dst.Protocol = common.HexToAddress(e.Protocol)
	dst.DelegateAddress = common.HexToAddress(e.DelegateAddress)
	dst.TxHash = common.HexToHash(e.TxHash)
	dst.BlockNumber = big.NewInt(e.BlockNumber)
	dst.TxLogIndex = e.LogIndex
	dst.Cutoff = big.NewInt(e.Cutoff)
	dst.BlockTime = e.CreateTime
	dst.Status = types.TxStatus(e.Status)
	dst.OrderHashList = UnmarshalStrToHashList(e.OrderHashList)

	return nil
}

func (s *RdsService) GetCutoffEvent(txhash common.Hash) (CutOffEvent, error) {
	var event CutOffEvent
	err := s.Db.Where("tx_hash=?", txhash.Hex()).Where("fork=?", false).First(&event).Error
	return event, err
}

func (s *RdsService) GetCutoffForkEvents(from, to int64) ([]CutOffEvent, error) {
	var (
		list []CutOffEvent
		err  error
	)

	err = s.Db.Where("block_number > ? and block_number <= ?", from, to).
		Where("fork=?", false).
		Find(&list).Error

	return list, err
}

func (s *RdsService) RollBackCutoff(from, to int64) error {
	return s.Db.Model(&CutOffEvent{}).Where("block_number > ? and block_number <= ?", from, to).Update("fork", true).Error
}
