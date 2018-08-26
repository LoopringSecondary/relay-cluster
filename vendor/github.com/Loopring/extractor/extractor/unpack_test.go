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

package extractor_test

import (
	"encoding/json"
	"github.com/Loopring/relay-lib/eth/abi"
	"github.com/Loopring/relay-lib/eth/contract"
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"testing"
)

const (
	erc20AbiStr         = "[{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"who\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"}]"
	wethAbiStr          = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"guy\",\"type\":\"address\"},{\"name\":\"wad\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"src\",\"type\":\"address\"},{\"name\":\"dst\",\"type\":\"address\"},{\"name\":\"wad\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"wad\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"dst\",\"type\":\"address\"},{\"name\":\"wad\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"src\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"guy\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"wad\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"src\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"dst\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"wad\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"dst\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"wad\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"src\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"wad\",\"type\":\"uint256\"}],\"name\":\"Withdrawal\",\"type\":\"event\"}]"
	implAbiStr          = "[{\"constant\":true,\"inputs\":[],\"name\":\"MARGIN_SPLIT_PERCENTAGE_BASE\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ringIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"RATE_RATIO_SCALE\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lrcTokenAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"tokenRegistryAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"delegateAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"orderOwner\",\"type\":\"address\"},{\"name\":\"token1\",\"type\":\"address\"},{\"name\":\"token2\",\"type\":\"address\"}],\"name\":\"getTradingPairCutoffs\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token1\",\"type\":\"address\"},{\"name\":\"token2\",\"type\":\"address\"},{\"name\":\"cutoff\",\"type\":\"uint256\"}],\"name\":\"cancelAllOrdersByTradingPair\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addresses\",\"type\":\"address[5]\"},{\"name\":\"orderValues\",\"type\":\"uint256[6]\"},{\"name\":\"buyNoMoreThanAmountB\",\"type\":\"bool\"},{\"name\":\"marginSplitPercentage\",\"type\":\"uint8\"},{\"name\":\"v\",\"type\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"cancelOrder\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_RING_SIZE\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"cutoff\",\"type\":\"uint256\"}],\"name\":\"cancelAllOrders\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"rateRatioCVSThreshold\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addressList\",\"type\":\"address[4][]\"},{\"name\":\"uintArgsList\",\"type\":\"uint256[6][]\"},{\"name\":\"uint8ArgsList\",\"type\":\"uint8[1][]\"},{\"name\":\"buyNoMoreThanAmountBList\",\"type\":\"bool[]\"},{\"name\":\"vList\",\"type\":\"uint8[]\"},{\"name\":\"rList\",\"type\":\"bytes32[]\"},{\"name\":\"sList\",\"type\":\"bytes32[]\"},{\"name\":\"feeRecipient\",\"type\":\"address\"},{\"name\":\"feeSelections\",\"type\":\"uint16\"}],\"name\":\"submitRing\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"walletSplitPercentage\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_ringIndex\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"_ringHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_miner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_feeRecipient\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_orderInfoList\",\"type\":\"bytes32[]\"}],\"name\":\"RingMined\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_orderHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_amountCancelled\",\"type\":\"uint256\"}],\"name\":\"OrderCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_address\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_cutoff\",\"type\":\"uint256\"}],\"name\":\"AllOrdersCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_address\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_token1\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_token2\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_cutoff\",\"type\":\"uint256\"}],\"name\":\"OrdersCancelled\",\"type\":\"event\"}]"
	tokenRegistryAbiStr = "[{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"unregisterToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"getAddressBySymbol\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addressList\",\"type\":\"address[]\"}],\"name\":\"areAllTokensRegistered\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isTokenRegistered\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"start\",\"type\":\"uint256\"},{\"name\":\"count\",\"type\":\"uint256\"}],\"name\":\"getTokens\",\"outputs\":[{\"name\":\"addressList\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"claimOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"registerToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"addresses\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"isTokenRegisteredBySymbol\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"TokenRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"TokenUnregistered\",\"type\":\"event\"}]"
	delegateAbiStr      = "[{\"constant\":true,\"inputs\":[{\"name\":\"owners\",\"type\":\"address[]\"},{\"name\":\"tradingPairs\",\"type\":\"bytes20[]\"},{\"name\":\"validSince\",\"type\":\"uint256[]\"}],\"name\":\"checkCutoffsBatch\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"resume\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"max\",\"type\":\"uint256\"}],\"name\":\"getLatestAuthorizedAddresses\",\"outputs\":[{\"name\":\"addresses\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"orderHash\",\"type\":\"bytes32\"},{\"name\":\"cancelOrFillAmount\",\"type\":\"uint256\"}],\"name\":\"addCancelledOrFilled\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"cancelled\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"kill\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"lrcTokenAddress\",\"type\":\"address\"},{\"name\":\"miner\",\"type\":\"address\"},{\"name\":\"feeRecipient\",\"type\":\"address\"},{\"name\":\"walletSplitPercentage\",\"type\":\"uint8\"},{\"name\":\"batch\",\"type\":\"bytes32[]\"}],\"name\":\"batchTransferToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"authorizeAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"tokenPair\",\"type\":\"bytes20\"},{\"name\":\"t\",\"type\":\"uint256\"}],\"name\":\"setTradingPairCutoffs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"claimOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"cancelledOrFilled\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"suspended\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"batch\",\"type\":\"bytes32[]\"}],\"name\":\"batchAddCancelledOrFilled\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"bytes20\"}],\"name\":\"tradingPairCutoffs\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"orderHash\",\"type\":\"bytes32\"},{\"name\":\"cancelAmount\",\"type\":\"uint256\"}],\"name\":\"addCancelled\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addressInfos\",\"outputs\":[{\"name\":\"previous\",\"type\":\"address\"},{\"name\":\"index\",\"type\":\"uint32\"},{\"name\":\"authorized\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isAddressAuthorized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"cutoffs\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"suspend\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"deauthorizeAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"t\",\"type\":\"uint256\"}],\"name\":\"setCutoffs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"number\",\"type\":\"uint32\"}],\"name\":\"AddressAuthorized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"number\",\"type\":\"uint32\"}],\"name\":\"AddressDeauthorized\",\"type\":\"event\"}]"

	protocolStr = "0x781870080C8C24a2FD6882296c49c837b06A65E6"
	delegateStr = "0xC533531f4f291F036513f7Abb41bfcCc62475486"
)

