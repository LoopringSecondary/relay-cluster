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

package loopringaccessor

import (
	"errors"
	"github.com/Loopring/relay-lib/eth/abi"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func Erc20Balance(tokenAddress, ownerAddress common.Address, blockParameter string) (*big.Int, error) {
	var balance types.Big
	callMethod := accessor.ContractCallMethod(loopringParams.Erc20Abi, tokenAddress)
	if err := callMethod(&balance, "balanceOf", blockParameter, ownerAddress); nil != err {
		return nil, err
	} else {
		return balance.BigInt(), err
	}
}

func Erc20Allowance(tokenAddress, ownerAddress, spenderAddress common.Address, blockParameter string) (*big.Int, error) {
	var allowance types.Big
	callMethod := accessor.ContractCallMethod(loopringParams.Erc20Abi, tokenAddress)
	if err := callMethod(&allowance, "allowance", blockParameter, ownerAddress, spenderAddress); nil != err {
		return nil, err
	} else {
		return allowance.BigInt(), err
	}
}

func GetCancelledOrFilled(contractAddress common.Address, orderhash common.Hash, blockNumStr string) (*big.Int, error) {
	var amount types.Big
	if _, ok := loopringParams.DelegateAddresses[contractAddress]; !ok {
		return nil, errors.New("accessor: contract address invalid -> " + contractAddress.Hex())
	}
	callMethod := accessor.ContractCallMethod(loopringParams.DelegateAbi, contractAddress)
	if err := callMethod(&amount, "cancelledOrFilled", blockNumStr, orderhash); err != nil {
		return nil, err
	}

	return amount.BigInt(), nil
}

func GetCancelled(contractAddress common.Address, orderhash common.Hash, blockNumStr string) (*big.Int, error) {
	var amount types.Big
	if _, ok := loopringParams.DelegateAddresses[contractAddress]; !ok {
		return nil, errors.New("accessor: contract address invalid -> " + contractAddress.Hex())
	}
	callMethod := accessor.ContractCallMethod(loopringParams.DelegateAbi, contractAddress)
	if err := callMethod(&amount, "cancelled", blockNumStr, orderhash); err != nil {
		return nil, err
	}

	return amount.BigInt(), nil
}

func GetCutoff(result interface{}, contractAddress, owner common.Address, blockNumStr string) error {
	if _, ok := loopringParams.DelegateAddresses[contractAddress]; !ok {
		return errors.New("accessor: contract address invalid -> " + contractAddress.Hex())
	}
	callMethod := accessor.ContractCallMethod(loopringParams.DelegateAbi, contractAddress)
	if err := callMethod(result, "cutoffs", blockNumStr, owner); err != nil {
		return err
	}
	return nil
}

func GetCutoffPair(result interface{}, contractAddress, owner, token1, token2 common.Address, blockNumStr string) error {
	if _, ok := loopringParams.DelegateAddresses[contractAddress]; !ok {
		return errors.New("accessor: contract address invalid -> " + contractAddress.Hex())
	}
	callMethod := accessor.ContractCallMethod(loopringParams.DelegateAbi, contractAddress)
	if err := callMethod(result, "getTradingPairCutoffs", blockNumStr, owner, token1, token2); err != nil {
		return err
	}
	return nil
}

//func BatchErc20BalanceAndAllowance(routeParam string, reqs []*BatchErc20Req) error {
//	reqElems := make([]rpc.BatchElem, 2*len(reqs))
//
//
//	erc20Abi := loopringParams.Erc20Abi
//
//	balanceReqs := BatchBalanceReqs{}
//	allowanceReqs := BatchErc20AllowanceReqs{}
//	for idx, req := range reqs {
//		balanceReq := &BatchBalanceReq{}
//		balanceReq.BlockParameter = "latest"
//		balanceReq.Token = req.Token
//		balanceReq.Owner = req.Owner
//		balanceReqs = append(balanceReqs, balanceReq)
//
//		//balanceOfData, _ := erc20Abi.Pack("balanceOf", req.Owner)
//		//balanceOfArg := &relayethtyp.CallArg{}
//		//balanceOfArg.To = req.Token
//		//balanceOfArg.Data = common.ToHex(balanceOfData)
//
//
//		allowanceData, _ := erc20Abi.Pack("allowance", req.Owner, req.Spender)
//		allowanceArg := &relayethtyp.CallArg{}
//		allowanceArg.To = req.Token
//		allowanceArg.Data = common.ToHex(allowanceData)
//		reqElems[2*idx] = rpc.BatchElem{
//			Method: "eth_call",
//			Args:   []interface{}{balanceOfArg, req.BlockParameter},
//			Result: &req.Balance,
//		}
//		reqElems[2*idx+1] = rpc.BatchElem{
//			Method: "eth_call",
//			Args:   []interface{}{allowanceArg, req.BlockParameter},
//			Result: &req.Allowance,
//		}
//	}
//
//	if err := accessor.BatchCall(routeParam, reqElems); err != nil {
//		return err
//	}
//
//	for idx, req := range reqs {
//		req.BalanceErr = reqElems[2*idx].Error
//		req.AllowanceErr = reqElems[2*idx+1].Error
//	}
//	return nil
//}

func BatchErc20Allowance(routeParam string, reqs BatchErc20AllowanceReqs) error {
	if err := accessor.BatchCall(routeParam, []accessor.BatchReq{reqs}); err != nil {
		return err
	}
	return nil
}

func BatchErc20Balance(routeParam string, reqs BatchBalanceReqs) error {
	if err := accessor.BatchCall(routeParam, []accessor.BatchReq{reqs}); err != nil {
		return err
	}
	return nil
}

func ProtocolImplAbi() *abi.ABI {
	return loopringParams.ProtocolImplAbi
}

func Erc20Abi() *abi.ABI {
	return loopringParams.Erc20Abi
}

func WethAbi() *abi.ABI {
	return loopringParams.WethAbi
}

func TokenRegistryAbi() *abi.ABI {
	return loopringParams.TokenRegistryAbi
}

func DelegateAbi() *abi.ABI {
	return loopringParams.DelegateAbi
}

func IsSpenderAddress(spender common.Address) bool {
	_, exists := loopringParams.DelegateAddresses[spender]
	return exists
}

func ProtocolAddresses() map[common.Address]*ProtocolAddress {
	return loopringParams.ProtocolAddresses
}

func DelegateAddresses() map[common.Address]bool {
	return loopringParams.DelegateAddresses
}

func SupportedDelegateAddress(delegate common.Address) bool {
	return loopringParams.DelegateAddresses[delegate]
}

func IsRelateProtocol(protocol, delegate common.Address) bool {
	protocolAddress, ok := loopringParams.ProtocolAddresses[protocol]
	if ok {
		return protocolAddress.DelegateAddress == delegate
	} else {
		return false
	}
}

func GetSpenderAddress(protocol common.Address) (common.Address, error) {
	impl, ok := loopringParams.ProtocolAddresses[protocol]
	if !ok {
		return common.Address{}, errors.New("accessor method:invalid protocol address")
	}

	return impl.DelegateAddress, nil
}
