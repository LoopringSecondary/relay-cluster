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
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/sns"
	"github.com/Loopring/relay-lib/types"
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
)

var GM *GlobalMarket

const GlobalTickerKey = "GTKPK"
const GlobalTrendKey = "GTDPK"
const GlobalMarketTickerKey = "GMTKPK"
const GMCLock = "globalMarketZkLock"

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
	Symbol            string `json:"symbol"`
	Price             string `json:"price"`
	PriceUsd          string `json:"price_usd"`
	PriceCnyUtc0      string `json:"price_cny_utc0"`
	PriceCny          string `json:"price_cny"`
	Volume24hUsd      string `json:"volume_24h_usd"`
	Volume24h         string `json:"volume_24h"`
	Volume24hFrom     string `json:"volume_24h_from"`
	PercentChangeUtc0 string `json:"percent_change_utc0"`
	Alias             string `json:"alias"`
	PriceUpdatedAt    string `json:"price_updated_at"`
}

type GlobalMarketTicker struct {
	MarketName        string `json:"market_name"`
	Symbol            string `json:"symbol"`
	Anchor            string `json:"anchor"`
	Pair              string `json:"pair"`
	Price             string `json:"price"`
	PriceUsd          string `json:"price_usd"`
	PriceCny          string `json:"price_cny"`
	Volume24hUsd      string `json:"volume_24h_usd"`
	Volume24h         string `json:"volume_24h"`
	Volume24hFrom     string `json:"volume_24h_from"`
	PercentChangeUtc0 string `json:"percent_change_utc0"`
	Alias             string `json:"alias"`
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
	Limit       int64  `json:"limit"`
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
	Anchor    string `json:"anchor"`
	NameId    string `json:"name_id"`
	Symbol    string `json:"symbol"`
	SortType  string `json:"sort_type"`
	SortField string `json:"sort_field"`
}

type GlobalMarketTickerResp struct {
	MyTokenResp
	Data map[string][]GlobalMarketTicker `json:"data"`
}

func NewGlobalMarket(config MyTokenConfig) GlobalMarket {
	GM = &GlobalMarket{config: config, cron: cron.New()}
	return *GM
}

func (g *GlobalMarket) GetGlobalTrendCache(token string) (trends map[string][]GlobalTrend, err error) {
	fields := getAllTokenField(token)
	if len(fields) == 0 {
		trends = make(map[string][]GlobalTrend)
		trends[strings.ToUpper(token)] = make([]GlobalTrend, 0)
		return
	}
	trendBytes, err := cache.HMGet(GlobalTrendKey, fields...)
	if err != nil || len(trendBytes) == 0 {
		return trends, err
	}

	trends = make(map[string][]GlobalTrend)

	for k, v := range fields {
		var trend []GlobalTrend
		errUnmarshal := json.Unmarshal(trendBytes[k], &trend)
		if errUnmarshal != nil {
			continue
		}
		trends[string(v)] = trend

	}

	return trends, err
}

func (g *GlobalMarket) GetGlobalTickerCache(token string) (tickers map[string]GlobalTicker, err error) {
	fields := getAllTokenField(token)
	if len(fields) == 0 {
		return
	}
	tickerBytes, err := cache.HMGet(GlobalTickerKey, fields...)
	if err != nil || len(tickerBytes) == 0 {
		return tickers, err
	}
	tickers = make(map[string]GlobalTicker)

	for k, v := range fields {
		var ticker GlobalTicker
		errUnmarshal := json.Unmarshal(tickerBytes[k], &ticker)
		if errUnmarshal != nil {
			continue
		}
		tickers[string(v)] = ticker

	}

	return tickers, err
}

func (g *GlobalMarket) GetGlobalMarketTickerCache(token string) (tickers map[string][]GlobalMarketTicker, err error) {
	fields := getAllTokenField(token)
	if len(fields) == 0 {
		tickers = make(map[string][]GlobalMarketTicker)
		tickers[strings.ToUpper(token)] = make([]GlobalMarketTicker, 0)
		return
	}
	tickerBytes, err := cache.HMGet(GlobalMarketTickerKey, fields...)
	if err != nil || len(tickerBytes) == 0 {
		return tickers, err
	}
	tickers = make(map[string][]GlobalMarketTicker)

	for k, v := range fields {
		var ticker []GlobalMarketTicker
		errUnmarshal := json.Unmarshal(tickerBytes[k], &ticker)
		if errUnmarshal != nil {
			continue
		}
		tickers[string(v)] = ticker

	}
	return tickers, err
}

