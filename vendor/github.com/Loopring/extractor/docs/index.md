## About
The Loopring Extractor is responsible for maintaining data in the Loopring ecosystem, and provides data support for the Relay and miner.
The extractor relays support for contract events and methods by unpacking Ethereum blocks and transactions
It is extracted from the transaction and converted to the data type used by the relay. It is then sent to the relay and completed using the kafka message queue.

## Documents in Other Languages
- [Chinese (中文文档)](chinese.md)

## Extracting Process
* Determine the starting block - after the extractor starts, it first determines the starting extractor blockNumber based on the parameters of the configuration file and the block data stored in the database.
* Get node data - analyze Ethereum node block from the starting block and get the transaction & transactionReceipt in mass.
* Extractor events and methods - unpack transaction.Input&transactionReceipt.logs according to the contract method/event data structure and then convert to data structures needed for relaying and completion
* Bifurcation detection - based on if there is a fork based on the block number and parent hash, if there is a fork, generate a fork notification with a data type supported by the relay/miner
* Kafka message queue - send extracted data and forked data using kfaka message queue

## Environment

* gcc
* golang(above v1.9.0)

## Dependencies

* mysql
* redis
* zookeeper-kafka
* ethereum-node/cluster

## Configuration

- [reference](https://loopring.github.io/relay-cluster/deploy/deploy_index.html#%E9%83%A8%E7%BD%B2%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)

## Quick start
After pulling the code from github, run:
```
go build -ldflags -s -v  -o build/bin/extractor cmd/main.go
extractor --config=your_config_file_path
```

## Deploy
- [reference](https://loopring.github.io/relay-cluster/deploy/deploy_index.html#)
- [docker](docker.md)



## Support
Please visit the official website for contact information and help: https://loopring.org
