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
	"encoding/json"
	"errors"
	rcache "github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
)

const (
	UnlockedPrefix    = "unlock_"
	BalancePrefix     = "balance_"
	AllowancePrefix   = "allowance_"
	CustomTokenPrefix = "customtoken_"
)

type AccountBase struct {
	Owner        common.Address
	CustomTokens []types.Token
}

type Balance struct {
	LastBlock *types.Big `json:"last_block"`
	Balance   *types.Big `json:"balance"`
}

type Allowance struct {
	LastBlock *types.Big `json:"last_block"`
	Allowance *types.Big `json:"allowance"`
}

type AccountBalances struct {
	AccountBase
	Balances map[common.Address]Balance
}

func balanceCacheKey(owner common.Address) string {
	return BalancePrefix + strings.ToLower(owner.Hex())
}

func unlockCacheKey(owner common.Address) string {
	return UnlockedPrefix + strings.ToLower(owner.Hex())
}

func balanceCacheField(token common.Address) []byte {
	return []byte(strings.ToLower(token.Hex()))
}

func parseCacheField(field []byte) common.Address {
	return common.HexToAddress(string(field))
}

//todo:tokens
func (b AccountBalances) batchReqs(tokens ...common.Address) loopringaccessor.BatchBalanceReqs {
	reqs := loopringaccessor.BatchBalanceReqs{}
	for _, token := range util.AllTokens {
		req := &loopringaccessor.BatchBalanceReq{}
		req.BlockParameter = "latest"
		req.Token = token.Protocol
		req.Owner = b.Owner
		reqs = append(reqs, req)
	}
	for _, token := range b.CustomTokens {
		req := &loopringaccessor.BatchBalanceReq{}
		req.BlockParameter = "latest"
		req.Token = token.Protocol
		req.Owner = b.Owner
		reqs = append(reqs, req)
	}
	req := &loopringaccessor.BatchBalanceReq{}
	req.BlockParameter = "latest"
	req.Token = common.HexToAddress("0x")
	req.Owner = b.Owner
	reqs = append(reqs, req)
	return reqs
}

func (accountBalances AccountBalances) save(ttl int64) error {
	data := [][]byte{}
	for token, balance := range accountBalances.Balances {
		//log.Debugf("balance owner:%s, token:%s, amount:", accountBalances.Owner.Hex(), token.Hex(), balance.Balance.BigInt().String())
		if balanceData, err := json.Marshal(balance); nil == err {
			data = append(data, balanceCacheField(token), balanceData)
		} else {
			log.Errorf("accountmanager er:%s", err.Error())
		}
	}
	err := rcache.HMSet(balanceCacheKey(accountBalances.Owner), ttl, data...)
	return err
}

func (accountBalances AccountBalances) applyData(cachedFieldData, balanceData []byte) error {
	if len(balanceData) <= 0 {
		return errors.New("not in cache")
	} else {
		tokenAddress := parseCacheField(cachedFieldData)
		balance := Balance{}
		if err := json.Unmarshal(balanceData, &balance); nil != err {
			log.Errorf("accountmanager, syncFromCache err:%s", err.Error())
			return err
		} else {
			accountBalances.Balances[tokenAddress] = balance
		}
		return nil
	}
}

func (accountBalances AccountBalances) syncFromCache(tokens ...common.Address) error {
	missedTokens := []common.Address{}
	if len(tokens) > 0 {
		tokensBytes := [][]byte{}
		for _, token := range tokens {
			tokensBytes = append(tokensBytes, balanceCacheField(token))
		}
		if balancesData, err := rcache.HMGet(balanceCacheKey(accountBalances.Owner), tokensBytes...); nil != err {
			return err
		} else {
			if len(balancesData) > 0 {
				for idx, data := range balancesData {
					if len(data) > 0 {
						if err := accountBalances.applyData(tokensBytes[idx], data); nil != err {
							missedTokens = append(missedTokens, tokens[idx])
						}
					} else {
						missedTokens = append(missedTokens, tokens[idx])
						return errors.New("this address not in cache")
					}
				}
			} else {
				return errors.New("this address not in cache")
			}
		}
	} else {
		if balancesData, err := rcache.HGetAll(balanceCacheKey(accountBalances.Owner)); nil != err {
			return err
		} else {
			if len(balancesData) > 0 {
				idx := 0
				for idx < len(balancesData) {
					if err := accountBalances.applyData(balancesData[idx], balancesData[idx+1]); nil != err {
						return err
					}
					idx = idx + 2
				}
			} else {
				return errors.New("this address not in cache")
			}
		}
	}

	return nil
}

