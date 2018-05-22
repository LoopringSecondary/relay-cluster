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
	"github.com/Loopring/relay-lib/eth/abi"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/log"
	"github.com/ethereum/go-ethereum/common"
)

var loopringParams *LoopringParams

type LoopringParams struct {
	Erc20Abi         *abi.ABI
	ProtocolImplAbi  *abi.ABI
	DelegateAbi      *abi.ABI
	TokenRegistryAbi *abi.ABI
	//NameRegistryAbi   *abi.ABI
	WethAbi *abi.ABI
	//WethAddress       common.Address
	ProtocolAddresses map[common.Address]*ProtocolAddress
	DelegateAddresses map[common.Address]bool
}

func IsInit() bool {
	return nil != loopringParams
}

func Initialize(options LoopringProtocolOptions) error {
	if !accessor.IsInit() {
		log.Fatalf("must init accessor first")
	}
	var err error
	loopringParams = &LoopringParams{}

	if loopringParams.Erc20Abi, err = abi.New(Erc20AbiStr); nil != err {
		return err
	}

	if loopringParams.WethAbi, err = abi.New(WethAbiStr); nil != err {
		return err
	}

	loopringParams.ProtocolAddresses = make(map[common.Address]*ProtocolAddress)
	loopringParams.DelegateAddresses = make(map[common.Address]bool)

	if "" == options.ImplAbi {
		//todo:
		for _, address := range options.Address {
			data, _ := abiStrFromEthscan(common.HexToAddress(address))
			options.ImplAbi = string(data)
			break
		}
	}

	if protocolImplAbi, err := abi.New(options.ImplAbi); nil != err {
		return err
	} else {
		loopringParams.ProtocolImplAbi = protocolImplAbi
	}

	for version, address := range options.Address {
		impl := &ProtocolAddress{Version: version, ContractAddress: common.HexToAddress(address)}
		callMethod := accessor.ContractCallMethod(loopringParams.ProtocolImplAbi, impl.ContractAddress)
		var addr string
		if err := callMethod(&addr, "lrcTokenAddress", "latest"); nil != err {
			return err
		} else {
			log.Debugf("version:%s, contract:%s, lrcTokenAddress:%s", version, address, addr)
			impl.LrcTokenAddress = common.HexToAddress(addr)
		}
		if err := callMethod(&addr, "tokenRegistryAddress", "latest"); nil != err {
			return err
		} else {
			log.Debugf("version:%s, contract:%s, tokenRegistryAddress:%s", version, address, addr)
			impl.TokenRegistryAddress = common.HexToAddress(addr)
		}
		if err := callMethod(&addr, "delegateAddress", "latest"); nil != err {
			return err
		} else {
			log.Debugf("version:%s, contract:%s, delegateAddress:%s", version, address, addr)
			impl.DelegateAddress = common.HexToAddress(addr)
		}
		//if err := callMethod(&addr, "nameRegistryAddress", "latest"); nil != err {
		//	return err
		//} else {
		//	log.Debugf("version:%s, contract:%s, nameRegistryAddress:%s", version, address, addr)
		//	impl.NameRegistryAddress = common.HexToAddress(addr)
		//}
		loopringParams.ProtocolAddresses[impl.ContractAddress] = impl
		loopringParams.DelegateAddresses[impl.DelegateAddress] = true
	}

	if "" == options.DelegateAbi {
		for _, impl := range loopringParams.ProtocolAddresses {
			//todo:error
			data, _ := abiStrFromEthscan(impl.DelegateAddress)
			options.DelegateAbi = string(data)
			break
		}
	}
	if delegateAbi, err := abi.New(options.DelegateAbi); nil != err {
		return err
	} else {
		loopringParams.DelegateAbi = delegateAbi
	}

	if "" == options.TokenRegistryAbi {
		for _, impl := range loopringParams.ProtocolAddresses {
			data, _ := abiStrFromEthscan(impl.TokenRegistryAddress)
			options.TokenRegistryAbi = string(data)
			break
		}
	}
	if tokenRegistryAbi, err := abi.New(options.TokenRegistryAbi); nil != err {
		return err
	} else {
		loopringParams.TokenRegistryAbi = tokenRegistryAbi
	}
	return nil
}