var (
	erc20Abi         *abi.ABI
	wethAbi          *abi.ABI
	implAbi          *abi.ABI
	tokenRegistryAbi *abi.ABI
	delegateAbi      *abi.ABI
	protocol         common.Address
	delegate         common.Address
)

func init() {
	erc20Abi, _ = abi.New(erc20AbiStr)
	wethAbi, _ = abi.New(wethAbiStr)
	implAbi, _ = abi.New(implAbiStr)
	tokenRegistryAbi, _ = abi.New(tokenRegistryAbiStr)
	delegateAbi, _ = abi.New(delegateAbiStr)
	protocol = common.HexToAddress(protocolStr)
	delegate = common.HexToAddress(delegateStr)
}

func TestExtractorServiceImpl_UnpackSubmitRingMethod(t *testing.T) {
	input := "0xe78aadb20000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000024000000000000000000000000000000000000000000000000000000000000003e0000000000000000000000000000000000000000000000000000000000000044000000000000000000000000000000000000000000000000000000000000004a0000000000000000000000000000000000000000000000000000000000000054000000000000000000000000000000000000000000000000000000000000005e0000000000000000000000000b94065482ad64d4c2b9252358d746b39e820a58200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000f1bd6422e4420cfd9759f660d739d102328187a5000000000000000000000000ef68e7c694f40c8202821edf525de3782458639f000000000000000000000000b94065482ad64d4c2b9252358d746b39e820a5820000000000000000000000008c4f5e19695fbbfd92d027c1d9ef35a296d539c9000000000000000000000000a3ae668b6239fa3eb1dc26daabb03f244d0259f0000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000b94065482ad64d4c2b9252358d746b39e820a582000000000000000000000000dd859de34ff674050b4961aa51aa74467cb0f03a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000002b5e3af16b188000000000000000000000000000000000000000000000000000000b1a2bc2ec50000000000000000000000000000000000000000000000000000000000005b3f6231000000000000000000000000000000000000000000000000000000005b3f70410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002b5e3af16b188000000000000000000000000000000000000000000000000000000b1a2bc2ec50000000000000000000000000000000000000000000000000002b5e3af16b1880000000000000000000000000000000000000000000000000000000000005b3f623a000000000000000000000000000000000000000000000000000000005b3f704a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b1a2bc2ec500000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000003200000000000000000000000000000000000000000000000000000000000000320000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000001c000000000000000000000000000000000000000000000000000000000000001c000000000000000000000000000000000000000000000000000000000000001c000000000000000000000000000000000000000000000000000000000000001b0000000000000000000000000000000000000000000000000000000000000004132027c19b311dae47345ed7def3533ef46872a7b462395f66dae3f8276ee9528d67b3ffc5aaa419e45fb9959073f88151cb2d2fdb07b5f7d55934911f9bf85c46dc5adf72e5643c544cdfa48502beec19c73c5b0ec317f6d0c6591a26e94213e479d38f02fced9c5ff65934b4ec09d70a42d6aa8f8ae1b92d7e95954d00314700000000000000000000000000000000000000000000000000000000000000044f40d6573cb002321e45dcc2f7f467a605306449dbbf5121d54bd95293c0aafd6c687963b8beca7845c65ad56f3e65666e48840ced6af3990424dcdeefb66ae95199431284c7e697954a9c1ec31328c81f85e487d40a37055de894b0e91890841133b74e888deb375deb1ba93eb39a2ca4ff6aeba080db95e5950e5463fd458c"

	var ring contract.SubmitRingMethodInputs
	//ring.Protocol = protocol
	ring.Protocol = common.HexToAddress("0x8d8812b72d1e4ffCeC158D25f56748b7d67c1e78")
	delegateAddr := common.HexToAddress("0x5567ee920f7E62274284985D793344351A00142B")

	data := hexutil.MustDecode("0x" + input[10:])

	for i := 0; i < len(data)/32; i++ {
		t.Logf("index:%d -> %s", i*32, common.ToHex(data[i*32:(i+1)*32]))
	}

	if err := implAbi.UnpackMethod(&ring, "submitRing", data); err != nil {
		t.Fatalf(err.Error())
	}

	event, err := ring.ConvertDown(ring.Protocol, delegateAddr)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, v := range event.OrderList {
		//t.Log(k, "orderhash", v.Hash.Hex())
		//t.Log(k, "protocol", v.Protocol.Hex())
		//t.Log(k, "tokenS", v.TokenS.Hex())
		//t.Log(k, "tokenB", v.TokenB.Hex())
		//
		//t.Log(k, "amountS", v.AmountS.String())
		//t.Log(k, "amountB", v.AmountB.String())
		//t.Log(k, "validSince", v.ValidSince.String())
		//t.Log(k, "validUntil", v.ValidUntil.String())
		//t.Log(k, "lrcFee", v.LrcFee.String())
		//t.Log(k, "rateAmountS", ring.UintArgsList[k][5].String())
		//
		//t.Log(k, "marginSplitpercentage", v.MarginSplitPercentage)
		//t.Log(k, "feeSelectionList", ring.Uint8ArgsList[k][0])
		//
		//t.Log(k, "buyNoMoreThanAmountB", v.BuyNoMoreThanAmountB)
		//
		//t.Log(k, "v", v.V)
		//t.Log(k, "s", v.S.Hex())
		//t.Log(k, "r", v.R.Hex())

		// var order types.Order
		bs, _ := json.Marshal(v)
		t.Log(string(bs))
	}

	t.Log("ring.v1", ring.VList[2])
	t.Log("ring.r1", types.Bytes32(ring.RList[2]).Hex())
	t.Log("ring.s1", types.Bytes32(ring.SList[2]).Hex())
	t.Log("ring.v2", ring.VList[3])
	t.Log("ring.r2", types.Bytes32(ring.RList[3]).Hex())
	t.Log("ring.s2", types.Bytes32(ring.SList[3]).Hex())
	t.Log("ring.feeReceipt", event.FeeReceipt.Hex())
	t.Log("ring.feeSelection", event.FeeSelection)
}

