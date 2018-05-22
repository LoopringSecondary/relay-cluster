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
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"sync"
	"time"
)

/*todo:
1、通过zk设置配置文件
2、节点加入与退出事件
*/

type ZkLockConfig struct {
	ZkServers      string
	ConnectTimeOut int
}

type ZkLock struct {
	zkClient *zk.Conn
	lockMap  map[string]*zk.Lock
	mutex    sync.Mutex
}

var zl *ZkLock

const basePath = "/loopring_lock"

func Initialize(config ZkLockConfig) (*ZkLock, error) {
	if config.ZkServers == "" || len(config.ZkServers) < 10 {
		return nil, fmt.Errorf("Zookeeper server list config invalid: %s\n", config.ZkServers)
	}
	zkClient, _, err := zk.Connect(strings.Split(config.ZkServers, ","), time.Second*time.Duration(config.ConnectTimeOut))
	if err != nil {
		return nil, fmt.Errorf("Connect zookeeper error: %s\n", err.Error())
	}
	zl = &ZkLock{zkClient, make(map[string]*zk.Lock), sync.Mutex{}}
	return zl, nil
}

//when get err, should send sns message
func TryLock(lockName string) error {
	zl.mutex.Lock()
	if _, ok := zl.lockMap[lockName]; !ok {
		acls := zk.WorldACL(zk.PermAll)
		zl.lockMap[lockName] = zk.NewLock(zl.zkClient, fmt.Sprintf("%s/%s", basePath, lockName), acls)
	}
	zl.mutex.Unlock()
	return zl.lockMap[lockName].Lock()
}

func ReleaseLock(lockName string) error {
	if innerLock, ok := zl.lockMap[lockName]; ok {
		innerLock.Unlock()
		return nil
	} else {
		return fmt.Errorf("Try release not exists lock: %s\n", lockName)
	}
}

func IsInit() bool {
	return nil != zl
}
