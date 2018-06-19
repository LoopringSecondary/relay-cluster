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

package zklock

import (
	"fmt"
	"github.com/Loopring/relay-lib/log"
	"github.com/samuel/go-zookeeper/zk"
)

const configShareBasePath = "/loopring_config"

type HandlerFunc func(value []byte) error

func RegisterConfigHandler(namespace string, key string, action HandlerFunc) error {
	if !IsLockInitialed() {
		return fmt.Errorf("zkClient is not intiliazed")
	}
	if _, err := CreatePath(configShareBasePath); err != nil {
		return fmt.Errorf("create config base path failed %s with error %s", configShareBasePath, err.Error())
	}
	ns := fmt.Sprintf("%s/%s", configShareBasePath, namespace)
	if _, err := CreatePath(ns); err != nil {
		return fmt.Errorf("create config namespace path failed %s with error %s", ns, err.Error())
	}
	keyPath := fmt.Sprintf("%s/%s", ns, key)
	if _, err := CreatePath(keyPath); err != nil {
		return fmt.Errorf("create config key path failed %s with error %s", keyPath, err.Error())
	}

	if data, _, ch, err := ZkClient.GetW(keyPath); err != nil {
		log.Errorf("Get config %s data failed with error : %s", keyPath, err.Error())
	} else {
		log.Infof("Get config %s data success", keyPath)
		action(data)
		go func() {
			for {
				select {
				case evt := <-ch:
					if evt.Type == zk.EventNodeDataChanged {
						if data, _, chx, err := ZkClient.GetW(keyPath); err != nil {
							log.Errorf("Get config %s data failed with error : %s", keyPath, err.Error())
						} else {
							log.Infof("config %s data changed to value %s", keyPath, string(data[:]))
							ch = chx
							action(data)
						}
					}
				}
			}
		}()
	}
	return nil
}
