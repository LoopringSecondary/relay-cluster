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
	"github.com/Shopify/sarama"
)

type MessageProducer struct {
	pd sarama.SyncProducer
}

func (md *MessageProducer) Initialize(brokerList []string) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err == nil {
		md.pd = producer
	}
	return err
}

func (md *MessageProducer) SendMessage(topic string, data interface{}, key string) (partition int32, offset int64, sendErr error) {
	if data == nil {
		return -1, -1, fmt.Errorf("kafka message to send is null for topic : %s", topic)
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return -1, -1, fmt.Errorf("failed to Marshal kafka Msg %+v for topic : %s", data, topic)
	}
	return md.pd.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(bytes),
		Key:   sarama.StringEncoder(key),
	})
}

func (md *MessageProducer) Close() error {
	return md.pd.Close()
}
