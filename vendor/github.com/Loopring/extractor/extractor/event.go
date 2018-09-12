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
	"github.com/Loopring/relay-lib/eth/abi"
	"github.com/Loopring/relay-lib/eth/contract"
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	//"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

type EventData struct {
	types.TxInfo
	Event interface{}
	Abi   *abi.ABI
	Id    common.Hash
	Name  string
}

func newEventData(event *abi.Event, cabi *abi.ABI) EventData {
	var c EventData

	c.Id = event.Id()
	c.Name = event.Name
	c.Abi = cabi

	return c
}

func (e *EventData) handleEvent(tx *ethtyp.Transaction, evtLog *ethtyp.Log, gasUsed, blockTime *big.Int, methodName string) error {
	if err := e.beforeUnpack(tx, evtLog, gasUsed, blockTime, methodName); err != nil {
		return err
	}
	if err := e.unpack(evtLog); err != nil {
		return err
	}
	if err := e.afterUnpack(); err != nil {
		return err
	}

	return nil
}

func (e *EventData) beforeUnpack(tx *ethtyp.Transaction, evtLog *ethtyp.Log, gasUsed, blockTime *big.Int, methodName string) error {
	e.TxInfo = setTxInfo(tx, gasUsed, blockTime, methodName)
	e.Protocol = common.HexToAddress(evtLog.Address)
	e.TxLogIndex = evtLog.LogIndex.Int64()
	e.Status = types.TX_STATUS_SUCCESS

	return nil
}

func (e *EventData) unpack(evtLog *ethtyp.Log) (err error) {
	var decodedValues [][]byte
	data := hexutil.MustDecode(evtLog.Data)
	for _, topic := range evtLog.Topics {
		decodeBytes := hexutil.MustDecode(topic)
		decodedValues = append(decodedValues, decodeBytes)
	}
	return e.Abi.UnpackEvent(e.Event, e.Name, data, decodedValues)
}

func (e *EventData) afterUnpack() error {
	if e.Name == contract.EVENT_RING_MINED {
		ringmined, fills, err := e.getRingMinedEvents()
		if err != nil {
			return err
		}
		Produce(ringmined)
		for _, fill := range fills {
			Produce(fill)
		}
		return nil
	}

	var (
		event interface{}
		err   error
	)

	switch e.Name {
	case contract.EVENT_ORDER_CANCELLED:
		event, err = e.getOrderCancelledEvent()
	case contract.EVENT_CUTOFF_ALL:
		event, err = e.getCutoffAllEvent()
	case contract.EVENT_CUTOFF_PAIR:
		event, err = e.getCutoffPairEvent()
	case contract.EVENT_TRANSFER:
		event, err = e.getTransferEvent()
	case contract.EVENT_APPROVAL:
		event, err = e.getApprovalEvent()
	case contract.EVENT_WETH_DEPOSIT:
		event, err = e.getDepositEvent()
	case contract.EVENT_WETH_WITHDRAWAL:
		event, err = e.getWithdrawalEvent()
	case contract.EVENT_TOKEN_REGISTERED:
		event, err = e.getTokenRegisteredEvent()
	case contract.EVENT_TOKEN_UNREGISTERED:
		event, err = e.getTokenUnRegisteredEvent()
	case contract.EVENT_ADDRESS_AUTHORIZED:
		event, err = e.getAddressAuthorizedEvent()
	case contract.EVENT_ADDRESS_DEAUTHORIZED:
		event, err = e.getAddressDeAuthorizedEvent()
	}

	if err != nil {
		return err
	}

	return Produce(event)
}

func (e *EventData) getRingMinedEvents() (*types.RingMinedEvent, []*types.OrderFilledEvent, error) {
	src := e.Event.(*contract.RingMinedEvent)

	event, fills, err := src.ConvertDown()
	if err != nil {
		return nil, fills, err
	}

	event.TxInfo = e.TxInfo
	for _, v := range fills {
		v.TxInfo = e.TxInfo
	}

	//log.Debugf("extractor,tx:%s ringMined event logIndex:%d, ringhash:%s, ringIndex:%s", event.TxHash.Hex(), event.TxLogIndex, event.Ringhash.Hex(), event.RingIndex.String())
	return event, fills, nil
}

