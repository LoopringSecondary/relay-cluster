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

package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Loopring/relay-lib/log"
	"github.com/bsm/sarama-cluster"
	"reflect"
	"sync"
)

type ConsumerRegister struct {
	brokers     []string
	conf        *cluster.Config
	consumerMap map[string]map[string]*cluster.Consumer
	mutex       sync.Mutex
}

type HandlerFunc func(event interface{}) error

func (cr *ConsumerRegister) Initialize(brokerList []string) {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	cr.conf = config
	cr.brokers = brokerList
	cr.consumerMap = make(map[string]map[string]*cluster.Consumer) //map[topic][groupId]
	cr.mutex = sync.Mutex{}
}

func (cr *ConsumerRegister) RegisterTopicAndHandler(topic string, groupId string, data interface{}, action HandlerFunc) error {
	cr.mutex.Lock()
	groupConsumerMap, ok := cr.consumerMap[topic]
	if ok {
		_, ok1 := groupConsumerMap[groupId]
		if ok1 {
			cr.mutex.Unlock()
			return fmt.Errorf("kafka consumer alreay registered for [%s, %s]!!\n", topic, groupId)
		}
	} else {
		cr.consumerMap[topic] = make(map[string]*cluster.Consumer)
	}
	consumer, err := cluster.NewConsumer(cr.brokers, groupId, []string{topic}, cr.conf)
	if err != nil {
		cr.mutex.Unlock()
		return err
	}
	cr.consumerMap[topic][groupId] = consumer
	log.Infof("Register kafka consumer success for [%s, %s]\n", topic, groupId)
	cr.mutex.Unlock()

	go func() {
		for err := range consumer.Errors() {
			log.Errorf("kafka consumer error [%s, %s]: %s\n", topic, groupId, err.Error())
		}
	}()

	// consume notifications
	go func() {
		for ntf := range consumer.Notifications() {
			log.Infof("Notification for [%s, %s] : %+v\n", topic, groupId, ntf)
		}
	}()

	go func() {
		for {
			select {
			case msg, ok := <-consumer.Messages():
				if ok {
					data := (reflect.New(reflect.TypeOf(data))).Interface()
					err := json.Unmarshal(msg.Value, data)
					if err != nil {
						log.Errorf("Kafka consumer for [%s, %s] failed Unmarshal data for data type : %s\n", topic, groupId, reflect.TypeOf(data).Name())
					} else {
						err := action(data)
						if err != nil {
							log.Errorf("Kafka consumer for [%s, %s], message handler execute failed : %s\n", topic, groupId, err.Error())
						}
					}
					consumer.MarkOffset(msg, "") // mark message as processed
				} else {
					log.Errorf("Kafka consumer for [%s, %s] receive message failed\n", topic, groupId)
				}
			}
		}
	}()

	return nil
}

func (cr *ConsumerRegister) Close() {
	for _, mp := range cr.consumerMap {
		for _, cm := range mp {
			cm.Close()
		}
	}
}