func (accountBalances AccountBalances) syncFromEthNode(tokens ...common.Address) error {
	reqs := accountBalances.batchReqs(tokens...)
	if err := accessor.BatchCall("latest", []accessor.BatchReq{reqs}); nil != err {
		return err
	}
	for _, req := range reqs {
		if nil != req.BalanceErr {
			log.Errorf("get balance failed, owner:%s, token:%s, err:%s", req.Owner.Hex(), req.Token.Hex(), req.BalanceErr.Error())
		} else {
			balance := Balance{}
			balance.Balance = &req.Balance
			//balance.LastBlock =
			accountBalances.Balances[req.Token] = balance
		}
	}
	return nil
}

func (accountBalances AccountBalances) getOrSave(ttl int64, tokens ...common.Address) error {
	if err := accountBalances.syncFromCache(tokens...); nil != err {
		if err := accountBalances.syncFromEthNode(tokens...); nil != err {
			return err
		} else {
			go accountBalances.save(ttl)
		}
	}
	return nil
}

type AccountAllowances struct {
	AccountBase
	Allowances map[common.Address]map[common.Address]Allowance //token -> spender
}

func allowanceCacheKey(owner common.Address) string {
	return AllowancePrefix + strings.ToLower(owner.Hex())
}

func allowanceCacheField(token common.Address, spender common.Address) []byte {
	return []byte(strings.ToLower(token.Hex() + spender.Hex()))
}

func parseAllowanceCacheField(data []byte) (token common.Address, spender common.Address) {
	return common.HexToAddress(string(data[0:42])), common.HexToAddress(string(data[42:]))
}

//todo:tokens
func (accountAllowances *AccountAllowances) batchReqs(tokens, spenders []common.Address) loopringaccessor.BatchErc20AllowanceReqs {
	reqs := loopringaccessor.BatchErc20AllowanceReqs{}
	for _, v := range util.AllTokens {
		for _, impl := range loopringaccessor.ProtocolAddresses() {
			req := &loopringaccessor.BatchErc20AllowanceReq{}
			req.BlockParameter = "latest"
			req.Spender = impl.DelegateAddress
			req.Token = v.Protocol
			req.Owner = accountAllowances.Owner
			reqs = append(reqs, req)
		}
	}
	for _, v := range accountAllowances.CustomTokens {
		for _, impl := range loopringaccessor.ProtocolAddresses() {
			req := &loopringaccessor.BatchErc20AllowanceReq{}
			req.BlockParameter = "latest"
			req.Spender = impl.DelegateAddress
			req.Token = v.Protocol
			req.Owner = accountAllowances.Owner
			reqs = append(reqs, req)
		}
	}
	return reqs
}

func (accountAllowances *AccountAllowances) save(ttl int64) error {
	data := [][]byte{}
	for token, spenderMap := range accountAllowances.Allowances {
		for spender, allowance := range spenderMap {
			if allowanceData, err := json.Marshal(allowance); nil == err {
				data = append(data, allowanceCacheField(token, spender), allowanceData)
			} else {
				log.Errorf("accountmanager allowance.save err:%s", err.Error())
			}
		}
	}
	return rcache.HMSet(allowanceCacheKey(accountAllowances.Owner), ttl, data...)
}

