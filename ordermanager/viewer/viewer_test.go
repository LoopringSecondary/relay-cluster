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

package viewer_test

import (
	"github.com/Loopring/relay-cluster/ordermanager/viewer"
	"github.com/Loopring/relay-lib/motan"
	"testing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/Loopring/relay-cluster/dao"
	libdao "github.com/Loopring/relay-lib/dao"
)

func TestOrderViewerImpl_MotanRpcServer(t *testing.T) {
	serverInstance := &viewer.OrderViewerImpl{}
	options := motan.MotanServerOptions{}
	options.ConfFile = "/Users/fukun/projects/gohome/src/github.com/Loopring/relay-cluster/config/test-ordermanager.yaml"
	options.ServerInstance = serverInstance
	motan.RunServer(options)
}

func TestOrderViewerImpl_GetOrderByHash(t *testing.T) {
	//path := "/Users/fukun/projects/gohome/src/github.com/Loopring/relay-cluster/config/relay.toml"
	//cfg := node.LoadConfig(path)
	//log.Initialize(cfg.Log)

	opt := libdao.MysqlOptions{}
	opt.Hostname = "loopring-relay.cfsiqsz1ae0c.ap-northeast-1.rds.amazonaws.com"
	opt.Port = "3306"
	opt.User = "root"
	opt.Password = "s4sfo1q}W+$^%]E8"
	opt.DbName = "loopring_relay_v1_5"
	opt.TablePrefix = "lpr_"
	opt.Debug = false
	rds := dao.NewDb(&opt)

	order, err := rds.GetOrderByHash(common.HexToHash("0x827945f687d4105f179e371e3be12f6729058ed5bf481d53a561175cb3259da6"))
	if err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Log(order.OrderHash)
	}
}