func getAllTokenField(token string) [][]byte {

	token = strings.ToUpper(token)

	fields := [][]byte{}

	if strings.ToUpper(token) == "ETH" {
		token = "WETH"
	}

	_, ok := util.AllTokens[token]

	if ok {
		fields = append(fields, []byte(strings.ToUpper(token)))
	} else if token == "" {
		for k := range util.AllTokens {
			fields = append(fields, []byte(strings.ToUpper(k)))
		}
	}
	return fields
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
	nameId, err := getNameId(token)
	if err != nil {
		return trend, err
	}
	request := GlobalTrendReq{TrendAnchor: "usd", NameId: strings.ToLower(nameId), Limit: int64(90), Period: "1d"}
	urlParam, err := g.Sign(request)
	if err != nil {
		return trend, err
	}

	body, err := getResp(url + urlParam)
	if err != nil {
		return trend, err
	}

	var resp GlobalTrendResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return trend, err
	}

	return resp.Data["trend"], nil
}

func (g *GlobalMarket) GetGlobalTicker(token string) (ticker GlobalTicker, err error) {
	url := g.config.BaseUrl + "ticker/global?"
	nameId, err := getNameId(token)
	if err != nil {
		return ticker, err
	}

	request := GlobalTickerReq{NameId: strings.ToLower(nameId)}
	urlParam, err := g.Sign(request)
	if err != nil {
		return ticker, err
	}

	body, err := getResp(url + urlParam)
	if err != nil {
		return ticker, err
	}

	var resp GlobalTickerResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return ticker, err
	}

	return resp.Data, nil
}

func (g *GlobalMarket) GetGlobalMarketTicker(symbol string) (trend []GlobalMarketTicker, err error) {

	url := g.config.BaseUrl + "ticker/paironmarket?"

	token, ok := util.AllTokens[strings.ToUpper(symbol)]
	if !ok {
		return trend, errors.New("unsupported token " + symbol)
	}
	nameId := token.Source
	var request GlobalMarketTickerReq
	if strings.ToUpper(symbol) == "ETH" || strings.ToUpper(symbol) == "WETH" {
		request = GlobalMarketTickerReq{NameId: "ethereum", Anchor: "usd", Symbol: "eth", SortType: "desc", SortField: "volume_24h_usd"}
	} else {
		request = GlobalMarketTickerReq{NameId: nameId, Anchor: "usd", Symbol: strings.ToLower(token.Symbol), SortType: "desc", SortField: "volume_24h_usd"}
	}
	urlParam, err := g.Sign(request)
	if err != nil {
		return trend, err
	}

	body, err := getResp(url + urlParam)
	if err != nil {
		return trend, err
	}

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
			return urlParam, errors.New("unsupported data type " + reflect.TypeOf(v).String())
		}
	}

	waitToSign := strings.Join(signatureList, "&")
	sign := computeHmac256(waitToSign+"&app_secret="+g.config.AppSecret, g.config.AppSecret)

	return waitToSign + "&sign=" + strings.ToUpper(sign), nil
}

func (g *GlobalMarket) Start() {
	go func() {
		if zklock.TryLock(GMCLock) == nil {
			syncGlobalTrend()
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

func syncData(redisKey string, syncFunc func(token string) ([]byte, []byte, error)) {
	if GM == nil {
		return
	}

	data := [][]byte{}
	for k := range util.AllTokens {
		kF, vF, err := syncFunc(k)
		if err != nil {
			continue
		}
		data = append(data, kF, vF)
	}
	cache.HMSet(redisKey, -1, data...)
}

func syncGlobalTrend() {
	syncData(GlobalTrendKey, GM.syncGlobalTrend)
}

func syncGlobalTicker() {
	syncData(GlobalTickerKey, GM.syncGlobalTicker)
}

func syncGlobalMarketTicker() {
	syncData(GlobalMarketTickerKey, GM.syncPairOnMarket)
}

func (g *GlobalMarket) syncGlobalTrend(token string) ([]byte, []byte, error) {

	trends, err := g.GetGlobalTrend(token)
	if err != nil {
		return nil, nil, err
	}

	trendsInByte, err := json.Marshal(trends)
	if err != nil {
		return nil, nil, err
	}

	return []byte(strings.ToUpper(token)), trendsInByte, nil
}

func (g *GlobalMarket) syncGlobalTicker(token string) ([]byte, []byte, error) {
	ticker, err := g.GetGlobalTicker(token)
	if err != nil {
		return nil, nil, err
	}

	tickerInByte, err := json.Marshal(ticker)
	if err != nil {
		return nil, nil, err
	}

	return []byte(strings.ToUpper(token)), tickerInByte, nil
}

func (g *GlobalMarket) syncPairOnMarket(token string) ([]byte, []byte, error) {
	marketTickers, err := g.GetGlobalMarketTicker(token)
	if err != nil {
		return nil, nil, err
	}

	tickersByte, err := json.Marshal(marketTickers)
	if err != nil {
		return nil, nil, err
	}
	return []byte(strings.ToUpper(token)), tickersByte, nil
}

func computeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return strings.ToUpper(strings.TrimLeft(types.BytesToBytes32(h.Sum(nil)).Hex(), "0x"))
}

func getNameId(symbol string) (nameId string, err error) {
	token, ok := util.AllTokens[symbol]
	if !ok {
		return nameId, errors.New("unsupported token " + symbol)
	}

	return token.Source, nil
}
