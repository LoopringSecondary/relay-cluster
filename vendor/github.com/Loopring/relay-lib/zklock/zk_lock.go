package zklock

import (
	"github.com/samuel/go-zookeeper/zk"
	"fmt"
	"strings"
	"time"
)

type ZkLockConfig struct {
	NeedZkLock   bool
	ZkServers       string
	ConnectTimeOut  int
}

type ZkLock struct {
	zkClient *zk.Conn
	lockMap map[string]*zk.Lock
}

const basePath = "/loopring_lock"

func NewLock(config ZkLockConfig) (*ZkLock, error) {
	if !config.NeedZkLock {
		return nil, fmt.Errorf("NeedZkLock is set to false")
	}
	if config.ZkServers == "" || len(config.ZkServers) < 10 {
		return nil, fmt.Errorf("Zookeeper server list config invalid:%s", config.ZkServers)
	}
	zkClient, _, err := zk.Connect(strings.Split(config.ZkServers,","), time.Second * time.Duration(config.ConnectTimeOut))
	if err != nil {
		return nil, fmt.Errorf("Connect zookeeper error:%s", err.Error())
	}
	return &ZkLock{zkClient, make(map[string]*zk.Lock)}, nil
}

func (l *ZkLock) TryLock(lockName string) {
	if _, ok := l.lockMap[lockName]; !ok {
		acls := zk.WorldACL(zk.PermAll)
		l.lockMap[lockName] = zk.NewLock(l.zkClient, fmt.Sprintf("%s/%s", basePath, lockName), acls)
	}
	l.lockMap[lockName].Lock()
}

func (l *ZkLock) ReleaseLock(lockName string) (error) {
	if innerLock, ok := l.lockMap[lockName]; ok {
		innerLock.Unlock()
		return nil
	} else {
		return fmt.Errorf("Try release not exists lock:%s", lockName)
	}
}