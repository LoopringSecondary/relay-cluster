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
	NewOrder = "NewOrder"

	WethDeposit         = "WethDeposit"
	WethWithdrawal      = "WethWithdrawal"
	Approve             = "Approve"
	Transfer            = "Transfer"
	EthTransfer         = "EthTransfer"
	UnsupportedContract = "UnsupportedContract"

	RingMined           = "RingMined"
	OrderFilled         = "OrderFilled"
	CancelOrder         = "CancelOrder"
	CutoffAll           = "Cutoff"
	CutoffPair          = "CutoffPair"
	TokenRegistered     = "TokenRegistered"
	TokenUnRegistered   = "TokenUnRegistered"
	RingHashSubmitted   = "RingHashSubmitted"
	AddressAuthorized   = "AddressAuthorized"
	AddressDeAuthorized = "AddressDeAuthorized"

	MinedOrderState            = "MinedOrderState" //orderbook send orderstate to miner
	WalletTransactionSubmitted = "WalletTransactionSubmitted"

	ExtractorFork   = "ExtractorFork" //chain forked
	Transaction     = "Transaction"
	GatewayNewOrder = "GatewayNewOrder"

	//Miner
	Miner_DeleteOrderState           = "Miner_DeleteOrderState"
	Miner_NewOrderState              = "Miner_NewOrderState"
	Miner_NewRing                    = "Miner_NewRing"
	Miner_RingMined                  = "Miner_RingMined"
	Miner_RingSubmitResult           = "Miner_RingSubmitResult"
	Miner_SubmitRing_Method          = "Miner_SubmitRing_Method"
	Miner_SubmitRingHash_Method      = "Miner_SubmitRingHash_Method"
	Miner_BatchSubmitRingHash_Method = "Miner_BatchSubmitRingHash_Method"

	// Block
	Block_New = "Block_New"
	Block_End = "Block_End"

	// Extractor
	SyncChainComplete = "SyncChainComplete"
	ChainForkDetected = "ChainForkDetected"
	ExtractorWarning  = "ExtractorWarning"

	// Transaction
	TransactionEvent   = "TransactionEvent"
	PendingTransaction = "PendingTransaction"

	// socketio notify event types
	LoopringTickerUpdated = "LoopringTickerUpdated"
	TrendUpdated          = "TrendUpdated"
	PortfolioUpdated      = "PortfolioUpdated"
	BalanceUpdated        = "BalanceUpdated"
	DepthUpdated          = "DepthUpdated"
	TransactionUpdated    = "TransactionUpdated"
)