func (accountAllowances *AccountAllowances) applyData(cacheFieldData, allowanceData []byte) error {
	if len(allowanceData) <= 0 {
		return errors.New("invalid allowanceData")
	} else {
		allowance := Allowance{}
		if err := json.Unmarshal(allowanceData, &allowance); nil != err {
			log.Errorf("accountmanager syncFromCache err:%s", err.Error())
			return err
		} else {
			token, spender := parseAllowanceCacheField(cacheFieldData)
			if _, exists := accountAllowances.Allowances[token]; !exists {
				accountAllowances.Allowances[token] = make(map[common.Address]Allowance)
			}
			accountAllowances.Allowances[token][spender] = allowance
		}
	}
	return nil
}

func generateAllowanceCahceFieldList(tokens, spenders []common.Address) [][]byte {
	fields := [][]byte{}
	for _, token := range tokens {
		if !types.IsZeroAddress(token) {
			for _, spender := range spenders {
				if !types.IsZeroAddress(spender) {
					fields = append(fields, allowanceCacheField(token, spender))
				}
			}
		}
	}
	return fields
}

func (accountAllowances *AccountAllowances) syncFromCache(tokens, spenders []common.Address) error {
	fields := generateAllowanceCahceFieldList(tokens, spenders)
	if len(fields) > 0 {
		if allowanceData, err := rcache.HMGet(allowanceCacheKey(accountAllowances.Owner), fields...); nil != err {
			return err
		} else {
			if len(allowanceData) > 0 {
				for idx, data := range allowanceData {
					if len(data) > 0 {
						if err := accountAllowances.applyData(fields[idx], data); nil != err {
							return err
						}
					} else {
						return errors.New("allowance of this address not in cache")
					}
				}
			} else {
				return errors.New("allowance of this address not in cache")
			}
		}
	} else {
		if allowanceData, err := rcache.HGetAll(allowanceCacheKey(accountAllowances.Owner)); nil != err {
			return err
		} else {
			if len(allowanceData) > 0 {
				i := 0
				for i < len(allowanceData) {
					if err := accountAllowances.applyData(allowanceData[i], allowanceData[i+1]); nil != err {
						return err
					}
					i = i + 2
				}
			} else {
				return errors.New("this address not in cache")
			}
		}
	}
	return nil
}

func (accountAllowances *AccountAllowances) syncFromEthNode(tokens, spenders []common.Address) error {
	reqs := accountAllowances.batchReqs(tokens, spenders)
	if err := accessor.BatchCall("latest", []accessor.BatchReq{reqs}); nil != err {
		return err
	}
	for _, req := range reqs {
		if nil != req.AllowanceErr {
			log.Errorf("get balance failed, owner:%s, token:%s, err:%s", req.Owner.Hex(), req.Token.Hex(), req.AllowanceErr.Error())
		} else {
			allowance := Allowance{}
			allowance.Allowance = &req.Allowance
			//balance.LastBlock =
			if _, exists := accountAllowances.Allowances[req.Token]; !exists {
				accountAllowances.Allowances[req.Token] = make(map[common.Address]Allowance)
			}
			accountAllowances.Allowances[req.Token][req.Spender] = allowance
		}
	}

	return nil
}

func (accountAllowances *AccountAllowances) getOrSave(ttl int64, tokens, spenders []common.Address) error {
	if err := accountAllowances.syncFromCache(tokens, spenders); nil != err {
		if err := accountAllowances.syncFromEthNode(tokens, spenders); nil != err {
			return err
		} else {
			go accountAllowances.save(ttl)
		}
	}
	return nil
}

type ChangedOfBlock struct {
	currentBlockNumber *big.Int
	cachedDuration     *big.Int
}

func (b *ChangedOfBlock) saveBalanceKey(owner, token common.Address) error {
	err := rcache.SAdd(b.cacheBalanceKey(), int64(0), b.cacheBalanceField(owner, token))
	return err
}

