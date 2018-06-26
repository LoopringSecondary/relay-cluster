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

package matrix

import (
	"encoding/json"
	"fmt"
	"github.com/Loopring/relay-lib/broadcast"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/log"
	"time"
)

const (
	CacheKeyLastFrom = "matrix_sub_from_"
)

type MatrixSubscriberOption struct {
	MatrixClientOptions
	Rooms     []string
	CacheFrom bool
	CacheTtl  int64
}

type MatrixSubscriber struct {
	matrixClient *MatrixClient
	Room         string
	From         string
	CacheFrom    bool
	CacheTtl     int64
}

func (subscriber *MatrixSubscriber) Next() ([][]byte, error) {
	orderData := [][]byte{}
	eventTypes := []string{LoopringOrderType}
	roomEventFilter := RoomEventFilter{
		Types: eventTypes,
		Rooms: []string{subscriber.Room},
	}
	filterStr, err := json.Marshal(roomEventFilter)
	if nil != err {
		log.Errorf("err:%s", err.Error())
		return orderData, err
	}
	var res *RoomMessagesRes
	for {
		res, err = subscriber.matrixClient.RoomMessages(subscriber.Room, subscriber.From, "", "", "100", string(filterStr))
		if nil != err {
			log.Errorf("err:%s", err.Error())
			return orderData, err
		}
		subscriber.From = res.End
		if subscriber.CacheFrom {
			cache.Set(CacheKeyLastFrom+subscriber.Room, []byte(subscriber.From), subscriber.CacheTtl)
		}
		if len(res.Chunk) == 0 {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	for _, chunk := range res.Chunk {
		orderData = append(orderData, []byte(chunk.Content.Body))
		//if content, ok := chunk.Content.(map[string]string); ok {
		//	orderData = append(orderData, []byte(content["body"]))
		//}
	}
	return orderData, nil
}

func (subscriber *MatrixSubscriber) Name() string {
	return "matrixSubscriber"
}

func NewSubscribers(options []MatrixSubscriberOption) ([]broadcast.Subscriber, error) {
	subscribers := []broadcast.Subscriber{}
	var err error
	for _, option := range options {
		for _, room := range option.Rooms {
			subscriber := &MatrixSubscriber{}
			subscriber.Room = room
			subscriber.CacheFrom = option.CacheFrom
			subscriber.CacheTtl = option.CacheTtl
			subscriber.matrixClient, err = NewMatrixClient(option.MatrixClientOptions)
			if nil != err {
				return nil, fmt.Errorf("hsurl:%s, room:%s, err:%s", option.MatrixClientOptions.HSUrl, room, err.Error())
			}
			if subscriber.CacheFrom {
				if data, err := cache.Get(CacheKeyLastFrom + subscriber.Room); nil != err {
					log.Errorf("err:%s", err.Error())
					subscriber.From = string(data)
				} else {
					subscriber.From = string(data)
				}
			} else {
				subscriber.From = ""
			}
			subscribers = append(subscribers, subscriber)
		}
	}
	return subscribers, nil
}
