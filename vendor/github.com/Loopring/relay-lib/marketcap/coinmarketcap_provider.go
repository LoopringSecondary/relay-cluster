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

package marketcap

import (
	"github.com/Loopring/relay-lib/zklock"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"encoding/json"
	"errors"
	"fmt"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/cloudwatch"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/sns"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"net/http"
)

const (
	CACHEKEY_COIN_MARKETCAP  = "coin_marketcap_"
	ZKNAME_COIN_MARKETCAP    = "coin_marketcap_"
	HEARTBEAT_COIN_MARKETCAP = "coin_marketcap"
)

type MarketCap struct {
	Price            *big.Rat
	Volume24H        *big.Rat
	MarketCap        *big.Rat
	PercentChange1H  *big.Rat
	PercentChange24H *big.Rat
	PercentChange7D  *big.Rat
}

func (cap *MarketCap) UnmarshalJSON(input []byte) error {
	type Cap struct {
		Price            float64 `json:"price"`
		Volume24H        float64 `json:"24h_volume"`
		MarketCap        float64 `json:"market_cap"`
		PercentChange1H  float64 `json:"percent_change_1h"`
		PercentChange24H float64 `json:"percent_change_24h"`
		PercentChange7D  float64 `json:"percent_change_7d"`
	}
	c := &Cap{}
	if err := json.Unmarshal(input, c); nil != err {
		return err
	} else {

		cap.Price = new(big.Rat).SetFloat64(c.Price)
		cap.Volume24H = new(big.Rat).SetFloat64(c.Volume24H)
		cap.MarketCap = new(big.Rat).SetFloat64(c.MarketCap)
		cap.PercentChange1H = new(big.Rat).SetFloat64(c.PercentChange1H)
		cap.PercentChange24H = new(big.Rat).SetFloat64(c.PercentChange24H)
		cap.PercentChange7D = new(big.Rat).SetFloat64(c.PercentChange7D)
	}
	return nil
}
func (cap *MarketCap) MarshalJSON() ([]byte, error) {
	type Cap struct {
		Price            float64 `json:"price"`
		Volume24H        float64 `json:"24h_volume"`
		MarketCap        float64 `json:"market_cap"`
		PercentChange1H  float64 `json:"percent_change_1h"`
		PercentChange24H float64 `json:"percent_change_24h"`
		PercentChange7D  float64 `json:"percent_change_7d"`
	}
	c := &Cap{}
	c.Price, _ = cap.Price.Float64()
	c.Volume24H, _ = cap.Volume24H.Float64()
	c.MarketCap, _ = cap.MarketCap.Float64()
	c.PercentChange1H, _ = cap.PercentChange1H.Float64()
	c.PercentChange24H, _ = cap.PercentChange24H.Float64()
	c.PercentChange7D, _ = cap.PercentChange7D.Float64()
	return json.Marshal(c)
}

type CoinMarketCap struct {
	Id                int                   `json:"id"`
	Address           common.Address        `json:"-"`
	Name              string                `json:"name"`
	Symbol            string                `json:"symbol"`
	WebsiteSlug       string                `json:"website_slug"`
	Rank              int                   `json:"rank"`
	CirculatingSupply float64               `json:"circulating_supply"`
	TotalSupply       float64               `json:"total_supply"`
	MaxSupply         float64               `json:"max_supply"`
	Quotes            map[string]*MarketCap `json:"quotes"`
	LastUpdated       int64                 `json:"last_updated"`
	Decimals          *big.Int
}

type CoinMarketCapResult struct {
	Data     map[string]CoinMarketCap `json:"data"`
	Metadata struct {
		Timestamp           int64  `json:"timestamp"`
		NumCryptocurrencies int    `json:"num_cryptocurrencies"`
		Error               string `json:"error"`
	} `json:"metadata"`
}

type icoTokens []common.Address

func (tokens icoTokens) contains(addr common.Address) bool {
	for _, token := range tokens {
		if token == addr {
			return true
		}
	}
	return false
}

type CapProvider_CoinMarketCap struct {
	baseUrl         string
	tokenMarketCaps map[common.Address]*CoinMarketCap
	icoTokens       icoTokens
	notSupportTokens map[common.Address]bool
	//icoTokens	noSupportTokens
	slugToAddress   map[string]common.Address
	currency        string
	duration        int
	dustValue       *big.Rat
	stopFuncs       []func()
}

func (p *CapProvider_CoinMarketCap) LegalCurrencyValue(tokenAddress common.Address, amount *big.Rat) (*big.Rat, error) {
	return p.LegalCurrencyValueByCurrency(tokenAddress, amount, p.currency)
}

func (p *CapProvider_CoinMarketCap) LegalCurrencyValueOfEth(amount *big.Rat) (*big.Rat, error) {
	tokenAddress := util.AllTokens["WETH"].Protocol
	return p.LegalCurrencyValueByCurrency(tokenAddress, amount, p.currency)
}

