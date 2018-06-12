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

package cache

import (
	"encoding/json"
	"fmt"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-lib/cache"
	"math/big"
)

// 该缓存模块用于批量获取transactionEntity缓存

const (
	TxEntityPrefix = "txm_entity_" // txm_entity_blocknumber_txhash_logIndex,不用hash结构,避免不同用户数据在同一个key的情况
	TxEntityTtl    = 864000
)

func RollbackEntityCache(from, to int64) error {
	if from+1 > to {
		return fmt.Errorf("rollbackCache error, from + 1 > to")
	}

	for i := from + 1; i <= to; i++ {
		format := generateTxEntityBlockFormat(i)
		keysbytes, err := cache.Keys(format)
		if err != nil {
			continue
		}
		var keys []string
		for _, v := range keysbytes {
			keys = append(keys, string(v))
		}
		if err := cache.Dels(keys); err != nil {
			return err
		}
	}

	return nil
}

// 根据transactionView数据批量获取缓存
func GetEntityCache(views []dao.TransactionView) TransactionEntityMap {
	var (
		uncachedTxHashList []string
		entityMap          = make(TransactionEntityMap)
	)

	// get entity already exist in cache
	for _, v := range views {
		key := generateTxEntityKey(v.TxHash, v.BlockNumber, v.LogIndex)

		bs, err := cache.Get(key)
		if err != nil {
			uncachedTxHashList = append(uncachedTxHashList, v.TxHash)
			continue
		}

		var entity dao.TransactionEntity
		if err := json.Unmarshal(bs, &entity); err == nil {
			entityMap.SaveEntity(entity)
		}
	}

	// get entity from db
	models, _ := rds.GetTxEntity(uncachedTxHashList)
	if len(models) == 0 {
		return entityMap
	}

	// save entity in cache
	for _, model := range models {
		for _, v := range views {
			if _, ok := entityMap.GetEntity(v.TxHash, v.LogIndex); !ok {
				if bs, err := json.Marshal(&model); err == nil {
					key := generateTxEntityKey(model.TxHash, model.BlockNumber, model.LogIndex)
					cache.Set(key, bs, TxEntityTtl)
				}
				entityMap.SaveEntity(model)
			}
		}
	}

	return entityMap
}

type TransactionEntityMap map[string]map[int64]dao.TransactionEntity

func (m TransactionEntityMap) SaveEntity(entity dao.TransactionEntity) {
	if _, ok := m[entity.TxHash]; !ok {
		m[entity.TxHash] = make(map[int64]dao.TransactionEntity)
	}
	if _, ok := m[entity.TxHash][entity.LogIndex]; !ok {
		m[entity.TxHash][entity.LogIndex] = entity
	}
}

func (m TransactionEntityMap) GetEntity(txhash string, logindex int64) (dao.TransactionEntity, bool) {
	var (
		entity dao.TransactionEntity
		ok     bool
	)

	txs, logsok := m[txhash]
	if !logsok {
		return entity, false
	}

	entity, ok = txs[logindex]
	return entity, ok
}

func generateTxEntityKey(txhash string, blockNumber, logIndex int64) string {
	blockStr := big.NewInt(blockNumber).String()
	logIdxStr := big.NewInt(logIndex).String()
	return TxEntityPrefix + blockStr + "_" + txhash + "_" + logIdxStr
}

func generateTxEntityBlockFormat(blockNumber int64) string {
	blockStr := big.NewInt(blockNumber).String()
	return TxEntityPrefix + blockStr + "_*"
}
