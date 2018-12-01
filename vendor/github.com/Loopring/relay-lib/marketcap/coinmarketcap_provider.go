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
	"encoding/json"
	"errors"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
	"time"
)

const (
	marketTickerCachePreKey = "COINMARKETCAP_TICKER_NEW_"
)

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
	baseUrl          string
	tokenMarketCaps  map[common.Address]*types.CMCTicker
	icoTokens        icoTokens
	notSupportTokens map[common.Address]bool
	slugToAddress    map[string]common.Address
	currency         string
	duration         int
	dustValue        *big.Rat
	stopFuncs        []func()
}

func (p *CapProvider_CoinMarketCap) LegalCurrencyValue(tokenAddress common.Address, amount *big.Rat) (*big.Rat, error) {
	return p.LegalCurrencyValueByCurrency(tokenAddress, amount, p.currency)
}

func (p *CapProvider_CoinMarketCap) LegalCurrencyValueOfEth(amount *big.Rat) (*big.Rat, error) {
	tokenAddress := util.AllTokens["WETH"].Protocol
	return p.LegalCurrencyValueByCurrency(tokenAddress, amount, p.currency)
}

func (p *CapProvider_CoinMarketCap) LegalCurrencyValueByCurrency(tokenAddress common.Address, amount *big.Rat, currencyStr string) (*big.Rat, error) {
	if _, exists := p.notSupportTokens[tokenAddress]; exists {
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

func (p *CapProvider_CoinMarketCap) GetMarketCapByCurrency(tokenAddress common.Address, currencyStr string) (*big.Rat, error) {
	if _, exists := p.notSupportTokens[tokenAddress]; exists {
		return big.NewRat(int64(0), int64(1)), nil
	} else if c, exists := p.tokenMarketCaps[tokenAddress]; exists {
		var v *big.Rat
		v = new(big.Rat).SetFloat64(c.Price)
		if v == nil {
			return nil, errors.New("tokenCap is nil")
		} else {
			return v, nil
		}
	} else {
		tickerMap, _ := getTickersFromRedis(currencyStr)
		var v *big.Rat
		webSiteSlug := util.AddressToSource(tokenAddress.Hex())
		if c, exists := tickerMap[webSiteSlug]; exists {
			v = new(big.Rat).SetFloat64(c.Price)
		}
		if v == nil {
			return nil, errors.New("tokenCap is nil")
		} else {
			return v, nil
		}
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

}

func (p *CapProvider_CoinMarketCap) syncMarketCapFromRedis() error {

	notSupportTokens := make(map[common.Address]bool)
	tickerMap, _ := getTickersFromRedis(p.currency)
	for tokenAddr, c1 := range p.tokenMarketCaps {
		if cmcTicker, exists := tickerMap[c1.WebsiteSlug]; exists {
			p.tokenMarketCaps[tokenAddr].Price = cmcTicker.Price
			p.tokenMarketCaps[tokenAddr].CmcRank = cmcTicker.CmcRank
			p.tokenMarketCaps[tokenAddr].TokenName = cmcTicker.TokenName
			p.tokenMarketCaps[tokenAddr].LastUpdated = cmcTicker.LastUpdated
			p.tokenMarketCaps[tokenAddr].PercentChange7D = cmcTicker.PercentChange7D
			p.tokenMarketCaps[tokenAddr].PercentChange24H = cmcTicker.PercentChange24H
			p.tokenMarketCaps[tokenAddr].PercentChange1H = cmcTicker.PercentChange1H
			p.tokenMarketCaps[tokenAddr].MarketCap = cmcTicker.MarketCap
			p.tokenMarketCaps[tokenAddr].Market = cmcTicker.Market
			p.tokenMarketCaps[tokenAddr].MaxSupply = cmcTicker.MaxSupply
			p.tokenMarketCaps[tokenAddr].CirculatingSupply = cmcTicker.CirculatingSupply
			p.tokenMarketCaps[tokenAddr].TotalSupply = cmcTicker.TotalSupply
			p.tokenMarketCaps[tokenAddr].Volume24H = cmcTicker.Volume24H

		} else {
			notSupportTokens[tokenAddr] = true
		}
	}
	p.notSupportTokens = notSupportTokens

	return nil
}

func (p *CapProvider_CoinMarketCap) IsSupport(token common.Address) bool {
	_, exists := p.notSupportTokens[token]
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
	provider.tokenMarketCaps = make(map[common.Address]*types.CMCTicker)
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
		c := &types.CMCTicker{}
		c.Address = v.Protocol
		c.WebsiteSlug = v.Source
		c.Symbol = v.Symbol
		c.Decimals = v.Decimals
		provider.tokenMarketCaps[c.Address] = c
		provider.slugToAddress[strings.ToUpper(c.WebsiteSlug)] = c.Address

	}

	if err := provider.syncMarketCapFromRedis(); nil != err {
		log.Fatalf("can't sync marketcap with error:%s", err.Error())
	}

	return provider
}

func (p *CapProvider_CoinMarketCap) IsOrderValueDust(state *types.OrderState) bool {
	remainedAmountS, remainedAmountB := state.RemainedAmount()

	remainedValue := new(big.Rat)
	if p.IsSupport(state.RawOrder.TokenS) {
		remainedValue, _ = p.LegalCurrencyValue(state.RawOrder.TokenS, remainedAmountS)
	} else {
		remainedValue, _ = p.LegalCurrencyValue(state.RawOrder.TokenB, remainedAmountB)
	}

	return p.IsValueDusted(remainedValue)
}

func (p *CapProvider_CoinMarketCap) IsValueDusted(value *big.Rat) bool {
	return p.dustValue.Cmp(value) > 0
}

func getTickersFromRedis(market string) (tickerMap map[string]*types.CMCTicker, err error) {
	tickerMap = make(map[string]*types.CMCTicker)
	if ticketData, err := cache.HGetAll(marketTickerCachePreKey + market); nil != err {
		log.Debug(">>>>>>>> get ticker data from redis error " + err.Error())
		return tickerMap, err
	} else {
		if len(ticketData) > 0 {
			idx := 0
			for idx < len(ticketData) {
				ticker := &types.CMCTicker{}
				if err := json.Unmarshal(ticketData[idx+1], ticker); nil != err {
					log.Errorf("get marketcap of ticker data err:%s", err.Error())
					return nil, err
				} else {
					tickerMap[string(ticketData[idx])] = ticker
				}
				idx = idx + 2
			}
		}
	}

	return tickerMap, nil
}
