## 关于 Extractor

Loopring Extractor(解析器)，在路印生态系统中负责维护链上数据，给relay(中继)&miner(旷工)提供数据支持.<br>
解析器通过遍历以太坊block及transaction将路印中继支持的合约事件及方法<br>
从交易中提取出来，并转换成中继使用的数据类型，最后使用kafka消息队列发送给中继及旷工。

## 工作流程
* 确定起始块--解析器启动后首先根据配置文件参数及数据库存储的block数据确定起始解析blockNumber
* 获取节点数据--从起始块开始遍历以太坊节点block,并批量获取transaction&transactionReceipt
* 解析事件及方法--根据合约method/event数据结构解析transaction.Input&transactionReceipt.logs,并转换成中继及旷工需要的数据结构
* 分叉检测--根据块号及parent hash判断是否有分叉,如果有分叉,生成中继/矿工支持的分叉通知数据类型
* kafka消息队列--将解析的数据及分叉数据使用kfaka消息队列发送出去

## 环境
* gcc
* golang(v1.9.0以上)

## 依赖
* mysql数据库
* redis缓存
* 以太坊节点(集群)
* zookeeper-kafka消息队列

## 配置文件
- [参考](https://loopring.github.io/relay-cluster/deploy/deploy_extractor_cn.html#%E9%83%A8%E7%BD%B2%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)

## 快速开始
从github上拉取代码后,运行
```bash
cd $GOPATH/src/github.com/Loopring/extractor
go build -ldflags -s -v  -o build/bin/extractor cmd/main.go
```
将在项目build/bin目录下生成extractor可执行文件

```bash
extractor --config=you_config_file_path
```

## 部署
- [参考](https://loopring.github.io/relay-cluster/deploy/deploy_extractor_cn.html)
- [docker](docker_chinese)

## 支持
请访问官方网站获取联系方式，获得帮助: https://loopring.org