func (p *CapProvider_CoinMarketCap) LegalCurrencyValueByCurrency(tokenAddress common.Address, amount *big.Rat, currencyStr string) (*big.Rat, error) {
	if _,exists := p.notSupportTokens[tokenAddress]; exists {
		return big.NewRat(int64(0), int64(1)), nil
	} else if c, exists := p.tokenMarketCaps[tokenAddress]; !exists {
		return nil, errors.New("not found tokenCap:" + tokenAddress.Hex())
	} else {
		v := new(big.Rat).SetInt(c.Decimals)
		v.Quo(amount, v)
		if price, err := p.GetMarketCapByCurrency(tokenAddress, currencyStr); nil != err {
			log.Errorf("err:%s", err.Error())
			return nil, err
		} else {
			log.Debugf("LegalCurrencyValueByCurrency token:%s,decimals:%s, amount:%s, currency:%s, price:%s", tokenAddress.Hex(), c.Decimals.String(), amount.FloatString(2), currencyStr, price.FloatString(2))
			v.Mul(price, v)
			return v, nil
		}
	}
}

func (p *CapProvider_CoinMarketCap) GetMarketCap(tokenAddress common.Address) (*big.Rat, error) {
	return p.GetMarketCapByCurrency(tokenAddress, p.currency)
}

func (p *CapProvider_CoinMarketCap) GetEthCap() (*big.Rat, error) {
	return p.GetMarketCapByCurrency(util.AllTokens["WETH"].Protocol, p.currency)
}

func (p *CapProvider_CoinMarketCap) getMarketCapFromRedis(websiteSlug string, currencyStr string) (cap *CoinMarketCap, err error) {
	cap = &CoinMarketCap{}
	if data, err := cache.Get(p.cacheKey(websiteSlug, currencyStr)); nil != err {
		log.Errorf("err:%s", err.Error())
		return nil, err
	} else {
		if err := json.Unmarshal(data, cap); nil != err {
			log.Errorf("get marketcap of token err:%s", err.Error())
			return nil, err
		} else {
			return cap, err
		}
	}
}

func (p *CapProvider_CoinMarketCap) GetMarketCapByCurrency(tokenAddress common.Address, currencyStr string) (*big.Rat, error) {
	if _,exists := p.notSupportTokens[tokenAddress]; exists {
		return big.NewRat(int64(0), int64(1)), nil
	} else if c, exists := p.tokenMarketCaps[tokenAddress]; exists {
		var v *big.Rat
		if quote, exists := c.Quotes[currencyStr]; exists {
			v = quote.Price
		} else {
			if p.icoTokens.contains(tokenAddress) {
				wethCap, err := p.getMarketCapFromRedis(util.AllTokens["WETH"].Source, currencyStr)
				if nil == err {
					if quote, exists := wethCap.Quotes[currencyStr]; exists {
						v = new(big.Rat).Set(quote.Price)
						v.Mul(v, util.AllTokens[c.Symbol].IcoPrice)
					}
				}
			} else {
				var err error
				c, err = p.getMarketCapFromRedis(c.WebsiteSlug, currencyStr)
				if nil == err {
					if quote, exists := c.Quotes[currencyStr]; exists {
						v = quote.Price
					}
				}
			}
		}
		if v == nil {
			return nil, errors.New("tokenCap is nil")
		} else {
			return new(big.Rat).Set(v), nil
		}
	} else {
		err := errors.New("not found tokenCap:" + tokenAddress.Hex())
		res := new(big.Rat).SetInt64(int64(1))
		if nil != err {
			log.Errorf("get MarketCap of token:%s, occurs error:%s. the value will be default value:%s", tokenAddress.Hex(), err.Error(), res.String())
		}
		return res, err
	}
}

func (p *CapProvider_CoinMarketCap) Stop() {
	for _, f := range p.stopFuncs {
		f()
	}
}

func (p *CapProvider_CoinMarketCap) Start() {
	stopChan := make(chan bool)
	p.stopFuncs = append(p.stopFuncs, func() {
		stopChan <- true
	})
	go func() {
		for {
			select {
			case <-time.After(time.Duration(p.duration) * time.Minute):
				log.Debugf("sync marketcap from redis...")
				if err := p.syncMarketCapFromRedis(); nil != err {
					log.Errorf("can't sync marketcap, time:%d", time.Now().Unix())
				}
			case stopped := <-stopChan:
				if stopped {
					return
				}
			}
		}
	}()

	go p.syncMarketCapFromAPIWithZk()
}

func (p *CapProvider_CoinMarketCap) zklockName() string {
	return ZKNAME_COIN_MARKETCAP + p.currency
}

