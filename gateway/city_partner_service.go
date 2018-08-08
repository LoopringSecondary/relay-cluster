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
	"fmt"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"math/rand"
	"sync"
	"time"
	"qiniupkg.com/x/log.v7"
)

type CityPartnerStatus struct {
	CustomerCount int               `json:"customer_count"`
	Received      map[string]string `json:"received"`
	WalletAddress string `json:"walletAddress"`
}

func (w *WalletServiceImpl) CreateCityPartner(req *dao.CityPartner) (isSuccessed bool, err error) {
	req.WalletAddress = common.HexToAddress(req.WalletAddress).Hex()
	isSuccessed, err = w.rds.SaveCityPartner(req)
	return
}

var activateMtx sync.Mutex

type ExcludeCodes []string

func (codes ExcludeCodes) contains(code string) bool {
	for _, c := range codes {
		if c == code {
			return true
		}
	}
	return false
}

func (w *WalletServiceImpl) CreateCustumerInvitationInfo(req *dao.CustumerInvitationInfo) (activateCode string, err error) {
	activateMtx.Lock()
	defer activateMtx.Unlock()

	info := &dao.CustumerInvitationInfo{}
	info.InvitationCode = req.InvitationCode
	info.Activate = 0
	activateCodes, err := w.rds.GetAllActivateCode(info.InvitationCode)
	if nil != err {
		log.Errorf("err:%s", err.Error())
	}
	activateCode = generateactivateCode(activateCodes, 10, 5)
	info.ActivateCode = activateCode
	err = w.rds.SaveCustumerInvitationInfo(info)
	return
}

func generateactivateCode(excludeCodes ExcludeCodes, count, halfCount int) string {
	activateCode := fmt.Sprintf("%d", rand.Intn(900000)+100000)
	if count <= halfCount {
		activateCode = fmt.Sprintf("%d", rand.Intn(90000000)+10000000)
	}
	count = count - 1
	if excludeCodes.contains(activateCode) {
		return generateactivateCode(excludeCodes, count, halfCount)
	} else {
		return activateCode
	}
}

func (w *WalletServiceImpl) ActivateCustumerInvitation(req *dao.CustumerInvitationInfo) (res *dao.CityPartner, err error) {
	res = &dao.CityPartner{}
	info, err := w.rds.FindCustumerInvitationInfo(req)
	if nil != err {
		return res, err
	} else {
		info.Activate = info.Activate + 1
		err = w.rds.AddCustumerInvitationActivate(info)
		res,err = w.rds.FindCityPartnerByInvitationCode(info.InvitationCode)
		return res, err
	}
}

func (w *WalletServiceImpl) GetCityPartnerStatus(req *dao.CityPartner) (*CityPartnerStatus, error) {
	var err error
	var cityPartner *dao.CityPartner
	invitationCode := req.InvitationCode
	cityPartner, err = w.rds.FindCityPartnerByInvitationCode(invitationCode)
	if nil == cityPartner || nil != err {
		return nil, err
	}
	status := &CityPartnerStatus{}
	status.WalletAddress = cityPartner.WalletAddress
	status.CustomerCount, err = w.rds.GetCityPartnerCustomerCount(invitationCode)
	status.Received = make(map[string]string)
	for _, token := range marketutil.AllTokens {
		status.Received[token.Symbol] = "0"
	}
	if allReceived, err := w.rds.GetAllReceivedByWallet(cityPartner.WalletAddress); nil != err {
		return status, err
	} else {
		for _, received := range allReceived {
			amount := new(big.Int)
			amount.SetBytes(common.Hex2Bytes(received.Amount))
			status.Received[received.TokenSymbol] = amount.String()
		}
		return status, nil
	}
}

func (w *WalletServiceImpl) Start() {
	activateMtx = sync.Mutex{}
	orderFilledEventWatcher := &eventemitter.Watcher{Concurrent: false, Handle: w.HandleFilledEventForCityPartner}
	eventemitter.On(eventemitter.OrderFilled, orderFilledEventWatcher)
}

func (w *WalletServiceImpl) HandleFilledEventForCityPartner(input eventemitter.EventData) error {
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
		//receivedDetail.TokenAddress = order.TokenB
		//receivedDetail.TokenSymbol = marketutil.AddressToAlias(order.TokenB)
		amount.SetInt(event.SplitB)
	} else if event.SplitS.Sign() > 0 {
		if token, err = marketutil.AddressToToken(common.HexToAddress(order.TokenS)); nil != err {
			token.Protocol = common.HexToAddress(order.TokenS)
			token.Symbol = "Unknown"
			token.Decimals = big.NewInt(1)
		}
		//receivedDetail.TokenAddress = order.TokenS
		//receivedDetail.TokenSymbol = marketutil.AddressToAlias(order.TokenS)
		amount.SetInt(event.SplitS)
	}
	receivedDetail.TokenAddress = token.Protocol.Hex()
	receivedDetail.TokenSymbol = token.Symbol
	amount.Quo(amount, splitRate)
	amountInt := new(big.Int)
	amountInt.SetString(amount.FloatString(0), 10)
	receivedDetail.Amount = "0x" + common.Bytes2Hex(amountInt.Bytes())
	receivedDetail.CreateTime = time.Now().Unix()
	w.rds.Add(receivedDetail)

	if received, err := w.rds.FindReceivedByWalletAndToken(common.HexToAddress(order.WalletAddress), common.HexToAddress(receivedDetail.TokenAddress)); nil != err && "record not found" != err.Error() {
		return err
	} else {
		if nil == received || (nil != err && "record not found" == err.Error()) {
			received = &dao.CityPartnerReceived{}
			received.CreateTime = time.Now().Unix()
			received.TokenAddress = receivedDetail.TokenAddress
			received.WalletAddress = receivedDetail.WalletAddress
			received.Amount = receivedDetail.Amount
			sumAmount := new(big.Int)
			sumAmount.SetBytes(common.Hex2Bytes(received.Amount))
			humanAmount := new(big.Rat)
			humanAmount.SetFrac(sumAmount, token.Decimals)
			received.HumanAmount = humanAmount.FloatString(8)
			received.TokenSymbol = receivedDetail.TokenSymbol
			return w.rds.Add(received)
		} else {
			sumAmount := new(big.Int)
			sumAmount.SetBytes(common.Hex2Bytes(received.Amount))
			sumAmount.Add(sumAmount, amountInt)
			received.Amount = "0x" + common.Bytes2Hex(sumAmount.Bytes())
			humanAmount := new(big.Rat)
			humanAmount.SetFrac(sumAmount, token.Decimals)
			received.HumanAmount = humanAmount.FloatString(8)
			return w.rds.UpdateCityPartnerReceived(received)
		}
	}

	return nil
}
