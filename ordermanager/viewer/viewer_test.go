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
	"github.com/Loopring/relay-cluster/test"
	"github.com/Loopring/relay-lib/motan"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

func TestOrderViewerImpl_MotanRpcServer(t *testing.T) {
	serverInstance := &viewer.OrderViewerImpl{}
	options := motan.MotanServerOptions{}
	options.ConfFile = "/Users/fukun/projects/gohome/src/github.com/Loopring/relay-cluster/config/ordermanager.yaml"
	options.ServerInstance = serverInstance
	motan.RunServer(options)
}

func TestOrderViewerImpl_FlexCancelOrder(t *testing.T) {
	data := &types.FlexCancelOrderEvent{
		Owner:      common.HexToAddress("0x1B978a1D302335a6F2Ebe4B8823B5E17c3C84135"),
		OrderHash:  common.HexToHash("0xceb13a7678b7a24ab1ab54cfd429dbe4bf31bbf647ff6c01b781b72c058ab9c9"),
		CutoffTime: 0,
		TokenS:     types.NilAddress,
		TokenB:     types.NilAddress,
		Type:       types.FLEX_CANCEL_BY_HASH,
	}

	v := test.GenerateOrderView()
	if err := v.FlexCancelOrder(data); err != nil {
		t.Logf(err.Error())
	}
}
