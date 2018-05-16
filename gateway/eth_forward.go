package gateway

import (
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/eth/types"
	"github.com/ethereum/go-ethereum/common"
)

type EthForwarder struct {
}

func (e *EthForwarder) GetBalance(address, blockNumber string) (result string, err error) {
	err = accessor.GetBalance(&result, common.HexToAddress(address), blockNumber)
	//err = e.Accessor.RetryCall("latest", 2, &result, "eth_getBalance", common.HexToAddress(address), blockNumber)
	return
}

func (e *EthForwarder) SendRawTransaction(tx string) (result string, err error) {
	err = accessor.SendRawTransaction(&result, tx)
	//err = e.Accessor.RetryCall("latest", 2, &result, "eth_sendRawTransaction", tx)
	return
}

func (e *EthForwarder) GetTransactionCount(address, blockNumber string) (result string, err error) {
	err = accessor.GetTransactionCount(&result, common.HexToAddress(address), blockNumber)
	return
	//err = e.Accessor.RetryCall("latest", 2, &result, "eth_getTransactionCount", common.HexToAddress(address), blockNumber)
	//return result, nil
}

func (e *EthForwarder) Call(ethCall *types.CallArg, blockNumber string) (result string, err error) {
	err = accessor.Call(&result, ethCall, blockNumber)
	return
	//err = e.Accessor.RetryCall("latest", 2, &result, "eth_call", ethCall, blockNumber)
	//return result, nil
}
