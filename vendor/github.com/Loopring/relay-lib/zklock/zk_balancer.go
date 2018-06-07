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
	"encoding/json"
	"fmt"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/utils"
	"github.com/samuel/go-zookeeper/zk"
	"sort"
	"sync"
	"time"
)

type Task struct {
	Payload   string //Native info for bussiness structs
	Path      string // friendly string for zk path composite, default is .substr(Payload, 0, 10)
	Weight    int    //bi
	Status    int
	Owner     string
	Timestamp int64
}

const balancerShareBasePath = "/loopring_balancer"
const worker_path = "worker"
const event_path = "event"

const localIpPrefix = "172.31"
const releasingTimeout = 120

type ZkBalancer struct {
	name                   string
	workerBasePath         string
	eventBasePath          string
	tasks                  map[string]*Task
	isMaster               bool
	mutex                  sync.Mutex
	onAssignFunc           func([]Task) error
	workerPath             string
	postponedGlobalBalance bool
}

type Status int

const (
	Init      = iota
	Assigned
	Releasing
	Deleting
)

//TODO check whether need load tasks schedule from zk to keep data consistency

type tasksForSort []*Task

func (a tasksForSort) Len() int      { return len(a) }
func (a tasksForSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a tasksForSort) Less(i, j int) bool {
	if a[i].Weight > a[j].Weight {
		return true
	} else {
		return false
	}
}

func (zb *ZkBalancer) Init(name string, initTasks []Task) error {
	if name != "" {
		zb.name = name
	} else if zb.name == "" {
		return fmt.Errorf("balancer Name is empty")
	}
	if len(initTasks) > 0 {
		zb.tasks = make(map[string]*Task)
		for i := 0; i < len(initTasks); i++ {
			if initTasks[i].Payload == "" {
				return fmt.Errorf("task payload is empty")
			}
			if initTasks[i].Path == "" {
				if len(initTasks[i].Payload) <= 10 {
					initTasks[i].Path = initTasks[i].Payload
				} else {
					initTasks[i].Path = initTasks[i].Payload[:10]
				}
				log.Warn("task not specify Path use payload as default\n")
			}
			initTasks[i].Status = Init
			zb.tasks[initTasks[i].Path] = &initTasks[i]
		}
	} else {
		return fmt.Errorf("no initTasks to balance")
	}
	if !IsLockInitialed() {
		return fmt.Errorf("zkClient is not intiliazed")
	}
	var err error
	if _, err = zb.createPath(balancerShareBasePath); err != nil {
		return fmt.Errorf("create balancer base path failed %s", balancerShareBasePath)
	}
	blp := fmt.Sprintf("%s/%s", balancerShareBasePath, zb.name)
	if _, err = zb.createPath(blp); err != nil {
		return fmt.Errorf("create balancer path for %s failed with error : %s", blp, err.Error())
	}
	wp := fmt.Sprintf("%s/%s", blp, worker_path)
	if zb.workerBasePath, err = zb.createPath(wp); err != nil {
		return fmt.Errorf("create balancer worker path for %s failed with error : %s", wp, err.Error())
	}
	ep := fmt.Sprintf("%s/%s", blp, event_path)
	if zb.eventBasePath, err = zb.createPath(ep); err != nil {
		return fmt.Errorf("create balancer event path for %s failed with error : %s", ep, err.Error())
	}
	zb.workerPath = utils.GetLocalIpByPrefix(localIpPrefix)
	zb.mutex = sync.Mutex{}
	return nil
}

func (zb *ZkBalancer) Start() {
	zb.startMaster()
	time.Sleep(time.Second * time.Duration(2))
	zb.registerWorker()
}

func (zb *ZkBalancer) Stop() {
	zb.mutex.Lock()
	if zb.isMaster {
		ReleaseLock(zb.masterLockName())
	}
	zb.mutex.Unlock()
	zb.unRegisterWorker()
}

// worker callback
// WARN : as assignFunc may be called multi times, every time missing tasks should handle by client safely as accumulated releasing tasks,
// 		  those task may occur by history assignFunc calls, and should be finally released somehow
func (zb *ZkBalancer) OnAssign(assignFunc func(tasks []Task) error) {
	zb.onAssignFunc = assignFunc
}

// worker call
func (zb *ZkBalancer) Released(tasks []Task) error {
	if len(tasks) == 0 {
		log.Errorf("released tasks is empty\n")
		return fmt.Errorf("released tasks is empty")
	}
	if data, err := json.Marshal(tasks); err != nil {
		return fmt.Errorf("marshal released tasks failed %s", err.Error())
	} else {
		if _, err := ZkClient.CreateProtectedEphemeralSequential(fmt.Sprintf("%s/released-", zb.eventBasePath), data, zk.WorldACL(zk.PermAll)); err != nil {
			log.Errorf("released event node failed create with error : %s\n", err.Error())
			return err
		}
		return nil
	}
}

