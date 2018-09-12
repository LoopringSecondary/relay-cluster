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
	"github.com/Loopring/relay-lib/eth/contract"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	//"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func setTxInfo(tx *ethtyp.Transaction, gasUsed, blockTime *big.Int, methodName string) types.TxInfo {
	var txinfo types.TxInfo

	txinfo.Protocol = common.HexToAddress(tx.To)
	txinfo.From = common.HexToAddress(tx.From)
	txinfo.To = common.HexToAddress(tx.To)

	if impl, ok := loopringaccessor.ProtocolAddresses()[txinfo.To]; ok {
		txinfo.DelegateAddress = impl.DelegateAddress
	} else {
		txinfo.DelegateAddress = types.NilAddress
	}

	txinfo.BlockNumber = tx.BlockNumber.BigInt()
	txinfo.BlockTime = blockTime.Int64()
	txinfo.BlockHash = common.HexToHash(tx.BlockHash)
	txinfo.TxHash = common.HexToHash(tx.Hash)
	txinfo.TxIndex = tx.TransactionIndex.Int64()
	txinfo.Value = tx.Value.BigInt()

	txinfo.GasLimit = tx.Gas.BigInt()
	txinfo.GasUsed = gasUsed
	txinfo.GasPrice = tx.GasPrice.BigInt()
	txinfo.Nonce = tx.Nonce.BigInt()

	txinfo.Identify = methodName

	return txinfo
}

func handleOtherTransaction(tx *ethtyp.Transaction, receipt *ethtyp.TransactionReceipt, time *big.Int) error {
	gasUsed := getGasUsed(receipt)
	txinfo := setTxInfo(tx, gasUsed, time, contract.METHOD_UNKNOWN)

	if tx.Value.BigInt().Cmp(big.NewInt(0)) > 0 {
		event := &types.EthTransferEvent{}
		event.TxInfo = txinfo
		event.Amount = tx.Value.BigInt()
		event.TxLogIndex = 0
		event.Status = getStatus(tx, receipt)
		event.Sender = common.HexToAddress(tx.From)
		event.Receiver = common.HexToAddress(tx.To)

		//log.Debugf("extractor,tx:%s handleEthTransfer sender:%s, receiver:%s, value:%s, gasUsed:%s, status:%d", event.TxHash.Hex(), event.Sender.Hex(), event.Receiver.Hex(), event.Amount.String(), event.GasUsed.String(), event.Status)

		return Produce(event)
	} else {
		event := &types.UnsupportedContractEvent{}
		event.TxInfo = txinfo
		event.TxLogIndex = 0
		event.Status = getStatus(tx, receipt)

		//log.Debugf("extractor,tx:%s handleUnSupportedContract from:%s, to:%s, gasUsed:%s, status:%d", event.TxHash.Hex(), event.From.Hex(), event.To.Hex(), event.GasUsed.String(), event.Status)

		return Produce(event)
	}
}

func getGasUsed(receipt *ethtyp.TransactionReceipt) *big.Int {
	var gasUsed *big.Int

	if receipt == nil {
		gasUsed = big.NewInt(0)
	} else {
		gasUsed = receipt.GasUsed.BigInt()
	}

	return gasUsed
}

func getStatus(tx *ethtyp.Transaction, receipt *ethtyp.TransactionReceipt) types.TxStatus {
	var status types.TxStatus

	if receipt == nil {
		status = types.TX_STATUS_PENDING
	} else if receipt.Failed(tx) {
		status = types.TX_STATUS_FAILED
	} else {
		status = types.TX_STATUS_SUCCESS
	}

	return status
}
