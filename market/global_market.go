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
	"fmt"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/sns"
	"github.com/Loopring/relay-lib/types"
	"github.com/Loopring/relay-lib/zklock"
	"github.com/Loopring/relay/cache"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var globalMarket *GlobalMarket

const GlobalTickerPreKey = "GTKPK_"
const GlobalTrendPreKey = "GTDPK_"
const GlobalMarketTickerPreKey = "GMTKPK_"

type GlobalMarket struct {
	config MyTokenConfig
	cron   *cron.Cron
}

type GlobalTrend struct {
	Price      string `json:"price"`
	Time       string `json:"time"`
	VolumeFrom string `json:"volumefrom"`
}

type GlobalTicker struct {
	Symbol            string  `json:"symbol"`
	Price             string `json:"price"`
	PriceUsd          string `json:"price_usd"`
	PriceCnyUtc0      string `json:"price_cny_utc0"`
	PriceCny          string `json:"price_cny"`
	Volume24hUsd      string  `json:"volume_24h_usd"`
	Volume24h         string  `json:"volume_24h"`
	Volume24hFrom     string   `json:"volume_24h_from"`
	PercentChangeUtc0 string `json:"percent_change_utc0"`
	Alias             string  `json:"alias"`
	PriceUpdatedAt    string   `json:"price_updated_at"`
}

type GlobalMarketTicker struct {
	MarketName        string  `json:"market_name"`
	Symbol            string  `json:"symbol"`
	Anchor            string  `json:"anchor"`
	Pair              string  `json:"pair"`
	Price             string  `json:"price"`
	PriceUsd          string  `json:"price_usd"`
	PriceCny          string  `json:"price_cny"`
	Volume24hUsd      string  `json:"volume_24h_usd"`
	Volume24h         string  `json:"volume_24h"`
	Volume24hFrom     string  `json:"volume_24h_from"`
	PercentChangeUtc0 string  `json:"percent_change_utc0"`
	Alias             string  `json:"alias"`
}

type MyTokenConfig struct {
	AppId     string
	AppSecret string
	BaseUrl   string
}

type MyTokenResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type GlobalTrendReq struct {
	TrendAnchor string `json:"trend_anchor"`
	NameId      string `json:"name_id"`
	Limit      int64 `json:"limit"`
	Period      string `json:"period"`
}

type GlobalTrendResp struct {
	MyTokenResp
	Data map[string][]GlobalTrend `json:"data"`
}

type GlobalTickerReq struct {
	NameId string `json:"name_id"`
}

type GlobalTickerResp struct {
	MyTokenResp
	Data GlobalTicker `json:"data"`
}

type GlobalMarketTickerReq struct {
	Anchor string `json:"anchor"`
	NameId      string `json:"name_id"`
	Symbol      string `json:"symbol"`
	SortType      string `json:"sort_type"`
	SortField      string `json:"sort_field"`
}

type GlobalMarketTickerResp struct {
	MyTokenResp
	Data map[string][]GlobalMarketTicker `json:"data"`
}

func NewGlobalMarket(config MyTokenConfig) GlobalMarket {
	return GlobalMarket{config: config}
}

func (g *GlobalMarket) GetGlobalTrendCache(token string) (trend []GlobalTrend, err error) {
	trendBytes, err := cache.Get(GlobalTrendPreKey + strings.ToUpper(token)); if err != nil {
		return trend, err
	}
	trend = make([]GlobalTrend, 0)
	err = json.Unmarshal(trendBytes, &trend)
	return trend, err
}

func (g *GlobalMarket) GetGlobalTickerCache(token string) (ticker GlobalTicker, err error) {
	tickerBytes, err := cache.Get(GlobalTickerPreKey + strings.ToUpper(token)); if err != nil {
		return ticker, err
	}
	err = json.Unmarshal(tickerBytes, &ticker)
	return ticker, err
}

func (g *GlobalMarket) GetGlobalMarketTickerCache(token string) (tickers []GlobalMarketTicker, err error) {
	tickerBytes, err := cache.Get(GlobalMarketTickerPreKey + strings.ToUpper(token)); if err != nil {
		return tickers, err
	}
	err = json.Unmarshal(tickerBytes, &tickers)
	return tickers, err
}

func getResp(url string) (body []byte, err error) {

	resp, err := http.Get(url)
	if err != nil {
		return body, err
	}
	if resp != nil && resp.Status != "200 OK" {
		return body, errors.New("response error, resp.Status : " + resp.Status)
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
	nameId, err := getNameId(token); if err != nil {
		return trend, err
	}
	request := GlobalTrendReq{TrendAnchor: "usd", NameId: strings.ToLower(nameId), Limit : int64(90), Period: "1d"}
	urlParam, err := g.Sign(request)
	if err != nil {
		return trend, err
	}

	fmt.Println(url + urlParam)
	body, err := getResp(url + urlParam)
	if err != nil {
		return trend, err
	}

	var resp GlobalTrendResp
	fmt.Println(string(body))
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return trend, err
	}

	return resp.Data["trend"], nil
}

