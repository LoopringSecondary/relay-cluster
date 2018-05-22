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

package accessor

import (
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	"github.com/ethereum/go-ethereum/rpc"
)

type BatchReq interface {
	ToBatchElem() []rpc.BatchElem
	FromBatchElem(batchElems []rpc.BatchElem)
}

type BatchTransactionReq struct {
	TxHash    string
	TxContent ethtyp.Transaction
	Err       error
}

type BatchTransactionRecipientReq struct {
	TxHash    string
	TxContent ethtyp.TransactionReceipt
	Err       error
}
