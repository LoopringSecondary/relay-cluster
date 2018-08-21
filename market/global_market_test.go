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

package market_test

import (
	//"github.com/Loopring/relay-cluster/test"
	"testing"

	//"time"
	"math/big"
)

func TestGlobalMarket_Sign(t *testing.T) {

	//fmt.Println("11111")
	////test.LoadConfig()
	////marketutil.Initialize(&globalConfig.Market)
	//
	//cache.NewCache(redis.RedisOptions{Host: "13.112.62.24", Port: "6379", Password: "", IdleTimeout: 20, MaxIdle: 50, MaxActive: 50})
	//
	//config := market.MyTokenConfig{}
	//config.AppId = "83ga_-yxA_yKiFyL"
	//config.AppSecret = "glQVQRP8ro-QRN59CpXj12TzwgJ1rM8w"
	//config.BaseUrl = "https://open.api.mytoken.io/"
	//
	//g := market.NewGlobalMarket(config)
	//fmt.Println(g)
	////g.Start()
	//
	////fmt.Println("12344")
	//
	//req := market.GlobalTrendReq{}
	//req.TrendAnchor = "USDT"
	////req.Symbol = "LRC"
	////fmt.Println(g.Sign(req))
	////fmt.Println(g.GetGlobalTicker("LRC"))
	////fmt.Println("111112222")
	////time.Sleep(50 * time.Second)
	////fmt.Println("111112222333")
	////fmt.Println(market.GM.GetGlobalMarketTickerCache(""))
	////fmt.Println(market.GM.GetGlobalTickerCache("LRC"))
	////fmt.Println(market.GM.GetGlobalTrendCache("LRC"))
	//fmt.Println(g.GetGlobalTicker("vite"))
	//time.Sleep(30* time.Second)

	amountS1 := new(big.Int)
	amountS1.SetString("311024990000000000", 10)
	amountS2 := new(big.Int)
	amountS2.SetString("1982724182260000256000", 10)
	rate1 := new(big.Rat)
	rate2 := new(big.Rat)

	amountB1 := new(big.Int)
	amountB1.SetString("1019887854140000000000", 10)
	rate1.Quo(new(big.Rat).SetInt(amountS1), new(big.Rat).SetInt(amountB1))
	amountB2 := new(big.Int)
	amountB2.SetString("604651569999999897", 10)
	//amountB2.Add(amountB2, new(big.Int).SetInt64(1000))
	amountS := new(big.Int)
	amountS.Mul(amountS1, amountS2)
	amountB := new(big.Int)
	amountB.Mul(amountB1, amountB2)
	rate2.Quo(new(big.Rat).SetInt(amountB2), new(big.Rat).SetInt(amountS2))

	rate2.SetString("0.00030496")
	as := new(big.Rat)
	as.SetString("311024990000000000")
	ab := new(big.Rat)
	ab.Quo(as, rate2)
	ab1 := new(big.Rat)
	ab1.SetString(ab.FloatString(0))
	rate := new(big.Rat)
	rate.Quo(as, ab1)
	println(ab1.FloatString(0), as.FloatString(0), rate.FloatString(20))

	//616676792229681894751551023580000000000
	//616676768960174757022397440000000000000
	println(amountS.Cmp(amountB), amountS.String(), amountB.String(), rate1.FloatString(20), rate2.FloatString(20))

}
