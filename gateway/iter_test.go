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

package gateway

import (
	"github.com/Loopring/relay-cluster/test"
	"github.com/Loopring/relay-lib/eth/accessor"
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	"github.com/Loopring/relay-lib/kafka"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"math/big"
	"testing"
	"time"
)

const (
	suffix = "0000000000000000" //0.01
)

func TestRing(t *testing.T) {
	entity := test.Entity()

	lrc := util.AllTokens["LRC"].Protocol
	eth := util.AllTokens["WETH"].Protocol

	account1 := entity.Accounts[0]
	account2 := entity.Accounts[1]

	// 卖出0.1个eth， 买入300个lrc,lrcFee为20个lrc
	amountS1, _ := new(big.Int).SetString("10"+suffix, 0)
	amountB1, _ := new(big.Int).SetString("30000"+suffix, 0)
	lrcFee1 := new(big.Int).Mul(big.NewInt(1e18), big.NewInt(5)) // 20个lrc
	test.CreateOrder(eth, lrc, account1.Address, amountS1, amountB1, lrcFee1)

	// 卖出1000个lrc,买入0.1个eth,lrcFee为20个lrc
	amountS2, _ := new(big.Int).SetString("30000"+suffix, 0)
	amountB2, _ := new(big.Int).SetString("10"+suffix, 0)
	lrcFee2 := new(big.Int).Mul(big.NewInt(1e18), big.NewInt(3))
	test.CreateOrder(lrc, eth, account2.Address, amountS2, amountB2, lrcFee2)
}

func TestUnlockWallet(t *testing.T) {
	manager := test.GenerateAccountManager()
	accounts := []string{
		test.Entity().Accounts[0].Address.Hex(),
		test.Entity().Accounts[1].Address.Hex(),
		test.Entity().Creator.Address.Hex(),
	}
	for _, account := range accounts {
		manager.UnlockedWallet(account)
	}
	time.Sleep(1 * time.Second)
}

func TestPendingTransaction(t *testing.T) {
	var tx ethtyp.Transaction
	if err := accessor.GetTransactionByHash(&tx, "0x4cd20fa3854420f497d3d943469becbef14f368be46d65e955aad6f34bb0d30d", "latest"); err != nil {
		t.Fatalf(err.Error())
	} else {
		producer.SendMessage(kafka.Kafka_Topic_Extractor_PendingTransaction, &tx, "extractor")
	}
}

var producer = &kafka.MessageProducer{}

func init() {
	if err := producer.Initialize(test.Cfg().Kafka.Brokers); err != nil {
		log.Fatal(err.Error())
	}
}

func TestPrepare(t *testing.T) {
	test.PrepareTestData()
}