func (g *GlobalMarket) GetGlobalTicker(token string) (ticker GlobalTicker, err error) {
	url := g.config.BaseUrl + "ticker/global?"
	nameId, err := getNameId(token); if err != nil {
		return ticker, err
	}

	request := GlobalTickerReq{NameId: strings.ToLower(nameId)}
	urlParam, err := g.Sign(request)
	if err != nil {
		return ticker, err
	}

	fmt.Println(url + urlParam)
	body, err := getResp(url + urlParam)
	if err != nil {
		return ticker, err
	}

	fmt.Println(string(body))
	var resp GlobalTickerResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return ticker, err
	}

	return resp.Data, nil
}

func (g *GlobalMarket) GetGlobalMarketTicker(symbol string) (trend []GlobalMarketTicker, err error) {

	url := g.config.BaseUrl + "ticker/paironmarket?"

	token, ok := util.AllTokens[strings.ToUpper(symbol)]; if !ok {
		return trend, errors.New("unsupported token " + symbol)
	}
	nameId := token.Source
	request := GlobalMarketTickerReq{NameId: nameId, Anchor: "eth", Symbol : strings.ToLower(token.Symbol), SortType:"desc", SortField: "volume_24h_usd"}
	urlParam, err := g.Sign(request)
	if err != nil {
		return trend, err
	}

	fmt.Println(url + urlParam)
	body, err := getResp(url + urlParam)
	if err != nil {
		return trend, err
	}

	fmt.Println(string(body))
	var resp GlobalMarketTickerResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
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
		} else if reflect.TypeOf(v).Kind() == reflect.Float64 {
			signatureList = append(signatureList, k+"="+strconv.FormatInt(int64(v.(float64)), 10))
		} else {
			fmt.Println(v)
			return urlParam, errors.New("unsupported data type " + reflect.TypeOf(v).String())
		}
	}

	waitToSign := strings.Join(signatureList, "&")
	sign := computeHmac256(waitToSign+"&app_secret="+g.config.AppSecret, g.config.AppSecret)

	return waitToSign + "&sign=" + strings.ToUpper(sign), nil
}

func (g *GlobalMarket) Start() {
	go func() {
		if zklock.TryLock(tickerCollectorCronJobZkLock) == nil {
			g.cron.AddFunc("@every 5s", syncGlobalTicker)
			g.cron.AddFunc("@every 10s", syncGlobalMarketTicker)
			g.cron.AddFunc("@every 1h", syncGlobalTrend)
			log.Info("start mytoken global market cron jobs......... ")
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

func syncGlobalTicker() {
	if globalMarket == nil {
		return
	}
	for k := range util.AllTokens {
		globalMarket.syncGlobalTicker(k)
	}
}

func syncGlobalMarketTicker() {
	if globalMarket == nil {
		return
	}
	for t := range util.AllTokens {
		globalMarket.syncPairOnMarket(t)
	}
}

func (g *GlobalMarket) syncGlobalTrend(token string) error {

	trends, err := g.GetGlobalTrend(token)
	if err != nil {
		return err
	}

	trendsInByte, err := json.Marshal(trends)
	if err != nil {
		return err
	}
	return cache.Set(GlobalTrendPreKey+strings.ToUpper(token), trendsInByte, -1)
}

func (g *GlobalMarket) syncGlobalTicker(token string) error {
	ticker, err := g.GetGlobalTicker(token)
	if err != nil {
		return err
	}

	tickerInByte, err := json.Marshal(ticker)
	if err != nil {
		return err
	}
	return cache.Set(GlobalTickerPreKey+strings.ToUpper(token), tickerInByte, -1)
}

func (g *GlobalMarket) syncPairOnMarket(token string) error {
	marketTickers, err := g.GetGlobalMarketTicker(token)
	if err != nil {
		return err
	}

	tickersByte, err := json.Marshal(marketTickers)
	if err != nil {
		return err
	}
	return cache.Set(GlobalMarketTickerPreKey+strings.ToUpper(token), tickersByte, -1)
}

func computeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return strings.ToUpper(strings.TrimLeft(types.BytesToBytes32(h.Sum(nil)).Hex(), "0x"))
}

func getNameId(symbol string) (nameId string, err error) {
	token, ok := util.AllTokens[symbol]; if !ok {
		return nameId, errors.New("unsupported token " + symbol)
	}
	return token.Source, nil
}
