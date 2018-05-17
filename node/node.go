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

package node

import (
	"sync"

	"fmt"
	"github.com/Loopring/relay-cluster/accountmanager"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/gateway"
	"github.com/Loopring/relay-cluster/market"
	"github.com/Loopring/relay-cluster/ordermanager"
	"github.com/Loopring/relay-cluster/txmanager"
	"github.com/Loopring/relay-cluster/usermanager"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/crypto"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/eth/gasprice_evaluator"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/zklock"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"go.uber.org/zap"
)

type Node struct {
	globalConfig *GlobalConfig
	rdsService   dao.RdsService
	//ipfsSubService    gateway.IPFSSubService
	orderManager      ordermanager.OrderManager
	userManager       usermanager.UserManager
	marketCapProvider marketcap.MarketCapProvider
	accountManager    accountmanager.AccountManager
	trendManager      market.TrendManager
	tickerCollector   market.CollectorImpl
	jsonRpcService    gateway.JsonrpcServiceImpl
	websocketService  gateway.WebsocketServiceImpl
	socketIOService   gateway.SocketIOServiceImpl
	walletService     gateway.WalletServiceImpl
	txManager         txmanager.TransactionManager

	stop   chan struct{}
	lock   sync.RWMutex
	logger *zap.Logger
}

func NewNode(logger *zap.Logger, globalConfig *GlobalConfig) *Node {
	n := &Node{}
	n.logger = logger
	n.globalConfig = globalConfig

	// register
	n.registerZklock()

	n.registerMysql()
	cache.NewCache(n.globalConfig.Redis)

	util.Initialize(&n.globalConfig.Market)
	n.registerMarketCap()
	n.registerAccessor()
	n.registerUserManager()
	n.registerOrderManager()
	n.registerAccountManager()
	n.registerGateway()
	n.registerCrypto(nil)

	n.registerTransactionManager()
	n.registerTrendManager()
	n.registerTickerCollector()
	n.registerWalletService()
	n.registerJsonRpcService()
	n.registerWebsocketService()
	n.registerSocketIOService()
	txmanager.NewTxView(n.rdsService)

	return n
}

func (n *Node) Start() {
	n.orderManager.Start()
	n.marketCapProvider.Start()
	n.accountManager.Start()
	n.txManager.Start()
	//gateway.NewJsonrpcService("8080").Start()
	fmt.Println("step in relay node start")
	n.tickerCollector.Start()
	go n.jsonRpcService.Start()
	//n.websocketService.Start()
	go n.socketIOService.Start()
	go gasprice_evaluator.InitGasPriceEvaluator()
}

func (n *Node) Wait() {
	n.lock.RLock()

	// TODO(fk): states should be judged

	stop := n.stop
	n.lock.RUnlock()

	<-stop
}

// todo
func (n *Node) Stop() {
}

func (n *Node) registerCrypto(ks *keystore.KeyStore) {
	c := crypto.NewKSCrypto(true, ks)
	crypto.Initialize(c)
}

func (n *Node) registerMysql() {
	n.rdsService = dao.NewRdsService(&n.globalConfig.Mysql)
	n.rdsService.Prepare()
}

func (n *Node) registerAccessor() {
	err := accessor.Initialize(n.globalConfig.Accessor)
	err = loopringaccessor.Initialize(n.globalConfig.LoopringProtocol)
	if nil != err {
		log.Fatalf("err:%s", err.Error())
	}
}

//func (n *Node) registerIPFSSubService() {
//	n.ipfsSubService = gateway.NewIPFSSubService(&n.globalConfig.Ipfs)
//}

func (n *Node) registerOrderManager() {
	n.orderManager = ordermanager.NewOrderManager(&n.globalConfig.OrderManager, n.rdsService, n.userManager, n.marketCapProvider)
}

func (n *Node) registerTrendManager() {
	n.trendManager = market.NewTrendManager(n.rdsService, n.globalConfig.Market.CronJobLock)
}

func (n *Node) registerAccountManager() {
	n.accountManager = accountmanager.Initialize(&n.globalConfig.AccountManager, n.globalConfig.Kafka.Brokers)
}

func (n *Node) registerTransactionManager() {
	n.txManager = txmanager.NewTxManager(n.rdsService)
}

func (n *Node) registerTickerCollector() {
	n.tickerCollector = *market.NewCollector(n.globalConfig.Market.CronJobLock)
}

func (n *Node) registerWalletService() {
	n.walletService = *gateway.NewWalletService(n.trendManager, n.orderManager,
		n.accountManager, n.marketCapProvider, n.tickerCollector, n.rdsService, n.globalConfig.Market.OldVersionWethAddress)
}

func (n *Node) registerJsonRpcService() {
	n.jsonRpcService = *gateway.NewJsonrpcService(n.globalConfig.Jsonrpc.Port, &n.walletService)
}

func (n *Node) registerWebsocketService() {
	n.websocketService = *gateway.NewWebsocketService(n.globalConfig.Websocket.Port, n.trendManager, n.accountManager, n.marketCapProvider)
}

func (n *Node) registerSocketIOService() {
	n.socketIOService = *gateway.NewSocketIOService(n.globalConfig.Websocket.Port, n.walletService)
}

func (n *Node) registerGateway() {
	gateway.Initialize(&n.globalConfig.GatewayFilters, &n.globalConfig.Gateway, &n.globalConfig.Ipfs, n.orderManager, n.marketCapProvider, n.accountManager)
}

func (n *Node) registerUserManager() {
	n.userManager = usermanager.NewUserManager(&n.globalConfig.UserManager, n.rdsService)
}

func (n *Node) registerMarketCap() {
	n.marketCapProvider = marketcap.NewMarketCapProvider(&n.globalConfig.MarketCap)
}

func (n *Node) registerZklock() {
	if _, err := zklock.Initialize(n.globalConfig.ZkLock); err != nil {
		log.Fatalf("node start, register zklock error:%s", err.Error())
	}
}
