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

package test

import (
	"fmt"
	"github.com/Loopring/relay-cluster/accountmanager"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/node"
	ordermanager "github.com/Loopring/relay-cluster/ordermanager/manager"
	"github.com/Loopring/relay-cluster/ordermanager/viewer"
	"github.com/Loopring/relay-cluster/txmanager"
	"github.com/Loopring/relay-cluster/usermanager"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/crypto"
	"github.com/Loopring/relay-lib/eth/abi"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/naoina/toml"
	"math/big"
	"os"
	"strings"
	"time"
)

type AccountEntity struct {
	Address    common.Address
	Passphrase string
}

type TestEntity struct {
	Tokens          map[string]common.Address
	Accounts        []AccountEntity
	Creator         AccountEntity
	KeystoreDir     string
	AllowanceAmount int64
	PrivateKey      crypto.EthPrivateKeyCrypto
}

const (
	Version      = "v1.5.1"
	DebugFile    = "debug.toml"
	KeystorePath = "/Users/fukun/projects/gohome/src/github.com/Loopring/relay-cluster/ks_dir"
)

var (
	cfg           *node.GlobalConfig
	rds           *dao.RdsService
	entity        *TestEntity
	orderAccounts = []accounts.Account{}
	creator       accounts.Account
	protocol      common.Address
	delegate      common.Address
	Path          string
)

func init() {
	Path = strings.TrimSuffix(os.Getenv("GOPATH"), "/") + "/src/github.com/Loopring/relay-cluster/config/" + DebugFile

	cfg = loadConfig()
	util.Initialize(&cfg.Market)
	rds = dao.NewDb(&cfg.Mysql)
	cache.NewCache(cfg.Redis)
	entity = loadTestData()

	txmanager.NewTxView(rds)
	accessor.Initialize(cfg.Accessor)
	loopringaccessor.Initialize(cfg.LoopringProtocol)
	unlockAccounts()
	protocol = common.HexToAddress(cfg.LoopringProtocol.Address[Version])
	delegate = loopringaccessor.ProtocolAddresses()[protocol].DelegateAddress
}

func loadConfig() *node.GlobalConfig {
	c := node.LoadConfig(Path)
	log.Initialize(c.Log)

	return c
}

func LoadConfig() *node.GlobalConfig {
	c := node.LoadConfig(Path)
	log.Initialize(c.Log)

	return c
}

func LoadTestData() *TestEntity {
	return loadTestData()
}

func loadTestData() *TestEntity {
	e := new(TestEntity)

	type Account struct {
		Address    string
		Passphrase string
	}

	type AuthKey struct {
		Address common.Address
		Privkey string
	}

	type TestData struct {
		Accounts        []Account
		Creator         Account
		AllowanceAmount int64
		Auth            AuthKey
	}

	file := strings.TrimSuffix(os.Getenv("GOPATH"), "/") + "/src/github.com/Loopring/relay-cluster/test/testdata.toml"
	io, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer io.Close()

	var testData TestData
	if err := toml.NewDecoder(io).Decode(&testData); err != nil {
		log.Fatalf(err.Error())
	}

	e.Accounts = make([]AccountEntity, 0)
	for _, v := range testData.Accounts {
		var acc AccountEntity
		acc.Address = common.HexToAddress(v.Address)
		acc.Passphrase = v.Passphrase
		e.Accounts = append(e.Accounts, acc)
	}

	e.Tokens = make(map[string]common.Address)
	for symbol, token := range util.AllTokens {
		e.Tokens[symbol] = token.Protocol
	}

	e.Creator = AccountEntity{Address: common.HexToAddress(testData.Creator.Address), Passphrase: testData.Creator.Passphrase}
	e.KeystoreDir = KeystorePath
	e.AllowanceAmount = testData.AllowanceAmount

	e.PrivateKey, _ = crypto.NewPrivateKeyCrypto(false, testData.Auth.Privkey)
	return e
}