func (zb *ZkBalancer) startMaster() {
	go func() {
		firstTime := true
		for {
			zb.mutex.Lock()
			zb.isMaster = false
			zb.mutex.Unlock()
			if !firstTime {
				time.Sleep(time.Second * time.Duration(3))
			} else {
				firstTime = false
			}
			if err := TryLock(zb.masterLockName()); err != nil {
				log.Errorf("master failed get lock for %s, with error %s, try again\n", zb.masterLockName(), err.Error())
				continue
			} else {
				log.Info("get master lock success !!!!! \n")
			}
			zb.isMaster = true
			if workers, _, ch, err := ZkClient.ChildrenW(zb.workerBasePath); err == nil {
				if err != nil {
					log.Errorf("master get workers failed %s\n", err.Error())
					ReleaseLock(zb.masterLockName())
					continue
				}
				if err := zb.loadTasks(workers); err != nil {
					log.Errorf("master load tasks from worker failed with error %s\n", err.Error())
					ReleaseLock(zb.masterLockName())
					continue
				}
				if err := zb.deprecateTasks(); err != nil {
					log.Errorf("master failed deprecateTasks with error %s\n", err.Error())
					ReleaseLock(zb.masterLockName())
					continue
				}
				log.Info("master watch worker nodes success !!!\n")
				zb.releaseOrphanTasks(workers)
				zb.balanceTasks(workers)
				go func() {
					for {
						select {
						case evt := <-ch:
							if evt.Type == zk.EventNodeChildrenChanged {
								log.Info("worker nodes changed !!!!!")
								if workers, _, chx, err := ZkClient.ChildrenW(zb.workerBasePath); err == nil {
									ch = chx
									zb.releaseOrphanTasks(workers)
									zb.balanceTasks(workers)
								}
							}
						}
					}
				}()
				zb.handleReleasedEvents()
				zb.scheduleCheckTasks()
				return
			} else {
				log.Errorf("master watch workers failed with error : %s\n", err.Error())
			}
		}
	}()
}

func (zb *ZkBalancer) registerWorker() error {
	log.Info("begin registerWorker !!\n")
	for {
		loaded := false
		if _, err := ZkClient.Create(zb.workerEphemeralPath(), nil, zk.FlagEphemeral, zk.WorldACL(zk.PermAll)); err == nil || err == zk.ErrNodeExists {
			if err == zk.ErrNodeExists {
				log.Warnf("worker node not released, register worker use former node!!\n")
			}
			if err == zk.ErrNodeExists && !loaded {
				if err := zb.loadTasksForWorker(zb.workerPath); err != nil {
					loaded = true
				}
			}
			_, _, ch, err := ZkClient.GetW(zb.workerEphemeralPath())
			if err != nil {
				log.Errorf("watch worker %s failed with error %s\n", zb.workerEphemeralPath(), err.Error())
				continue
			} else {
				log.Info("registerWorker success!!\n")
				go func() {
					for {
						select {
						case evt := <-ch:
							if evt.Type == zk.EventNodeDataChanged {
								if data, _, chx, err := ZkClient.GetW(zb.workerEphemeralPath()); err != nil {
									log.Errorf("Get worker %s data failed with error : %s\n", zb.workerEphemeralPath(), err.Error())
								} else {
									log.Infof("Get worker %s data success\n", zb.workerEphemeralPath())
									ch = chx
									if workerTasks, err := decodeData(data); err == nil {
										log.Infof("worker watched new assigndTask %+v\n", workerTasks)
										zb.onAssignFunc(workerTasks)
									} else {
										log.Errorf("worker watcher decodeData failed with err : %s\n", err.Error())
									}
								}
							}
						}
					}
				}()
			}
			return nil
		} else {
			log.Errorf("registerWorker failed when create zk node with error : %s\n", err.Error())
		}
	}
}

func (zb *ZkBalancer) unRegisterWorker() error {
	if err := ZkClient.Delete(zb.workerEphemeralPath(), -1); err != nil {
		log.Errorf("failed unRegister worker with error : %s\n", err.Error())
		return err
	}
	return nil
}

func (zb *ZkBalancer) loadTasks(workers []string) error {
	for _, worker := range workers {
		if err := zb.loadTasksForWorker(worker); err != nil {
			return err
		}
	}
	return nil
}

