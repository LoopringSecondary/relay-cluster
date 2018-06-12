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
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

const (
	MaxNoncePrefix     = "txm_nonce_max_"
	SuccessNoncePrefix = "txm_nonce_success_"
	NonceTtl           = 86400 // todo 临时数据,只存储10分钟,系统性宕机后无法重启后丢失?
)

func GetMaxNonceValue(owner common.Address) (*big.Int, error) {
	key := generateNonceKey(owner)
	if bs, err := cache.Get(key); err == nil {
		return new(big.Int).SetBytes(bs), nil
	}

	nonce, err := rds.GetMaxNonce(owner)
	if err != nil {
		var result types.Big
		if err := accessor.GetTransactionCount(&result, owner, "latest"); err != nil {
			return big.NewInt(0), err
		}
		nonce = result.BigInt()
	}

	bs := nonce.Bytes()
	cache.Set(key, bs, NonceTtl)

	return nonce, nil
}

func GetSuccessNonceValue(owner common.Address) (*big.Int, error) {
	key := generateSuccessNonceKey(owner)
	if bs, err := cache.Get(key); err == nil {
		return new(big.Int).SetBytes(bs), nil
	}

	nonce, err := rds.GetMaxSuccessNonce(owner)
	if err != nil {
		var result types.Big
		if err := accessor.GetTransactionCount(&result, owner, "latest"); err != nil {
			return big.NewInt(0), err
		}
		nonce = result.BigInt()
	}

	bs := nonce.Bytes()
	cache.Set(key, bs, NonceTtl)

	return nonce, nil
}

func generateNonceKey(owner common.Address) string {
	return MaxNoncePrefix + owner.Hex()
}

func generateSuccessNonceKey(owner common.Address) string {
	return SuccessNoncePrefix + owner.Hex()
}
