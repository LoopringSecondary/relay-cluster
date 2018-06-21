# Loopring Relay Cluster


Loopring Relay is an important part of Loopring’s technology and ecosystem. It centralizes the management of offline order pools, broadcasts and matches Loopring’s orders at the same time, and provides complete back-end services for exchanges and wallets. Based on the centralized system, we use the order broadcasting system to share orders to multiple relays to implement a network-wide order pool. This document describes the three parts of Loopring’s relay, being; how the relay functions, how third-party partners can access the relay, and how the relay deploys orders to the miners.


***


## Table of Contents

- [Glossary](#glossary)
- [How the Relay works](#how-the-relay-works)
   * [Loopring Ecosystem](#loopring-ecosystem)
   * [System Structure](#system-structure)
   * [Wallet Background Services](#wallet-background-services)
   * [Exchange Background Services](#exchange-background-services)
   * [Basic Services](#basic-services)
   * [Ethereum Services Analysis](#ethereum-analysis-services)
   * [Mining & Relaying Services](#mining-and-relaying-services)
- [Features](#features)
   * [Order Management](#order-management)
   * [Account Management](#account-management)
   * [Transactions & Matching](#transactions-and-matching)
   * [Market Exchange Rates](#market-exchange-rates)
   * [Meta information Management](#meta-information-management)
- [How to access the Relay](#how-to-access-the-relay)
   * [Jsonrpc](#jsonrpc)
   * [Socketio](#socketio)
   * [Sdk](#sdk)
   * [Test Environment](#test-environment)
- [How to deploy your own Relay](#how-to-deploy-your-own-relay)
   * [Code Compilation](#code-compilation)
   * [Profile](#profile)
   * [Docker Mirror](#docker-mirror)
- [Distributed Relay](#distributed-relay)
   * [Github code structure](#github)
   * [Microservice introduction](#microservice-introduction)
   * [Who is better for deploying a distributed version](#who-is-better-for-deploying-a-distributed-version)
   * [Distribute Version Docker Mirror](#distributed-version-mirroring)
   * [Distribute Version Profile](#distributed-relay-configuration)
- [Get help](#get-help)


## Glossary
| Type | Name | Explanation | 
|------|------|------|
| x | y | z|

| Type | Name | Explanation | 
|------|------|------|
|xxx| yyy | zzz |
| Order | Order | Order data that conforms to the Loopring protocol format |

---

| t1 | t2 | t3 |
|------|------|------|
|xxx| yyy | aaa |

## How the Relay Works

### Loopring Ecosystem

Loopring protocol currently based on the Ethereum smart contract has a chain order pool that is individually maintained by one relay, working together to achieve a total network pool integration with higher liquidity. We broadcast each order and connect all the relays to achieve an entire network of exchange orders. The Relay can use the order broadcast hub to exchange orders with other relays. At the same time the Loopring Miner can commit to multiple Relay orders. The order broadcast hub connects all the relays and the miner to form a larger (relative to a single relay) order pool.

___

### System Structure

The relay is a layered architecture system common in a centralized exchange, like a wallet, that runs in the background:
* System access layer, which is an external interface layer providing wallet and exchange services, supporting jsonrpc 2.0 and socketio
* Business logic layer, including various business logic interfaces and specific implementation
* Storage layer, including caching and constant storage, cache access to Redis. We chose a kv storage system (Redis) and a relational storage system (Mysql) to save and maximize storage. Redis is used to store data suitable for kv queries and queries that have higher performance requirements. One example of this could be the user’s account balance. MySQL is used to store relatively complex data queries such as the user's order and transaction information.
* Broadcast layer, used to spread out orders

___

### Wallet Background Services

Wallet background is used to support wallet-related services, including user account balance, authorizations, transaction data, wallet nonce, and cutoff information.

___

### Exchange Background Services

Exchange background services are used to support exchange-related functions, including: exchange order pools, order information, loop information, market trend charts/k-line data, in-depth/orderbook data, and global/third-party market exchange prices.

___

### Basic Services
Provides some basic configuration information for relays and Ethereum, such as Ethereum's estimated gasPrice, relay-supported contract list, relay-supported token list, and market-pair list.

___

### Ethereum Analysis Services

Ethereum analysis services are used to identify data on the chain and transfer it to the relay. This includes analyzing raw data such as blocks, transactions, and events, forming service categories, transferring to relays, and then completing transactions or events. It also updates user balances, orders, and transaction information. Lastly, resolution services deal with the frequent occurrence of Ethereum bifurcation.

___

### Mining And Relaying Services

According to the Loopring protocol white paper, Loopring orders are linked together on the chain, through the chained liquidation system, the miner discovers the loop in the chain, and the miner clears the chain, submitting the discovered loop to the Loopring smart contract. The liquidation is completed and the token received from the transaction is sent to the user and the sub-run address. The relaying service discovers loops and submits smart contracts.

___

## Features

### Order Management

This manages order life cycles, handles operations of user and Ethereum order updates, and provides different types of query interfaces for users and other subsystems.

#### Submitting an Order
The submit order interface handles orders from users and transmissions to other relay sources.

#### Checking
Before an order can be successfully processed, it needs to go through the following series of checks:
```
1. Workload verification check (Workload verification check detailed below)
2. Basic verification rules, including the minimum amount of amountS and the minimum amount of legal currency; the latest effective time limit (not later than a certain point of time, the current setting can’t be later than the order submission time of 10 hours); the proportion of sub-runs must be between 0 and 1; Validity check of account address and Token address, etc.
3. Verify the signature, and obtain the signed address based on v, r, s and wallet address
4. Is it supported by the token and the market
5. Whether the order has expired
```

#### Completed Order Information
After a series of verification checks, the orderHash, price, market and other fields of the order will be filled to facilitate the follow-up relational query.

#### Transmission Broadcast Policy
The last order will be transmitted to the database. If the relay sets a broadcast policy, the order will be broadcasted according to the policy. At this point, the order submission process ends.

#### Order Tracking

We provide a multi-dimensional query interface for orders, in addition to the most basic order list query (see API documentation), we also provide an order-based market depth query interface, and an orderbook query interface (market depth is the result of similar-price aggregation, and the Orderbook does not aggregate orders).

---

### Account Management

This is the user's Ethereum wallet address account, we support a series of write operations for the wallet address: transfer/authorization/WETH conversion, etc., transmit the user account balance to the in-memory database, and update the user account balance in real-time. The Relay centrally maintains the user's latest nonce and ensures the highest accuracy, while the nonce provided by the Relay can minimize the failure rate of the user's Ethereum activity. Account management also maintains all transaction records of the user to help the user understand their activity details.

---

### Transactions And Matching

The Loopring match service (Miner) obtains unfinished orders through internal RPC interfaces, finds loops, submits Ethereum smart contracts, and eventually resolves services through Ethereum to obtain match results and update order status.

---

### Market Exchange Rates

The relay provides the market information necessary for the exchange, including in-depth/orderbook/news/ticker/trend/kline, etc., and integrates coinmarketcap global market conditions. Recently, we teamed up with MyToken to access the MyToken open platform to obtain more comprehensive market data and to provide users with comprehensive transaction reference information.

---

### Meta Information Management 

The meta-information management here includes information on the best estimated gasPrice, the contract information currently supported by the relay (Delegate and Protocol correspondence), the list of transaction pairs supported by the relay, and the Token list.

---

## How to Access the Relay

There are currently 3 ways to access the relay: jsonrpc, socketio, and sdk.

### jsonrpc

JSON-RPC is a cross-language Remote Call Protocol based on JSON. JSON-RPC is very simple. The format of the data transfer to the server requested is as follows (based on JSON2.0):
```
{
   "jsonrpc" : 2.0,
   "method" : "helloLoopring", 
   "params" : [{"key" : "value"}], 
   "id" : 347579
}
```

For specific JSON RPC interface access info, please refer to the API documentation.

---

### socketio

SocketIO encapsulates the underlying communication as an event programming model and provides event-based long connection communication. In order to improve the real-time performance of data, the relay adopts the same technical means as the traditional centralized exchange to enhance the user experience. For specific socket access methods, refer to the API documentation.

--- 

### sdk
Based on the network interface, we further encapsulated the relay interface calls to form multiple platforms of sdk, allowing developers to get access through method calls. Currently, the JavaScript version of Loopring.js is open source, and IOS and Java SDK are under development. The SDK is a development from the JSONRPC and Socketio interface calls.

---

### Test Environment

The relay currently provides a complete set of testing environments for the partners to develop and debug their Dapps. To use the relay test environment, you need to know this relevant information:
1. The test environment address is 13.112.62.24
 - the jsonrpc entry is: http://13.112.62.24/rpc/v2
 - socketio entry is: http://13.112.62.24/socket.io
 - ethereum test node entry is : http://13.112.62.24/eth
2. 13.112.62.24:8000 is the test environment entrance for loopr web wallet
3. To use relays for wallets, some Relay and Ethereum node configurations are needed. These configuration items are different in the main network and our test environment. Here are the configuration items and test environment configurations:
```
DelegateAddress, please refer to glossary, test environment address: 0xa0af16edd397d9e826295df9e564b10d57e3c457
ProtocolImplAddress, please refer to glossary, test environment address: 0x456044789a41b277f033e4d79fab2139d69cd154
walletAddress, wallet sub-address, wallet set by yourself
chainId，Ethereum EIP155 introduced, in order to prevent repeated attacks, test environment values: 7107171
List of tokens, list of tokens for wallet configuration, including token details, can also be passed
loopring_getLooprSupportedTokens, get list of tokens supported by Relay
```

4. Online Relay version only supports https, test environment only supports http

---

## How to deploy your own Relay

The Loopring Foundation implements two versions of the relay: standalone and cluster. Below is how to compile and deploy the standalone version.

---

### Code Compilation
Relays are developed using the golang language. Version 1.9.2 is used when developing relays. Version 1.9 is recommended for compiler relays. Go build your own environment. Please make your own google source code address: https://github.com/Loopring/relay, please use the master branch code to compile and run, and please refer to README

--- 

### Profile

The relay contains two configuration files: relay.toml and tokens.json. The first is a relay global configuration and the second is used to specify the list of Tokens supported by the relay and the market. There are many relay configuration items. Here is a description of the more important ones in relay.toml:
```
websocket      - Relay’s external socketio ports
jsonrpc        - Relay’s external jsonrpc ports
ipfs & gateway - The configuration of the message broadcast, which has been annotated, and cannot be configured by you
accessor       - Ethereum network node url, an array that can support multi-access simultaneously
extractor.start_block_number   - Which block to start processing when the relay starts, it is recommended to change to the latest block when starting the first time, reducing unnecessary block synchronization
extractor.confirm_block_number - After a certain number of blocks are confirmed, the extractor sends the transaction message for use by other modules.
common abi     - Contains ERC20 standard ABI, WETH ABI and Loopring contract ABI configuration, if it is linked to the main network, you will not be able to modify
gateway filter - Used to configure various validation rules for submitting orders
miner          - When mining in miner mode, configure mining parameters
```

We will further update the relay.toml configuration file, add comments, and do self-explanations.

---

### Docker Mirror

docker miner address: https://hub.docker.com/r/loopring/relay, the current version is older, we will update the image as soon as possible.

---

### Distributed Relay
The distributed relay is a distributed version of the relay. Although the single node relay can fully support the wallet and the DEX service, there is a single point of failure in the architecture, the upgrade and expansion are difficult, and the performance bottleneck cannot be downgraded when system problems are encountered. In order to make relaying an enterprise-class application, and to provide high-performance wallets and DEX services, we have reconfigured the relays to split the original system into three services of relays, bindings, and parsers. This is a distributed version of microservices.

---

### github
The distributed relay includes the following github libraries: relay-cluster/miner/extractor/relay-lib, which gives a brief introduction to the role of each library:
* Relay-cluster: Relay-cluster version, providing wallet and DEX background services
* miner: cooperative services
* extractor: Ethereum analysis service
* relay-lib: Basic library for use on the above microservices system

---

### Microservice Introduction

#### relay-cluster
The relay-cluster function is basically the same as that of the previous single-node version relay function. It provides background services for wallets and DEX applications, and strips off the matching service and the Ethereum resolution service originally used in the single-node version relay. This is the function provided by the current relay-cluster.

#### miner
Match services, receive orders, find loops, submit loops to Ethereum networks, serve relay-cluster communications via rpc

#### extractor
Ethereum analyzing services. Every transaction, every event, etc. that occur on Ethereum is analyzed by the extractor and then put into events. It is provided to the relay-cluster and miner through the message queue (currently kafka).

---

## Who is better for deploying a distributed version
Distributed relay is a complete enterprise-level distributed application. The current system volume is relatively small and only contains 3 micro-services, but with other distributed applications, we have done the following:
* Based on aws ALB to provide a complete load-balancing strategy
* We deployed a highly available kafka&zookeeper cluster
* Adopted aws redis active and standby mode cache clusters and mysql storage cluster
* Has 200+ perfect cloudwatch monitoring configuration items
* Supported aws codedeploy one-click deployment script
* There is disaster recovery and fault tolerance in place for each single point where there may be failure or other problems in the system.

Therefore, if partners want to deploy a highly capable relay, they need to do a lot of system work and need very professional engineers to maintain it for a long time. Our proposal is to deploy distributed relays with partners with certain R&D strengths. Of course, we will also provide deployment assistance as much as possible.

---

### Distributed Version Mirroring
In order to facilitate the deployment of distributed relays by partners, we will provide the following mirrors (in production):
* relay-cluster mirror
* Miner mirror
* Extractor mirror
* zookeeper&kafka
* mysql & redis, please download the official mirror yourself

---

### Distributed Relay Configuration
The configuration files of each microservice and middleware need to be independent of the mirror configuration. We will update the configuration file description and increase self-explanation. This part of the work will be completed together with the dockerization and then provided to the partners for use.

---

## Get Help
Please visit the official website for contact information and help: https://loopring.org