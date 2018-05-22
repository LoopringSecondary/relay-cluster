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

package contract

import (
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// sateHash contract bytes32 to common.hash
func safeHash(bytes [32]uint8) common.Hash {
	return common.Hash(bytes)
}

// safeAbsBig return abs value and isNeg
func safeBig(bytes [32]uint8) *big.Int {
	num := new(big.Int).SetBytes(bytes[:])
	if bytes[0] > uint8(128) {
		num.Xor(types.MaxUint256, num)
		num.Not(num)
	}
	return num
}

// safeAddress contract bytes32 to common.address
func safeAddress(bytes [32]uint8) common.Address {
	var newbytes []byte = bytes[0:]
	return common.BytesToAddress(newbytes)
}
