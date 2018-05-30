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

package market

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/sns"
	"github.com/Loopring/relay-lib/zklock"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"github.com/Loopring/relay-lib/types"
	"fmt"
)

var globalMarket *GlobalMarket

type GlobalMarket struct {
	config MyTokenConfig
	cron   *cron.Cron
}

type GlobalTrend struct {
	Price string `json:"price"`
	Time string `json:"time"`
	VolumeFrom string `json:"volumefrom"`
}

type GlobalTicker struct {
	Symbol string `json:"symbol"`
	Price float64 `json:"price"`
	PriceUsd float64 `json:"price_usd"`
	PriceCnyUtc0 float64 `json:"price_cny_utc0"`
	PriceCny float64 `json:"price_cny"`
	Volume24hUsd string `json:"volume_24h_usd"`
	Volume24h string `json:"volume_24h"`
	Volume24hFrom int64 `json:"volume_24h_from"`
	PercentChangeUtc0 float64 `json:"percent_change_utc0"`
	Alias string `json:"alias"`
	PriceUpdatedAt int64 `json:"price_updated_at"`
}

type GlobalMarketTicker struct {
	Symbol string `json:"symbol"`
	Price string `json:"price_usd"`
	PriceUsd string `json:"price"`
	PriceCnyUtc0 float64 `json:"price_cny_utc0"`
	PriceCny string `json:"price_cny"`
	Volume24hUsd string `json:"volume_24h_usd"`
	Volume24h string `json:"volume_24h"`
	Volume24hFrom int64 `json:"volume_24h_from"`
	PercentChangeUtc0 float64 `json:"percent_change_utc0"`
	Alias string `json:"alias"`
	PriceUpdatedAt int64 `json:"price_updated_at"`
}

type MyTokenConfig struct {
	AppId     string
	AppSecret string
	BaseUrl   string
}

type MyTokenResp struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Timestamp int64 `json:"timestamp"`
}

type GlobalTrendReq struct {
	TrendAnchor string `json:"trend_anchor"`
	Symbol      string `json:"symbol"`
}

type GlobalTrendResp struct {
	MyTokenResp
	Data map[string][]GlobalTrend `json:"data"`
}

type GlobalTickerReq struct {
	Symbol string `json:"symbol"`
}

type GlobalTickerResp struct {
	MyTokenResp
	Data GlobalTicker `json:"data"`
}

type GlobalMarketTickerReq struct {
	Pair string `json:"pair"`
}

type GlobalMarketTickerResp struct {
	MyTokenResp
	Data map[string][]GlobalMarketTicker `json:"data"`
}


func NewGlobalMarket(config MyTokenConfig) GlobalMarket {
	return GlobalMarket{config: config}
}

func getResp(url string) (body []byte, err error) {

	resp, err := http.Get(url)
	if err != nil || (resp != nil && resp.Status != "200") {

		return body, errors.New("response error, resp.Status")
	}
	if err != nil || (resp != nil && resp.Status != "200") {

		return body, errors.New("response error, resp.Status")
	}

	defer func() {
		if nil != resp && nil != resp.Body {
			resp.Body.Close()
		}
	}()

	return ioutil.ReadAll(resp.Body)
}

func (g *GlobalMarket) GetGlobalTrend(token string) (trend []GlobalTrend, err error) {

	url := g.config.BaseUrl + "symbol/trend?"
	request := GlobalTrendReq{TrendAnchor: "USDT", Symbol: strings.ToUpper(token)}
	urlParam, err := g.Sign(request)
	if err != nil {
		return trend, err
	}

	fmt.Println(url + urlParam)
	body, err := getResp(url+urlParam); if err != nil {
		return trend, err
	}

	var resp GlobalTrendResp
	fmt.Println(string(body))
	err = json.Unmarshal(body, &resp); if err != nil {
		return trend, err
	}

	return resp.Data["trend"], nil
}

func (g *GlobalMarket) GetGlobalTicker(token string) (ticker GlobalTicker, err error) {
	url := g.config.BaseUrl + "ticker/global?"
	request := GlobalTickerReq{Symbol: strings.ToUpper(token)}
	urlParam, err := g.Sign(request)
	if err != nil {
		return ticker, err
	}

	fmt.Println(url + urlParam)
	body, err := getResp(url+urlParam); if err != nil {
		return ticker, err
	}

	fmt.Println(string(body))
	var resp GlobalTickerResp
	err = json.Unmarshal(body, &resp); if err != nil {
		return ticker, err
	}

	return resp.Data, nil
}

func (g *GlobalMarket) GetGlobalMarketTicker(market string) (trend []GlobalMarketTicker, err error) {

	url := g.config.BaseUrl + "ticker/paironmarket?"
	market = strings.ToUpper(market)
	market = strings.Replace(market, "WETH", "ETH", 0)
	market = strings.Replace(market, "-", "_", 0)
	request := GlobalMarketTickerReq{Pair: market}
	urlParam, err := g.Sign(request)
	if err != nil {
		return trend, err
	}

	fmt.Println(url + urlParam)
	body, err := getResp(url+urlParam); if err != nil {
		return trend, err
	}

	fmt.Println(string(body))
	var resp GlobalMarketTickerResp
	err = json.Unmarshal(body, &resp); if err != nil {
		return trend, err
	}

	return resp.Data["market_list"], nil
}

func (g *GlobalMarket) Sign(param interface{}) (urlParam string, err error) {

	jsonStr, err := json.Marshal(param)
	if err != nil {
		return urlParam, err
	}
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonStr, &jsonMap)
	if err != nil {
		return urlParam, err
	}

	jsonMap["app_id"] = g.config.AppId
	jsonMap["timestamp"] = time.Now().Unix()

	var keys []string
	for k := range jsonMap {
		keys = append(keys, k)
	}

	signatureList := make([]string, 0)
	sort.Strings(keys)
	for _, k := range keys {
		v := jsonMap[k]
		if reflect.TypeOf(v).Kind() == reflect.String {
			signatureList = append(signatureList, k+"="+v.(string))
		} else if reflect.TypeOf(v).Kind() == reflect.Int64 {
			signatureList = append(signatureList, k+"="+strconv.FormatInt(v.(int64), 10))
		} else {
			return urlParam, errors.New("unsupported data type " + reflect.TypeOf(v).String())
		}
	}

	waitToSign := strings.Join(signatureList, "&")
	sign := computeHmac256(waitToSign + "&app_secret=" + g.config.AppSecret, g.config.AppSecret)

	return waitToSign + "&sign=" + strings.ToUpper(sign), nil
}

func (g *GlobalMarket) Start() {
	go func() {
		if zklock.TryLock(tickerCollectorCronJobZkLock) == nil {
			updateBinanceCache()
			updateOkexCache()
			updateHuobiCache()
			g.cron.AddFunc("@every 20s", updateBinanceCache)
			log.Info("start collect cron jobs......... ")
			g.cron.Start()
		} else {
			err := sns.PublishSns(tryLockFailedMsg, tryLockFailedMsg)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}()
}

func syncGlobalTrend() {
	if globalMarket == nil {
		return
	}
	for k := range util.AllTokens {
		globalMarket.syncGlobalTrend(k)
	}
}

func (g *GlobalMarket) syncGlobalTrend(token string) error {

	return nil
}

func (g *GlobalMarket) syncGlobalTicker(token string) error {
	return nil
}

func (g *GlobalMarket) syncPairOnMarket(market string) error {
	return nil
}

func computeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return strings.ToUpper(strings.TrimLeft(types.BytesToBytes32(h.Sum(nil)).Hex(), "0x"))
}
