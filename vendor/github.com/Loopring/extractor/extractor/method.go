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
	"github.com/Loopring/relay-lib/eth/abi"
	"github.com/Loopring/relay-lib/eth/contract"
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	//"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

type MethodData struct {
	types.TxInfo
	Method interface{}
	Abi    *abi.ABI
	Id     string
	Name   string
}

func newMethodData(method *abi.Method, cabi *abi.ABI) MethodData {
	var c MethodData

	c.Id = common.ToHex(method.Id())
	c.Name = method.Name
	c.Abi = cabi

	return c
}

// UnpackMethod v should be ptr
func (m *MethodData) handleMethod(tx *ethtyp.Transaction, gasUsed, blockTime *big.Int, status types.TxStatus, methodName string) error {
	if err := m.beforeUnpack(tx, gasUsed, blockTime, status, methodName); err != nil {
		return err
	}
	if m.Name != contract.METHOD_WETH_DEPOSIT {
		if err := m.unpack(tx); err != nil {
			return err
		}
	}
	if err := m.afterUnpack(); err != nil {
		return err
	}

	return nil
}

// beforeUnpack full fill method txinfo and set status...
func (m *MethodData) beforeUnpack(tx *ethtyp.Transaction, gasUsed, blockTime *big.Int, status types.TxStatus, methodName string) error {
	m.TxInfo = setTxInfo(tx, gasUsed, blockTime, methodName)
	m.TxLogIndex = 0
	m.Status = status

	switch m.Name {
	case contract.METHOD_CANCEL_ORDER:
		if m.DelegateAddress == types.NilAddress {
			return fmt.Errorf("cancelOrder method cann't get delegate address")
		}
	}

	return nil
}

func (m *MethodData) unpack(tx *ethtyp.Transaction) error {
	const methodStartIdx = 10

	if len([]byte(tx.Input)) < methodStartIdx {
		return fmt.Errorf("unpack method error, input length invalid:%d", len([]byte(tx.Input)))
	}

	data := hexutil.MustDecode("0x" + tx.Input[methodStartIdx:])
	return m.Abi.UnpackMethod(m.Method, m.Name, data)
}

// afterUnpack set special fields in internal event
func (m *MethodData) afterUnpack() error {
	var (
		event interface{}
		err   error
	)

	switch m.Name {
	case contract.METHOD_SUBMIT_RING:
		event, err = m.getSubmitRingEvent()
	case contract.METHOD_CANCEL_ORDER:
		event, err = m.getOrderCancelledEvent()
	case contract.METHOD_CUTOFF_ALL:
		event, err = m.getCutoffAllEvent()
	case contract.METHOD_CUTOFF_PAIR:
		event, err = m.getCutoffPairEvent()
	case contract.METHOD_APPROVE:
		event, err = m.getApproveEvent()
	case contract.METHOD_TRANSFER:
		event, err = m.getTransferEvent()
	case contract.METHOD_WETH_DEPOSIT:
		event, err = m.getDepositEvent()
	case contract.METHOD_WETH_WITHDRAWAL:
		event, err = m.getWithdrawalEvent()
	}

	if err != nil {
		return err
	}

	return Produce(event)
}

func (m *MethodData) getSubmitRingEvent() (*types.SubmitRingMethodEvent, error) {
	var (
		event = &types.SubmitRingMethodEvent{}
		err   error
	)

	src, ok := m.Method.(*contract.SubmitRingMethodInputs)
	if !ok {
		return nil, fmt.Errorf("submitRing method inputs type error")
	}

	if event, err = src.ConvertDown(m.Protocol, m.DelegateAddress); err != nil {
		return event, fmt.Errorf("submitRing method inputs convert error:%s", err.Error())
	}

	// set txinfo for event
	event.TxInfo = m.TxInfo
	if event.Status == types.TX_STATUS_FAILED {
		event.Err = fmt.Sprintf("method %s transaction failed", contract.METHOD_SUBMIT_RING)
	}

	// 不需要发送订单到gateway
	//for _, v := range event.OrderList {
	//	v.Hash = v.GenerateHash()
	//	log.Debugf("extractor,tx:%s submitRing method orderHash:%s,owner:%s,tokenS:%s,tokenB:%s,amountS:%s,amountB:%s", event.TxHash.Hex(), v.Hash.Hex(), v.Owner.Hex(), v.TokenS.Hex(), v.TokenB.Hex(), v.AmountS.String(), v.AmountB.String())
	//	eventemitter.Produce(eventemitter.GatewayNewOrder, v)
	//}

	//log.Debugf("extractor,tx:%s submitRing method gas:%s, gasprice:%s, status:%s", event.TxHash.Hex(), event.GasUsed.String(), event.GasPrice.String(), types.StatusStr(event.Status))

	return event, nil
}

