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

const (
	Kafka_Topic_Extractor_PendingTransaction  = "Kafka_Topic_Extractor_PendingTransaction"
	Kafka_Topic_Extractor_EventOnChain        = "Kafka_Topic_Extractor_EventOnChain"
	Kafka_Topic_AccountManager_BalanceUpdated = "Kafka_Topic_AccountManager_BalanceUpdated"
	Kafka_Topic_RelayCluster_BlockEnd         = "Kafka_Topic_RelayCluster_BlockEnd"

	Kafka_Group_Extractor_PendingTransaction = "Kafka_Group_Extractor_PendingTransaction"
	Kafka_Group_RelayCluster_EventOnChain    = "Kafka_Group_RelayCluster_EventOnChain"
	Kafka_Group_Extractor_EventOnChain       = "Kafka_Group_Extractor_EventOnChain"

	// socket io messages
	Kafka_Topic_SocketIO_Loopring_Ticker_Updated = "Kafka_Topic_SocketIO_Loopring_Ticker_Updated"
	Kafka_Topic_SocketIO_Tickers_Updated         = "Kafka_Topic_SocketIO_Tickers_Updated"
	Kafka_Topic_SocketIO_Trades_Updated          = "Kafka_Topic_SocketIO_Trades_Updated"
	Kafka_Topic_SocketIO_Trends_Updated          = "Kafka_Topic_SocketIO_Trends_Updated"

	Kafka_Topic_SocketIO_Order_Updated = "Kafka_Topic_SocketIO_Order_Updated"
	Kafka_Topic_SocketIO_Cutoff        = "Kafka_Topic_SocketIO_Cutoff"
	Kafka_Topic_SocketIO_Cutoff_Pair   = "Kafka_Topic_SocketIO_Cutoff_Pair"

	Kafka_Topic_SocketIO_BalanceUpdated      = "Kafka_Topic_SocketIO_BalanceUpdated"
	Kafka_Topic_SocketIO_Transaction_Updated = "Kafka_Topic_SocketIO_Transactions_Updated"
)
