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

package accountmanager_test

import (
	"testing"
	"github.com/Loopring/relay-cluster/accountmanager"
	"go.uber.org/zap"
	"encoding/json"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/cache/redis"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/ethereum/go-ethereum/common"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/zklock"
)

func init() {
	logConfig := `{
	  "level": "debug",
	  "development": false,
	  "encoding": "json",
	  "outputPaths": ["stdout"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`
	rawJSON := []byte(logConfig)

	var (
		cfg zap.Config
		err error
	)
	if err = json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	log.Initialize(cfg)
	cache.NewCache(redis.RedisOptions{Host: "127.0.0.1", Port: "6379"})

	accessor.Initialize(accessor.AccessorOptions{RawUrls: []string{"http://13.230.23.98:8545"}})

	options := loopringaccessor.LoopringProtocolOptions{}
	options.Address = make(map[string]string)
	options.Address["1.5"] = "0x8d8812b72d1e4ffCeC158D25f56748b7d67c1e78"
	loopringaccessor.InitLoopringAccessor(options)

	utiloptions := marketutil.MarketOptions{}
	utiloptions.TokenFile = "/Users/yuhongyu/Desktop/service/go/src/github.com/Loopring/relay/config/tokens.json"
	marketutil.Initialize(&utiloptions)

	zklock.Initialize(zklock.ZkLockConfig{ZkServers: "127.0.0.1:2181", ConnectTimeOut: 10000})
	acctOptions := &accountmanager.AccountManagerOptions{}
	//options.CacheDuration =
	accountmanager.Initialize(acctOptions)

}

func TestAccountManager_GetBalance(t *testing.T) {

	balances, err := accountmanager.GetBalanceWithSymbolResult(common.HexToAddress("0x634211e61Ac19baF71153d193b3482d12d669c5C"))
	if nil != err {
		t.Fatalf("err:%s", err.Error())
	}
	for k, v := range balances {
		t.Logf("token:%s, balance:%s", k, v.String())
	}
}

func TestAccountManager_UnlockedWallet(t *testing.T) {
	//owner := common.HexToAddress("0x750ad4351bb728cec7d639a9511f9d6488f1e259")
	//data := append(append(owner.Bytes(), owner.Bytes()...), owner.Bytes()...)
	//t.Log(common.BytesToAddress(data[0:20]).Hex(), common.BytesToAddress(data[20:40]).Hex(), common.BytesToAddress(data[40:]).Hex())
	//accManager := test.GenerateAccountManager()
	//
	//if err := accManager.UnlockedWallet("0x750ad4351bb728cec7d639a9511f9d6488f1e259"); nil != err {
	//	t.Errorf("err:%s", err.Error())
	//} else {
	//	t.Log("##ooooo")
	//}
	//balances,err := accManager.GetBalanceWithSymbolResult(common.HexToAddress("0x750ad4351bb728cec7d639a9511f9d6488f1e259"))
	//if nil != err {
	//	t.Fatalf("err:%s", err.Error())
	//}
	//for k,v := range balances {
	//	t.Logf("token:%s, balance:%s", k, v.Balance.String())
	//}
}
func TestAccountManager_GetBAndAllowance(t *testing.T) {
	balance, allowance, err := accountmanager.GetBalanceAndAllowance(common.HexToAddress("0x750ad4351bb728cec7d639a9511f9d6488f1e259"), common.HexToAddress("0x"), common.HexToAddress("0x"))
	if nil != err {
		t.Fatalf("err:%s", err.Error())
	}
	t.Logf("balance:%s, allowance:%s", balance.String(), allowance.String())
}

func TestAccountManager_GetAllowances(t *testing.T) {
	spender := common.HexToAddress("0x17233e07c67d086464fD408148c3ABB56245FA64")
	allowances, err := accountmanager.GetAllowanceWithSymbolResult(common.HexToAddress("0x634211e61Ac19baF71153d193b3482d12d669c5C"), spender)
	if nil != err {
		t.Fatalf("err:%s", err.Error())
	}
	for k, v := range allowances {
		t.Logf("token:%s, spener:%s, allowance:%s", k, spender.Hex(), v.String())
	}

}