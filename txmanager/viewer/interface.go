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
	"errors"
	txtyp "github.com/Loopring/relay-cluster/txmanager/types"
	"math/big"
)

func GetPendingTransactions(owner string) ([]txtyp.TransactionJsonResult, error) {
	return impl.GetPendingTransactions(owner)
}
func GetTransactionsByHash(owner string, hashList []string) ([]txtyp.TransactionJsonResult, error) {
	return impl.GetTransactionsByHash(owner, hashList)
}
func GetAllTransactionCount(ownerStr, symbol, status, typ string) (int, error) {
	return impl.GetAllTransactionCount(ownerStr, symbol, status, typ)
}
func GetAllTransactions(owner, symbol, status, typ string, limit, offset int) ([]txtyp.TransactionJsonResult, error) {
	return impl.GetAllTransactions(owner, symbol, status, typ, limit, offset)
}
func GetNonce(owner string) (*big.Int, error) {
	return impl.GetNonce(owner)
}
func ValidateNonce(owner string, nonce *big.Int) error {
	return impl.ValidateNonce(owner, nonce)
}

type TransactionViewer interface {
	GetPendingTransactions(owner string) ([]txtyp.TransactionJsonResult, error)
	GetAllTransactionCount(owner, symbol, status, typ string) (int, error)
	GetAllTransactions(owner, symbol, status, typ string, limit, offset int) ([]txtyp.TransactionJsonResult, error)
	GetTransactionsByHash(owner string, hashList []string) ([]txtyp.TransactionJsonResult, error)
	GetNonce(owner string) (*big.Int, error)
	ValidateNonce(owner string, nonce *big.Int) error
}

var (
	ErrOwnerAddressInvalid error = errors.New("owner address invalid")
	ErrHashListEmpty       error = errors.New("hash list is empty")
	ErrNonTransaction      error = errors.New("no transaction found")
	ErrNonceNotExist       error = errors.New("nonce not exist")
	ErrNonceInvalid        error = errors.New("user nonce invalid")
)
