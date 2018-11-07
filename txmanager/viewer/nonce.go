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

package viewer

import (
	"github.com/Loopring/relay-cluster/txmanager/cache"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// 该模块用于管理用户nonce.这里我们不区分tx状态,持续记录用户
// 同时,我们将该nonce作为一个中心化的数据,如果在其他钱包拥有高于本模块的处于pending的nonce,该模块是无法处理的

// 返回用户处于success/pending状态下的nonce值
func (impl *TransactionViewerImpl) GetNonce(ownerStr string) (*big.Int, error) {
	ownerStr = safeOwner(ownerStr)
	owner := common.HexToAddress(ownerStr)

	if nonce, err := cache.GetMaxNonceValue(owner); err != nil {
		return big.NewInt(0), ErrNonceNotExist
	} else if nonce.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0), nil
	} else {
		return new(big.Int).Add(nonce, big.NewInt(1)), nil
	}
}

// 校验用户nonce是否可用,用户提交的pending tx nonce值不允许小于已经成功了的tx nonce值
func (impl *TransactionViewerImpl) ValidateNonce(ownerStr string, nonce *big.Int) error {
	if nonce.Cmp(big.NewInt(0)) < 0 {
		return ErrNonceInvalid
	}

	ownerStr = safeOwner(ownerStr)
	owner := common.HexToAddress(ownerStr)
	successNonce, err := cache.GetTxMinedMaxNonceValue(owner)
	if err != nil {
		return ErrNonceNotExist
	}
	if nonce.Cmp(successNonce) < 1 {
		return ErrNonceInvalid
	}

	return nil
}
