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

package gateway

import (
	"encoding/json"
	"github.com/Loopring/relay-lib/broadcast"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
)

//to broadcast
func listenOrderForBroadcast() error {
	gatewayOrderWatcher := &eventemitter.Watcher{Concurrent: true, Handle: handleGatewayOrder}
	eventemitter.On(eventemitter.NewOrderForBroadcast, gatewayOrderWatcher)
	return nil
}

//
func listenOrderFromBroacast() error {
	if orderChan, err := broadcast.SubOrderNext(); nil != err {
		return err
	} else {
		go func() {
			for {
				select {
				case dataI := <-orderChan:
					if data, ok := dataI.([]byte); ok {
						order := &types.Order{}
						if err := json.Unmarshal(data, order); nil != err {
							log.Errorf("err:%s", err.Error())
						} else {
							if _, err := HandleInputOrder(order); nil != err {
								log.Errorf("err:%s", err.Error())
							}
						}
					}
				}
			}
		}()
		return nil
	}
}

func handleGatewayOrder(input eventemitter.EventData) error {
	if order, ok := input.(*types.Order); ok {
		data, err := json.Marshal(order)
		if nil != err {
			log.Errorf("err:%s", err.Error())
			return err
		}
		if err1 := broadcast.PubOrder(order.Hash.Hex(), data); nil != err1 {
			log.Errorf("err:%s", err1.Error())
			return err1
		}
	}
	return nil
}
