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

package contract

import (
	"fmt"
	"github.com/Loopring/relay-lib/crypto"
	"github.com/Loopring/relay-lib/eth/abi"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

const (
	METHOD_UNKNOWN              = ""
	METHOD_SUBMIT_RING          = "submitRing"
	METHOD_CANCEL_ORDER         = "cancelOrder"
	METHOD_CUTOFF_ALL           = "cancelAllOrders"
	METHOD_CUTOFF_PAIR          = "cancelAllOrdersByTradingPair"
	METHOD_WETH_DEPOSIT         = "deposit"
	METHOD_WETH_WITHDRAWAL      = "withdraw"
	METHOD_APPROVE              = "approve"
	METHOD_TRANSFER             = "transfer"
	METHOD_TOKEN_REGISTRY       = "registerToken"
	METHOD_TOKEN_UNREGISTRY     = "unregisterToken"
	METHOD_ADDRESS_AUTHORIZED   = "authorizeAddress"
	METHOD_ADDRESS_DEAUTHORIZED = "deauthorizeAddress"
)

func TxIsSubmitRing(methodName string) bool {
	if methodName == METHOD_SUBMIT_RING {
		return true
	}

	return false
}

type SubmitRingMethodInputs struct {
	AddressList        [][4]common.Address `fieldName:"addressList" fieldId:"0"`   // owner,tokenS, wallet, authAddress
	UintArgsList       [][6]*big.Int       `fieldName:"uintArgsList" fieldId:"1"`  // amountS, amountB, validSince (second),validUntil (second), lrcFee, rateAmountS.
	Uint8ArgsList      [][1]uint8          `fieldName:"uint8ArgsList" fieldId:"2"` // marginSplitPercentageList
	BuyNoMoreThanBList []bool              `fieldName:"buyNoMoreThanAmountBList" fieldId:"3"`
	VList              []uint8             `fieldName:"vList" fieldId:"4"`
	RList              [][32]byte          `fieldName:"rList" fieldId:"5"`
	SList              [][32]byte          `fieldName:"sList" fieldId:"6"`
	FeeRecipient       common.Address      `fieldName:"feeRecipient" fieldId:"7"`
	FeeSelections      uint16              `fieldName:"feeSelections" fieldId:"8"`
	Protocol           common.Address
	FeeReceipt         common.Address
}

func GenerateSubmitRingMethodInputsData(ring *types.Ring, feeReceipt common.Address, protocolAbi *abi.ABI) ([]byte, error) {
	inputs := &SubmitRingMethodInputs{}
	inputs = emptySubmitRingInputs(feeReceipt)
	authVList := []uint8{}
	authRList := [][32]byte{}
	authSList := [][32]byte{}
	if types.IsZeroHash(ring.Hash) {
		ring.Hash = ring.GenerateHash(feeReceipt)
	}
	for _, filledOrder := range ring.Orders {
		order := filledOrder.OrderState.RawOrder
		inputs.AddressList = append(inputs.AddressList, [4]common.Address{order.Owner, order.TokenS, order.WalletAddress, order.AuthAddr})
		rateAmountS, _ := new(big.Int).SetString(filledOrder.RateAmountS.FloatString(0), 10)
		inputs.UintArgsList = append(inputs.UintArgsList, [6]*big.Int{order.AmountS, order.AmountB, order.ValidSince, order.ValidUntil, order.LrcFee, rateAmountS})
		inputs.Uint8ArgsList = append(inputs.Uint8ArgsList, [1]uint8{order.MarginSplitPercentage})

		inputs.BuyNoMoreThanBList = append(inputs.BuyNoMoreThanBList, order.BuyNoMoreThanAmountB)

		inputs.VList = append(inputs.VList, order.V)
		inputs.RList = append(inputs.RList, order.R)
		inputs.SList = append(inputs.SList, order.S)

		//sign By authPrivateKey
		if signBytes, err := order.AuthPrivateKey.Sign(ring.Hash.Bytes(), order.AuthPrivateKey.Address()); nil == err {
			v, r, s := crypto.SigToVRS(signBytes)
			authVList = append(authVList, v)
			authRList = append(authRList, types.BytesToBytes32(r).Bytes32())
			authSList = append(authSList, types.BytesToBytes32(s).Bytes32())
		} else {
			return []byte{}, err
		}
	}

	inputs.VList = append(inputs.VList, authVList...)
	inputs.RList = append(inputs.RList, authRList...)
	inputs.SList = append(inputs.SList, authSList...)

	inputs.FeeSelections = uint16(ring.FeeSelections().Uint64())

	return protocolAbi.Pack("submitRing",
		inputs.AddressList,
		inputs.UintArgsList,
		inputs.Uint8ArgsList,
		inputs.BuyNoMoreThanBList,
		inputs.VList,
		inputs.RList,
		inputs.SList,
		inputs.FeeReceipt,
		inputs.FeeSelections,
	)
	//if err := ring.GenerateAndSetSignature(miner); nil != err {
	//	return nil, err
	//} else {
	//	ringSubmitArgs.VList = append(ringSubmitArgs.VList, ring.V)
	//	ringSubmitArgs.RList = append(ringSubmitArgs.RList, ring.R)
	//	ringSubmitArgs.SList = append(ringSubmitArgs.SList, ring.S)
	//}
}

func emptySubmitRingInputs(feeReceipt common.Address) *SubmitRingMethodInputs {
	return &SubmitRingMethodInputs{
		AddressList:        [][4]common.Address{},
		UintArgsList:       [][6]*big.Int{},
		Uint8ArgsList:      [][1]uint8{},
		BuyNoMoreThanBList: []bool{},
		VList:              []uint8{},
		RList:              [][32]byte{},
		SList:              [][32]byte{},
		FeeReceipt:         feeReceipt,
	}
}

// should add protocol, miner, feeRecipient
func (m *SubmitRingMethodInputs) ConvertDown() (*types.SubmitRingMethodEvent, error) {
	var (
		list  []types.Order
		event types.SubmitRingMethodEvent
	)

	length := len(m.AddressList)
	vrsLength := 2 * length

	orderLengthInvalid := length < 2
	argLengthInvalid := length != len(m.UintArgsList) || length != len(m.Uint8ArgsList)
	vrsLengthInvalid := vrsLength != len(m.VList) || vrsLength != len(m.RList) || vrsLength != len(m.SList)
	if orderLengthInvalid || argLengthInvalid || vrsLengthInvalid {
		return nil, fmt.Errorf("submitRing method unpack error:orders length invalid")
	}

	for i := 0; i < length; i++ {
		var order types.Order

		order.Protocol = m.Protocol
		order.Owner = m.AddressList[i][0]
		order.TokenS = m.AddressList[i][1]
		if i == length-1 {
			order.TokenB = m.AddressList[0][1]
		} else {
			order.TokenB = m.AddressList[i+1][1]
		}
		order.WalletAddress = m.AddressList[i][2]
		order.AuthAddr = m.AddressList[i][3]

		order.AmountS = m.UintArgsList[i][0]
		order.AmountB = m.UintArgsList[i][1]
		order.ValidSince = m.UintArgsList[i][2]
		order.ValidUntil = m.UintArgsList[i][3]
		order.LrcFee = m.UintArgsList[i][4]
		// order.rateAmountS

		order.MarginSplitPercentage = m.Uint8ArgsList[i][0]

		order.BuyNoMoreThanAmountB = m.BuyNoMoreThanBList[i]

		order.V = m.VList[i]
		order.R = m.RList[i]
		order.S = m.SList[i]

		list = append(list, order)
	}

	event.OrderList = list
	event.FeeReceipt = m.FeeRecipient
	event.FeeSelection = m.FeeSelections
	event.Err = ""

	return &event, nil
}

type CancelOrderMethod struct {
	AddressList    [5]common.Address `fieldName:"addresses" fieldId:"0"`   //  owner, tokenS, tokenB, authAddr
	OrderValues    [6]*big.Int       `fieldName:"orderValues" fieldId:"1"` //  amountS, amountB, validSince (second), validUntil (second), lrcFee, and cancelAmount
	BuyNoMoreThanB bool              `fieldName:"buyNoMoreThanAmountB" fieldId:"2"`
	MarginSplit    uint8             `fieldName:"marginSplitPercentage" fieldId:"3"`
	V              uint8             `fieldName:"v" fieldId:"4"`
	R              [32]byte          `fieldName:"r" fieldId:"5"`
	S              [32]byte          `fieldName:"s" fieldId:"6"`
}

// todo(fuk): modify internal cancelOrderMethod and implement related functions
func (m *CancelOrderMethod) ConvertDown() (*types.Order, *big.Int, error) {
	var order types.Order

	order.Owner = m.AddressList[0]
	order.TokenS = m.AddressList[1]
	order.TokenB = m.AddressList[2]
	order.WalletAddress = m.AddressList[3]
	order.AuthAddr = m.AddressList[4]

	order.AmountS = m.OrderValues[0]
	order.AmountB = m.OrderValues[1]
	order.ValidSince = m.OrderValues[2]
	order.ValidUntil = m.OrderValues[3]
	order.LrcFee = m.OrderValues[4]
	cancelAmount := m.OrderValues[5]

	order.BuyNoMoreThanAmountB = bool(m.BuyNoMoreThanB)
	order.MarginSplitPercentage = m.MarginSplit

	order.V = m.V
	order.S = m.S
	order.R = m.R

	return &order, cancelAmount, nil
}

type CutoffMethod struct {
	Cutoff *big.Int `fieldName:"cutoff" fieldId:"0"`
}

func (method *CutoffMethod) ConvertDown() *types.CutoffEvent {
	evt := &types.CutoffEvent{}
	evt.Cutoff = method.Cutoff

	return evt
}

type CutoffPairMethod struct {
	Token1 common.Address `fieldName:"token1" fieldId:"0"`
	Token2 common.Address `fieldName:"token2" fieldId:"1"`
	Cutoff *big.Int       `fieldName:"cutoff" fieldId:"2"`
}

func (method *CutoffPairMethod) ConvertDown() *types.CutoffPairEvent {
	evt := &types.CutoffPairEvent{}
	evt.Cutoff = method.Cutoff
	evt.Token1 = method.Token1
	evt.Token2 = method.Token2

	return evt
}

type WethDepositMethod struct {
	Value *big.Int `fieldId:"0"`
}

func (e *WethDepositMethod) ConvertDown() *types.WethDepositEvent {
	evt := &types.WethDepositEvent{}
	return evt
}

type WethWithdrawalMethod struct {
	Value *big.Int `fieldName:"wad" fieldId:"0"`
}

func (e *WethWithdrawalMethod) ConvertDown() *types.WethWithdrawalEvent {
	evt := &types.WethWithdrawalEvent{}
	evt.Amount = e.Value

	return evt
}

type ApproveMethod struct {
	Spender common.Address `fieldName:"spender" fieldId:"0"`
	Value   *big.Int       `fieldName:"value" fieldId:"1"`
}

func (e *ApproveMethod) ConvertDown() *types.ApprovalEvent {
	evt := &types.ApprovalEvent{}
	evt.Spender = e.Spender
	evt.Amount = e.Value

	return evt
}

// function transfer(address to, uint256 value) public returns (bool);
type TransferMethod struct {
	Receiver common.Address `fieldName:"to" fieldId:"0"`
	Value    *big.Int       `fieldName:"value" fieldId:"1"`
}

func (e *TransferMethod) ConvertDown() *types.TransferEvent {
	evt := &types.TransferEvent{}
	evt.Receiver = e.Receiver
	evt.Amount = e.Value

	return evt
}
