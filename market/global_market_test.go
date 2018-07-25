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
	"fmt"
	"github.com/Loopring/relay-cluster/market"
	//"github.com/Loopring/relay-cluster/test"
	"testing"
	//"time"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/cache/redis"
	//"time"
)

func TestGlobalMarket_Sign(t *testing.T) {

	fmt.Println("11111")
	//test.LoadConfig()
	//marketutil.Initialize(&globalConfig.Market)

	cache.NewCache(redis.RedisOptions{Host: "13.112.62.24", Port: "6379", Password: "", IdleTimeout: 20, MaxIdle: 50, MaxActive: 50})

	config := market.MyTokenConfig{}
	config.AppId = "83ga_-yxA_yKiFyL"
	config.AppSecret = "glQVQRP8ro-QRN59CpXj12TzwgJ1rM8w"
	config.BaseUrl = "https://open.api.mytoken.io/"

	g := market.NewGlobalMarket(config)
	fmt.Println(g)
	//g.Start()

	//fmt.Println("12344")

	req := market.GlobalTrendReq{}
	req.TrendAnchor = "USDT"
	//req.Symbol = "LRC"
	//fmt.Println(g.Sign(req))
	//fmt.Println(g.GetGlobalTicker("LRC"))
	//fmt.Println("111112222")
	//time.Sleep(50 * time.Second)
	//fmt.Println("111112222333")
	//fmt.Println(market.GM.GetGlobalMarketTickerCache(""))
	//fmt.Println(market.GM.GetGlobalTickerCache("LRC"))
	//fmt.Println(market.GM.GetGlobalTrendCache("LRC"))
	fmt.Println(g.GetGlobalTicker("vite"))
	//time.Sleep(30* time.Second)
}