func (zb *ZkBalancer) loadTasksForWorker(worker string) error {
	if data, _, err := ZkClient.Get(fmt.Sprintf("%s/%s", zb.workerBasePath, worker)); err != nil && err != zk.ErrNoNode {
		log.Errorf("loadTasksForWorker failed on get worker data for %s\n, with error : %s\n", worker, err.Error())
		return err
	} else if err == zk.ErrNoNode {
		log.Warnf("expected worker %s not exists when try load data\n", worker)
		return nil
	} else {
		if workerTasks, err := decodeData(data); err != nil {
			log.Errorf("loadTasksForWorker failed on decode tasks from worker %s, with error %s\n", worker, err.Error())
			return err
		} else {
			for _, wt := range workerTasks {
				if v, ok := zb.tasks[wt.Path]; ok {
					v.Owner = worker
					v.Status = Assigned
					v.Timestamp = time.Now().Unix()
				} else {
					wt.Status = Deleting
					wt.Timestamp = time.Now().Unix()
					zb.tasks[wt.Path] = &wt
				}
			}
		}
		return nil
	}
}

func (zb *ZkBalancer) deprecateTasks() error {
	needDeleteWorker := make(map[string]string)
	var err error = nil
	for _, task := range zb.tasks {
		if task.Status == Deleting {
			if _, ok := needDeleteWorker[task.Owner]; !ok {
				needDeleteWorker[task.Owner] = "-"
				if e := zb.assignedTasks(task.Owner); e != nil {
					err = e
				}
			}
		}
	}
	return err
}

func (zb *ZkBalancer) assignedTasks(worker string) error {
	//log.Info("showTasks in assignedTasks\n")
	//zb.showTasks()
	assignedTasks := make([]*Task, 0, 10)
	for _, task := range zb.tasks {
		//log.Infof("assignedTasks woker %s , task %+v\n", worker, *task)
		if task.Owner == worker && task.Status == Assigned {
			assignedTasks = append(assignedTasks, task)
		}
	}
	//log.Infof("assignedTasks for %s with tasks %+v\n", worker, assignedTasks)
	if data, err := encodeData(assignedTasks); err != nil {
		log.Errorf("assignedTasks encode worker %s data failed, with error %s\n", worker, err.Error())
		return err
	} else {
		if _, err := ZkClient.Set(fmt.Sprintf("%s/%s", zb.workerBasePath, worker), data, -1); err != nil {
			log.Errorf("assignedTasks  set data for worker %s failed, with error %s\n", worker, err.Error())
			return err
		} else {
			//log.Infof("assignedTasks success with len(data) %d \n", len(data))
			return nil
		}
	}
}

func (zb *ZkBalancer) handleReleasedEvents() {
	_, _, ch, err := ZkClient.ChildrenW(zb.eventBasePath)
	if err != nil {
		log.Errorf("Watch event children failed with error : %s\n", err.Error())
	} else {
		log.Info("master handleReleasedEvents success !!!\n")
		go func() {
			for {
				select {
				case evt := <-ch:
					if evt.Type == zk.EventNodeChildrenChanged {
						if events, _, chx, err := ZkClient.ChildrenW(zb.eventBasePath); err == nil {
							ch = chx
							if len(events) > 0 {
								releaseMap := make(map[string]Task)
								for _, event := range events {
									if data, _, err := ZkClient.Get(fmt.Sprintf("%s/%s", zb.eventBasePath, event)); err != nil {
										log.Errorf("failed get event data from zk with error : %s\n", err.Error())
									} else {
										if releasedTasks, err := decodeData(data); err == nil {
											for _, v := range releasedTasks {
												releaseMap[v.Path] = v
											}
										} else {
											log.Errorf("decode data failed when handleReleasedEvents with error : %s", err.Error())
										}
									}
								}
								zb.innerOnReleased(releaseMap)
								for _, event := range events {
									if err := ZkClient.Delete(fmt.Sprintf("%s/%s", zb.eventBasePath, event), -1); err != nil {
										log.Errorf("failed delete event node %s with error : %s\n", event, err.Error())
									}
								}
							}
						} else {
							log.Errorf("get event nodes failed with error %s", err.Error())
						}
					}
				}
			}
		}()
	}
}

func (zb *ZkBalancer) scheduleCheckTasks() {
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(60))
			log.Info("scheduleCheckTasks every minute >>>>>>>>>> \n")
			zb.handleReleaseTimeoutAndInitTasks()
			zb.showTasks()
		}
	}()
}

