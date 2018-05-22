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

package util

import (
	"github.com/Loopring/relay-cluster/txmanager/types"
	"github.com/Loopring/relay-lib/kafka"
	libTypes "github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

// todo delete return after test

func NotifyGatewayOrder(owner common.Address, market string) error {
	return nil
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Depth_Updated, market)
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Orderbook_Updated, market)
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Orders_Updated, owner)

	return nil
}

func NotifyOrderFilled(owner common.Address, market string) error {
	return nil
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Depth_Updated, market)
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Orderbook_Updated, market)
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Orders_Updated, owner)
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Trades_Updated, market)

	return nil
}

func NotifyOrderCancelled(owner common.Address, market string) error {
	return nil
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Depth_Updated, market)
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Orderbook_Updated, market)
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Orders_Updated, owner)

	return nil
}

func NotifyCutoff(owner common.Address) error {
	return nil
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Depth_Updated, "")
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Orderbook_Updated, "")
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Orders_Updated, owner)

	return nil
}

func NotifyCutoffPair(owner common.Address, market string) error {
	return nil

	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Depth_Updated, market)
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Orderbook_Updated, market)
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Orders_Updated, owner)

	return nil
}

func NotifyTransactionView(tx *types.TransactionView) error {
	return nil
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Transactions_Updated, tx)
	ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_PendingTx_Updated, tx)

	return nil
}

func NotifyAccountBalanceUpdate(event *libTypes.BalanceUpdateEvent) error {
	//todo:
	//ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_BalanceUpdated, event)
	return nil
}
