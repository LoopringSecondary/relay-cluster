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

package gateway

import (
	"errors"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type CityPartnerStatus struct {
	CustomerCount int
	Received      map[string]string
}

func (w *WalletServiceImpl) CreateCityPartner(req dao.CityPartner) (isSuccessed bool, err error) {
	isSuccessed, err = w.rds.SaveCityPartner(req)
	return
}

func (w *WalletServiceImpl) CreateCustumerInvitationInfo(req dao.CustumerInvitationInfo) (isSuccessed bool, err error) {
	isSuccessed, err = w.rds.SaveCustumerInvitationInfo(req)
	return
}

func (w *WalletServiceImpl) ActivateCustumerInvitation(req dao.CustumerInvitationInfo) (invitationCode string, err error) {
	info, err := w.rds.FindCustumerInvitationInfo(req)
	if nil != err {
		return "", err
	} else {
		info.Activate = info.Activate + 1
		w.rds.AddCustumerInvitationActivate(info)
		return info.InvitationCode, err
	}
}

func (w *WalletServiceImpl) GetCityPartnerStatus(invitationCode string) (*CityPartnerStatus, error) {
	var err error
	var cityPartner *dao.CityPartner
	cityPartner, err = w.rds.FindCityPartnerByInvitationCode(invitationCode)
	if nil == cityPartner || nil != err {
		return nil, err
	}
	status := &CityPartnerStatus{}
	status.CustomerCount, err = w.rds.GetCityPartnerCustomerCount(invitationCode)
	status.Received = make(map[string]string)
	if allReceived, err := w.rds.GetAllReceivedByWallet(cityPartner.WalletAddress); nil != err {
		return status, err
	} else {
		for _, received := range allReceived {
			status.Received[received.TokenSymbol] = received.HumanAmount
		}
		return status, nil
	}
}

func (w *WalletServiceImpl) Start() {
	orderFilledEventWatcher := &eventemitter.Watcher{Concurrent: false, Handle: w.handleFilledEventForCityPartner}
	eventemitter.On(eventemitter.OrderFilled, orderFilledEventWatcher)
}

func (w *WalletServiceImpl) handleFilledEventForCityPartner(input eventemitter.EventData) error {
	event := input.(*types.OrderFilledEvent)
	order, err := w.rds.GetOrderByHash(event.OrderHash)
	if nil != err {
		return err
	}
	if cityPartner, err := w.rds.FindCityPartnerByWalletAddress(common.HexToAddress(order.WalletAddress)); nil == cityPartner || nil != err {
		return errors.New("no this city partner")
	}

	receivedDetail := &dao.CityPartnerReceivedDetail{}
	splitRate := new(big.Rat).SetFrac64(20, 100)
	var token *types.Token
	amount := new(big.Rat)
	receivedDetail.WalletAddress = order.WalletAddress
	receivedDetail.Ringhash = event.Ringhash.Hex()
	receivedDetail.Orderhash = event.OrderHash.Hex()
	if event.LrcFee.Sign() > 0 {
		if token, err = marketutil.AddressToToken(common.HexToAddress("0xef68e7c694f40c8202821edf525de3782458639f")); nil != err {
			token.Protocol = common.HexToAddress("0xef68e7c694f40c8202821edf525de3782458639f")
			token.Symbol = "LRC"
			token.Decimals = big.NewInt(100000000000000000)
		}
		amount.SetInt(event.LrcFee)
	} else if event.SplitB.Sign() > 0 {
		if token, err = marketutil.AddressToToken(common.HexToAddress(order.TokenB)); nil != err {
			token.Protocol = common.HexToAddress(order.TokenB)
			token.Symbol = "Unknown"
			token.Decimals = big.NewInt(1)
		}
		receivedDetail.TokenAddress = order.TokenB
		receivedDetail.TokenSymbol = marketutil.AddressToAlias(order.TokenB)
		amount.SetInt(event.SplitB)
	} else if event.SplitS.Sign() > 0 {
		if token, err = marketutil.AddressToToken(common.HexToAddress(order.TokenS)); nil != err {
			token.Protocol = common.HexToAddress(order.TokenS)
			token.Symbol = "Unknown"
			token.Decimals = big.NewInt(1)
		}
		receivedDetail.TokenAddress = order.TokenS
		receivedDetail.TokenSymbol = marketutil.AddressToAlias(order.TokenS)
		amount.SetInt(event.SplitS)
	}
	amount.Quo(amount, splitRate)
	amountInt := new(big.Int)
	amountInt.SetString(amount.FloatString(0), 10)
	receivedDetail.Amount = "0x" + common.Bytes2Hex(amountInt.Bytes())
	w.rds.Add(receivedDetail)

	if received, err := w.rds.FindReceivedByWalletAndToken(common.HexToAddress(order.WalletAddress), common.HexToAddress(receivedDetail.TokenAddress)); nil != err {
		return err
	} else {
		if nil == received {
			received.TokenAddress = receivedDetail.TokenAddress
			received.WalletAddress = receivedDetail.WalletAddress
			received.Amount = receivedDetail.Amount
			received.TokenSymbol = receivedDetail.TokenSymbol
			w.rds.Add(received)
		} else {
			sumAmount := new(big.Int)
			sumAmount.SetBytes(common.Hex2Bytes(received.Amount))
			sumAmount.Add(sumAmount, amountInt)
			received.Amount = "0x" + common.Bytes2Hex(sumAmount.Bytes())
			humanAmount := new(big.Rat)
			humanAmount.SetFrac(sumAmount, token.Decimals)
			received.HumanAmount = humanAmount.FloatString(8)
			w.rds.Add(received)
		}
	}

	return nil
}
