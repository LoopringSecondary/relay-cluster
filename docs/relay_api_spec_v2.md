
# Relay API Spec V2.0

Loopring Relays are nodes that act as a bridge between Ethereum nodes and Loopring compatible wallets. A relay maintains global order-books for all trading pairs and is resposible for broadcasting orders trustlessly to selected peer-to-peer networks. 

Wallets can host their own relay nodes to facilitate trading using Loopring, but can also take advantage of public relays provided by the Loopring foundation or other third-parties. Order-book visualization services or order browsers can also set up their own relay nodes to display Loopring order-books to their users -- in such a senario, wallet-compatible APIs can be disabled so the relay will run in a read-only mode. 

This document describes the relay's public APIs v2.0 (JSON_RPC and SocketIO), but doesn't articulate how order-books and trading history is maintained.

As v1.0 supports arrays and json request format, v2.0 unifies the request params to only support json format, and adds socketIO support.

This document contains the following sections:
- Endport
- JSON-RPC Methods
- SocketIO Events


## Endport
```
JSON-RPC : http://{hostname}:{port}/rpc/v2/
JSON-RPC(mainnet) : https://relay1.loopring.io/rpc/v2/ or https://relay1.loopr.io/rpc/v2/ (better for china 4G network)
Ethereum standard JSON-RPC : https://relay1.loopring.io/eth or https://relay1.loopr.io/eth (better for china 4G network)
SocketIO(local|test) : https://{hostname}:{port}/socket.io
SocketIO(mainnet) : https://relay1.loopring.io/socket.io or https://relay1.loopr.io/socket.io (better for china 4G network)
*** Some socketio client make append '/socket.io' path in the end of the URL automatically. 
```

## JSON-RPC Methods 