func TestExtractorServiceImpl_UnpackWethWithdrawalMethod(t *testing.T) {
	input := "0x2e1a7d4d0000000000000000000000000000000000000000000000000000000000000064"
	txfrom := common.HexToAddress("")

	var withdrawal contract.WethWithdrawalMethod

	data := hexutil.MustDecode("0x" + input[10:])

	if err := wethAbi.UnpackMethod(&withdrawal, contract.METHOD_WETH_WITHDRAWAL, data); err != nil {
		t.Fatalf(err.Error())
	}

	evt := withdrawal.ConvertDown(txfrom)
	t.Logf("withdrawal event value:%s", evt.Amount)
}

func TestExtractorServiceImpl_UnpackCancelOrderMethod(t *testing.T) {
	input := "0x8c59f7ca000000000000000000000000b1018949b241d76a1ab2094f473e9befeabb5ead000000000000000000000000480037780d0b0e766941b8c5e99e685bf8812c39000000000000000000000000f079e0612e869197c5f4c7d0a95df570b163232b000000000000000000000000b1018949b241d76a1ab2094f473e9befeabb5ead00000000000000000000000047fe1648b80fa04584241781488ce4c0aaca23e400000000000000000000000000000000000000000000003635c9adc5dea00000000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000000000000000000000000000000000005ad8a62f000000000000000000000000000000000000000000000000000000005b5c7c2f00000000000000000000000000000000000000000000000029a2241af62c00000000000000000000000000000000000000000000000000001bc16d674ec8000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001b39026cca9b4e4e42ac957182e6bbeebd88d327c9368f905620b8edbf2be687af12e190eb0ec2fc5b337487834aeb9ce9df2f0275f281b3e7ca5bdec13246444f"

	var method contract.CancelOrderMethod

	data := hexutil.MustDecode("0x" + input[10:])

	//for i := 0; i < len(data)/32; i++ {
	//	t.Logf("index:%d -> %s", i, common.ToHex(data[i*32:(i+1)*32]))
	//}

	if err := implAbi.UnpackMethod(&method, "cancelOrder", data); err != nil {
		t.Fatalf(err.Error())
	}

	order, cancelAmount, err := method.ConvertDown(protocol, delegate)
	if err != nil {
		t.Fatalf(err.Error())
	}

	t.Log("delegate", order.DelegateAddress.Hex())
	t.Log("orderHash", order.Hash.Hex())
	t.Log("owner", order.Owner.Hex())
	t.Log("wallet", order.WalletAddress.Hex())
	t.Log("auth", order.AuthAddr.Hex())
	t.Log("tokenS", order.TokenS.Hex())
	t.Log("tokenB", order.TokenB.Hex())
	t.Log("amountS", order.AmountS.String())
	t.Log("amountB", order.AmountB.String())
	t.Log("validSince", order.ValidSince.String())
	t.Log("validUntil", order.ValidUntil.String())
	t.Log("lrcFee", order.LrcFee.String())
	t.Log("cancelAmount", method.OrderValues[5].String())
	t.Log("buyNoMoreThanAmountB", order.BuyNoMoreThanAmountB)
	t.Log("marginSplitpercentage", order.MarginSplitPercentage)
	t.Log("v", order.V)
	t.Log("s", order.S.Hex())
	t.Log("r", order.R.Hex())
	t.Log("cancelAmount", cancelAmount)
}