func unlockAccounts() {
	ks := keystore.NewKeyStore(KeystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	c := crypto.NewKSCrypto(false, ks)
	crypto.Initialize(c)

	creator = accounts.Account{Address: entity.Creator.Address}
	if err := ks.Unlock(creator, entity.Creator.Passphrase); err != nil {
		fmt.Printf(err.Error())
	}

	for _, accTmp := range entity.Accounts {
		account := accounts.Account{Address: accTmp.Address}
		orderAccounts = append(orderAccounts, account)
		if err := ks.Unlock(account, accTmp.Passphrase); nil != err {
			log.Fatalf("unlock account:%s error:%s", accTmp.Address.Hex(), err.Error())
		} else {
			log.Debugf("unlocked:%s", accTmp.Address.Hex())
		}
	}
}

func Rds() *dao.RdsService       { return rds }
func Cfg() *node.GlobalConfig    { return cfg }
func Entity() *TestEntity        { return entity }
func Protocol() common.Address   { return protocol }
func Delegate() common.Address   { return delegate }
func TokenRegisterAbi() *abi.ABI { return loopringaccessor.TokenRegistryAbi() }
func DelegateAbi() *abi.ABI      { return loopringaccessor.DelegateAbi() }
func Erc20Abi() *abi.ABI         { return loopringaccessor.Erc20Abi() }
func WethAbi() *abi.ABI          { return loopringaccessor.WethAbi() }
func LprAbi() *abi.ABI           { return loopringaccessor.ProtocolImplAbi() }

func TokenRegisterAddress() common.Address {
	return loopringaccessor.ProtocolAddresses()[protocol].TokenRegistryAddress
}
func DelegateAddress() common.Address {
	return loopringaccessor.ProtocolAddresses()[protocol].DelegateAddress
}
func LrcAddress() common.Address {
	return loopringaccessor.ProtocolAddresses()[protocol].LrcTokenAddress
}

func GenerateOrderManager() *ordermanager.OrderManagerImpl {
	mc := GenerateMarketCap()
	ob := ordermanager.NewOrderManager(&cfg.OrderManager, rds, mc, cfg.Kafka.Brokers)
	return ob
}

func GenerateOrderView() *viewer.OrderViewerImpl {
	mc := GenerateMarketCap()
	um := GenerateUserManager()
	ov := viewer.NewOrderViewer(&cfg.OrderManager, rds, mc, um)
	return ov
}

func GenerateAccountManager() accountmanager.AccountManager {
	return accountmanager.Initialize(&cfg.AccountManager, cfg.Kafka.Brokers)
}

func GenerateUserManager() *usermanager.UserManagerImpl {
	return usermanager.NewUserManager(&cfg.UserManager, rds)
}

func GenerateMarketCap() *marketcap.CapProvider_CoinMarketCap {
	return marketcap.NewMarketCapProvider(&cfg.MarketCap)
}

func CreateOrder(tokenS, tokenB, owner common.Address, amountS, amountB, lrcFee *big.Int) *types.Order {
	var (
		order types.Order
		state types.OrderState
		model dao.Order
	)
	order.Protocol = protocol
	order.DelegateAddress = delegate
	order.TokenS = tokenS
	order.TokenB = tokenB
	order.AmountS = amountS
	order.AmountB = amountB
	order.ValidSince = big.NewInt(time.Now().Unix())
	order.ValidUntil = big.NewInt(time.Now().Unix() + 8640000)
	order.LrcFee = lrcFee
	order.BuyNoMoreThanAmountB = false
	order.MarginSplitPercentage = 0
	order.Owner = owner
	order.PowNonce = 1
	order.AuthPrivateKey = entity.PrivateKey
	order.AuthAddr = order.AuthPrivateKey.Address()
	order.WalletAddress = owner
	order.Hash = order.GenerateHash()
	order.GeneratePrice()
	if err := order.GenerateAndSetSignature(owner); nil != err {
		log.Fatalf(err.Error())
	}
	market, err := util.WrapMarketByAddress(order.TokenB.Hex(), order.TokenS.Hex())
	if err != nil {
		log.Fatalf("get market error:%s", err.Error())
	}
	order.Market = market
	order.OrderType = "market_order"
	order.Side = util.GetSide(order.TokenS.Hex(), order.TokenB.Hex())

	state.RawOrder = order
	state.DealtAmountS = big.NewInt(0)
	state.DealtAmountB = big.NewInt(0)
	state.SplitAmountS = big.NewInt(0)
	state.SplitAmountB = big.NewInt(0)
	state.CancelledAmountB = big.NewInt(0)
	state.CancelledAmountS = big.NewInt(0)
	state.UpdatedBlock = big.NewInt(0)
	state.Status = types.ORDER_NEW

	model.ConvertDown(&state)

	rds.Add(&model)

	return &order
}

func getCallArg(a *abi.ABI, protocol common.Address, methodName string, args ...interface{}) *ethtyp.CallArg {
	if callData, err := a.Pack(methodName, args...); nil != err {
		panic(err)
	} else {
		arg := ethtyp.CallArg{}
		arg.From = protocol
		arg.To = protocol
		arg.Data = common.ToHex(callData)
		return &arg
	}
}

func PrepareTestData() {
	// name registry
	// nameRegistryAbi := ethaccessor.nam

	//delegate registry
	delegateAbi := loopringaccessor.DelegateAbi()
	delegateAddress := delegate
	var res types.Big
	if err := accessor.Call(&res, getCallArg(delegateAbi, delegateAddress, "isAddressAuthorized", protocol), "latest"); nil != err {
		log.Errorf("err:%s", err.Error())
	} else {
		if res.Int() <= 0 {
			delegateCallMethod := accessor.ContractSendTransactionMethod("latest", delegateAbi, delegateAddress)
			if hash, _, err := delegateCallMethod(creator.Address, "authorizeAddress", big.NewInt(106762), big.NewInt(21000000000), nil, protocol); nil != err {
				log.Errorf("delegate add version error:%s", err.Error())
			} else {
				log.Infof("delegate add version hash:%s", hash)
			}
		} else {
			log.Infof("delegate had added this version")
		}
	}

	log.Infof("tokenregistry")
	//tokenregistry
	tokenRegisterAbi := loopringaccessor.TokenRegistryAbi()
	tokenRegisterAddress := loopringaccessor.ProtocolAddresses()[protocol].TokenRegistryAddress
	for symbol, tokenAddr := range entity.Tokens {
		log.Infof("token:%s addr:%s", symbol, tokenAddr.Hex())
		callMethod := accessor.ContractCallMethod(tokenRegisterAbi, tokenRegisterAddress)
		var res types.Big
		if err := callMethod(&res, "isTokenRegistered", "latest", tokenAddr); nil != err {
			log.Errorf("err:%s", err.Error())
		} else {
			if res.Int() <= 0 {
				registryMethod := accessor.ContractSendTransactionMethod("latest", tokenRegisterAbi, tokenRegisterAddress)
				if hash, _, err := registryMethod(creator.Address, "registerToken", big.NewInt(106762), big.NewInt(21000000000), nil, tokenAddr, symbol); nil != err {
					log.Errorf("token registry error:%s", err.Error())
				} else {
					log.Infof("token registry hash:%s", hash)
				}
			} else {
				log.Infof("token %s had registered, res:%s", res.BigInt().String())
			}
		}
	}

	//approve
	for _, tokenAddr := range entity.Tokens {
		erc20SendMethod := accessor.ContractSendTransactionMethod("latest", loopringaccessor.Erc20Abi(), tokenAddr)
		for _, acc := range orderAccounts {
			approval := new(big.Int).Mul(big.NewInt(1e18), big.NewInt(1000000))
			if hash, _, err := erc20SendMethod(acc.Address, "approve", big.NewInt(106762), big.NewInt(21000000000), nil, delegateAddress, approval); nil != err {
				log.Errorf("token approve error:%s", err.Error())
			} else {
				log.Infof("token approve hash:%s", hash)
			}
		}
	}
}

func humanNumber(amount *big.Int) string {
	base := new(big.Int).SetInt64(1e18)
	ret := new(big.Rat).SetFrac(amount, base)
	return ret.FloatString(6)
}

func AllowanceToLoopring(tokens1 []common.Address, orderAccounts1 []accounts.Account) {
	if nil == tokens1 {
		for _, v := range entity.Tokens {
			tokens1 = append(tokens1, v)
		}
	}
	if nil == orderAccounts1 {
		for _, v := range orderAccounts {
			orderAccounts1 = append(orderAccounts1, v)
		}
	}

	for _, tokenAddr := range tokens1 {
		for _, account := range orderAccounts1 {
			if balance, err := loopringaccessor.Erc20Balance(tokenAddr, account.Address, "latest"); err != nil {
				log.Errorf("err:%s", err.Error())
			} else {
				log.Infof("token:%s, owner:%s, balance:%s", tokenAddr.Hex(), account.Address.Hex(), humanNumber(balance))
			}

			for _, impl := range loopringaccessor.ProtocolAddresses() {
				if allowance, err := loopringaccessor.Erc20Allowance(tokenAddr, account.Address, impl.DelegateAddress, "latest"); nil != err {
					log.Error(err.Error())
				} else {
					log.Infof("token:%s, owner:%s, spender:%s, allowance:%s", tokenAddr.Hex(), account.Address.Hex(), impl.DelegateAddress.Hex(), humanNumber(allowance))
				}
			}
		}
	}
}

//setbalance after deploy token by protocol
//不能设置weth
func SetTokenBalances() {
	dummyTokenAbiStr := `[{"constant":true,"inputs":[],"name":"mintingFinished","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_amount","type":"uint256"}],"name":"mint","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_subtractedValue","type":"uint256"}],"name":"decreaseApproval","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"finishMinting","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_addedValue","type":"uint256"}],"name":"increaseApproval","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_target","type":"address"},{"name":"_value","type":"uint256"}],"name":"setBalance","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[{"name":"_name","type":"string"},{"name":"_symbol","type":"string"},{"name":"_decimals","type":"uint8"},{"name":"_totalSupply","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"amount","type":"uint256"}],"name":"Mint","type":"event"},{"anonymous":false,"inputs":[],"name":"MintFinished","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"previousOwner","type":"address"},{"indexed":true,"name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`
	dummyTokenAbi := &abi.ABI{}
	dummyTokenAbi.UnmarshalJSON([]byte(dummyTokenAbiStr))

	sender := accounts.Account{Address: entity.Creator.Address}
	amount := new(big.Int).Mul(big.NewInt(1e18), big.NewInt(2000000))
	//wethAmount, _ := new(big.Int).SetString("79992767978000000000", 0)

	// deposit weth
	//wethToken := entity.Tokens["WETH"]
	//for _, v := range entity.Accounts {
	//	owner := accounts.Account{Address: v.Address}
	//	sendTransactionMethod := ethaccessor.ContractSendTransactionMethod("latest", ethaccessor.WethAbi(), wethToken)
	//	hash, err := sendTransactionMethod(owner, "deposit", nil, nil, wethAmount)
	//	if nil != err {
	//		log.Fatalf("call method weth-deposit error:%s", err.Error())
	//	} else {
	//		log.Debugf("weth-deposit txhash:%s", hash)
	//	}
	//}

	// other token set balance
	for symbol, tokenAddress := range entity.Tokens {
		if symbol == "WETH" {
			continue
		}
		sendTransactionMethod := accessor.ContractSendTransactionMethod("latest", dummyTokenAbi, tokenAddress)
		for _, acc := range orderAccounts {
			if balance, err := loopringaccessor.Erc20Balance(tokenAddress, acc.Address, "latest"); nil != err {
				fmt.Errorf(err.Error())
			} else if balance.Cmp(big.NewInt(int64(0))) <= 0 {
				hash, _, err := sendTransactionMethod(sender.Address, "setBalance", big.NewInt(106762), big.NewInt(21000000000), nil, acc.Address, amount)
				if nil != err {
					fmt.Errorf(err.Error())
				}
				fmt.Printf("sendhash:%s", hash)
			} else {
				fmt.Printf("tokenAddress:%s, useraddress:%s, balance:%s", tokenAddress.Hex(), acc.Address.Hex(), balance.String())
			}
		}
	}
}

// 给lrc，rdn等dummy合约支持的代币充值
func SetTokenBalance(account, tokenAddress common.Address, amount *big.Int) {
	dummyTokenAbi := &abi.ABI{}
	dummyTokenAbiStr := `[{"constant":true,"inputs":[],"name":"mintingFinished","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_amount","type":"uint256"}],"name":"mint","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_subtractedValue","type":"uint256"}],"name":"decreaseApproval","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"finishMinting","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_addedValue","type":"uint256"}],"name":"increaseApproval","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_target","type":"address"},{"name":"_value","type":"uint256"}],"name":"setBalance","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[{"name":"_name","type":"string"},{"name":"_symbol","type":"string"},{"name":"_decimals","type":"uint8"},{"name":"_totalSupply","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"amount","type":"uint256"}],"name":"Mint","type":"event"},{"anonymous":false,"inputs":[],"name":"MintFinished","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"previousOwner","type":"address"},{"indexed":true,"name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`
	dummyTokenAbi.UnmarshalJSON([]byte(dummyTokenAbiStr))

	sender := accounts.Account{Address: entity.Creator.Address}
	sendTransactionMethod := accessor.ContractSendTransactionMethod("latest", dummyTokenAbi, tokenAddress)

	hash, _, err := sendTransactionMethod(sender.Address, "setBalance", big.NewInt(1000000), big.NewInt(21000000000), nil, account, amount)
	if nil != err {
		fmt.Errorf(err.Error())
	}
	fmt.Printf("sendhash:%s", hash)
}
