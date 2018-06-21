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
	"github.com/Loopring/relay-lib/kafka"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestFlexCancelOrder(t *testing.T) {
	brokers := test.Cfg().Kafka.Brokers
	producer := &kafka.MessageProducer{}
	if err := producer.Initialize(brokers); err != nil {
		t.Fatalf(err.Error())
	}

	key := "ordermanager_test"

	data := &types.FlexCancelOrderEvent{
		Owner:      common.HexToAddress("0x1B978a1D302335a6F2Ebe4B8823B5E17c3C84135"),
		OrderHash:  common.HexToHash("0x6186ae3094494d59bc52340e231e04be40c63b3750ddb5f3022c00c8c126c414"),
		CutoffTime: 0,
		TokenS:     types.NilAddress,
		TokenB:     types.NilAddress,
		Type:       types.FLEX_CANCEL_BY_HASH,
	}

	producer.SendMessage(kafka.Kafka_Topic_OrderManager_FlexCancelOrder, data, key)
}

func TestNewOrderEntity(t *testing.T) {
	entity := test.Entity()

	lrc := util.AllTokens["LRC"].Protocol
	eth := util.AllTokens["WETH"].Protocol

	account := entity.Accounts[0]

	// 卖出0.1个eth， 买入300个lrc,lrcFee为20个lrc
	amountS1, _ := new(big.Int).SetString("100000000000000000", 0)
	amountB1, _ := new(big.Int).SetString("300000000000000000000", 0)
	lrcFee1 := new(big.Int).Mul(big.NewInt(1e18), big.NewInt(5)) // 20个lrc
	order := test.CreateOrder(eth, lrc, account.Address, amountS1, amountB1, lrcFee1)
	state := &types.OrderState{RawOrder: order}
	if entity, err := manager.NewOrderEntity(state, nil); err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Logf("dealtAmountS:%s, dealtAmountB:%s, cancelAmountS:%s, cancelAmountB:%s", entity.DealtAmountS, entity.DealtAmountB, entity.CancelledAmountS, entity.CancelledAmountB)
	}

}