func TestExtractorServiceImpl_UnpackApproveMethod(t *testing.T) {
	input := "0x095ea7b300000000000000000000000045aa504eb94077eec4bf95a10095a8e3196fc5910000000000000000000000000000000000000000000000008ac7230489e80000"
	txfrom := common.HexToAddress("")

	var method contract.ApproveMethod

	data := hexutil.MustDecode("0x" + input[10:])
	for i := 0; i < len(data)/32; i++ {
		t.Logf("index:%d -> %s", i, common.ToHex(data[i*32:(i+1)*32]))
	}

	if err := erc20Abi.UnpackMethod(&method, "approve", data); err != nil {
		t.Fatalf(err.Error())
	}

	approve := method.ConvertDown(txfrom)
	t.Logf("approve spender:%s, value:%s", approve.Spender.Hex(), approve.Amount.String())
}

func TestExtractorServiceImpl_UnpackTransferMethod(t *testing.T) {
	input := "0xa9059cbb0000000000000000000000008311804426a24495bd4306daf5f595a443a52e32000000000000000000000000000000000000000000000000000000174876e800"
	txfrom := common.HexToAddress("")

	data := hexutil.MustDecode("0x" + input[10:])

	var method contract.TransferMethod
	if err := erc20Abi.UnpackMethod(&method, "transfer", data); err != nil {
		t.Fatalf(err.Error())
	}
	transfer := method.ConvertDown(txfrom)

	t.Logf("transfer receiver:%s, value:%s", transfer.Receiver.Hex(), transfer.Amount.String())
}

