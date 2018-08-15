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
	"time"

	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"strings"
)

type CityPartnerStatus struct {
	CustomerCount int               `json:"customer_count"`
	Received      map[string]string `json:"received"`
	WalletAddress string            `json:"walletAddress"`
}

func (w *WalletServiceImpl) CreateCityPartner(req *dao.CityPartner) (cityPartner *dao.CityPartner, err error) {
	req.WalletAddress = strings.TrimSpace(req.WalletAddress)
	if !common.IsHexAddress(req.WalletAddress) {
		return nil, errors.New(req.WalletAddress + " isn't an ethereum address.")
	}
	req.WalletAddress = strings.ToLower(common.HexToAddress(req.WalletAddress).Hex())
	if req.CityPartner, err = generateCityPartner(req.WalletAddress, w.rds); nil != err {
		return nil, err
	}
	if cp, err := w.rds.SaveCityPartner(req); nil != err {
		return nil, err
	} else {
		return cp, nil
	}
}

func generateCityPartner(walletAddress string, rds *dao.RdsService) (string, error) {
	res := walletAddress[34:]
	if count, err := rds.GetCityPartnerCount(res); nil == err {
		if count <= 0 {
			return res, nil
		} else {
			return res + getRandomString(4), nil
		}
	} else {
		return "", err
	}

}

func getRandomString(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	bytesLen := len(bytes)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(bytesLen)])
	}
	return string(result)
}

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

func clientIp(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

type JsonRpcError struct {
	Message string `json:"message"`
}
type JsonRpcRes struct {
	JsonRPc string        `json:"jsonrpc"`
	Id      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JsonRpcError `json:"error,omitempty"`
}

func NewJsonRpcRes() JsonRpcRes {
	return JsonRpcRes{JsonRPc: "2.0"}
}

func (w *WalletServiceImpl) CreateCustomerInvitationInfo(writer http.ResponseWriter, req *http.Request) {
	res := NewJsonRpcRes()
	ip := clientIp(req)
	res.Result = ip
	cityPartner := strings.Split(req.URL.Path, "/")[3]
	if _, err := w.rds.FindCityPartnerByCityPartner(cityPartner); nil != err {
		res.Error = &JsonRpcError{}
		res.Error.Message = err.Error()
	} else {
		info := &dao.CustumerInvitationInfo{}
		info.CityPartner = cityPartner
		info.ActivateCode = ip
		info.Activate = 0
		err := w.rds.SaveCustomerInvitationInfo(info)
		if nil != err {
			res.Error = &JsonRpcError{}
			res.Error.Message = err.Error()
		}
	}
	if data, err1 := json.Marshal(res); nil != err1 {
		writer.Write([]byte("{\"error\":{\"message\":\"" + err1.Error() + "\"}}"))
	} else {
		writer.Write(data)
	}
}

func (w *WalletServiceImpl) ActivateCustomerInvitation(writer http.ResponseWriter, req *http.Request) {
	res := NewJsonRpcRes()
	ip := clientIp(req)
	res.Result = ip
	cityPartner := &dao.CityPartner{}
	info, err := w.rds.FindCustomerInvitationInfo(ip)
	if nil != err {
		res.Error = &JsonRpcError{}
		res.Error.Message = err.Error()
	} else {
		err = w.rds.AddCustomerInvitationActivate(info)
		cityPartner, err = w.rds.FindCityPartnerByCityPartner(info.CityPartner)
		if nil != err {
			res.Error = &JsonRpcError{}
			res.Error.Message = err.Error()
		} else {
			res.Result = cityPartner
		}
	}

	if data, err1 := json.Marshal(res); nil != err1 {
		writer.Write([]byte("{\"error\":{\"message\":\"" + err.Error() + "\"}}"))
	} else {
		writer.Write(data)
	}
}

func (w *WalletServiceImpl) GetCityPartnerStatus(req *dao.CityPartner) (*CityPartnerStatus, error) {
	req.CityPartner = strings.TrimSpace(req.CityPartner)
	var err error
	var cityPartner *dao.CityPartner
	cityPartner, err = w.rds.FindCityPartnerByCityPartner(req.CityPartner)
	if nil == cityPartner || nil != err {
		return nil, err
	}
	status := &CityPartnerStatus{}
	status.WalletAddress = cityPartner.WalletAddress
	status.CustomerCount, err = w.rds.GetCityPartnerCustomerCount(req.CityPartner)
	if nil != err {
		return status, err
	}
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
	//activateMtx = sync.Mutex{}
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
