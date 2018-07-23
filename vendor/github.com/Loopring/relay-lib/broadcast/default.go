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

package broadcast

import (
	"errors"
	"github.com/Loopring/relay-lib/log"
	"time"
)

type DefaultPublisher struct {
}

func (publisher *DefaultPublisher) PubOrder(hash string, orderData []byte) error {
	log.Debug("default broadcaster doesn't publish msg")
	return nil
}

func (publisher *DefaultPublisher) Name() string {
	return "defaultPublisher"
}

type DefaultSubscriber struct {
}

func (subscriber *DefaultSubscriber) Next() ([]byte, error) {
	log.Debug("default broadcaster doesn't subscribe msg")
	time.Sleep(10 * time.Minute)
	return []byte{}, errors.New("default broadcaster doesn't subscribe msg")
}

func (subscriber *DefaultSubscriber) Name() string {
	return "defaultSubscriber"
}