func (m *MethodData) getOrderCancelledEvent() (*types.OrderCancelledEvent, error) {
	src, ok := m.Method.(*contract.CancelOrderMethod)
	if !ok {
		return nil, fmt.Errorf("cancelOrder method inputs type error")
	}

	order, cancelAmount, _ := src.ConvertDown(m.Protocol, m.DelegateAddress)

	event := &types.OrderCancelledEvent{}
	event.TxInfo = m.TxInfo
	event.OrderHash = order.Hash
	event.AmountCancelled = cancelAmount

	//log.Debugf("extractor,tx:%s cancelOrder method order tokenS:%s,tokenB:%s,amountS:%s,amountB:%s", event.TxHash.Hex(), order.TokenS.Hex(), order.TokenB.Hex(), order.AmountS.String(), order.AmountB.String())

	return event, nil
}

func (m *MethodData) getCutoffAllEvent() (*types.CutoffEvent, error) {
	src, ok := m.Method.(*contract.CutoffMethod)
	if !ok {
		return nil, fmt.Errorf("cutoffAll method inputs type error")
	}

	event := src.ConvertDown(m.From)
	event.TxInfo = m.TxInfo

	//log.Debugf("extractor,tx:%s cutoff method owner:%s, cutoff:%d, status:%d", event.TxHash.Hex(), event.Owner.Hex(), event.Cutoff.Int64(), event.Status)

	return event, nil
}

func (m *MethodData) getCutoffPairEvent() (*types.CutoffPairEvent, error) {
	src, ok := m.Method.(*contract.CutoffPairMethod)
	if !ok {
		return nil, fmt.Errorf("cutoffPair method inputs type error")
	}

	event := src.ConvertDown(m.From)
	event.TxInfo = m.TxInfo

	//log.Debugf("extractor,tx:%s cutoffpair method owenr:%s, token1:%s, token2:%s, cutoff:%d", event.TxHash.Hex(), event.Owner.Hex(), event.Token1.Hex(), event.Token2.Hex(), event.Cutoff.Int64())

	return event, nil
}

func (m *MethodData) getApproveEvent() (*types.ApprovalEvent, error) {
	src, ok := m.Method.(*contract.ApproveMethod)
	if !ok {
		return nil, fmt.Errorf("approve method inputs type error")
	}

	event := src.ConvertDown(m.From)
	event.TxInfo = m.TxInfo

	//log.Debugf("extractor,tx:%s approve method owner:%s, spender:%s, value:%s", event.TxHash.Hex(), event.Owner.Hex(), event.Spender.Hex(), event.Amount.String())

	return event, nil
}

func (m *MethodData) getTransferEvent() (*types.TransferEvent, error) {
	src, ok := m.Method.(*contract.TransferMethod)
	if !ok {
		return nil, fmt.Errorf("transfer method inputs type error")
	}

	event := src.ConvertDown(m.From)
	event.TxInfo = m.TxInfo

	//log.Debugf("extractor,tx:%s transfer method sender:%s, receiver:%s, value:%s", event.TxHash.Hex(), event.Sender.Hex(), event.Receiver.Hex(), event.Amount.String())

	return event, nil
}

func (m *MethodData) getDepositEvent() (*types.WethDepositEvent, error) {
	event := &types.WethDepositEvent{}
	event.Dst = m.From
	event.Amount = m.Value
	event.TxInfo = m.TxInfo

	//log.Debugf("extractor,tx:%s wethDeposit method to:%s, value:%s", event.TxHash.Hex(), event.Dst.Hex(), event.Amount.String())

	return event, nil
}

func (m *MethodData) getWithdrawalEvent() (*types.WethWithdrawalEvent, error) {
	src, ok := m.Method.(*contract.WethWithdrawalMethod)
	if !ok {
		return nil, fmt.Errorf("wethWithdrawal method inputs type error")
	}

	event := src.ConvertDown(m.From)
	event.TxInfo = m.TxInfo

	//log.Debugf("extractor,tx:%s wethWithdrawal method from:%s, value:%s", event.TxHash.Hex(), event.Src.Hex(), event.Amount.String())

	return event, nil
}