func (p *CapProvider_CoinMarketCap) heartBeatName() string {
	return HEARTBEAT_COIN_MARKETCAP + p.currency
}

func (p *CapProvider_CoinMarketCap) cacheKey(websiteSlug string, currency string) string {
	if "" != currency {
		return CACHEKEY_COIN_MARKETCAP + strings.ToLower(websiteSlug) + "_" + currency
	} else {
		return CACHEKEY_COIN_MARKETCAP + strings.ToLower(websiteSlug) + "_" + p.currency
	}
}

func (p *CapProvider_CoinMarketCap) syncMarketCapFromAPIWithZk() {
	if err := zklock.TryLock(p.zklockName()); nil != err {
		log.Errorf("err:%s", err.Error())
		sns.PublishSns("marketcap failed", "try to get zklock err:"+err.Error())
		return
	}
	log.Infof("MarketCap has gotten zklock....")
	stopChan := make(chan bool)
	p.stopFuncs = append(p.stopFuncs, func() {
		stopChan <- true
	})

	go func() {
		for {
			select {
			case <-time.After(time.Duration(p.duration) * time.Minute):
				log.Debugf("sync marketcap(key:%s) from api...", p.zklockName())
				p.syncMarketCapFromAPI()
				if err := cloudwatch.PutHeartBeatMetric(p.heartBeatName()); nil != err {
					log.Errorf("err:%s", err.Error())
				}
			case stopped := <-stopChan:
				if stopped {
					zklock.ReleaseLock(p.zklockName())
					return
				}
			}
		}
	}()
}

func (p *CapProvider_CoinMarketCap) syncMarketCapFromAPI() error {
	log.Debugf("syncMarketCapFromAPI...")
	//https://api.coinmarketcap.com/v2/ticker/?convert=%s&start=%d&limit=%d
	numCryptocurrencies := 105
	start := 0
	limit := 100
	for numCryptocurrencies > 0 {
		url := fmt.Sprintf(p.baseUrl, p.currency, start, limit)
		println(url)
		resp, err := http.Get(url)
		if err != nil {
			log.Errorf("err:%s", err.Error())
			return err
		}
		defer func() {
			if nil != resp && nil != resp.Body {
				resp.Body.Close()
			}
		}()

		body, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			log.Errorf("err:%s", err.Error())
			return err
		} else {
			result := &CoinMarketCapResult{}
			if err1 := json.Unmarshal(body, result); nil != err1 {
				log.Errorf("err1:%s", err1.Error())
				return err1
			} else {

				if "" == result.Metadata.Error {
					for _, cap1 := range result.Data {
						if data, err2 := json.Marshal(cap1); nil != err2 {
							log.Errorf("err:%s", err2.Error())
							return err2
						} else {
							err = cache.Set(p.cacheKey(cap1.WebsiteSlug, p.currency), data, int64(43200))
							if nil != err {
								log.Errorf("err:%s", err.Error())
								return err
							}
						}
					}
					start = start + len(result.Data)
					numCryptocurrencies = result.Metadata.NumCryptocurrencies - start
				} else {
					log.Errorf("err:%s", result.Metadata.Error)
				}
			}
		}

	}

	return nil
}

func (p *CapProvider_CoinMarketCap) syncMarketCapFromRedis() error {
	//todo:use zk to keep
	//tokenMarketCaps := make(map[common.Address]*CoinMarketCap)
	syncedFromApi := false
	notSupportTokens := make(map[common.Address]bool)
	for tokenAddr, c1 := range p.tokenMarketCaps {
		if p.icoTokens.contains(tokenAddr) {
			continue
		}
		data, err := cache.Get(p.cacheKey(c1.WebsiteSlug, p.currency))
		if nil != err && !syncedFromApi {
			if err1 := p.syncMarketCapFromAPI(); nil != err1 {
				return err1
			}
			syncedFromApi = true
			data, err = cache.Get(p.cacheKey(c1.WebsiteSlug, p.currency))
		}
		if nil != err {
			notSupportTokens[tokenAddr] = true
			log.Errorf("can't get marketcap of token:%s", tokenAddr.Hex())
		} else {
			c := &CoinMarketCap{}
			if err := json.Unmarshal(data, c); nil != err {
				notSupportTokens[tokenAddr] = true
				log.Errorf("get marketcap of token err:%s", err.Error())
			} else {
				p.tokenMarketCaps[tokenAddr].Quotes = c.Quotes
				p.tokenMarketCaps[tokenAddr].CirculatingSupply = c.CirculatingSupply
				p.tokenMarketCaps[tokenAddr].Rank = c.Rank
				p.tokenMarketCaps[tokenAddr].TotalSupply = c.TotalSupply
				p.tokenMarketCaps[tokenAddr].MaxSupply = c.MaxSupply
				p.tokenMarketCaps[tokenAddr].LastUpdated = c.LastUpdated
			}
		}
	}
	p.notSupportTokens = notSupportTokens
	wethAddress := util.AllTokens["WETH"].Protocol

	for currency, wethCap := range p.tokenMarketCaps[wethAddress].Quotes {
		for _, tokenAddr := range p.icoTokens {
			if c, exists := p.tokenMarketCaps[tokenAddr]; exists {
				if nil == c.Quotes {
					c.Quotes = make(map[string]*MarketCap)
				}
				if _, exists := c.Quotes[currency]; !exists {
					c.Quotes[currency] = &MarketCap{}
				}
				c.Quotes[currency].Price = new(big.Rat).Mul(wethCap.Price, util.AllTokens[c.Symbol].IcoPrice)
			}
		}
	}

	return nil
}

