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
	"github.com/Loopring/relay-lib/crypto"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

//go:generate gencodec -type P2POrderJsonRequest -field-override p2pOrderJsonRequestMarshaling -out gen_p2pOrder_request_json.go
type P2POrderJsonRequest struct {
	Protocol        common.Address             `json:"protocol" gencodec:"required"`        // 智能合约地址
	DelegateAddress common.Address             `json:"delegateAddress" gencodec:"required"` // 智能合约地址
	TokenS          common.Address             `json:"tokenS" gencodec:"required"`          // 卖出erc20代币智能合约地址
	TokenB          common.Address             `json:"tokenB" gencodec:"required"`          // 买入erc20代币智能合约地址
	AuthAddr        common.Address             `json:"authAddr" gencodec:"required"`        //
	AuthPrivateKey  crypto.EthPrivateKeyCrypto `json:"authPrivateKey"`                      //
	WalletAddress   common.Address             `json:"walletAddress" gencodec:"required"`
	AmountS         *big.Int                   `json:"amountS" gencodec:"required"`    // 卖出erc20代币数量上限
	AmountB         *big.Int                   `json:"amountB" gencodec:"required"`    // 买入erc20代币数量上限
	ValidSince      *big.Int                   `json:"validSince" gencodec:"required"` //
	ValidUntil      *big.Int                   `json:"validUntil" gencodec:"required"` // 订单过期时间
	// Salt                  int64          `json:"salt" gencodec:"required"`
	LrcFee                *big.Int       `json:"lrcFee" ` // 交易总费用,部分成交的费用按该次撮合实际卖出代币额与比例计算
	BuyNoMoreThanAmountB  bool           `json:"buyNoMoreThanAmountB" gencodec:"required"`
	MarginSplitPercentage uint8          `json:"marginSplitPercentage" gencodec:"required"` // 不为0时支付给交易所的分润比例，否则视为100%
	V                     uint8          `json:"v" gencodec:"required"`
	R                     Bytes32        `json:"r" gencodec:"required"`
	S                     Bytes32        `json:"s" gencodec:"required"`
	Price                 *big.Rat       `json:"price"`
	Owner                 common.Address `json:"owner"`
	Hash                  common.Hash    `json:"hash"`
	CreateTime            int64          `json:"createTime"`
	PowNonce              uint64         `json:"powNonce"`
	Side                  string         `json:"side"`
	OrderType             string         `json:"orderType"`
	MakerOrderHash        common.Hash    `json:"makerOrderHash"`
	P2PSide               string         `json:"p2pSide"`
	SourceId              string         `json:"sourceId"`
}

type p2pOrderJsonRequestMarshaling struct {
	AmountS    *Big
	AmountB    *Big
	ValidSince *Big
	ValidUntil *Big
	LrcFee     *Big
}

func ToP2POrder(request *P2POrderJsonRequest) *Order {
	order := &Order{}
	order.Protocol = request.Protocol
	order.DelegateAddress = request.DelegateAddress
	order.TokenS = request.TokenS
	order.TokenB = request.TokenB
	order.AmountS = request.AmountS
	order.AmountB = request.AmountB
	order.ValidSince = request.ValidSince
	order.ValidUntil = request.ValidUntil
	order.AuthAddr = request.AuthAddr
	order.AuthPrivateKey = request.AuthPrivateKey
	order.LrcFee = request.LrcFee
	order.BuyNoMoreThanAmountB = request.BuyNoMoreThanAmountB
	order.MarginSplitPercentage = request.MarginSplitPercentage
	order.V = request.V
	order.R = request.R
	order.S = request.S
	order.Owner = request.Owner
	order.WalletAddress = request.WalletAddress
	order.PowNonce = request.PowNonce
	order.OrderType = request.OrderType
	order.P2PSide = request.P2PSide
	if "" == request.SourceId {
		order.SourceId = "unknown"
	} else {
		order.SourceId = request.SourceId
	}
	return order
}