func (b *ChangedOfBlock) cacheBalanceKey() string {
	if nil == b.currentBlockNumber {
		log.Error("b.currentBlockNumber is nil")
	}
	return "block_balance_" + b.currentBlockNumber.String()
}

func (b *ChangedOfBlock) cacheBalanceField(owner, token common.Address) []byte {
	return append(owner.Bytes(), token.Bytes()...)
}
func (b *ChangedOfBlock) parseCacheBalanceField(data []byte) (owner, token common.Address) {
	return common.BytesToAddress(data[0:20]), common.BytesToAddress(data[20:])
}

func (b *ChangedOfBlock) cacheAllowanceKey() string {
	if nil == b.currentBlockNumber {
		log.Error("b.currentBlockNumber is nil")
	}
	return "block_allowance_" + b.currentBlockNumber.String()
}

func (b *ChangedOfBlock) cacheAllowanceField(owner, token, spender common.Address) []byte {
	return append(append(owner.Bytes(), token.Bytes()...), spender.Bytes()...)
}

func (b *ChangedOfBlock) parseCacheAllowanceField(data []byte) (owner, token, spender common.Address) {
	return common.BytesToAddress(data[0:20]), common.BytesToAddress(data[20:40]), common.BytesToAddress(data[40:])
}

func (b *ChangedOfBlock) saveAllowanceKey(owner, token, spender common.Address) error {
	err := rcache.SAdd(b.cacheAllowanceKey(), int64(0), b.cacheAllowanceField(owner, token, spender))
	return err
}

func removeExpiredBlock(blockNumber, duration *big.Int) error {
	nb := &ChangedOfBlock{}
	nb.currentBlockNumber = new(big.Int)
	nb.currentBlockNumber.Sub(blockNumber, duration)
	log.Debugf("removeExpiredBlock cacheAllowanceKey ")

	if err := rcache.Del(nb.cacheAllowanceKey()); nil != err {
		log.Errorf("removeExpiredBlock cacheAllowanceKey err:%s", err.Error())
	}
	if err := rcache.Del(nb.cacheBalanceKey()); nil != err {
		log.Errorf("removeExpiredBlock cacheBalanceKey err:%s", err.Error())
	}
	return nil
}

func (b *ChangedOfBlock) syncAndSaveBalances() (map[common.Address]bool, error) {
	changedAddrs := make(map[common.Address]bool)
	reqs := b.batchBalanceReqs()
	if err := accessor.BatchCall("latest", []accessor.BatchReq{reqs}); nil != err {
		return changedAddrs, err
	}
	accounts := make(map[common.Address]*AccountBalances)
	for _, req := range reqs {
		if nil != req.BalanceErr {
			log.Errorf("get balance failed, owner:%s, token:%s, err:%s", req.Owner.Hex(), req.Token.Hex(), req.BalanceErr.Error())
		} else {
			if _, exists := accounts[req.Owner]; !exists {
				accounts[req.Owner] = &AccountBalances{}
				accounts[req.Owner].Owner = req.Owner
				accounts[req.Owner].Balances = make(map[common.Address]Balance)
			}
			balance := Balance{}
			balance.LastBlock = types.NewBigPtr(b.currentBlockNumber)
			balance.Balance = &req.Balance
			accounts[req.Owner].Balances[req.Token] = balance
		}
	}
	for _, balances := range accounts {
		balances.save(int64(0))
		changedAddrs[balances.Owner] = true
	}

	return changedAddrs, nil
}

