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
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/txmanager/types"
	"github.com/Loopring/relay-lib/kafka"
	"github.com/Loopring/relay-lib/log"
	libTypes "github.com/Loopring/relay-lib/types"
)

// todo delete return after test

func NotifyOrderUpdate(o *libTypes.OrderState) error {
	err := ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Order_Updated, o)
	if err != nil {
		log.Error("notify new order failed. " + o.RawOrder.Hash.Hex())
	} else {
		log.Info("notify new order success")
	}
	return err
}

func NotifyOrderFilled(f *dao.FillEvent) error {
	err := ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Trades_Updated, f)
	if err != nil {
		log.Error("notify order fill failed. " + f.OrderHash)
	}
	return err
}

func NotifyCutoff(evt *libTypes.CutoffEvent) error {

	err := ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Cutoff, evt)
	if err != nil {
		log.Error("notify cutoff failed. " + evt.Owner.Hex())
	}
	return err
}

func NotifyCutoffPair(evt *libTypes.CutoffPairEvent) error {
	err := ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Cutoff_Pair, evt)
	if err != nil {
		log.Error("notify cutoff pair failed. " + evt.Owner.Hex())
	}
	return err
}

func NotifyTransactionView(tx *types.TransactionView) error {
	err := ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Transaction_Updated, tx)
	if err != nil {
		log.Error("notify cutoff failed. " + tx.TxHash.Hex())
	}
	return err
}

func NotifyAccountBalanceUpdate(event *libTypes.BalanceUpdateEvent) error {
	err := ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_BalanceUpdated, event)
	if err != nil {
		log.Error("notify cutoff failed. " + event.Owner)
	}
	return err
}
