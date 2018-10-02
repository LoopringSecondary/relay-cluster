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
	"encoding/json"
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

func TestOrderFilled(t *testing.T) {
	jsonstr := `{"protocol":"0x8d8812b72d1e4ffcec158d25f56748b7d67c1e78","delegate":"0x17233e07c67d086464fd408148c3abb56245fa64","from":"0x5552dcfba48c94544beaaf26470df9898e050ac2","to":"0x8d8812b72d1e4ffcec158d25f56748b7d67c1e78","block_hash":"0xe8182201bb26aecbd3307be1047026c0bee0b7c8282f2745ec5aefb00839ae4d","block_number":6437994,"block_time":1538461276,"tx_hash":"0xd90c70eac06ff3f92cf93513afffe095f5b567d203b6bed4f83eb10c855f6c06","tx_index":19,"tx_log_index":40,"value":0,"status":2,"gas_limit":400000,"gas_used":347235,"gas_price":16418026158,"nonce":5475,"identify":"submitRing","ringhash":"0x8de312a3706e2e45bac048feda845a8e62fd431bfb9d0161964449fd067b4a56","pre_order_hash":"0x827945f687d4105f179e371e3be12f6729058ed5bf481d53a561175cb3259da6","order_hash":"0x1e90239afe8f83fcc785a84ebcccb030f33f21b9724f67e32e07c7136adbe341","next_order_hash":"0x827945f687d4105f179e371e3be12f6729058ed5bf481d53a561175cb3259da6","owner":"0x1b80346b14c766f7d17d2d8959206bce7c97d196","token_s":"0x1b793e49237758dbd8b752afc9eb4b329d5da016","token_b":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","sell_to":"0xa0a8972ea0c20ee201e36d283b2f13fa87af2e93","buy_from":"0xa0a8972ea0c20ee201e36d283b2f13fa87af2e93","ring_index":7104,"amount_s":300000000000000000000,"amount_b":58199999999999999,"lrc_reward":0,"lrc_fee":5000000000000000000,"split_s":0,"split_b":0,"market":"","fill_index":1}`
	fill := types.OrderFilledEvent{}
	if err := json.Unmarshal([]byte(jsonstr), &fill); err != nil {
		t.Fatalf(err.Error())
	}
	om := test.GenerateOrderManager()
	om.HandlerOrderRelatedEvent(&fill)
}
