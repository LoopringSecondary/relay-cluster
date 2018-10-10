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

package marketutil

import (
	"encoding/json"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/zklock"
	"github.com/ethereum/go-ethereum/common"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"math/big"
	"strings"
	"time"
)

type CustomToken struct {
	Address  common.Address `json:"address"`
	Symbol   string         `json:"symbol"`
	Decimals *big.Int       `json:"decimals"`
	Source   string         `json:"source"`
}

var localCache *gocache.Cache

const customTokensPreKey = "CTPK_"
const allCustomTokens = "ALLCT"
const zkLockCustomToken = "ZKLOCKCT"

func GetCustomTokenList(address common.Address) (tokens map[string]CustomToken, err error) {
	return getCustomTokenList(buildCacheKey(address))
}

func GetAllCustomTokenList() (tokens map[string]CustomToken, err error) {
	return getCustomTokenList(allCustomTokens)
}

func getCustomTokenList(key string) (tokens map[string]CustomToken, err error) {

	tokens = make(map[string]CustomToken)

	if localCache == nil {
		localCache = gocache.New(5*time.Second, 5*time.Minute)
	}

	tokensInLocal, ok := localCache.Get(key)
	if ok {
		return tokensInLocal.(map[string]CustomToken), err
	}

	customTokens, err := GetCustomTokensFromRedis(key)
	for _, ct := range customTokens {
		tokens[ct.Symbol] = ct
	}

	for _, v := range AllTokens {
		c := CustomToken{Address: v.Protocol, Symbol: v.Symbol, Decimals: v.Decimals}
		tokens[v.Symbol] = c
	}

	localCache.Set(key, tokens, 5*time.Second)
	return tokens, err
}

func GetCustomTokensFromRedis(key string) (customTokens map[string]CustomToken, err error) {
	tokenBytes, err := cache.Get(key)
	if err != nil || len(tokenBytes) == 0 {
		return make(map[string]CustomToken), nil
	}

	err = json.Unmarshal(tokenBytes, &customTokens)
	if err != nil {
		return customTokens, err
	}
	return customTokens, err
}

func setTokenToRedis(key string, token CustomToken) (err error) {
	err = zklock.TryLock(zkLockCustomToken)
	if err == nil {
		getAndSetCustomToken(key, token)
		getAndSetCustomToken(allCustomTokens, token)
	}

	defer zklock.ReleaseLock(zkLockCustomToken)
	return err
}

func getAndSetCustomToken(key string, token CustomToken) (err error) {
	ct, err := GetCustomTokensFromRedis(key)
	if err != nil {
		return err
	}

	ct[strings.ToUpper(token.Symbol)] = token

	ctByte, err := json.Marshal(ct)
	if err != nil {
		return err
	}
	return cache.Set(key, ctByte, -1)
}

func AddToken(address common.Address, token CustomToken) error {

	tokens, err := GetCustomTokenList(address)
	if err != nil {
		return err
	}

	token.Symbol = strings.ToUpper(token.Symbol)

	_, ok := tokens[token.Symbol]
	if ok {
		return errors.New("same symbol exist")
	}

	if hadRegistedInner(tokens, token.Address) {
		return errors.New("same address was registed")
	}

	decimals, err := loopringaccessor.Erc20Decimals(token.Address, "latest")
	if err == nil {
		token.Decimals = decimals
	} else {
		log.Errorf("get decimal failed from address : %s", token.Address.Hex())
	}

	return setTokenToRedis(buildCacheKey(address), token)
}

func DeleteToken(address common.Address, symbol string) error {

	tokens, err := GetCustomTokenList(address)
	if err != nil {
		return err
	}

	symbol = strings.ToUpper(symbol)

	_, ok := tokens[symbol]
	if !ok {
		return errors.New("not exist token symbol")
	}

	delete(tokens, symbol)

	ctByte, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	return cache.Set(buildCacheKey(address), ctByte, -1)
}

func HadRegistedByAddress(address common.Address, token common.Address) bool {
	tokenMap, err := GetCustomTokenList(address)
	if err != nil {
		return false
	}
	return hadRegistedInner(tokenMap, token)
}

func HadRegisted(token common.Address) bool {
	tokenMap, err := GetAllCustomTokenList()
	if err != nil {
		return false
	}
	return hadRegistedInner(tokenMap, token)
}

func hadRegistedInner(tokenMap map[string]CustomToken, address common.Address) bool {
	for _, v := range tokenMap {
		if strings.ToUpper(v.Address.Hex()) == strings.ToUpper(address.Hex()) {
			return true
		}
	}
	return false
}

func AddressToSymbol(address, token common.Address) (symbol string, err error) {

	tokens, err := GetCustomTokenList(address)
	if err != nil {
		return symbol, err
	}

	for k, v := range tokens {
		if strings.ToLower(token.Hex()) == strings.ToLower(v.Address.Hex()) {
			return k, nil
		}
	}
	return symbol, errors.New("address had not register this token")
}

func buildCacheKey(address common.Address) string {
	return customTokensPreKey + strings.ToUpper(address.Hex())
}
