package gateway_test

import (
	"testing"
	"github.com/Loopring/relay-cluster/test"
	"github.com/Loopring/relay-cluster/gateway"
	"github.com/Loopring/relay-cluster/accountmanager"
	"github.com/Loopring/relay-lib/types"
	"encoding/json"
	orderviewer "github.com/Loopring/relay-cluster/ordermanager/viewer"
	"time"
)

func TestHandleInputOrder(t *testing.T) {
	cfg := test.LoadConfig()

	rds := test.Rds()
	marketCap := test.GenerateMarketCap()
	accountmanager.Initialize(&cfg.AccountManager, cfg.Kafka.Brokers)
	viewer := orderviewer.NewOrderViewer(&cfg.OrderManager, rds , marketCap)
	gateway.Initialize(&cfg.GatewayFilters, &cfg.Gateway, viewer, marketCap, accountmanager.AccountManager{})

	s := `{"protocol":"0x456044789a41b277f033e4d79fab2139d69cd154","delegateAddress":"0xa0af16edd397d9e826295df9e564b10d57e3c457","authAddr":"0x47fe1648b80fa04584241781488ce4c0aaca23e4","authPrivateKey":"0x5a12849ba30a17144288161d348094588ade48a3eeb3c80fcfecd8f43934f15b","walletAddress":"0x251f3bd45b06a8b29cb6d171131e192c1254fec1","tokenS":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","tokenB":"0xef68e7c694f40c8202821edf525de3782458639f","amountS":"0x16345785d8a0000","amountB":"0x1043561a8829300000","validSince":"0x5b334c3d","validUntil":"0x5bb7223d","lrcFee":"0x4563918244f40000","buyNoMoreThanAmountB":false,"marginSplitPercentage":0,"v":27,"r":"0x98d8e21f487d4713be6cc86ea2f408dbfde407ccaf83e3404c248066ec05ba3b","s":"0x1a40d060a90d264b9ca5eef687f8bfe0eae3a91e6c2fb7250e0f10fae29df54d","price":"1/3000","owner":"0x251f3bd45b06a8b29cb6d171131e192c1254fec1","hash":"0xb012700e2923383340976b907d57194f4da26826dc3a5d55403c527fa46017da","market":"LRC-WETH","createTime":0,"powNonce":1,"side":"buy","orderType":"market_order"}`
	order := &types.Order{}
	json.Unmarshal([]byte(s), order)

	gateway.HandleInputOrder(order)

	time.Sleep(5 * time.Second)
}