func (e *EventData) getOrderCancelledEvent() (*types.OrderCancelledEvent, error) {
	src := e.Event.(*contract.OrderCancelledEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s orderCancelled event orderhash:%s, cancelAmount:%s", event.TxHash.Hex(), event.OrderHash.Hex(), event.AmountCancelled.String())

	return event, nil
}

func (e *EventData) getCutoffAllEvent() (*types.CutoffEvent, error) {
	src := e.Event.(*contract.CutoffEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s cutoffTimestampChanged event ownerAddress:%s, cutOffTime:%s, status:%d", event.TxHash.Hex(), event.Owner.Hex(), event.Cutoff.String(), event.Status)

	return event, nil
}

func (e *EventData) getCutoffPairEvent() (*types.CutoffPairEvent, error) {
	src := e.Event.(*contract.CutoffPairEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s cutoffPair event ownerAddress:%s, token1:%s, token2:%s, cutOffTime:%s", event.TxHash.Hex(), event.Owner.Hex(), event.Token1.Hex(), event.Token2.Hex(), event.Cutoff.String())

	return event, nil
}

func (e *EventData) getTransferEvent() (*types.TransferEvent, error) {
	src := e.Event.(*contract.TransferEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s tokenTransfer event, methodName:%s, logIndex:%d, from:%s, to:%s, value:%s", event.TxHash.Hex(), event.Identify, event.TxLogIndex, event.Sender.Hex(), event.Receiver.Hex(), event.Amount.String())

	return event, nil
}

func (e *EventData) getApprovalEvent() (*types.ApprovalEvent, error) {
	src := e.Event.(*contract.ApprovalEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s approval event owner:%s, spender:%s, value:%s", event.TxHash.Hex(), event.Owner.Hex(), event.Spender.Hex(), event.Amount.String())

	return event, nil
}

func (e *EventData) getDepositEvent() (*types.WethDepositEvent, error) {
	src := e.Event.(*contract.WethDepositEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s wethDeposit event deposit to:%s, number:%s", event.TxHash.Hex(), event.Dst.Hex(), event.Amount.String())

	return event, nil
}

func (e *EventData) getWithdrawalEvent() (*types.WethWithdrawalEvent, error) {
	src := e.Event.(*contract.WethWithdrawalEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s wethWithdrawal event withdrawal from:%s, number:%s", event.TxHash.Hex(), event.Src.Hex(), event.Amount.String())

	return event, nil
}

func (e *EventData) getTokenRegisteredEvent() (*types.TokenRegisterEvent, error) {
	src := e.Event.(*contract.TokenRegisteredEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s tokenRegistered event address:%s, symbol:%s", event.TxHash.Hex(), event.Token.Hex(), event.Symbol)

	return event, nil
}

func (e *EventData) getTokenUnRegisteredEvent() (*types.TokenUnRegisterEvent, error) {
	src := e.Event.(*contract.TokenUnRegisteredEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s tokenUnregistered event address:%s, symbol:%s", event.TxHash.Hex(), event.Token.Hex(), event.Symbol)

	return event, nil
}

func (e *EventData) getAddressAuthorizedEvent() (*types.AddressAuthorizedEvent, error) {
	src := e.Event.(*contract.AddressAuthorizedEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s addressAuthorized event address:%s, number:%d", event.TxHash.Hex(), event.Protocol.Hex(), event.Number)

	return event, nil
}

func (e *EventData) getAddressDeAuthorizedEvent() (*types.AddressDeAuthorizedEvent, error) {
	src := e.Event.(*contract.AddressDeAuthorizedEvent)

	event := src.ConvertDown()
	event.TxInfo = e.TxInfo

	//log.Debugf("extractor,tx:%s addressDeAuthorized event address:%s, number:%d", event.TxHash.Hex(), event.Protocol.Hex(), event.Number)

	return event, nil
}