func TestExtractorServiceImpl_UnpackTransferEvent(t *testing.T) {
	inputs := []string{
		"0x00000000000000000000000000000000000000000000001d2666491321fc5651",
		"0x0000000000000000000000000000000000000000000000008ac7230489e80000",
		"0x0000000000000000000000000000000000000000000000004c0303a413a39039",
		"0x000000000000000000000000000000000000000000000000016345785d8a0000",
	}
	transfer := &contract.TransferEvent{}

	for _, input := range inputs {
		data := hexutil.MustDecode(input)

		if err := erc20Abi.UnpackEvent(transfer, "Transfer", []byte{}, [][]byte{data}); err != nil {
			t.Fatalf(err.Error())
		}

		t.Logf("transfer value:%s", transfer.Value.String())
	}
}

func TestExtractorServiceImpl_UnpackRingMinedEvent(t *testing.T) {
	//input := "0x000000000000000000000000000000000000000000000000000000000000048f0000000000000000000000005552dcfba48c94544beaaf26470df9898e050ac20000000000000000000000005552dcfba48c94544beaaf26470df9898e050ac20000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000e3d810d2f26d07d7648026676d0c8c5cd6624f1e0c200c73ab794208929dcedab00000000000000000000000080679a2c82ab82f1e73e14c4bec4ba1992f9f25a000000000000000000000000ef68e7c694f40c8202821edf525de3782458639f00000000000000000000000000000000000000000000000051ce0cb2c0ca25290000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003100000000000000000000000000000000000000000000000000000000000000009efbfe62e7ed85de35906c81edea7138e44ee1c86dca2f1da03feca5e9a695ae000000000000000000000000b94065482ad64d4c2b9252358d746b39e820a582000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc20000000000000000000000000000000000000000000000000010f68b1ca0e40a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffc668cc0d6bfcf5"
	//topics := []string{"0x4d2a4adf7c5f6cf35d97aecc1919897bf86299dccd9b5e19b2b38ebebf07add0", "0x9c140a64afcf21b373ce11bbef7240c3b9c056b39c0f50235d28a50b257a99fc"}

	input := "0x000000000000000000000000000000000000000000000000000000000000048e0000000000000000000000005552dcfba48c94544beaaf26470df9898e050ac20000000000000000000000005552dcfba48c94544beaaf26470df9898e050ac20000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000e3d810d2f26d07d7648026676d0c8c5cd6624f1e0c200c73ab794208929dcedab00000000000000000000000080679a2c82ab82f1e73e14c4bec4ba1992f9f25a000000000000000000000000ef68e7c694f40c8202821edf525de3782458639f0000000000000000000000000000000000000000000000056bc75e2d63100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003bf3b91c95b0000000000000000000000000000000000000000000000000000000000000000000075d2760bde911c1a6e9d5a0ee5f5e3de64a247517b1b4c06fa73b4ede974cd86000000000000000000000000b94065482ad64d4c2b9252358d746b39e820a582000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000011fc51222ce800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000fffffffffffffffffffffffffffffffffffffffffffffffff88188dc6d4a0000"
	topics := []string{"0x4d2a4adf7c5f6cf35d97aecc1919897bf86299dccd9b5e19b2b38ebebf07add0", "0xf41455c7ba73138952db2311eb5cc682718ac0f5fd11e9e347542b9fc895fc86"}

	ringmined := &contract.RingMinedEvent{}

	data := hexutil.MustDecode(input)
	var decodedValues [][]byte

	for _, v := range topics {
		decodedValues = append(decodedValues, hexutil.MustDecode(v))
	}

	for i := 0; i < len(data)/32; i++ {
		t.Logf("index:%d -> %s", i, common.ToHex(data[i*32:(i+1)*32]))
	}

	if err := implAbi.UnpackEvent(ringmined, "RingMined", data, decodedValues); err != nil {
		t.Fatalf(err.Error())
	}

	evt, fills, err := ringmined.ConvertDown()
	if err != nil {
		t.Fatalf(err.Error())
	}

	for k, fill := range fills {
		t.Logf("k:%d --> ringindex:%s", k, fill.RingIndex.String())
		t.Logf("k:%d --> fillIndex:%s", k, fill.FillIndex.String())
		t.Logf("k:%d --> orderhash:%s", k, fill.OrderHash.Hex())
		t.Logf("k:%d --> preorder:%s", k, fill.PreOrderHash.Hex())
		t.Logf("k:%d --> nextorder:%s", k, fill.NextOrderHash.Hex())
		t.Logf("k:%d --> owner:%s", k, fill.Owner.Hex())
		t.Logf("k:%d --> tokenS:%s", k, fill.TokenS.Hex())
		t.Logf("k:%d --> tokenB:%s", k, fill.TokenB.Hex())
		t.Logf("k:%d --> amountS:%s", k, fill.AmountS.String())
		t.Logf("k:%d --> amountB:%s", k, fill.AmountB.String())
		t.Logf("k:%d --> lrcReward:%s", k, fill.LrcReward.String())
		t.Logf("k:%d --> lrcFee:%s", k, fill.LrcFee.String())
		t.Logf("k:%d --> splitS:%s", k, fill.SplitS.String())
		t.Logf("k:%d --> splitB:%s", k, fill.SplitB.String())
	}

	t.Logf("totalLrcFee:%s", evt.TotalLrcFee.String())
	t.Logf("tradeAmount:%d", evt.TradeAmount)
}

