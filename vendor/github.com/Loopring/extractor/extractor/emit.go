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

package extractor

import (
	"fmt"
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	ex "github.com/Loopring/relay-lib/extractor"
	"github.com/Loopring/relay-lib/kafka"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/zklock"
)

const (
	ZKNAME_EXTRACTOR   = "extractor"
	KAFKA_PRODUCER_KEY = "extractor"
)

var (
	producer *kafka.MessageProducer
	register *kafka.ConsumerRegister
)

func RegistryEmitter(producerOpt, consumerOpt kafka.KafkaOptions, service ExtractorService) error {
	brokers := producerOpt.Brokers
	producer = &kafka.MessageProducer{}
	if err := producer.Initialize(brokers); err != nil {
		return err
	}

	if len(consumerOpt.Brokers) < 1 {
		return fmt.Errorf("kafka consumer brokers should not be empty")
	}

	register = &kafka.ConsumerRegister{}
	register.Initialize(consumerOpt.Brokers)
	err := register.RegisterTopicAndHandler(
		kafka.Kafka_Topic_Extractor_PendingTransaction,
		kafka.Kafka_Group_Extractor_PendingTransaction,
		ethtyp.Transaction{},
		service.WatchingPendingTransaction)

	return err
}

func UnRegistryEmitter() {
	zklock.ReleaseLock(ZKNAME_EXTRACTOR)
	producer.Close()
	register.Close()
}

func Produce(src interface{}) error {
	event, err := ex.Assemble(src)
	if err != nil {
		return err
	}
	producer.SendMessage(kafka.Kafka_Topic_Extractor_EventOnChain, event, KAFKA_PRODUCER_KEY)
	log.Debugf("emit topic:%s, data:%s", event.Topic, event.Data)

	return nil
}
