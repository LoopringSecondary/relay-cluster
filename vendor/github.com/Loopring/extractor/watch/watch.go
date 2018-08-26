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

package watch

import (
	"github.com/Loopring/relay-lib/cloudwatch"
	"github.com/Loopring/relay-lib/log"
)

const (
	Metric_OnChainEvent_Emitted = "extractor_onchain_event"
)

var watchOpen bool

func Initialize(conf cloudwatch.CloudWatchConfig) {
	if !conf.Enabled {
		watchOpen = false
		return
	}

	if err := cloudwatch.Initialize(conf); err != nil {
		log.Fatalf("node start, register cloud watch error:%s", err.Error())
	}

	watchOpen = true
}

func ReportHeartBeat() {
	if !watchOpen {
		return
	}
	cloudwatch.PutHeartBeatMetric(Metric_OnChainEvent_Emitted)
}
