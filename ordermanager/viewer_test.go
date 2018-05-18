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

package ordermanager_test

import (
	"testing"
	"github.com/Loopring/relay-cluster/ordermanager"
	"github.com/Loopring/relay-lib/motan"
)

func TestOrderViewerImpl_MotanRpcServer(t *testing.T) {
	serverInstance := &ordermanager.OrderViewerImpl{}
	options := motan.MotanServerOptions{}
	options.ConfFile = "/Users/fukun/projects/gohome/src/github.com/Loopring/relay-cluster/config/ordermanager.yaml"
	options.ServerInstance = serverInstance
	motan.RunServer(options)
}
