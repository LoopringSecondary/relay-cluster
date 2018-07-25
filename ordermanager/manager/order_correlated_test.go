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

package manager_test

import (
	"github.com/Loopring/relay-cluster/ordermanager/manager"
	"github.com/Loopring/relay-cluster/test"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestOrderTxHandler_HandlerOrderCorrelatedTx(t *testing.T) {
	test.GenerateOrderManager()

	txinfo := types.TxInfo{
		Protocol:        common.HexToAddress("0xcd36128815ebe0b44d0374649bad2721b8751bef"),
		DelegateAddress: types.NilAddress,
		From:            common.HexToAddress("0xb1018949b241D76A1AB2094f473E9bEfeAbB5Ead"),
		To:              common.HexToAddress("0xcd36128815ebe0b44d0374649bad2721b8751bef"),
		BlockHash:       common.HexToHash("0x36465444dbec326cf815973fc3064bce9c1f7ec22631d69462dea396cdadd730"),
		BlockNumber:     big.NewInt(43170),
		BlockTime:       1528950468,
		TxHash:          common.HexToHash("0x8f61c0913a96116b26b73fe05b099e2b2f54803ed15b3a4ca053ee5c2d2dd158"),
		TxIndex:         1,
		TxLogIndex:      1,
		Value:           big.NewInt(0),
		Status:          types.TX_STATUS_FAILED,
		Nonce:           big.NewInt(366),
	}
	handler := manager.BaseOrderTxHandler(txinfo)
	if err := handler.HandlerOrderCorrelatedTx(); err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Logf("success")
	}
}