func (b *ChangedOfBlock) batchBalanceReqs() loopringaccessor.BatchBalanceReqs {
	reqs := loopringaccessor.BatchBalanceReqs{}
	if balancesData, err := rcache.SMembers(b.cacheBalanceKey()); nil == err && len(balancesData) > 0 {
		for _, data := range balancesData {
			accountAddr, token := b.parseCacheBalanceField(data)
			//log.Debugf("1---batchBalanceReqsbatchBalanceReqsbatchBalanceReqs:%s,%s", accountAddr.Hex(), token.Hex())
			if exists, err := rcache.Exists(balanceCacheKey(accountAddr)); nil == err && exists {
				//log.Debugf("2---batchBalanceReqsbatchBalanceReqsbatchBalanceReqs:%s,%s", accountAddr.Hex(), token.Hex())
				if exists1, err1 := rcache.HExists(balanceCacheKey(accountAddr), balanceCacheField(token)); nil == err1 && exists1 {
					log.Debugf("3---batchBalanceReqsbatchBalanceReqsbatchBalanceReqs:%s,%s", accountAddr.Hex(), token.Hex())
					req := &loopringaccessor.BatchBalanceReq{}
					req.Owner = accountAddr
					req.Token = token
					req.BlockParameter = "latest"
					reqs = append(reqs, req)
				}
			}
		}
	}
	return reqs
}

func (b *ChangedOfBlock) batchAllowanceReqs() loopringaccessor.BatchErc20AllowanceReqs {
	reqs := loopringaccessor.BatchErc20AllowanceReqs{}
	if allowancesData, err := rcache.SMembers(b.cacheAllowanceKey()); nil == err && len(allowancesData) > 0 {
		for _, data := range allowancesData {
			owner, token, spender := b.parseCacheAllowanceField(data)
			//log.Debugf("1---batchAllowanceReqs owner:%s, t:%s, s:%s", owner.Hex(), token.Hex(), spender.Hex())
			if loopringaccessor.IsSpenderAddress(spender) {
				if exists, err := rcache.Exists(balanceCacheKey(owner)); nil == err && exists {
					//log.Debugf("2---batchAllowanceReqs owner:%s, t:%s, s:%s", owner.Hex(), token.Hex(), spender.Hex())
					if exists1, err1 := rcache.HExists(allowanceCacheKey(owner), allowanceCacheField(token, spender)); nil == err1 && exists1 {
						log.Debugf("3---batchAllowanceReqs owner:%s, t:%s, s:%s", owner.Hex(), token.Hex(), spender.Hex())
						req := &loopringaccessor.BatchErc20AllowanceReq{}
						req.BlockParameter = "latest"
						req.Spender = spender
						req.Token = token
						req.Owner = owner
						reqs = append(reqs, req)
					}
				}
			}

		}
	}
	return reqs
}

func (b *ChangedOfBlock) syncAndSaveAllowances() (map[common.Address]bool, error) {
	changedAddrs := make(map[common.Address]bool)

	reqs := b.batchAllowanceReqs()
	if err := accessor.BatchCall("latest", []accessor.BatchReq{reqs}); nil != err {
		return changedAddrs, err
	}
	accountAllowances := make(map[common.Address]*AccountAllowances)
	for _, req := range reqs {
		if nil != req.AllowanceErr {
			log.Errorf("get allowance failed, owner:%s, token:%s, err:%s", req.Owner.Hex(), req.Token.Hex(), req.AllowanceErr.Error())
		} else {
			if _, exists := accountAllowances[req.Owner]; !exists {
				accountAllowances[req.Owner] = &AccountAllowances{}
				accountAllowances[req.Owner].Owner = req.Owner
				accountAllowances[req.Owner].Allowances = make(map[common.Address]map[common.Address]Allowance)
			}
			allowance := Allowance{}
			allowance.LastBlock = types.NewBigPtr(b.currentBlockNumber)
			allowance.Allowance = &req.Allowance
			if _, exists := accountAllowances[req.Owner].Allowances[req.Token]; !exists {
				accountAllowances[req.Owner].Allowances[req.Token] = make(map[common.Address]Allowance)
			}
			accountAllowances[req.Owner].Allowances[req.Token][req.Spender] = allowance
		}
	}
	for _, allowances := range accountAllowances {
		allowances.save(int64(0))
		changedAddrs[allowances.Owner] = true
	}

	return changedAddrs, nil
}