func (p *CapProvider_CoinMarketCap) IsSupport(token common.Address) bool {
	_,exists := p.notSupportTokens[token]
	return !exists
}

func (p *CapProvider_CoinMarketCap) icoPriceTokens() []common.Address {
	tokenAddrs := []common.Address{}
	for _, token := range util.AllTokens {
		if nil != token.IcoPrice && token.IcoPrice.Cmp(big.NewRat(int64(0), int64(1))) > 0 {
			tokenAddrs = append(tokenAddrs, token.Protocol)
		}
	}
	//tokenAddrs = append(tokenAddrs, common.HexToAddress("0xbeb6fdf4ef6ceb975157be43cbe0047b248a8922"), common.HexToAddress("0x1b793E49237758dBD8b752AFC9Eb4b329d5Da016"))
	return tokenAddrs
}

type MarketCapOptions struct {
	BaseUrl   string
	Currency  string
	Duration  int
	IsSync    bool
	DustValue *big.Rat
}

func NewMarketCapProvider(options *MarketCapOptions) *CapProvider_CoinMarketCap {
	provider := &CapProvider_CoinMarketCap{}
	provider.baseUrl = options.BaseUrl
	provider.currency = options.Currency
	provider.tokenMarketCaps = make(map[common.Address]*CoinMarketCap)
	provider.notSupportTokens = make(map[common.Address]bool)
	provider.icoTokens = provider.icoPriceTokens()
	provider.slugToAddress = make(map[string]common.Address)
	provider.duration = options.Duration
	provider.dustValue = options.DustValue
	if provider.duration <= 0 {
		//default 5 min
		provider.duration = 5
	}
	provider.stopFuncs = []func(){}

	// default dust value is 1.0 usd/cny
	if provider.dustValue.Cmp(new(big.Rat).SetFloat64(0)) <= 0 {
		provider.dustValue = new(big.Rat).SetFloat64(1.0)
	}

	for _, v := range util.AllTokens {
		c := &CoinMarketCap{}
		c.Address = v.Protocol
		c.WebsiteSlug = v.Source
		c.Name = v.Symbol
		c.Symbol = v.Symbol
		c.Decimals = new(big.Int).Set(v.Decimals)
		provider.tokenMarketCaps[c.Address] = c
		provider.slugToAddress[strings.ToUpper(c.WebsiteSlug)] = c.Address
		//if "ARP" == v.Symbol || "VITE" == v.Symbol {
		//	c := &CoinMarketCap{}
		//	c.Address = v.Protocol
		//	c.WebsiteSlug = v.Source
		//	c.Name = v.Symbol
		//	c.Symbol = v.Symbol
		//	c.Decimals = new(big.Int).Set(v.Decimals)
		//	provider.tokenMarketCaps[c.WebsiteSlug] = c
		//} else {
		//	c := &CoinMarketCap{}
		//	c.Address = v.Protocol
		//	c.WebsiteSlug = v.Source
		//	c.Name = v.Symbol
		//	c.Symbol = v.Symbol
		//	c.Decimals = new(big.Int).Set(v.Decimals)
		//	provider.tokenMarketCaps[c.Address] = c
		//	provider.slugToAddress[strings.ToUpper(c.WebsiteSlug)] = c.Address
		//}
	}

	if err := provider.syncMarketCapFromRedis(); nil != err {
		log.Fatalf("can't sync marketcap with error:%s", err.Error())
	}

	return provider
}

func (p *CapProvider_CoinMarketCap) IsOrderValueDust(state *types.OrderState) bool {
	remainedAmountS, _ := state.RemainedAmount()

	remainedValue := new(big.Rat)
	remainedValue, _ = p.LegalCurrencyValue(state.RawOrder.TokenS, remainedAmountS)

	return p.IsValueDusted(remainedValue)
}

func (p *CapProvider_CoinMarketCap) IsValueDusted(value *big.Rat) bool {
	return p.dustValue.Cmp(value) > 0
}