* The relay supports all Ethereum standard JSON-RPCs, please refer to [eth JSON-RPC](https://github.com/ethereum/wiki/wiki/JSON-RPC).
* [loopring_getBalance](#loopring_getbalance)
* [loopring_submitOrder](#loopring_submitorder)
* [loopring_getOrders](#loopring_getorders)
* [loopring_getOrderByHash](#loopring_getorderbyhash)
* [loopring_getDepth](#loopring_getdepth)
* [loopring_getTicker](#loopring_getticker)
* [loopring_getTickers](#loopring_gettickers)
* [loopring_getFills](#loopring_getfills)
* [loopring_getTrend](#loopring_gettrend)
* [loopring_getRingMined](#loopring_getringmined)
* [loopring_getCutoff](#loopring_getcutoff)
* [loopring_getPriceQuote](#loopring_getpricequote)
* [loopring_getEstimatedAllocatedAllowance](#loopring_getestimatedallocatedallowance)
* [loopring_getGetFrozenLRCFee](#loopring_getgetfrozenlrcfee)
* [loopring_getSupportedMarket](#loopring_getsupportedmarket)
* [loopring_getSupportedTokens](#loopring_getsupportedtokens)
* [loopring_getContracts](#loopring_getcontracts)
* [loopring_getLooprSupportedMarket](#loopring_getlooprsupportedmarket)
* [loopring_getLooprSupportedTokens](#loopring_getlooprsupportedtokens)
* [loopring_getTransactions](#loopring_gettransactions)
* [loopring_unlockWallet](#loopring_unlockwallet)
* [loopring_notifyTransactionSubmitted](#loopring_notifytransactionsubmitted)
* [loopring_submitRingForP2P](#loopring_submitringforp2p)
* [loopring_getUnmergedOrderBook](#loopring_getunmergedorderbook)
* [loopring_flexCancelOrder](#loopring_flexcancelorder)
* [loopring_getNonce](#loopring_getnonce)
* [loopring_getTempStore](#loopring_gettempstore)
* [loopring_setTempStore](#loopring_settempstore)
* [loopring_notifyCirculr](#loopring_notifycirculr)
* [loopring_getEstimateGasPrice](#loopring_getestimategasprice)


## SocketIO Events

* [portfolio](#portfolio)
* [balance](#balance)
* [tickers](#tickers)
* [loopringTickers](#loopringtickers)
* [transactions](#transactions)
* [marketcap](#marketcap)
* [depth](#depth)
* [trends](#trends)
* [pendingTx](#pendingtx)
* [orderBook](#orderbook)
* [trades](#trades)
* [orders](#orders)
* [estimatedGasPrice](#estimatedgasprice)
* [addressUnlock](#addressUnlock)
* [circulrNotify](#circulrNotify)

## JSON RPC API Reference

### loopring_getBalance

Get user's balance and token allowance info.

#### Parameters

- `owner` - The address, if is null, will query all orders.
- `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).

```js
params: [{
  "owner" : "0x847983c3a34afa192cfee860698584c030f4c9db1",
  "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B"
}]
```

#### Returns

`Account` - Account balance info object.

- `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
2. `tokens` - Info on all token balance and allowance arrays.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_getBalance","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {
    "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
    "tokens": [
      {
          "token": "LRC",
          "balance": "0x000001234d",
          "allowance": "0x0000001233a"
      },
      {
          "token": "WETH",
          "balance": "0x00000012dae734",
          "allowance": "0x00000012aae734"
      }
    ]
  }
}
```

***

### loopring_submitOrder

Submits an order. The order is submitted to the relay as a JSON object, which will be broadcasted into a peer-to-peer network for off-chain order-book maintainance and ring-ming. Once mined, the ring will be serialized into a transaction and submitted to the Ethereum blockchain.

#### Parameters

`JSON Object` - The order object(refer to [LoopringProtocol](https://github.com/Loopring/protocol/blob/master/contracts/LoopringProtocol.sol))
  - `protocol` - Loopring contract address
  - `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
  - `walletAddress` - The wallet margin address.
  - `owner` - user's wallet address
  - `AuthAddr` - The wallet auth public key.
  - `AuthPrivateKey` - The wallet auth private key used to sign a ring when submitting.
  - `tokenS` - Token to sell.
  - `tokenB` - Token to buy.
  - `amountS` - Maximum amount of tokenS to sell.
  - `amountB` - Minimum amount of tokenB to buy if all amountS sold.
  - `validSince` - Indicating when this order is created.
  - `validUntil` - How long, in seconds, this order will be valid for.
  - `lrcFee` - Max amount of LRC to pay the miner. The real amount to pay is proportional to fill amount.
  - `buyNoMoreThanAmountB` - If true, this order does not allow a purchase of more than `amountB`.
  - `marginSplitPercentage` - The percentage of savings paid to miner.
  - `v` - ECDSA signature parameter v.
  - `r` - ECDSA signature parameter r.
  - `s` - ECDSA signature parameter s.
  - `powNonce` - Before an order is submitted, it must be verified by our pow check logic. If number of orders submitted is exceeded in a certain time frame, we will increase pow difficulty.
  - `orderType` - The order type, enum is (market_order|p2p_order), default is market_order.

```js
params: [{
  "protocol" : "0x847983c3a34afa192cfee860698584c030f4c9db1",
  "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
  "owner" : "0x847983c3a34afa192cfee860698584c030f4c9db1",
  "walletAddress" : "0x847983c3a34afa192cfee860698584c030f4c9db1",
  "authAddr" : "0xcE862ca5e8DE3c5258B05C558daFDC4B7703a217",
  "authPrivateKey" : "0xe84989447467e438565dd2715d93d7537e9bc07fe7dc3044d8cbf4bd10967a69",
  "tokenS" : "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
  "tokenB" : "0xEF68e7C694F40c8202821eDF525dE3782458639f",
  "amountS" : "0x0001234d234",
  "amountB" : "0x002a7d",
  "validSince" : "0x5af13e32",
  "valiUntil": "0x5af28fb2",
  "lrcFee" : "0x14",
  "buyNoMoreThanAmountB" : true,
  "marginSplitPercentage" : 50, // 0~100
  "v" : 112,
  "r" : "239dskjfsn23ck34323434md93jchek3",
  "s" : "dsfsdf234ccvcbdsfsdf23438cjdkldy",
  "powNonce" : 10,
  "orderType" : "market",
}]
```

#### Returns

`OrderHash` - The hash of the order.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_submitOrder","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": { "orderHash" : "0xc7756d5d556383b2f965094464bdff3ebe658f263f552858cc4eff4ed0aeafeb"}
}
```

***

### loopring_getOrders

Get loopring order list.

#### Parameters

- `owner` - The address, if is null, will query all orders.
- `orderHash` - The order hash.
- `status` - order status enum string.(status collection is : ORDER_OPENED(include ORDER_NEW and ORDER_PARTIAL), ORDER_NEW, ORDER_PARTIAL, ORDER_FINISHED, ORDER_CANCEL, ORDER_CUTOFF)
- `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
- `market` - The market of the order.(format is LRC-WETH)
- `side` - The side of order. only support "buy" and "sell".
- `orderType` - The type of order. only support "market_order" and "p2p_order", default is "market_order".
- `pageIndex` - The page want to query, default is 1.
- `pageSize` - The size per page, default is 50.

```js
params: [{
  "owner" : "0x847983c3a34afa192cfee860698584c030f4c9db1",
  "orderHash" : "0xf0b75ed18109403b88713cd7a1a8423352b9ed9260e39cb1ea0f423e2b6664f0",
  "status" : "ORDER_CANCEL",
  "side" : "buy",
  "orderType" : "market",
  "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
  "market" : "coss-weth",
  "pageIndex" : 2,
  "pageSize" : 40
}]
```

#### Returns

`PageResult of Order` - Order list with page info

1. `data` 
  - `orginalOrder` - The original order info when submitting.(refer to [LoopringProtocol](https://github.com/Loopring/protocol/blob/master/contracts/LoopringProtocol.sol))
  - `status` - The current order status.
  - `dealtAmountS` - Dealt amount of token S.
  - `dealtAmountB` - Dealt amount of token B.
  - `cancelledAmountS` - cancelled amount of token S.
  - `cancelledAmountB` - cancelled amount of token B.

2. `total` - Total amount of orders.
3. `pageIndex` - Index of page.
4. `pageSize` - Amount per page.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_getOrders","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {
    "data" : [
        {
             "originalOrder":{
                 "protocol":"0x8d8812b72d1e4ffCeC158D25f56748b7d67c1e78",
                 "delegateAddress":"0x17233e07c67d086464fD408148c3ABB56245FA64",
                 "address":"0x71C079107B5af8619D54537A93dbF16e5aab4900",
                 "hash":"0x52c90064a0503ce566a50876fc41e0d549bffd2ba757f859b1749a75be798819",
                 "tokenS":"LRC",
                 "tokenB":"WETH",
                 "amountS":"0x1b1ae4d6e2ef500000",
                 "amountB":"0xde0b6b3a7640000",
                 "validSince":"0x5aefd848",
                 "validUntil":"0x5af129c8",
                 "lrcFee":"0x19ac8532c2790000",
                 "buyNoMoreThanAmountB":false,
                 "marginSplitPercentage":"0x32",
                 "v":"0x1c",
                 "r":"0x8eb60e6b1ebfbb9ab7aaf1b54a78497f112cb1f6430cd414ffc2a1366639f35e",
                 "s":"0x1b65ca88a645d3540e8a89232b73e67818be5cd81c66fa0cc38802e7a8358226",
                 "walletAddress":"0xb94065482Ad64d4c2b9252358D746B39e820A582",
                 "authAddr":"0xEf04F928F89cFF2a86CB4C2086D2aDa7D3A29200",
                 "authPrivateKey":"0x94866e133eb0cc774ca09a9de59c4c671fee6f7e871104d5e14004ac46fcee2b",
                 "market":"LRC-WETH",
                 "side":"sell",
                 "createTime":1525667919
             },
             "dealtAmountS":"0x0",
             "dealtAmountB":"0x0",
             "cancelledAmountS":"0x0",
             "cancelledAmountB":"0x0",
             "status":"ORDER_OPENED",
        }
    ]
    "total" : 12,
    "pageIndex" : 1,
    "pageSize" : 10
  }
}
```

***

### loopring_getOrderByHash

Get loopring order by order hash.

#### Parameters

- `orderHash` - The order hash.

```js
params: [{
  "orderHash" : "0xf0b75ed18109403b88713cd7a1a8423352b9ed9260e39cb1ea0f423e2b6664f0",
}]
```

#### Returns

`Object of Order` - Order detail info.

- `orginalOrder` - The original order info when submitting.(refer to [LoopringProtocol](https://github.com/Loopring/protocol/blob/master/contracts/LoopringProtocol.sol))
- `status` - The current order status.
- `dealtAmountS` - Dealt amount of token S.
- `dealtAmountB` - Dealt amount of token B.
- `cancelledAmountS` - cancelled amount of token S.
- `cancelledAmountB` - cancelled amount of token B.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_getOrders","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {
     "originalOrder":{
         "protocol":"0x8d8812b72d1e4ffCeC158D25f56748b7d67c1e78",
         "delegateAddress":"0x17233e07c67d086464fD408148c3ABB56245FA64",
         "address":"0x71C079107B5af8619D54537A93dbF16e5aab4900",
         "hash":"0x52c90064a0503ce566a50876fc41e0d549bffd2ba757f859b1749a75be798819",
         "tokenS":"LRC",
         "tokenB":"WETH",
         "amountS":"0x1b1ae4d6e2ef500000",
         "amountB":"0xde0b6b3a7640000",
         "validSince":"0x5aefd848",
         "validUntil":"0x5af129c8",
         "lrcFee":"0x19ac8532c2790000",
         "buyNoMoreThanAmountB":false,
         "marginSplitPercentage":"0x32",
         "v":"0x1c",
         "r":"0x8eb60e6b1ebfbb9ab7aaf1b54a78497f112cb1f6430cd414ffc2a1366639f35e",
         "s":"0x1b65ca88a645d3540e8a89232b73e67818be5cd81c66fa0cc38802e7a8358226",
         "walletAddress":"0xb94065482Ad64d4c2b9252358D746B39e820A582",
         "authAddr":"0xEf04F928F89cFF2a86CB4C2086D2aDa7D3A29200",
         "authPrivateKey":"0x94866e133eb0cc774ca09a9de59c4c671fee6f7e871104d5e14004ac46fcee2b",
         "market":"LRC-WETH",
         "side":"sell",
         "createTime":1525667919
     },
     "dealtAmountS":"0x0",
     "dealtAmountB":"0x0",
     "cancelledAmountS":"0x0",
     "cancelledAmountB":"0x0",
     "status":"ORDER_OPENED",
  }
}
```

***

### loopring_getDepth

Get depth and accuracy by token pair

#### Parameters

1. `market` - The market pair.
2 `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
3. `length` - The length of the depth data. default is 20.


```js
params: [{
  "market" : "LRC-WETH",
  "delegateAddress": "0x5567ee920f7E62274284985D793344351A00142B",
  "length" : 10 // defalut is 50
}]
```

#### Returns

1. `depth` - The depth data, every depth element is an array of length three, which contains price, amount A, and amount B in market A-B in an order.
2. `market` - The market pair.
3. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_getDepth","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {
    "depth" : {
      "buy" : [
        ["0.0008666300","10000.0000000000","8.6663000000"]
      ],
      "sell" : [
        ["0.0008683300","900.0000000000","0.7814970000"],["0.0009000000","7750.0000000000","6.9750000000"],["0.0009053200","480.0000000000","0.4345536000"]
      ]
    },
    "market" : "LRC-WETH",
    "delegateAddress": "0x5567ee920f7E62274284985D793344351A00142B",
  }
}
```

***


### loopring_getTicker

Get info on Loopring's 24hr merged tickers from loopring relay.

#### Parameters
NULL


```js
params: [{}]
```

#### Returns

1. `high` - The 24hr highest price.
2. `low`  - The 24hr lowest price.
3. `last` - The newest dealt price.
4. `vol` - The 24hr exchange volume.
5. `amount` - The 24hr exchange amount.
5. `buy` - The highest buy price in the depth.
6. `sell` - The lowest sell price in the depth.
7. `change` - The 24hr change percent of price.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_getTicker","params":[{see above}],"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": [{
    "exchange" : "",
    "market":"EOS-WETH",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  {
    "exchange" : "",
    "market":"LRC-WETH",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  {
    "exchange" : "",
    "market":"RDN-WETH",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  {
    "exchange" : "",
    "market":"SAN-WETH",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  }]
}
```

***

### loopring_getTickers

Get the info on all the 24hr merged tickers in the market from loopring relay.

#### Parameters
1. `market` - The market info like LRC-WETH.


```js
params: [{
    "market" : "LRC-WETH"
}]
```

#### Returns

1. `high` - The 24hr highest price.
2. `low`  - The 24hr lowest price.
3. `last` - The newest dealt price.
4. `vol` - The 24hr exchange volume.
5. `amount` - The 24hr exchange amount.
5. `buy` - The highest buy price in the depth.
6. `sell` - The lowest sell price in the depth.
7. `change` - The 24hr change percent of price.

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getTickers","params":{see above}},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {"loopr" : {
    "exchange" : "loopr",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  "binance" : {
    "exchange" : "binance",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  "okEx" : {
    "exchange" : "okEx",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  "huobi" : {
    "exchange" : "huobi",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  }}
}
```

***

### loopring_getFills

Get order fill history. This history consists of OrderFilled events.

#### Parameters

1. `market` - The market of the order.(format is LRC-WETH)
2. `owner` - The address, if is null, will query all orders.
3. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
4. `orderHash` - The order hash.
5. `ringHash` - The order fill related ring's hash.
6. `pageIndex` - The page want to query, default is 1.
7. `pageSize` - The size per page, default is 50.

```js
params: [{
  "market" : "LRC-WETH",
  "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
  "owner" : "0x8888f1f195afa192cfee860698584c030f4c9db1",
  "orderHash" : "0xee0b482d9b704070c970df1e69297392a8bb73f4ed91213ae5c1725d4d1923fd",
  "ringHash" : "0x2794f8e4d2940a2695c7ecc68e10e4f479b809601fa1d07f5b4ce03feec289d5",
  "pageIndex" : 1,
  "pageSize" : 20 // max size is 50.
}]
```

#### Returns

`PAGE RESULT of OBJECT`
1. `ARRAY OF DATA` - The fills list.
  - `protocol` - The loopring contract address.
  - `owner` - The order owner address.
  - `ringIndex` - The index of the ring.
  - `createTime` - The timestamp of matching time.
  - `ringHash` - The hash of the matching ring.
  - `txHash` - The transaction hash.
  - `orderHash` - The order hash.
  - `orderHash` - The order hash.
  - `amountS` - The matched sell amount.
  - `amountB` - The matched buy amount.
  - `tokenS` - The matched sell token.
  - `tokenB` - The matched buy token.
  - `lrcFee` - The real amount of LRC to pay for miner.
  - `lrcReward` - The amount of LRC paid by miner to order owner in exchange for margin split.
  - `side` - Show the ordered to be filled as Buy or Sell.
  - `splitS` - The tokenS paid to miner.
  - `splitB` - The tokenB paid to miner.
2. `pageIndex`
3. `pageSize`
4. `total`

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getFills","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {
    "data" : [
      {
          "protocol":"0x4c44d51CF0d35172fCe9d69e2beAC728de980E9D",
          "owner":"0x66727f5DE8Fbd651Dc375BB926B16545DeD71EC9",
          "ringIndex":100,
          "createTime":1512631182,
          "ringHash":"0x2794f8e4d2940a2695c7ecc68e10e4f479b809601fa1d07f5b4ce03feec289d5",
          "txHash":"0x2794f8e4d2940a2695c7ecc68e10e4f479b809601fa1d07f5b4ce03feec289d5",
          "orderHash":"0x2794f8e4d2940a2695c7ecc68e10e4f479b809601fa1d07f5b4ce03feec289d5",
          "amountS":"0xde0b6b3a7640000",
          "amountB":"0xde0b6b3a7640001",
          "tokenS":"WETH",
          "tokenB":"COSS",
          "lrcReward":"0xde0b6b3a7640000",
          "lrcFee":"0xde0b6b3a7640000",
          "splitS":"0xde0b6b3a7640000",
          "splitB":"0x0",
          "market":"LRC-WETH"
      }
    ],
    "pageIndex" : 1,
    "pageSize" : 20,
    "total" : 212
  }
}
```

***

### loopring_getTrend

Get trend info per market. If you select 1Hr interval, this function will return a list(the length is 100 mostly). Each item represents a data point of the price change in 1Hr. The same goes for other intervals.

#### Parameters

1. `market` - The market type.
2. `interval` - The interval like 1Hr, 2Hr, 4Hr, 1Day, 1Week.

```js
params: {"market" : "LRC-WETH", "interval" : "2Hr"}

```

#### Returns

`ARRAY of JSON OBJECT`
  - `market` - The market type.
  - `high` - The 24hr highest price.
  - `low`  - The 24hr lowest price.
  - `vol` - The 24hr exchange volume.
  - `amount` - The 24hr exchange amount.
  - `open` - The opening price.
  - `close` - The closing price.
  - `start` - The statistical cycle start time.
  - `end` - The statistical cycle end time.

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getTrend","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {
    "data" : [
      {
        "market" : "LRC-WETH",
        "high" : 30384.2,
        "low" : 19283.2,
        "vol" : 1038,
        "amount" : 1003839.32,
        "open" : 122321.01,
        "close" : 12388.3,
        "start" : 1512646617,
        "end" : 1512726001
      }
    ]
  }
}
```

***

### loopring_getRingMined

Get all mined rings.

#### Parameters

1. `ringIndex` - The ring index
2. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
2. `protocolAddress` - The loopring [LoopringProtocolImpl](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
3. `pageIndex` - The page desired from query, default is 1.
4. `pageSize` - The size per page, default is 50.

```js
params: [{
  "ringIndex" : "0x15",
  "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
  "protocolAddress" : "0xb1170dE31c7f72aB62535862C97F5209E356991b",
  "pageIndex" : 1,
  "pageSize" : 20 // max size is 50.
}]
```

#### Returns

1. `data` - The ring info.(refer to [Ring&RingMined](https://github.com/Loopring/protocol/blob/3bdc40c4f319e8fe70f58f82563db49579094b5c/contracts/LoopringProtocolImpl.sol#L109)
  - `ringHash` - The ring hash.
  - `tradeAmount` - The number of orders to be filled in the ring.
  - `miner` - The miner that submits filled orders.
  - `feeRecepient` - The fee recepient address.
  - `txHash` - The ring match transaction hash.
  - `blockNumber` - The number of the block which contains the transaction.
  - `totalLrcFee` - The total lrc fee.
  - `time` - The ring fill time.
2. `total` - Total amount of orders.
3. `pageIndex` - Index of page.
4. `pageSize` - Amount per page.

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getRingMined","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {
     "data" : [
       {
        "ringhash" : "0xb903239f8543d04b5dc1ba6579132b143087c68db1b2168786408fcbce568238",
        "tradeAmount" : 3,
        "miner" : "0x8888f1f195afa192cfee860698584c030f4c9db1",
        "feeRecepient" : "0x8888f1f195afa192cfee860698584c030f4c9db1",
        "txHash" : "0x8888f1f195afa192cfee860698584c030f4c9db1",
        "blockNumber" : 10001,
        "totalLrcFee" : "0x101",
        "timestamp" : 1506114710,
       }
     ]
     "total" : 12,
     "pageIndex" : 1,
     "pageSize" : 10
  }
}
```
***

### loopring_getCutoff

Get cut off time of the address.

#### Parameters

1. `address` - The address.
2. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
3. `blockNumber` - "earliest", "latest" or "pending", default is "latest".

```js
params: [{
  "address": "0x8888f1f195afa192cfee860698584c030f4c9db1",
  "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
  "blockNumber": "latest"
}]
```

#### Returns
- `string` - the cutoff timestamp string.

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getCutoff","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": "1501232222"
```
***

### loopring_getPriceQuote

Get the USD/CNY/BTC quoted price of tokens

#### Parameters

1. `curreny` - The base currency desired from query, supported types are `CNY`, `USD`.

```js
params: [{ "currency" : "CNY" }]
```

#### Returns
- `currency` - The base currency, CNY or USD.
- `tokens` - Every token price int the currency.

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getPriceQuote","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {
    "currency" : "CNY",
    "tokens" : [
        {
          "token": "ETH",
          "price": 31022.12 // hopeful price :)
        },
        {
          "token": "LRC",
          "price": 100.86
        }
     ]
  }
}
```
***

### loopring_getEstimatedAllocatedAllowance

Get the total frozen amount of all unfinished orders

#### Parameters

1. `owner` - The address.
2. `token` - The specific token which you want to get.
3. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).

```js
params: [{
  "owner" : "0x8888f1f195afa192cfee860698584c030f4c9db1",
  "token" : "WETH",
  "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
}]
```

#### Returns
- `string` - The frozen amount in hex format.

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getEstimatedAllocatedAllowance","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": "0x2347ad6c"
}
```
***

### loopring_getGetFrozenLRCFee

Get the total frozen lrcFee of all unfinished orders

#### Parameters

1. `owner` - The address, if is null, will query all orders.
2. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).

```js
params: [{
  "owner" : "0x8888f1f195afa192cfee860698584c030f4c9db1",
  "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
}]
```

#### Returns
- `string` - The frozen amount in hex format.

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getGetFrozenLRCFee","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": "0x2347ad6c"
}
```
***

### loopring_getSupportedMarket

Get all relay-supported market pairs

#### Parameters
no input params.

```js
params: [{}]
```

#### Returns
- `array of string` - The array of all supported markets.

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getSupportedMarket","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": ["SAN-WETH","GNO-WETH","RLC-WETH","AST-WETH"]
}
```
***

### loopring_getSupportedTokens

Get all relay-supported tokens

#### Parameters
no input params.

```js
params: [{}]
```

#### Returns
- `array of string` - The array of all supported tokens.

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getSupportedTokens","params":[{}],"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": [
      {
        "protocol":"0xd26114cd6EE289AccF82350c8d8487fedB8A0C07",
        "symbol":"OMG",
        "source":"omisego",
        "deny":false,
        "decimals":1000000000000000000,
        "isMarket":false
      },....
  ]
}
```
***

### loopring_getContracts

Get all relay-supported contracts. The result is map[delegateAddress] List(loopringProtocol)

#### Parameters
no input params.

```js
params: [{}]
```

#### Returns
- `json object` - The map of delegateAddress with list of loopringProtocol.

#### Example
```js
// Request
curl -X GET --data '{"jsonrpc":"2.0","method":"loopring_getContracts","params":[{}],"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {
      "0x17233e07c67d086464fD408148c3ABB56245FA64": ["0x8d8812b72d1e4ffCeC158D25f56748b7d67c1e78"]
  }
}
```
***

### loopring_getLooprSupportedMarket

Get Loopr wallet supported market pairs. Exactly the same as loopring_getSupportedMarket but the name is different.

### loopring_getLooprSupportedTokens

Get Loopr wallet supported tokens. Exactly the same as loopring_getSupportedTokens but the name is different.

### loopring_getTransactions

Get user's latest transactions by owner.

#### Parameters

- `owner` - The owner address, must be applied.
- `thxHash` - The transaction hash.
- `symbol` - The token symbol like LRC,WETH.
- `status` - The transaction status, enum is (pending|success|failed).
- `txType` - The transaction type, enum is (send|receive|enable|convert).
- `pageIndex` - The page want to query, default is 1.
- `pageSize` - The size per page, default is 10.


```js
params: [{
  "owner" : "0x847983c3a34afa192cfee860698584c030f4c9db1",
  "thxHash" : "0xc7756d5d556383b2f965094464bdff3ebe658f263f552858cc4eff4ed0aeafeb",
  "symbol" : "RDN",
  "status" : "pending",
  "txType" : "receive",
  "pageIndex" : 2, // default is 1
  "pageSize" : 20 // default is 20
}]
```

#### Returns

`PAGE RESULT of OBJECT`
1. `ARRAY OF DATA` - The transaction list.
  - `from` - The transaction sender.
  - `to` - The transaction receiver.
  - `owner` - the transaction main owner.
  - `createTime` - The timestamp of transaction create time.
  - `updateTime` - The timestamp of transaction update time.
  - `hash` - The transaction hash.
  - `blockNumber` - The number of the block which contains the transaction.
  - `value` - The amount of transaction involved.
  - `type` - The transaction type, like wrap/unwrap, transfer/receive.
  - `status` - The current transaction status.
2. `pageIndex`
3. `pageSize`
4. `total`

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_getTransactions","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": {
      "data" : [
        {
          "owner":"0x66727f5DE8Fbd651Dc375BB926B16545DeD71EC9",
          "from":"0x66727f5DE8Fbd651Dc375BB926B16545DeD71EC9",
          "to":"0x23605cD09677600A91Df271C86E290cb09a17eeD",
          "createTime":150134131,
          "updateTime":150101931,
          "hash":"0xa226639a5852df7a61a19a473a5f6feb98be5247077a7b22b8c868178772d01e",
          "blockNumber":5029675,
          "value":"0x0000000a7640001",
          "type":"WRAP", // eth -> weth
          "status":"PENDING"
      }
    ],
    "pageIndex" : 1,
    "pageSize" : 20,
    "total" : 212
  }

}
```

***

### loopring_unlockWallet

Tell the relay the unlocked wallet info.

#### Parameters

- `owner` - The address, if is null, will query all orders.

```js
params: [{
  "owner" : "0x847983c3a34afa192cfee860698584c030f4c9db1",
}]
```

#### Returns

`Account` - Account balance info object.

1. `string` - Success or fail info.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_unlockWallet","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": ["unlock_notice_success"]
}
```

***

### loopring_notifyTransactionSubmitted

wallet should notify relay there was a transaction sending to eth network, then relay will get and save the pending transaction immediately.

#### Parameters

- `hash` - The txHash.
- `nonce` - The owner newest nonce.
- `to` - The target address to send.
- `value` - The value in transaction.
- `gasPrice`.
- `gas`.
- `input` - The value input in transaction.
- `from` - The transaction sender.
- `v` - ECDSA signature parameter v.
- `r` - ECDSA signature parameter r.
- `s` - ECDSA signature parameter s.


```js
params: [{
    "hash":"0xb98c216fd29b627a2845a9c3eb6e2ac591049c07c71cd4e4c0f00962adfb4409",
    "nonce":"0x66",
    "to":"0x07a7191de1ba70dbe875f12e744b020416a5712b",
    "value":"0x16345785d8a0000",
    "gasPrice":"0x4e3b29200",
    "gas":"0x5208",
    "input":"0x",
    "from":"0x71c079107b5af8619d54537a93dbf16e5aab4900",
    "v":"0x1c",
    "r":"0x8eb60e6b1ebfbb9ab7aaf1b54a78497f112cb1f6430cd414ffc2a1366639f35e",
    "s":"0x1b65ca88a645d3540e8a89232b73e67818be5cd81c66fa0cc38802e7a8358226",
  }]
```

#### Returns

`String` - txHash.

1. no result if failed, you can see error info in param.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_notifyTransactionSubmitted","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": "0xb98c216fd29b627a2845a9c3eb6e2ac591049c07c71cd4e4c0f00962adfb4409"
}
```


***

### loopring_submitRingForP2P

submit signed raw transaction of ring information, then relay can help submitting the ring while tracing the status of orders for wallet. 
please submit taker and maker order before invoking this method.

#### Parameters

- `takerOrderHash` - The taker order hash.
- `makerOrderHash` - The maker order hash.
- `rawTx` - The raw transaction.

```js
params: [{
  "takerOrderHash" : "0x52c90064a0503ce566a50876fc41e0d549bffd2ba757f859b1749a75be798819",
  "makerOrderHash" : "0x52c90064a0503ce566a50876fc41e0d549bffd2ba757f859b1749a75be798819",
  "rawTx" : "f889808609184e72a00082271094000000000000000000000000000000000000000080a47f74657374320000000000000000000000000000000000000000000000000000006000571ca08a8bbf888cfa37bbf0bb965423625641fc956967b81d12e23709cead01446075a01ce999b56a8a88504be365442ea61239198e23d1fce7d00fcfc5cd3b44b7215f",
}]
```

#### Returns

`txHash` - The transaction hash of eth_sendRawTransaction result. 

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_submitRingForP2P","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": "0xf0458d1a96ed7678f3abfe469c754fcb974b79aa632fc7da246fa983f37a49ce"
}
```

***

### loopring_getUnmergedOrderBook

get orderbook from relay. the difference of orderbook and depth is that orderbook doesn't merge amount of order, one orderbook record represents a order.

#### Parameters

1. `market` - The market pair.
2 `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).

```js
params: [{
  "market" : "LRC-WETH",
  "delegateAddress": "0x5567ee920f7E62274284985D793344351A00142B",
}]
```

#### Returns

1. `market` - The market pair.
2. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
3. `buy`  - buy list of orderbook element.
4. `sell` - sell list of orderbook element.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_getUnmergedOrderBook","params":{see above},"id":64}'

// Result
{
	"jsonrpc": "2.0",
	"id": 0,
	"result": {
		"delegateAddress": "0x17233e07c67d086464fD408148c3ABB56245FA64",
		"market": "LRC-WETH",
		"buy": [{
			"price": 0.00249499,
			"size": 0.0002,
			"amount": 0.08016064,
			"orderHash": "0xf4e92d5bd00db16fca1cda106285367e53d4e9d7f9b558aae53c04e6af6ddc4b",
			"lrcFee": 0.62,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529413285
		}, {
			"price": 0.0002,
			"size": 0.02841565,
			"amount": 142.07824071,
			"orderHash": "0xda13253eaab212edaf96e3cee28b41e672862b169291abb0e494ad7af0828855",
			"lrcFee": 10.52,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529467620
		}, {
			"price": 0.000089,
			"size": 0.02841565,
			"amount": 319.27694541,
			"orderHash": "0x24fef40a98793940be8db793ae2508a74fb015c32b2d74a063f66c0800db1d77",
			"lrcFee": 17.1,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529577356
		}, {
			"price": 0.00002,
			"size": 0.02841565,
			"amount": 1420.78240706,
			"orderHash": "0x49db31491b40840216e523499031365ccf3988c3759b99c1ed0bfa3ef3e4b69a",
			"lrcFee": 2.42,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529467558
		}],
		"sell": [{
			"price": 0.00249,
			"size": 22.03558119,
			"amount": 8849.631,
			"orderHash": "0x246e86a6d6130c931a420da60f3d7e74bbc8d77d176322531e3f3c02448be59f",
			"lrcFee": 60.01,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529968620
		}, {
			"price": 0.00249499,
			"size": 0.02930466,
			"amount": 11.7454,
			"orderHash": "0xabc93b8c2d8119876513b382adf6159bf9450330002d63ad126435f25094e485",
			"lrcFee": 0.002,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529466785
		}]
	}
}
```

***

### loopring_flexCancelOrder

flex cancel order, cancel order only in relay, will not use gas.

#### Parameters

- `sign` - The Sign Info with timestamp, Please see detail at params detail.
- `orderHash` - The order hash.
- `cutoffTime` - The cutoff time, if cancel by cutoff time.
- `tokenS` - The cutoff time, if cancel by cutoff time.
- `tokenB` - The cutoff time, if cancel by cutoff time.
- `type` - The cancel type, enum type is (1 : cancel by order hash | 2: cancel by owner | 3 : cancel by cutoff time | 4 : cancel by market).

```js
params: [{
  "orderHash" : "0x52c90064a0503ce566a50876fc41e0d549bffd2ba757f859b1749a75be798819", // if type = 1 , order hash must be applied.
  "cutoffTime" : 1332342342, // if type = 3, cutoff must be applied
  "tokenS" : "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // tokenS's token address, if type = 4, must be applied.
  "tokenB" : "0x86fa049857e0209aa7d9e616f7eb3b3b78ecfdb0", // tokenB's token address, if type = 4, must be applied.
  "type" : 2,
  "sign" : {
    // v, r, s = sign(keccak256(timestamp)) , please see web3j, same to loopring order sign, https://github.com/Loopring/loopring.js/wiki/%E8%B7%AF%E5%8D%B0%E5%8D%8F%E8%AE%AEv1.0.0%E8%AE%A2%E5%8D%95%E7%BB%93%E6%9E%84%E5%92%8C%E6%95%B0%E5%AD%97%E7%AD%BE%E5%90%8D
      "owner" : "0x71c079107b5af8619d54537a93dbf16e5aab4900", // owner address
      "v" : 27,
      "r" : "0xfc476be69f175c18f16cf72738cec0b810716a8e564914e8d6eb2f61e33ad454",
      "s" : "0x3570a561cb85cc65c969411dabfd470a436d3af2d04694a410f500f2a6238127",
      "timestamp" : 1444423423, // must be less than 10 minutes distance from the request sending time.
  }
}]
```

#### Returns

no result. if cancel failed, please see error message result.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_flexCancelOrder","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": ""
}
```

***

### loopring_getNonce

get newest nonce of user's address, plused on the pending transaction counts submitted to relay.

#### Parameters

- `owner` - The owner address.

```js
params: [{
  "owner" : "0x71c079107b5af8619d54537a93dbf16e5aab4900",
}]
```

#### Returns

`nonce` - The newest nonce value.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_getNonce","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": 23,
}
```

***

### loopring_getTempStore

a simple temporary string to string k/v store expire in 24hr, normally used when scaning QR intermediate data.

#### Parameters

- `key` - The temporacy data key.

```js
params: [{
  "key" : "testKey",
}]
```

#### Returns

`string` - The string value.

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_getTempStore","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": "test string value"
}
```

***

### loopring_setTempStore

a simple temporary string to string k/v store expire in 24hr, normally used when scaning QR intermediate data.

#### Parameters

- `key` - The temporacy data key.
- `value` - The temporacy data value.

```js
params: [{
  "key" : "testKey",
  "value" : "testValue",
}]
```

#### Returns

no result

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_setTempStore","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": ""
}
```

***

### loopring_notifyCirculr

notify the web wallet when wallet app scaning the QR code, do some action.

#### Parameters

- `owner` - The owner address to notify.
- `body` - The notify message body.

```js
params: [{
  "owner" : "0x71c079107b5af8619d54537a93dbf16e5aab4900",
  "body" : {"type" : "orderStatusUpdate", "status" : "submitted"}},
}]
```

#### Returns

no result

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_notifyCirculr","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": ""
}
```

***

### loopring_getEstimateGasPrice

get estimated gas price from Relay.

#### Parameters
no input param.

```js
params: [{}]
```

#### Returns

`hex string` - The hex string of gas price

#### Example
```js
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"loopring_getEstimateGasPrice","params":{see above},"id":64}'

// Result
{
  "id":64,
  "jsonrpc": "2.0",
  "result": "0x27372e63b",
}
```

***

## SocketIO Methods Reference

### balance

Get user's balance and token allowance info.

#### subscribe events
- balance_req : emit this event to receive push message.
- balance_res : subscribe this event to receive push message.
- balance_end : emit this event to stop receive push message.

#### Parameters

- `owner` - The wallet address
- `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).

```js
socketio.emit("balance_req", '{"owner" : "0x847983c3a34afa192cfee860698584c030f4c9db1", "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B"}', function(data) {
  // your business code
});
socketio.on("balance_res", function(data) {
  // your business code
});
```

#### Returns

`Account` - Account balance info object.

1. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
2. `tokens` - All token balance and allowance info array.

#### Example
```js
// Request
{
  "owner" : "0x847983c3a34afa192cfee860698584c030f4c9db1",
  "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B"
}

// Result
{
    "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
    "tokens": [
      {
          "token": "LRC",
          "balance": "0x000001234d",
          "allowance": "0x0000001233a"
      },
      {
          "token": "WETH",
          "balance": "0x00000012dae734",
          "allowance": "0x00000012aae734"
      }
    ]
}
```
***

### loopringTickers

Get 24hr merged tickers info from loopring relay.

#### subscribe events
- loopringTickers_req : emit this event to receive push message.
- loopringTickers_res : subscribe this event to receive push message.
- loopringTickers_end : emit this event to stop receive push message.

#### Parameters
NULL

```js
socketio.emit("loopringTickers_req", '{}', function(data) {
  // your business code
});
socketio.on("loopringTickers_res", function(data) {
  // your business code
});
```

#### Returns

1. `high` - The 24hr highest price.
2. `low`  - The 24hr lowest price.
3. `last` - The newest dealt price.
4. `vol` - The 24hr exchange volume.
5. `amount` - The 24hr exchange amount.
5. `buy` - The highest buy price in the depth.
6. `sell` - The lowest sell price in the depth.
7. `change` - The 24hr change percent of price.

#### Example
```js
// Request

{}

// Result
[
  {
    "exchange" : "",
    "market" : "LRC-WETH",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  {
    "exchange" : "",
    "market" : "RDN-WETH",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  {
    "market" : "ZRX-WETH",
    "exchange" : "",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  {
    "exchange" : "",
    "market" : "AUX-WETH"
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  }
]
```

### tickers

Get 24hr merged tickers reference info from other exchange like binance, huobi.

#### subscribe events
- tickers_req : emit this event to receive push message.
- tickers_res : subscribe this event to receive push message.
- tickers_end : emit this event to stop receive push message.

#### Parameters
1. `market` - The market selected.

```js
socketio.emit("tickers_req", '{"market" : "LRC-WETH"}', function(data) {
  // your business code
});
socketio.on("tickers_res", function(data) {
  // your business code
});
```

#### Returns

1. `high` - The 24hr highest price.
2. `low`  - The 24hr lowest price.
3. `last` - The newest dealt price.
4. `vol` - The 24hr exchange volume.
5. `amount` - The 24hr exchange amount.
5. `buy` - The highest buy price in the depth.
6. `sell` - The lowest sell price in the depth.
7. `change` - The 24hr change percent of price.

#### Example
```js
// Request

{"market" : "LRC-WETH"}

// Result
{
  "loopr" : {
    "exchange" : "loopr",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  "binance" : {
    "exchange" : "binance",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  "okEx" : {
    "exchange" : "okEx",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  },
  "huobi" : {
    "exchange" : "huobi",
    "high" : 30384.2,
    "low" : 19283.2,
    "last" : 28002.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "buy" : 122321,
    "sell" : 12388,
    "change" : "-50.12%"
  }
}
```

***

### transaction

push user's latest 20 transactions by owner.

#### subscribe events
- transaction_req : emit this event to receive push message.
- transaction_res : subscribe this event to receive push message.
- transaction_end : emit this event to stop receive push message.

#### Parameters

- `owner` - The owner address.
- `thxHash` - The transaction hash.
- `symbol` - The token symbol, like LRC, WETH....
- `status` - The transaction status enum(pending, success, failed).
- `txType` - The transaction type(approve, send, receive, convert...).
- `pageIndex` - The pageIndex.
- `pageSize`  - The pageSize.

```js
socketio.emit("transaction_req", '{see below}', function(data) {
  // your business code
});
socketio.on("transaction_res", function(data) {
  // your business code
});
```

#### Returns

`PAGE RESULT of OBJECT`
1. `ARRAY OF DATA` - The transaction list.
  - `from` - The transaction sender.
  - `to` - The transaction receiver.
  - `owner` - the transaction main owner.
  - `createTime` - The timestamp of transaction create time.
  - `updateTime` - The timestamp of transaction update time.
  - `hash` - The transaction hash.
  - `blockNumber` - The number of the block which contains the transaction.
  - `value` - The amount of transaction involved.
  - `type` - The transaction type, like convert, transfer/receive.
  - `status` - The current transaction status.
2. `pageIndex`
3. `pageSize`
4. `total`

#### Example
```js
// Request
params: {
  "owner" : "0x847983c3a34afa192cfee860698584c030f4c9db1",
  "thxHash" : "0x2794f8e4d2940a2695c7ecc68e10e4f479b809601fa1d07f5b4ce03feec289d5",
  "symbol" : "WETH",
  "status" : "pending",
  "txType" : "receive",
  "pageIndex" : 1,
  "pageSize" : 20
}

// Result
[
  {
      "owner":"0x66727f5DE8Fbd651Dc375BB926B16545DeD71EC9",
      "from":"0x66727f5DE8Fbd651Dc375BB926B16545DeD71EC9",
      "to":"0x23605cD09677600A91Df271C86E290cb09a17eeD",
      "createTime":150134131,
      "updateTime":150101931,
      "hash":"0xa226639a5852df7a61a19a473a5f6feb98be5247077a7b22b8c868178772d01e",
      "blockNumber":5029675,
      "value":"0x0000000a7640001",
      "type":"convert", // eth -> weth
      "status":"PENDING"
  },{}...
]

```

***

### marketcap

Get the USD/CNY/BTC quoted price of tokens.

#### subscribe events
- marketcap_req : emit this event to receive push message.
- marketcap_res : subscribe this event to receive push message.
- marketcap_end : emit this event to stop receive push message.

#### Parameters

1. `curreny` - The base currency want to query, supported types is `CNY`, `USD`.

```js
socketio.emit("marketcap_req", '{see below}', function(data) {
  // your business code
});
socketio.on("marketcap_res", function(data) {
  // your business code
});
```

#### Returns
- `currency` - The base currency, CNY or USD.
- `tokens` - Every token price int the currency.

#### Example
```js
// Request
{"currency" : "CNY"}

// Result
{
    "currency" : "CNY",
    "tokens" : [
        {
          "token": "ETH",
          "price": 31022.12 // hopeful price :)
        },
        {
          "token": "LRC",
          "price": 100.86
        }
     ]
}
```
***

### depth

Get depth and accuracy by token pair.

#### subscribe events
- depth_req : emit this event to receive push message.
- depth_res : subscribe this event to receive push message.
- depth_end : emit this event to stop receive push message.


#### Parameters

1. `market` - The market pair.
2. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
3. `length` - The length of the depth data. default is 20.


```js
socketio.emit("depth_req", '{see below}', function(data) {
  // your business code
});
socketio.on("depth_res", function(data) {
  // your business code
});
```

#### Returns

1. `depth` - The depth data, every depth element is a three length of array, which contain price, amount A and B in market A-B in order.
2. `market` - The market pair.
3. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).

#### Example
```js
// Request
{
  "market" : "LRC-WETH",
  "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
  "length" : 10 // defalut is 50
}

// Result
{
    "depth" : {
      "buy" : [
        ["0.0008666300","10000.0000000000","8.6663000000"]
      ],
      "sell" : [
        ["0.0008683300","900.0000000000","0.7814970000"],["0.0009000000","7750.0000000000","6.9750000000"],["0.0009053200","480.0000000000","0.4345536000"]
      ]
    },
    "market" : "LRC-WETH",
    "delegateAddress" : "0x5567ee920f7E62274284985D793344351A00142B",
  }
}
```

***

### trends

Get trend info per market.

#### subscribe events
- trends_req : emit this event to receive push message.
- trends_res : subscribe this event to receive push message.
- trends_end : emit this event to stop receive push message.

#### Parameters

1. `market` - The market type.
2. `interval` - The interval like 1Hr, 2Hr, 4Hr, 1Day, 1Week default is 1Hr.
```js
params: {"market" : "LRC-WETH", "interval" : "1Hr"}

```

#### Returns

`ARRAY of JSON OBJECT`
  - `market` - The market type.
  - `high` - The 24hr highest price.
  - `low`  - The 24hr lowest price.
  - `vol` - The 24hr exchange volume.
  - `amount` - The 24hr exchange amount.
  - `open` - The opening price.
  - `close` - The closing price.
  - `start` - The statistical cycle start time.
  - `end` - The statistical cycle end time.

#### Example
```js
// Request
{"market" : "LRC-WETH", "interval" : "4hr"}


// Result
[
  {
    "market" : "LRC-WETH",
    "high" : 30384.2,
    "low" : 19283.2,
    "vol" : 1038,
    "amount" : 1003839.32,
    "open" : 122321.01,
    "close" : 12388.3,
    "start" : 1512646617,
    "end" : 1512726001
  }.{}....
]

```

***

### pendingTx

Get pendingTx info per address.

#### subscribe events
- pendingTx_req : emit this event to receive push message.
- pendingTx_res : subscribe this event to receive push message.
- pendingTx_end : emit this event to stop receive push message.

#### Parameters

1. `owner` - The owner address.
```js
params: {"owner" : "0x5567ee920f7E62274284985D793344351A00142B"}

```

#### Returns

```js
socketio.emit("pendingTx_req", '{see below}', function(data) {
  // your business code
});
socketio.on("pendingTx_res", function(data) {
  // your business code
});
```

#### Returns

`PAGE RESULT of OBJECT`
`ARRAY OF DATA` - The transaction list.
  - `from` - The transaction sender.
  - `to` - The transaction receiver.
  - `owner` - the transaction main owner.
  - `createTime` - The timestamp of transaction create time.
  - `updateTime` - The timestamp of transaction update time.
  - `hash` - The transaction hash.
  - `blockNumber` - The number of the block which contains the transaction.
  - `value` - The amount of transaction involved.
  - `type` - The transaction type, like convert, transfer/receive.
  - `status` - The current transaction status.

#### Example
```js
// Request
params: {
  "owner" : "0x847983c3a34afa192cfee860698584c030f4c9db1"
}

// Result
[
  {
      "owner":"0x66727f5DE8Fbd651Dc375BB926B16545DeD71EC9",
      "from":"0x66727f5DE8Fbd651Dc375BB926B16545DeD71EC9",
      "to":"0x23605cD09677600A91Df271C86E290cb09a17eeD",
      "createTime":150134131,
      "updateTime":150101931,
      "hash":"0xa226639a5852df7a61a19a473a5f6feb98be5247077a7b22b8c868178772d01e",
      "blockNumber":5029675,
      "value":"0x0000000a7640001",
      "type":"convert", // eth -> weth
      "status":"PENDING"
  },{}...
]

```

***

### orderbook

The orderbook sync socketio event key. Please see detail at loopring_getUnmergedOrderBook.

#### subscribe events
emit with `_req` postfix and listen on `_res` postfix with the event key

#### Parameters

1. `market` - The market pair.
2. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).

```js
params: {"owner" : "0x5567ee920f7E62274284985D793344351A00142B", "delegateAddress" : "0x17233e07c67d086464fD408148c3ABB56245FA64"}

```

#### Returns

```js
socketio.emit("orderbook_req", '{see below}', function(data) {
  // your business code
});
socketio.on("orderbook_res", function(data) {
  // your business code
});
```

#### Returns

1. `market` - The market pair.
2. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
3. `buy`  - buy list of orderbook element.
4. `sell` - sell list of orderbook element.

#### Example
```js
// Request
params: {
  "owner" : "0x847983c3a34afa192cfee860698584c030f4c9db1",
  "delegateAddress" : "0x17233e07c67d086464fD408148c3ABB56245FA64",
}

// Result
{
		"delegateAddress": "0x17233e07c67d086464fD408148c3ABB56245FA64",
		"market": "LRC-WETH",
		"buy": [{
			"price": 0.00249499,
			"size": 0.0002,
			"amount": 0.08016064,
			"orderHash": "0xf4e92d5bd00db16fca1cda106285367e53d4e9d7f9b558aae53c04e6af6ddc4b",
			"lrcFee": 0.62,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529413285
		}, {
			"price": 0.0002,
			"size": 0.02841565,
			"amount": 142.07824071,
			"orderHash": "0xda13253eaab212edaf96e3cee28b41e672862b169291abb0e494ad7af0828855",
			"lrcFee": 10.52,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529467620
		}, {
			"price": 0.000089,
			"size": 0.02841565,
			"amount": 319.27694541,
			"orderHash": "0x24fef40a98793940be8db793ae2508a74fb015c32b2d74a063f66c0800db1d77",
			"lrcFee": 17.1,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529577356
		}, {
			"price": 0.00002,
			"size": 0.02841565,
			"amount": 1420.78240706,
			"orderHash": "0x49db31491b40840216e523499031365ccf3988c3759b99c1ed0bfa3ef3e4b69a",
			"lrcFee": 2.42,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529467558
		}],
		"sell": [{
			"price": 0.00249,
			"size": 22.03558119,
			"amount": 8849.631,
			"orderHash": "0x246e86a6d6130c931a420da60f3d7e74bbc8d77d176322531e3f3c02448be59f",
			"lrcFee": 60.01,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529968620
		}, {
			"price": 0.00249499,
			"size": 0.02930466,
			"amount": 11.7454,
			"orderHash": "0xabc93b8c2d8119876513b382adf6159bf9450330002d63ad126435f25094e485",
			"lrcFee": 0.002,
			"splitS": 0,
			"splitB": 0,
			"validUntil": 1529466785
		}]
	}

```

***

### trades

sync latest 40 trades per market.

#### subscribe events
emit with `_req` postfix and listen on `_res` postfix with the event key.

#### Parameters

1. `market` - The market of the order.(format is LRC-WETH)
2. `delegateAddress` - The loopring [TokenTransferDelegate Protocol](https://github.com/Loopring/token-listing/blob/master/ethereum/deployment.md).
3. `side` - The market side, buy or sell.

```js
params: {"market" : "LRC-WETH"}

```

#### Returns

```js
socketio.emit("trades_req", '{see below}', function(data) {
  // your business code
});
socketio.on("trades_res", function(data) {
  // your business code
});
```

#### Returns

`PAGE RESULT of OBJECT`
`ARRAY OF DATA` - The trade list.
  - `createTime` - The timestamp of trade create time.
  - `price` - The fill price.
  - `amount` - The fill amount.
  - `side`
  - `ringHash`
  - `lrcFee` - The lrcFee.
  - `splitS`
  - `splitB`

#### Example
```js
// Request
params: {
  "market" : "LRC-WETH",
  "delegateAddress" : "0x17233e07c67d086464fD408148c3ABB56245FA64",
}

// Result
[{
	"createTime": 1529379947,
	"price": 0.00249499,
	"amount": 1.369,
	"side": "sell",
	"ringHash": "0x14d8aa2eb24f917e1a23ca80956c68518f718b21c8d1275bdd57ba94ffaa5dd0",
	"lrcFee": "9281854027793469",
	"splitS": "0",
	"splitB": "0"
}, {
	"createTime": 1528885109,
	"price": 0.00082139,
	"amount": 3,
	"side": "sell",
	"ringHash": "0xd2830682254ec4bf71fae232da7fcd7e4c722031264fdbd7c1d67f98020b5b58",
	"lrcFee": "11884928508368094",
	"splitS": "0",
	"splitB": "11884928508368094"
}]

```

***

### orders

sync latest 40 orders per market.

#### subscribe events
emit with `_req` postfix and listen on `_res` postfix with the event key.

#### Parameters

1. `market` - The market of the order.(format is LRC-WETH)
2. `owner` - The owner address.
3. `orderType` - The type of order, market_order | p2p_order.

```js
params: {
    "market" : "LRC-WETH",
    "owner" : "0xA64B16a18885F00FA1AD6D3d3100C3E6F1CEf724",
    "orderType" : "market_order",
}

```

#### Returns

```js
socketio.emit("orders_req", '{see below}', function(data) {
  // your business code
});
socketio.on("orders_res", function(data) {
  // your business code
});
```

#### Returns

`PAGE RESULT of OBJECT`
`ARRAY OF DATA` - The order list.
  - same as loopring_getOrders result

#### Example
```js
// Request
params: {
  "market" : "LRC-WETH",
  "owner" : "0x17233e07c67d086464fD408148c3ABB56245FA64",
}

// Result
[{
	"originalOrder": {
		"protocol": "0x8d8812b72d1e4ffCeC158D25f56748b7d67c1e78",
		"delegateAddress": "0x17233e07c67d086464fD408148c3ABB56245FA64",
		"address": "0xA8E6dd605136cEEfC9daCEBE56E24d6aBb5B01d7",
		"hash": "0xcf4bd90ef91404aa020302f869cbbacc366ed868ba35c898bb1d77e526b49d72",
		"tokenS": "WETH",
		"tokenB": "LRC",
		"amountS": "0x43a1349385ba400",
		"amountB": "0x1741922b0f82940000",
		"validSince": "0x5b28f15f",
		"validUntil": "0x5b28f4e3",
		"lrcFee": "0x10a741a462780000",
		"buyNoMoreThanAmountB": true,
		"marginSplitPercentage": "0x32",
		"v": "0x1b",
		"r": "0x2d1507ed216c305d82e83c64aa0be7d8b8e69fcae2c2c7e6ce63c454378edeca",
		"s": "0x5a69b69055604d13a08c9cc8b67d9a44b66ea8906486003849ec8369fdedafdb",
		"walletAddress": "0xA8E6dd605136cEEfC9daCEBE56E24d6aBb5B01d7",
		"authAddr": "0x787B3C4c4B19209A20bD11ebcf279B64708F32ae",
		"authPrivateKey": "0xf5c1b07141a5198446bd73c1305ae6c07499866e52614ca1e31a04bdfac7a7ce",
		"market": "LRC-WETH",
		"side": "buy",
		"createTime": 1529409888,
		"orderType": "market_order"
	},
	"dealtAmountS": "0x0",
	"dealtAmountB": "0x0",
	"cancelledAmountS": "0x0",
	"cancelledAmountB": "0x0",
	"status": "ORDER_OPENED"
}, {
	"originalOrder": {
		"protocol": "0x8d8812b72d1e4ffCeC158D25f56748b7d67c1e78",
		"delegateAddress": "0x17233e07c67d086464fD408148c3ABB56245FA64",
		"address": "0x2E9f19B096069c2d93Dbc6FF3911f4e5ca0f6dD9",
		"hash": "0xd7bd822326c14b73f80bc5e365b290fccd1cfd4f186fa2caf82563b9046e5300",
		"tokenS": "LRC",
		"tokenB": "WETH",
		"amountS": "0x1741922b0f82940000",
		"amountB": "0x48df284c99c9000",
		"validSince": "0x5b28f15d",
		"validUntil": "0x5b28f4e1",
		"lrcFee": "0x10a741a462780000",
		"buyNoMoreThanAmountB": true,
		"marginSplitPercentage": "0x32",
		"v": "0x1b",
		"r": "0xef84e320d802fc5d626b61f8ca13fc6dca1ee46b0ed128aa22d813d438015dd5",
		"s": "0x44a7ef448941aafde2e064f4382cce8fc41826f78d2c9908f13f1902a2bc503f",
		"walletAddress": "0x2E9f19B096069c2d93Dbc6FF3911f4e5ca0f6dD9",
		"authAddr": "0x52b9BB323132241aC4973a6D918a01E90Ab690b5",
		"authPrivateKey": "0x3d8ccf6eea7717f12215d2f60fe585dc3587a88ea12bcdf375659ac882f6d1e8",
		"market": "LRC-WETH",
		"side": "sell",
		"createTime": 1529409887,
		"orderType": "market_order"
	},
	"dealtAmountS": "0x0",
	"dealtAmountB": "0x0",
	"cancelledAmountS": "0x0",
	"cancelledAmountB": "0x0",
	"status": "ORDER_OPENED"
}]

```

***

### estimatedGasPrice

sync estimated GasPrice from Relay.

#### subscribe events
emit with `_req` postfix and listen on `_res` postfix with the event key.

#### Parameters
no input params 

```js
params: {}

```

#### Returns

```js
socketio.emit("estimatedGasPrice_req", '{see below}', function(data) {
  // your business code
});
socketio.on("estimatedGasPrice_res", function(data) {
  // your business code
});
```

#### Returns

`hex string` - The hex string of gas price.

#### Example
```js
// Request
params: {}

// Result
"0x27372e63b"

```

***

### addressUnlock

listen the scan QR to login message notify.

#### subscribe events
emit with `_req` postfix and listen on `_res` postfix with the event key.

#### Parameters
1. `uuid` - The uuid to notify.

```js
params: {"uuid" : "dkx921"}

```

#### Returns

```js
socketio.emit("addressUnlock_req", '{see below}', function(data) {
  // your business code
});
socketio.on("addressUnlock_res", function(data) {
  // your business code
});
```

#### Returns

`string` - notify message body.

#### Example
```js
// Request
params: {}

// Result
{
    "uuid" : "dkx921",
    "selfDefinedParam" : "selfDefinedValues",
}

```

***

### circulrNotify

listen the scan QR message notify.

#### subscribe events
emit with `_req` postfix and listen on `_res` postfix with the event key.

#### Parameters
1. `owner` - The owner to notify.

```js
params: {"owner" : "0x71c079107b5af8619d54537a93dbf16e5aab4900"}

```

#### Returns

```js
socketio.emit("circulrNotify_req", '{see below}', function(data) {
  // your business code
});
socketio.on("circulrNotify_res", function(data) {
  // your business code
});
```

#### Returns

`string` - app and web negotiated notify message body.

#### Example
```js
// Request
params: {}

// Result
{
    "owner" : "0x71c079107b5af8619d54537a93dbf16e5aab4900",
    "body" : {"key1" : "value1"...},
}

```

***
