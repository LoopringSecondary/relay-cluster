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

import "math/big"

type OrderDifficultyEvaluator struct {
	currentDifficult *OrderDifficulty
	parentDifficult *OrderDifficulty
	baseDifficulty *big.Int

}

type OrderDifficulty struct {
	difficulty *big.Int
	ordersNum int64
	timeStamp int64
}

//add ordersNum
func HandleNewOrder() {

}

var evaluator *OrderDifficultyEvaluator

//控制订单的提交速度，随着订单的流量增大而增大
func (evaluator *OrderDifficultyEvaluator) CalcDifficulty() *big.Int {
	//以太坊
	/**
	1、时间
	2、上一难度
	3、当前时间
	4、当前高度
	5、diff = 上一难度+难度调整
	6、难度调整=上一难度/2048*Max(1-(时间)/10, -99)
	 */
	if evaluator.currentDifficult.ordersNum <= big.NewInt(int64(100)) {
		return big.NewInt(int64(0))
	}


	parentOrderDibbiculty := &OrderDifficulty{}
	parentOrderDibbiculty.difficulty
	return big.NewInt(int64(0))
}




