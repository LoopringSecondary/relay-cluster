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

package accountmanager

import (
	"errors"
	rcache "github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var accManager *AccountManager

func IsInit() bool {
	return nil != accManager
}

func GetBalanceWithSymbolResult(owner common.Address) (map[string]*big.Int, error) {
	accountBalances := AccountBalances{}
	accountBalances.Owner = owner
	accountBalances.Balances = make(map[common.Address]Balance)

	res := make(map[string]*big.Int)
	//err := accountBalances.getOrSave(common.HexToAddress("0x1fa02762bd046abd30f5bf3513f9347d5e6b4257"), common.HexToAddress("0x"), common.HexToAddress("0x3cbcee9ff904ee0351b0ff2c05e08e860c94a5ea"))
	err := accountBalances.getOrSave(accManager.cacheDuration)

	if nil == err {
		for tokenAddr, balance := range accountBalances.Balances {
			symbol := ""
			if types.IsZeroAddress(tokenAddr) {
				symbol = "ETH"
			} else {
				symbol = marketutil.AddressToAlias(tokenAddr.Hex())
			}
			res[symbol] = balance.Balance.BigInt()
		}
	}

	return res, err
}

func GetAllowanceWithSymbolResult(owner, spender common.Address) (map[string]*big.Int, error) {
	accountAllowances := &AccountAllowances{}
	accountAllowances.Owner = owner
	accountAllowances.Allowances = make(map[common.Address]map[common.Address]Allowance)

	res := make(map[string]*big.Int)
	err := accountAllowances.getOrSave(accManager.cacheDuration, []common.Address{}, []common.Address{spender})

	if nil == err {
		for tokenAddr, allowances := range accountAllowances.Allowances {
			symbol := ""
			if types.IsZeroAddress(tokenAddr) {
				symbol = "ETH"
				res[symbol] = big.NewInt(int64(0))
			} else {
				symbol = marketutil.AddressToAlias(tokenAddr.Hex())
				if _, exists := allowances[spender]; !exists || nil == allowances[spender].Allowance {
					res[symbol] = big.NewInt(int64(0))
				} else {
					res[symbol] = allowances[spender].Allowance.BigInt()
				}
			}
		}
	} else {
		log.Errorf("err:%s", err.Error())
	}

	return res, err
}

func GetBalanceAndAllowance(owner, token, spender common.Address) (balance, allowance *big.Int, err error) {
	accountBalances := &AccountBalances{}
	accountBalances.Owner = owner
	accountBalances.Balances = make(map[common.Address]Balance)
	accountBalances.getOrSave(accManager.cacheDuration, token)
	balance = accountBalances.Balances[token].Balance.BigInt()

	accountAllowances := &AccountAllowances{}
	accountAllowances.Owner = owner
	accountAllowances.Allowances = make(map[common.Address]map[common.Address]Allowance)
	accountAllowances.getOrSave(accManager.cacheDuration, []common.Address{token}, []common.Address{spender})
	allowance = accountAllowances.Allowances[token][spender].Allowance.BigInt()

	return
}

func GetCutoff(contract, address string) (int, error) {
	var cutoffTime types.Big
	err := loopringaccessor.GetCutoff(&cutoffTime, common.HexToAddress(contract), common.HexToAddress(address), "latest")
	return int(cutoffTime.Int64()), err
}

func HasUnlocked(owner string) (exists bool, err error) {
	if !common.IsHexAddress(owner) {
		return false, errors.New("owner isn't a valid hex-address")
	}
	return rcache.Exists(unlockCacheKey(common.HexToAddress(owner)))
}

func InitializeView(options *AccountViewOptions) AccountManager {
	if nil != accManager {
		log.Fatalf("AccountManager has been init")
	}
	if err := isPackegeReady(); nil != err {
		log.Fatalf(err.Error())
	}

	accountManager := AccountManager{}
	if options.CacheDuration > 0 {
		accountManager.cacheDuration = options.CacheDuration
	} else {
		accountManager.cacheDuration = 3600 * 24 * 100
	}
	//accountManager.maxBlockLength = 3000
	b := &ChangedOfBlock{}
	b.cachedDuration = big.NewInt(int64(500))
	accountManager.block = b

	accManager = &accountManager
	return accountManager
}
