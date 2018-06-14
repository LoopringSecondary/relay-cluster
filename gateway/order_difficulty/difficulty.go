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

package order_difficulty

import (
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/types"
	"github.com/Loopring/relay-lib/zklock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lydy/go-ethereum/common/math"
	"math/big"
	"qiniupkg.com/x/log.v7"
	"strconv"
	"time"
)

const (
	OrderCountPerSecond = "o_cnt_per_s_"
	OrderDifficulty     = "order_diff"
	ZklockDifficulty    = "zklock_diff"
)

type OrderDifficultyEvaluator struct {
	currentDifficult *OrderDifficulty
	parentDifficult  *OrderDifficulty
	baseDifficulty   *big.Int
	orderTraffic     int64
	triggerThreshold float64
	stopFuns         []func()
	calCount         int64 //must be odd
}

type OrderDifficulty struct {
	difficulty *big.Int
	ordersNum  int64
	timeStamp  int64
}

func (evaluator *OrderDifficultyEvaluator) getCacheKey(createTime int64) (key string, expireAt int64) {
	mod := createTime % evaluator.calCount
	orderSection := createTime - mod
	orderSectionStr := strconv.FormatInt(orderSection, 10)
	return OrderCountPerSecond + orderSectionStr, createTime + evaluator.calCount
}

func (evaluator *OrderDifficultyEvaluator) Start() {
	evaluator.HandleNewOrder()
	go func() {
		if err := zklock.TryLock(ZklockDifficulty); nil != err {
			log.Errorf("erro:%s", err.Error())
		} else {
			now := time.Now().Unix()
			orderCntList := []int64{}
			for i := evaluator.calCount; i > 0; i-- {
				t := now - i
				cacheKey, _ := evaluator.getCacheKey(t)
				if data, err := cache.Get(cacheKey); nil == err {
					cnt, _ := strconv.ParseInt(string(data), 10, 0)
					orderCntList = append(orderCntList, cnt)
				}
			}
			for {
				select {
				case <-time.After(2 * time.Second):
					cacheKey, _ := evaluator.getCacheKey(time.Now().Unix() - 1)
					if data, err := cache.Get(cacheKey); nil == err {
						cnt, _ := strconv.ParseInt(string(data), 10, 0)
						orderCntList = append(orderCntList, cnt)
					}
					diff := evaluator.CalcAndSaveDifficulty(orderCntList)
					diffHash := common.BytesToHash(diff.Bytes())
					cache.Set(OrderDifficulty, []byte(diffHash.Hex()), int64(0))
					orderCntList = orderCntList[1:]
				}
			}
		}
	}()
}

func (evaluator *OrderDifficultyEvaluator) Stop() {
	for _, f := range evaluator.stopFuns {
		f()
	}
}

//add ordersNum
func (evaluator *OrderDifficultyEvaluator) HandleNewOrder() {
	watcher := &eventemitter.Watcher{
		Concurrent: false, Handle: func(input eventemitter.EventData) error {
			state := input.(*types.OrderState)
			cacheKey, expireAt := evaluator.getCacheKey(state.RawOrder.CreateTime)
			_, err := cache.Incr(cacheKey)
			if nil == err {
				err = cache.ExpireAt(cacheKey, expireAt)
			}
			return err
		},
	}

	evaluator.stopFuns = append(evaluator.stopFuns, func() {
		eventemitter.Un(eventemitter.NewOrder, watcher)
	})
	eventemitter.On(eventemitter.NewOrder, watcher)
}

//控制订单的提交速度，随着订单的流量增大而增大
func (evaluator *OrderDifficultyEvaluator) CalcAndSaveDifficulty(orderCntList []int64) *big.Int {
	return math.MaxBig256
}

func GetDifficulty() (common.Hash, error) {
	if data, err := cache.Get(OrderDifficulty); nil == err {
		return common.HexToHash(string(data)), nil
	} else {
		return common.Hash{}, err
	}
}