func TestExtractorServiceImpl_UnpackOrderCancelledEvent(t *testing.T) {
	input := "0x0000000000000000000000000000000000000000000000001bc16d674ec80000"

	src := &contract.OrderCancelledEvent{}
	data := hexutil.MustDecode(input)

	var decodedValues [][]byte
	topics := []string{
		"0x3e1003227205ab9eb9b1652e25b2f6fc548ff55e94bf76a42aca90501c6c4e35",
		"0xc0d710b036a622871974e8cc28dd5abe4065dfeebfc3a2724f1294c554d70e9c",
	}
	for _, v := range topics {
		decodedValues = append(decodedValues, hexutil.MustDecode(v))
	}

	if err := implAbi.UnpackEvent(src, contract.EVENT_ORDER_CANCELLED, data, decodedValues); err != nil {
		t.Fatalf(err.Error())
	}

	event := src.ConvertDown()

	t.Logf("orderhash:%s", event.OrderHash.Hex())
	t.Logf("amount:%s", event.AmountCancelled.String())
}

func TestExtractorServiceImpl_UnpackDepositEvent(t *testing.T) {
	input := "0x0000000000000000000000000000000000000000000000001bc16d674ec80000"

	deposit := &contract.WethDepositEvent{}
	data := hexutil.MustDecode(input)

	decodedValues := [][]byte{}
	topics := []string{
		"0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c",
		"0x0000000000000000000000001b978a1d302335a6f2ebe4b8823b5e17c3c84135",
	}
	for _, v := range topics {
		decodedValues = append(decodedValues, hexutil.MustDecode(v))
	}

	if err := wethAbi.UnpackEvent(deposit, "Deposit", data, decodedValues); err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Logf("deposit value:%s", deposit.Value.String())
	}
}

