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
	"github.com/Loopring/relay-lib/eth/contract"
	lpraccessor "github.com/Loopring/relay-lib/eth/loopringaccessor"
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
)

type AbiProcessor struct {
	events      map[common.Hash]EventData
	methods     map[string]MethodData
	erc20Events map[common.Hash]bool
	protocols   map[common.Address]string
	delegates   map[common.Address]string
}

// 这里无需考虑版本问题，对解析来说，不接受版本升级带来数据结构变化的可能性
func newAbiProcessor() *AbiProcessor {
	processor := &AbiProcessor{}

	processor.events = make(map[common.Hash]EventData)
	processor.erc20Events = make(map[common.Hash]bool)
	processor.methods = make(map[string]MethodData)
	processor.protocols = make(map[common.Address]string)
	processor.delegates = make(map[common.Address]string)

	processor.loadProtocolAddress()
	processor.loadErc20Contract()
	processor.loadWethContract()
	processor.loadProtocolContract()
	processor.loadTokenRegisterContract()
	processor.loadTokenTransferDelegateProtocol()

	return processor
}

// GetEvent get EventData with id hash
func (processor *AbiProcessor) GetEvent(evtLog ethtyp.Log) (EventData, bool) {
	var (
		event EventData
		ok    bool
	)

	id := evtLog.EventId()
	if id == types.NilHash {
		return event, false
	}

	event, ok = processor.events[id]
	return event, ok
}

// GetMethod get MethodData with method id
func (processor *AbiProcessor) GetMethod(tx *ethtyp.Transaction) (MethodData, bool) {
	var (
		method MethodData
		ok     bool
	)

	id := tx.MethodId()
	if id == "" {
		return method, false
	}

	method, ok = processor.methods[id]
	return method, ok
}

// GetMethodName
func (processor *AbiProcessor) GetMethodName(tx *ethtyp.Transaction) string {
	if method, ok := processor.GetMethod(tx); ok {
		return method.Name
	}
	return contract.METHOD_UNKNOWN
}

// HaveSupportedEvents supported contract events and unsupported erc20 events
func (processor *AbiProcessor) HaveSupportedEvents(receipt *ethtyp.TransactionReceipt) bool {
	if receipt == nil || len(receipt.Logs) == 0 {
		return false
	}

	for _, evtlog := range receipt.Logs {
		if processor.IsSupportedEvent(&evtlog) {
			return true
		}
	}

	return false
}

// HaveSupportedEvents supported contract events and unsupported erc20 events
func (processor *AbiProcessor) IsSupportedEvent(evtlog *ethtyp.Log) bool {
	id := evtlog.EventId()
	if id == types.NilHash {
		return false
	}

	// unsupported contract
	if !processor.IsContractSupported(common.HexToAddress(evtlog.Address)) {
		return false
	}

	// supported impl event
	if _, implEventIdSupported := processor.events[id]; implEventIdSupported {
		return true
	}
	// supported erc20 event
	if _, erc20EventIdSupported := processor.erc20Events[id]; erc20EventIdSupported {
		return true
	}

	return false
}

// IsSupportedMethod only supported contracts method
func (processor *AbiProcessor) IsSupportedMethod(tx *ethtyp.Transaction) bool {
	protocol := common.HexToAddress(tx.To)
	if !processor.IsContractSupported(protocol) {
		return false
	}

	id := tx.MethodId()
	if id == "" {
		return false
	}

	_, ok := processor.methods[id]
	return ok
}