func (zb *ZkBalancer) innerOnReleased(releasedTasks map[string]Task) {
	log.Infof("innerOnReleased to release %+v\n", releasedTasks)
	zb.mutex.Lock()
	needAssignedTasks := make([]Task, 0, len(releasedTasks))
	for _, rlt := range releasedTasks {
		if origin, ok := zb.tasks[rlt.Path]; ok {
			switch origin.Status {
			case Init:
				log.Errorf("Try release task with status Init : %+v\n", origin)
				break
			case Assigned:
				log.Errorf("Try release task with status Assigned : %+v\n", origin)
				break
			case Deleting:
				delete(zb.tasks, origin.Path)
				break
			case Releasing:
				origin.Status = Assigned
				origin.Timestamp = time.Now().Unix()
				needAssignedTasks = append(needAssignedTasks, *origin)
				break
			}
		} else {
			log.Error("Try release task not exits, ignore\n")
		}
	}
	zb.mutex.Unlock()
	if len(needAssignedTasks) > 0 {
		zb.balanceTasksOnReleased(needAssignedTasks)
	}
}

func (zb *ZkBalancer) handleReleaseTimeoutAndInitTasks() {
	needAssignTasks := make([]Task, 0, len(zb.tasks)/2)
	hasInit := false
	hasReleasing := false
	zb.mutex.Lock()
	for _, t := range zb.tasks {
		switch t.Status {
		case Releasing:
			if time.Now().Unix()-t.Timestamp >= releasingTimeout {
				t.Timestamp = time.Now().Unix()
				t.Status = Assigned
				needAssignTasks = append(needAssignTasks, *t)
			} else {
				hasReleasing = true
			}
			break
		case Deleting:
			if time.Now().Unix()-t.Timestamp >= releasingTimeout {
				delete(zb.tasks, t.Path)
			}
			break
		case Init:
			hasInit = true
			t.Timestamp = time.Now().Unix()
			break
		}
	}
	zb.mutex.Unlock()
	if hasInit {
		if !hasReleasing {
			log.Infof("handleReleaseTimeoutAndInitTasks found Init tasks, tryGlobalBalance")
			if workers, _, err := ZkClient.Children(zb.workerBasePath); err == nil {
				if zb.balanceTasks(workers) {
					log.Info("handleReleaseTimeoutAndInitTasks global balance success\n")
					return
				}
			} else {
				log.Errorf("handleReleaseTimeoutAndInitTasks get workers failed with error : %s\n", err.Error())
			}
		} else {
			zb.mutex.Lock()
			zb.postponedGlobalBalance = true
			zb.mutex.Unlock()
		}
	}

	if len(needAssignTasks) > 0 {
		log.Infof("handleReleaseTimeoutAndInitTasks found should assigned tasks, try balanceTasksOnReleased")
		zb.balanceTasksOnReleased(needAssignTasks)
	}
}

func (zb *ZkBalancer) releaseOrphanTasks(workers []string) bool {
	validWorkers := make(map[string]string, len(workers))
	for _, v := range workers {
		validWorkers[v] = "-"
	}
	zb.mutex.Lock()
	defer zb.mutex.Unlock()
	hasOrphan := false
	for _, t := range zb.tasks {
		if _, ok := validWorkers[t.Owner]; !ok {
			t.Status = Init
			t.Owner = ""
			hasOrphan = true
		}
	}
	return hasOrphan
}

//return false if Releasing task exists and postpone this action
func (zb *ZkBalancer) balanceTasks(workers []string) bool {
	log.Infof("balanceTasks workers %+v\n", workers)
	if len(workers) == 0 {
		return false
	}
	zb.mutex.Lock()
	defer zb.mutex.Unlock()
	sortedTask := make([]*Task, 0, len(zb.tasks))
	for _, t := range zb.tasks {
		if t.Status == Assigned || t.Status == Init {
			sortedTask = append(sortedTask, t)
		} else if t.Status == Releasing {
			log.Info("task with releasing status, will not rebalance\n")
			zb.postponedGlobalBalance = true
			return false
		}
	}
	if len(sortedTask) == 0 {
		log.Infof("len(sortedTask) %d \n", len(sortedTask))
		zb.postponedGlobalBalance = false
		return true
	}
	sort.Sort(tasksForSort(sortedTask))
	workersWeight := make(map[string]int)
	for i := 0; i < len(workers) && i < len(sortedTask); i++ {
		workersWeight[workers[i]] = sortedTask[i].Weight
		zb.tryReAssignTask(sortedTask[i], workers[i])
	}
	//log.Infof("balanceTasks workersWeight: %+v", workersWeight)
	if len(sortedTask) > len(workers) {
		for i := len(workers); i < len(sortedTask); i++ {
			minWorker := minWeightWorker(workersWeight)
			//log.Infof("balanceTasks minWorker : %+v", minWorker)
			zb.tryReAssignTask(sortedTask[i], minWorker)
			workersWeight[minWorker] += workersWeight[minWorker] + sortedTask[i].Weight
		}
	}
	log.Infof("balanceTasks result >>>>>>>>>>>> \n")
	zb.showTasks()
	for _, worker := range workers {
		zb.assignedTasks(worker)
	}
	zb.postponedGlobalBalance = false
	return true
}

