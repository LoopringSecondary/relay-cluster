package zklock

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"sync"
	"time"
)

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

func TryLock(lockName string) {
	zl.mutex.Lock()
	if _, ok := zl.lockMap[lockName]; !ok {
		acls := zk.WorldACL(zk.PermAll)
		zl.lockMap[lockName] = zk.NewLock(zl.zkClient, fmt.Sprintf("%s/%s", basePath, lockName), acls)
	}
	zl.mutex.Unlock()
	zl.lockMap[lockName].Lock()
}

func ReleaseLock(lockName string) error {
	if innerLock, ok := zl.lockMap[lockName]; ok {
		innerLock.Unlock()
		return nil
	} else {
		return fmt.Errorf("Try release not exists lock: %s\n", lockName)
	}
}
