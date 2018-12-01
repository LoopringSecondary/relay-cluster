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
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
)

type LegalCurrency int

func StringToLegalCurrency(currency string) LegalCurrency {
	currency = strings.ToUpper(currency)
	switch currency {
	default:
		return CNY
	case "CNY":
		return CNY
	case "USD":
		return USD
	case "BTC":
		return BTC
	}
}

func LegalCurrencyToString(currency LegalCurrency) string {
	switch currency {
	default:
		return "CNY"
	case CNY:
		return "CNY"
	case USD:
		return "USD"
	case BTC:
		return "BTC"
	}
}

const (
	CNY LegalCurrency = iota
	USD
	EUR
	BTC
)

type MarketCapProvider interface {
	Start()
	Stop()

	LegalCurrencyValue(tokenAddress common.Address, amount *big.Rat) (*big.Rat, error)
	LegalCurrencyValueOfEth(amount *big.Rat) (*big.Rat, error)
	LegalCurrencyValueByCurrency(tokenAddress common.Address, amount *big.Rat, currencyStr string) (*big.Rat, error)
	GetMarketCap(tokenAddress common.Address) (*big.Rat, error)
	GetEthCap() (*big.Rat, error)
	GetMarketCapByCurrency(tokenAddress common.Address, currencyStr string) (*big.Rat, error)
	IsOrderValueDust(state *types.OrderState) bool
	IsValueDusted(value *big.Rat) bool
	IsSupport(token common.Address) bool
}

type CapProvider_LocalCap struct {
	tokenMarketCaps map[common.Address]*types.CurrencyMarketCap
	stopChan        chan bool
	stopFuncs       []func()
}

//todo:
func NewLocalCap() *CapProvider_LocalCap {
	localCap := &CapProvider_LocalCap{}
	localCap.stopChan = make(chan bool)
	localCap.stopFuncs = make([]func(), 0)
	localCap.tokenMarketCaps = make(map[common.Address]*types.CurrencyMarketCap)
	return localCap
}

func (cap *CapProvider_LocalCap) Start() {

	for _, marketStr := range util.AllMarkets {
		tokenAddress, _ := util.UnWrapToAddress(marketStr)
		token, _ := util.AddressToToken(tokenAddress)
		c := &types.CurrencyMarketCap{}
		c.Address = token.Protocol
		c.Id = token.Source
		c.Name = token.Symbol
		c.Symbol = token.Symbol
		c.Decimals = new(big.Int).Set(token.Decimals)
		cap.tokenMarketCaps[tokenAddress] = c
	}
	//if stopFunc,err := eventemitter.NewSerialWatcher(eventemitter.RingMined, cap.listenRingMinedEvent); nil != err {
	//	log.Debugf("err:%s", err.Error())
	//} else {
	//	cap.stopFuncs = append(cap.stopFuncs, stopFunc)
	//}
}

func (cap *CapProvider_LocalCap) Stop() {
	for _, stopFunc := range cap.stopFuncs {
		stopFunc()
	}
	cap.stopChan <- true
}

//func (cap *CapProvider_LocalCap) LegalCurrencyValue(tokenAddress common.Address, amount *big.Rat) (*big.Rat, error) {
//
//}
//
//func (cap *CapProvider_LocalCap) LegalCurrencyValueOfEth(amount *big.Rat) (*big.Rat, error) {
//
//}
//
//func (cap *CapProvider_LocalCap) LegalCurrencyValueByCurrency(tokenAddress common.Address, amount *big.Rat, currencyStr string) (*big.Rat, error) {
//
//}
//
//func (cap *CapProvider_LocalCap) GetMarketCap(tokenAddress common.Address) (*big.Rat, error) {
//
//}
//
//func (cap *CapProvider_LocalCap) GetEthCap() (*big.Rat, error) {
//
//}
//
//func (cap *CapProvider_LocalCap) GetMarketCapByCurrency(tokenAddress common.Address, currencyStr string) (*big.Rat, error) {
//
//}

type MixMarketCap struct {
	coinMarketProvider *CapProvider_CoinMarketCap
	localCap           *CapProvider_LocalCap
}

func (cap *MixMarketCap) Start() {
	cap.coinMarketProvider.Start()
	cap.localCap.Start()
}

func (cap *MixMarketCap) Stop() {
	cap.coinMarketProvider.Stop()
	cap.localCap.Stop()
}

func (cap *MixMarketCap) selectCap(tokenAddress common.Address) MarketCapProvider {
	return cap.coinMarketProvider
	//if _,exists := cap.coinMarketProvider.tokenMarketCaps[tokenAddress]; exists || types.IsZeroAddress(tokenAddress) {
	//	return cap.coinMarketProvider
	//} else {
	//	return cap.localCap
	//}
}

func (cap *MixMarketCap) LegalCurrencyValue(tokenAddress common.Address, amount *big.Rat) (*big.Rat, error) {
	return cap.selectCap(tokenAddress).LegalCurrencyValue(tokenAddress, amount)
}

func (cap *MixMarketCap) LegalCurrencyValueOfEth(amount *big.Rat) (*big.Rat, error) {
	return cap.selectCap(types.NilAddress).LegalCurrencyValueOfEth(amount)
}

func (cap *MixMarketCap) LegalCurrencyValueByCurrency(tokenAddress common.Address, amount *big.Rat, currencyStr string) (*big.Rat, error) {
	return cap.selectCap(tokenAddress).LegalCurrencyValueByCurrency(tokenAddress, amount, currencyStr)
}

func (cap *MixMarketCap) GetMarketCap(tokenAddress common.Address) (*big.Rat, error) {
	return cap.selectCap(tokenAddress).GetMarketCap(tokenAddress)
}

func (cap *MixMarketCap) GetEthCap() (*big.Rat, error) {
	return cap.selectCap(types.NilAddress).GetEthCap()
}

func (cap *MixMarketCap) GetMarketCapByCurrency(tokenAddress common.Address, currencyStr string) (*big.Rat, error) {
	return cap.selectCap(tokenAddress).GetMarketCapByCurrency(tokenAddress, currencyStr)
}