func (zb *ZkBalancer) balanceTasksOnReleased(tasksReleased []Task) {
	log.Infof("balanceTasksOnReleased for %+v\n", tasksReleased)
	zb.mutex.Lock()
	hasPostponedTask := zb.postponedGlobalBalance
	zb.mutex.Unlock()
	if hasPostponedTask {
		log.Info("has postponed global balance action, try it\n")
		if workers, _, err := ZkClient.Children(zb.workerBasePath); err == nil {
			if zb.balanceTasks(workers) {
				log.Info("balanceTasksOnReleased global balance success\n")
				return
			}
		} else {
			log.Errorf("try postponedGlobalBalance get workers failed with error : %s\n", err.Error())
		}
	}
	needReAssignWorker := make(map[string]string)
	for _, task := range tasksReleased {
		if _, ok := needReAssignWorker[task.Owner]; !ok {
			needReAssignWorker[task.Owner] = "-"
			zb.assignedTasks(task.Owner)
		}
	}
}

func (zb *ZkBalancer) tryReAssignTask(t *Task, newWorker string) {
	t.Timestamp = time.Now().Unix()
	if t.Status == Assigned {
		if t.Owner != newWorker {
			t.Owner = newWorker
			t.Status = Releasing
		}
	} else if t.Status == Init {
		t.Status = Assigned
		t.Owner = newWorker
	}
}

func minWeightWorker(workerWeight map[string]int) string {
	minWorker := ""
	minWeight := -1
	for worker, weight := range workerWeight {
		if minWeight == -1 || weight < minWeight {
			minWorker = worker
			minWeight = weight
		}
	}
	return minWorker
}

func (zb *ZkBalancer) createPath(path string) (string, error) {
	isExist, _, err := ZkClient.Exists(path)
	if err != nil {
		return "", fmt.Errorf("check zk path %s exists failed : %s", path, err.Error())
	}
	if !isExist {
		_, err := ZkClient.Create(path, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return "", fmt.Errorf("failed create balancer sub path %s ,with error : %s ", path, err.Error())
		}
	}
	return path, nil
}

func (zb *ZkBalancer) masterLockName() string {
	return fmt.Sprintf("balancer/%s", zb.name)
}

func (zb *ZkBalancer) workerEphemeralPath() string {
	return fmt.Sprintf("%s/%s", zb.workerBasePath, zb.workerPath)
}

func (zb *ZkBalancer) showTasks() {
	log.Infof("Begin showTasks by worker >>>>>>>>>>>>>>\n")
	tasks4worker := make(map[string][]*Task)
	noWorker := "noWorker"
	for _, t := range zb.tasks {
		if t.Owner == "" {
			if ts, ok := tasks4worker[noWorker]; ok {
				tasks4worker[noWorker] = append(ts, t)
			} else {
				tasks4worker[noWorker] = []*Task{t}
			}
		} else {
			if ts, ok := tasks4worker[t.Owner]; ok {
				tasks4worker[t.Owner] = append(ts, t)
			} else {
				tasks4worker[t.Owner] = []*Task{t}
			}
		}
	}
	for k, v := range tasks4worker {
		log.Infof("tasks for worker : %s\n", k)
		for _, t := range v {
			log.Infof("%+v\n", *t)
		}
	}
	log.Infof("End showTasks by worker >>>>>>>>>>>>>>\n")
}

func decodeData(data []byte) ([]Task, error) {
	releasedTasks := []Task{}
	if len(data) == 0 {
		return releasedTasks, nil
	}
	if err := json.Unmarshal(data, &releasedTasks); err != nil {
		return nil, err
	} else {
		return releasedTasks, nil
	}
}

func encodeData(tasks []*Task) ([]byte, error) {
	if len(tasks) == 0 {
		return []byte{}, nil
	}
	res := make([]Task, 0, len(tasks))
	for _, v := range tasks {
		res = append(res, *v)
	}
	return json.Marshal(res)
}
