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

package types

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TxStatus uint8

const (
	TX_STATUS_UNKNOWN TxStatus = 0
	TX_STATUS_PENDING TxStatus = 1
	TX_STATUS_SUCCESS TxStatus = 2
	TX_STATUS_FAILED  TxStatus = 3
)

func StatusStr(status TxStatus) string {
	var ret string
	switch status {
	case TX_STATUS_PENDING:
		ret = "pending"
	case TX_STATUS_SUCCESS:
		ret = "success"
	case TX_STATUS_FAILED:
		ret = "failed"
	default:
		ret = "unknown"
	}

	return ret
}

func StrToTxStatus(txType string) TxStatus {
	var ret TxStatus
	switch txType {
	case "pending":
		ret = TX_STATUS_PENDING
	case "success":
		ret = TX_STATUS_SUCCESS
	case "failed":
		ret = TX_STATUS_FAILED
	default:
		ret = TX_STATUS_UNKNOWN
	}

	return ret
}

type TxInfo struct {
	Protocol        common.Address `json:"protocol"`
	DelegateAddress common.Address `json:"delegate"`
	From            common.Address `json:"from"`
	To              common.Address `json:"to"`
	BlockHash       common.Hash    `json:"block_hash"`
	BlockNumber     *big.Int       `json:"block_number"`
	BlockTime       int64          `json:"block_time"`
	TxHash          common.Hash    `json:"tx_hash"`
	TxIndex         int64          `json:"tx_index"`
	TxLogIndex      int64          `json:"tx_log_index"`
	Value           *big.Int       `json:"value"`
	Status          TxStatus       `json:"status"`
	GasLimit        *big.Int       `json:"gas_limit"`
	GasUsed         *big.Int       `json:"gas_used"`
	GasPrice        *big.Int       `json:"gas_price"`
	Nonce           *big.Int       `json:"nonce"`
	Identify        string         `json:"identify"`
}

type TokenRegisterEvent struct {
	TxInfo
	Token  common.Address `json:"token"`
	Symbol string         `json:"symbol"`
}

type TokenUnRegisterEvent struct {
	TxInfo
	Token  common.Address `json:"token"`
	Symbol string         `json:"symbol"`
}

type AddressAuthorizedEvent struct {
	TxInfo
	Protocol common.Address `json:"protocol"`
	Number   int            `json:"number"`
}

type AddressDeAuthorizedEvent struct {
	TxInfo
	Protocol common.Address `json:"protocol"`
	Number   int            `json:"number"`
}

type TransferEvent struct {
	TxInfo
	Sender   common.Address `json:"sender"`
	Receiver common.Address `json:"receiver"`
	Amount   *big.Int       `json:"amount"`
}

type ApprovalEvent struct {
	TxInfo
	Owner   common.Address `json:"owner"`
	Spender common.Address `json:"spender"`
	Amount  *big.Int       `json:"amount"`
}

type OrderFilledEvent struct {
	TxInfo
	Ringhash      common.Hash    `json:"ringhash"`
	PreOrderHash  common.Hash    `json:"pre_order_hash"`
	OrderHash     common.Hash    `json:"order_hash"`
	NextOrderHash common.Hash    `json:"next_order_hash"`
	Owner         common.Address `json:"owner"`
	TokenS        common.Address `json:"token_s"`
	TokenB        common.Address `json:"token_b"`
	SellTo        common.Address `json:"sell_to"`
	BuyFrom       common.Address `json:"buy_from"`
	RingIndex     *big.Int       `json:"ring_index"`
	AmountS       *big.Int       `json:"amount_s"`
	AmountB       *big.Int       `json:"amount_b"`
	LrcReward     *big.Int       `json:"lrc_reward"`
	LrcFee        *big.Int       `json:"lrc_fee"`
	SplitS        *big.Int       `json:"split_s"`
	SplitB        *big.Int       `json:"split_b"`
	Market        string         `json:"market"`
	FillIndex     *big.Int       `json:"fill_index"`
}

type OrderCancelledEvent struct {
	TxInfo
	OrderHash       common.Hash `json:"order_hash"`
	AmountCancelled *big.Int    `json:"amount_cancelled"`
}

type CutoffEvent struct {
	TxInfo
	Owner         common.Address `json:"owner"`
	Cutoff        *big.Int       `json:"cutoff"`
	OrderHashList []common.Hash  `json:"order_hash_list"`
}

type CutoffPairEvent struct {
	TxInfo
	Owner         common.Address `json:"owner"`
	Token1        common.Address `json:"token_1"`
	Token2        common.Address `json:"token_2"`
	Cutoff        *big.Int       `json:"cutoff"`
	OrderHashList []common.Hash  `json:"order_hash_list"`
}

type RingMinedEvent struct {
	TxInfo
	RingIndex    *big.Int       `json:"ring_index"`
	TotalLrcFee  *big.Int       `json:"total_lrc_fee"`
	TradeAmount  int            `json:"trade_amount"`
	Ringhash     common.Hash    `json:"ringhash"`
	Miner        common.Address `json:"miner"`
	FeeRecipient common.Address `json:"fee_recipient"`
	Err          string         `json:"err"`
}

type WethDepositEvent struct {
	TxInfo
	Dst    common.Address `json:"dst"`
	Amount *big.Int       `json:"amount"`
}

type WethWithdrawalEvent struct {
	TxInfo
	Src    common.Address `json:"src"`
	Amount *big.Int       `json:"amount"`
}

type SubmitRingMethodEvent struct {
	TxInfo
	OrderList    []Order        `json:"order_list"`
	FeeReceipt   common.Address `json:"fee_receipt"`
	FeeSelection uint16         `json:"fee_selection"`
	Err          string         `json:"err"`
}

type RingSubmitResultEvent struct {
	RingHash     common.Hash `json:"ring_hash"`
	RingUniqueId common.Hash `json:"ring_unique_id"`
	TxHash       common.Hash `json:"tx_hash"`
	Status       TxStatus    `json:"status"`
	RingIndex    *big.Int    `json:"ring_index"`
	BlockNumber  *big.Int    `json:"block_number"`
	UsedGas      *big.Int    `json:"used_gas"`
	Err          string      `json:"err"`
}

type ForkedEvent struct {
	DetectedBlock *big.Int    `json:"detected_block"`
	DetectedHash  common.Hash `json:"detected_hash"`
	ForkBlock     *big.Int    `json:"fork_block"`
	ForkHash      common.Hash `json:"fork_hash"`
}

type BlockEvent struct {
	BlockNumber *big.Int    `json:"block_number"`
	BlockHash   common.Hash `json:"block_hash"`
	BlockTime   int64       `json:"block_time"`
	IsFinished  bool        `json:"is_finished"`
}

type ExtractorWarningEvent struct{}

type EthTransferEvent struct {
	TxInfo
	Sender   common.Address `json:"sender"`
	Receiver common.Address `json:"receiver"`
	Amount   *big.Int       `json:"amount"`
}

type TransactionEvent struct {
	TxInfo
}

type UnsupportedContractEvent struct {
	TxInfo
}

type SyncCompleteEvent struct {
	BlockNumber *big.Int `json:"block_number"`
}

type DepthUpdateEvent struct {
	DelegateAddress string `json:"delegate_address"`
	Market          string `json:"market"`
}

type BalanceUpdateEvent struct {
	DelegateAddress string `json:"delegate_address"`
	Owner           string `json:"owner"`
}

type KafkaOnChainEvent struct {
	Data  string `json:"data"`
	Topic string `json:"topic"`
}