func (processor *AbiProcessor) loadProtocolAddress() {
	for _, v := range util.AllTokens {
		processor.protocols[v.Protocol] = v.Symbol
		log.Infof("extractor,contract protocol %s->%s", v.Symbol, v.Protocol.Hex())
	}
	processor.AddCustomTokens()

	for _, v := range lpraccessor.ProtocolAddresses() {
		protocolSymbol := "loopring"
		delegateSymbol := "transfer_delegate"
		tokenRegisterSymbol := "token_register"

		processor.protocols[v.ContractAddress] = protocolSymbol
		processor.protocols[v.TokenRegistryAddress] = tokenRegisterSymbol
		processor.protocols[v.DelegateAddress] = delegateSymbol

		log.Infof("extractor,contract protocol %s->%s", protocolSymbol, v.ContractAddress.Hex())
		log.Infof("extractor,contract protocol %s->%s", tokenRegisterSymbol, v.TokenRegistryAddress.Hex())
		log.Infof("extractor,contract protocol %s->%s", delegateSymbol, v.DelegateAddress.Hex())
	}
}

func (processor *AbiProcessor) loadProtocolContract() {
	for name, event := range lpraccessor.ProtocolImplAbi().Events {
		if name != contract.EVENT_RING_MINED && name != contract.EVENT_ORDER_CANCELLED && name != contract.EVENT_CUTOFF_ALL && name != contract.EVENT_CUTOFF_PAIR {
			continue
		}

		eventData := newEventData(&event, lpraccessor.ProtocolImplAbi())

		switch eventData.Name {
		case contract.EVENT_RING_MINED:
			eventData.Event = &contract.RingMinedEvent{}
		case contract.EVENT_ORDER_CANCELLED:
			eventData.Event = &contract.OrderCancelledEvent{}
		case contract.EVENT_CUTOFF_ALL:
			eventData.Event = &contract.CutoffEvent{}
		case contract.EVENT_CUTOFF_PAIR:
			eventData.Event = &contract.CutoffPairEvent{}
		}

		processor.events[eventData.Id] = eventData
		log.Infof("extractor,contract event name:%s -> key:%s", eventData.Name, eventData.Id.Hex())
	}

	for name, method := range lpraccessor.ProtocolImplAbi().Methods {
		if name != contract.METHOD_SUBMIT_RING && name != contract.METHOD_CANCEL_ORDER && name != contract.METHOD_CUTOFF_ALL && name != contract.METHOD_CUTOFF_PAIR {
			continue
		}

		methodData := newMethodData(&method, lpraccessor.ProtocolImplAbi())

		switch methodData.Name {
		case contract.METHOD_SUBMIT_RING:
			methodData.Method = &contract.SubmitRingMethodInputs{}
		case contract.METHOD_CANCEL_ORDER:
			methodData.Method = &contract.CancelOrderMethod{}
		case contract.METHOD_CUTOFF_ALL:
			methodData.Method = &contract.CutoffMethod{}
		case contract.METHOD_CUTOFF_PAIR:
			methodData.Method = &contract.CutoffPairMethod{}
		}

		processor.methods[methodData.Id] = methodData
		log.Infof("extractor,contract method name:%s -> key:%s", methodData.Name, methodData.Id)
	}
}

func (processor *AbiProcessor) loadErc20Contract() {
	for name, event := range lpraccessor.Erc20Abi().Events {
		if name != contract.EVENT_TRANSFER && name != contract.EVENT_APPROVAL {
			continue
		}

		watcher := &eventemitter.Watcher{}
		eventData := newEventData(&event, lpraccessor.Erc20Abi())

		switch eventData.Name {
		case contract.EVENT_TRANSFER:
			eventData.Event = &contract.TransferEvent{}
		case contract.EVENT_APPROVAL:
			eventData.Event = &contract.ApprovalEvent{}
		}

		eventemitter.On(eventData.Id.Hex(), watcher)
		processor.events[eventData.Id] = eventData
		processor.erc20Events[eventData.Id] = true
		log.Infof("extractor,contract event name:%s -> key:%s", eventData.Name, eventData.Id.Hex())
	}

	for name, method := range lpraccessor.Erc20Abi().Methods {
		if name != contract.METHOD_TRANSFER && name != contract.METHOD_APPROVE {
			continue
		}

		methodData := newMethodData(&method, lpraccessor.Erc20Abi())

		switch methodData.Name {
		case contract.METHOD_TRANSFER:
			methodData.Method = &contract.TransferMethod{}
		case contract.METHOD_APPROVE:
			methodData.Method = &contract.ApproveMethod{}
		}

		processor.methods[methodData.Id] = methodData
		log.Infof("extractor,contract method name:%s -> key:%s", methodData.Name, methodData.Id)
	}
}

