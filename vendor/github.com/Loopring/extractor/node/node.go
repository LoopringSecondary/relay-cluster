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
	"github.com/Loopring/extractor/dao"
	"github.com/Loopring/extractor/extractor"
	"github.com/Loopring/extractor/watch"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"go.uber.org/zap"
	"sync"
)

type Node struct {
	globalConfig *GlobalConfig
	rdsService   dao.RdsService
	extractor    extractor.ExtractorService
	wg           *sync.WaitGroup
	logger       *zap.Logger
}

func NewNode(logger *zap.Logger, globalConfig *GlobalConfig) *Node {
	n := &Node{}
	n.logger = logger
	n.globalConfig = globalConfig
	n.wg = new(sync.WaitGroup)

	n.registerCache()
	n.registerMysql()
	n.registerMarketUtil()
	n.registerAccessor()
	n.registerZkLock()
	n.registerExtractor()
	n.registerEmitter()
	n.registerCloudWatch()

	return n
}

func (n *Node) Start() {
	n.extractor.Start()
	n.wg.Add(1)
}

func (n *Node) Wait() {
	n.wg.Wait()
}

func (n *Node) Stop() {
	// todo
	//extractor.UnRegistryEmitter()
	n.wg.Done()
}

func (n *Node) registerCache() {
	cache.NewCache(n.globalConfig.Redis)
}

func (n *Node) registerMysql() {
	n.rdsService = dao.NewDb(&n.globalConfig.Mysql)
}

func (n *Node) registerMarketUtil() {
	util.Initialize(&n.globalConfig.Market)
}

func (n *Node) registerAccessor() {
	if err := accessor.Initialize(n.globalConfig.Accessor); err != nil {
		log.Fatalf("node start, register accessor error:%s", err.Error())
	}
	if err := loopringaccessor.Initialize(n.globalConfig.LoopringProtocol); err != nil {
		log.Fatalf("node start, register loopring accessor error:%s", err.Error())
	}
}

func (n *Node) registerExtractor() {
	n.extractor = extractor.NewExtractorService(n.globalConfig.Extractor, n.rdsService)
}

func (n *Node) registerEmitter() {
	if err := extractor.RegistryEmitter(n.globalConfig.Kafka, n.globalConfig.Kafka, n.extractor); err != nil {
		log.Fatalf("node start, register emitter error:%s", err.Error())
	}
}

func (n *Node) registerZkLock() {
	if err := extractor.RegistryZkLock(n.globalConfig.ZkLock); err != nil {
		log.Fatalf("node start, register zklock error:%s", err.Error())
	}
}

func (n *Node) registerCloudWatch() {
	watch.Initialize(n.globalConfig.CloudWatch)
}