func TestExtractorServiceImpl_UnpackTokenRegistryEvent(t *testing.T) {
	input := "0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000034c52430000000000000000000000000000000000000000000000000000000000"

	tokenRegistry := &contract.TokenRegisteredEvent{}
	data := hexutil.MustDecode(input)

	decodedValues := [][]byte{}
	topics := []string{
		"0xaaed15520cc86e95b7c2522d968096283afbef7858bdf194b2f60d28a1a8d63e",
		"0x000000000000000000000000cd36128815ebe0b44d0374649bad2721b8751bef",
	}
	for _, v := range topics {
		decodedValues = append(decodedValues, hexutil.MustDecode(v))
	}

	if err := tokenRegistryAbi.UnpackEvent(tokenRegistry, contract.EVENT_TOKEN_REGISTERED, data, decodedValues); err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Logf("TokenRegistered symbol:%s, address:%s", tokenRegistry.Symbol, tokenRegistry.Token.Hex())
	}
}

func TestExtractorServiceImpl_UnpackTokenUnRegistryEvent(t *testing.T) {
	input := "0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000034c5243"

	tokenUnRegistry := &contract.TokenUnRegisteredEvent{}
	data := hexutil.MustDecode(input)

	decodedValues := [][]byte{}
	topics := []string{
		"0xee98311a96660ce4ab10cd82053f767653901305ec8acf91ec60311de919e28a",
		"0x000000000000000000000000cd36128815ebe0b44d0374649bad2721b8751bef",
	}
	for _, v := range topics {
		decodedValues = append(decodedValues, hexutil.MustDecode(v))
	}

	if err := tokenRegistryAbi.UnpackEvent(tokenUnRegistry, "TokenUnregistered", data, decodedValues); err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Logf("TokenUnregistered symbol:%s, address:%s", tokenUnRegistry.Symbol, tokenUnRegistry.Token.Hex())
	}
}

func TestExtractorServiceImpl_Compare(t *testing.T) {
	str1 := "547722557505166136913"
	str2 := "1000000000000000000000"
	num1, _ := big.NewInt(0).SetString(str1, 0)
	num2, _ := big.NewInt(0).SetString(str2, 0)
	if num1.Cmp(num2) > 0 {
		t.Logf("%s > %s", str1, str2)
	} else {
		t.Logf("%s <= %s", str1, str2)
	}
}

func TestExtractorServiceImpl_UnpackNumbers(t *testing.T) {
	str1 := "0xffffffffffffffffffffffffffffffffffffffffffffffffffa1d2c1fb1c2d9f"
	str2 := "0xffffffffffffffffffffffffffffffffffffffffffffffffff90c5f64e557fa4"
	str3 := "0x0000000000000000000000000000000000000000000000026508392204063330"
	str4 := "0x0000000000000000000000000000000000000000000000031307535724740700"
	list := []string{str1, str2, str3, str4}

	for _, v := range list {
		n1 := safeBig(v)
		t.Logf("init data:%s -> number:%s", v, n1.String())
	}
}

func safeBig(input string) *big.Int {
	bytes := hexutil.MustDecode(input)
	num := new(big.Int).SetBytes(bytes[:])
	if bytes[0] > uint8(128) {
		num.Xor(types.MaxUint256, num)
		num.Not(num)
	}
	return num
}

func TestTxReceipt(t *testing.T) {
	receipt := &ethtyp.TransactionReceipt{}
	if receipt.Status == nil {
		t.Log("it is nil")
	}
	bs, err := json.Marshal(receipt)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(string(bs))
}