func (processor *AbiProcessor) loadWethContract() {
	for name, method := range lpraccessor.WethAbi().Methods {
		if name != contract.METHOD_WETH_DEPOSIT && name != contract.METHOD_WETH_WITHDRAWAL {
			continue
		}

		methodData := newMethodData(&method, lpraccessor.WethAbi())

		switch methodData.Name {
		case contract.METHOD_WETH_DEPOSIT:
			methodData.Method = &contract.WethDepositMethod{}
		case contract.METHOD_WETH_WITHDRAWAL:
			methodData.Method = &contract.WethWithdrawalMethod{}
		}

		processor.methods[methodData.Id] = methodData
		log.Infof("extractor,contract method name:%s -> key:%s", methodData.Name, methodData.Id)
	}

	for name, event := range lpraccessor.WethAbi().Events {
		if name != contract.EVENT_WETH_DEPOSIT && name != contract.EVENT_WETH_WITHDRAWAL {
			continue
		}

		eventData := newEventData(&event, lpraccessor.WethAbi())

		switch eventData.Name {
		case contract.EVENT_WETH_DEPOSIT:
			eventData.Event = &contract.WethDepositEvent{}
		case contract.EVENT_WETH_WITHDRAWAL:
			eventData.Event = &contract.WethWithdrawalEvent{}
		}

		processor.events[eventData.Id] = eventData
		log.Infof("extractor,contract event name:%s -> key:%s", eventData.Name, eventData.Id.Hex())
	}
}

func (processor *AbiProcessor) loadTokenRegisterContract() {
	for name, event := range lpraccessor.TokenRegistryAbi().Events {
		if name != contract.EVENT_TOKEN_REGISTERED && name != contract.EVENT_TOKEN_UNREGISTERED {
			continue
		}

		eventData := newEventData(&event, lpraccessor.TokenRegistryAbi())

		switch eventData.Name {
		case contract.EVENT_TOKEN_REGISTERED:
			eventData.Event = &contract.TokenRegisteredEvent{}
		case contract.EVENT_TOKEN_UNREGISTERED:
			eventData.Event = &contract.TokenUnRegisteredEvent{}
		}

		processor.events[eventData.Id] = eventData
		log.Infof("extractor,contract event name:%s -> key:%s", eventData.Name, eventData.Id.Hex())
	}
}

func (processor *AbiProcessor) loadTokenTransferDelegateProtocol() {
	for name, event := range lpraccessor.DelegateAbi().Events {
		if name != contract.EVENT_ADDRESS_AUTHORIZED && name != contract.EVENT_ADDRESS_DEAUTHORIZED {
			continue
		}

		eventData := newEventData(&event, lpraccessor.DelegateAbi())

		switch eventData.Name {
		case contract.EVENT_ADDRESS_AUTHORIZED:
			eventData.Event = &contract.AddressAuthorizedEvent{}
		case contract.EVENT_ADDRESS_DEAUTHORIZED:
			eventData.Event = &contract.AddressDeAuthorizedEvent{}
		}

		processor.events[eventData.Id] = eventData
		log.Infof("extractor,contract event name:%s -> key:%s", eventData.Name, eventData.Id.Hex())
	}
}

func (processor *AbiProcessor) IsContractSupported(protocol common.Address) bool {
	_, ok := processor.protocols[protocol]
	return ok
}

func (processor *AbiProcessor) AddCustomTokens() {
	tokens, _ := util.GetAllCustomTokenList()
	for _, v := range tokens {
		if !processor.IsContractSupported(v.Address) {
			processor.protocols[v.Address] = v.Symbol
			log.Infof("extractor,contract protocol %s->%s", v.Symbol, v.Address.Hex())
		}
	}
}
