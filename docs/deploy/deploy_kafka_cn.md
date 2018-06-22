> kafka依赖zookeeper，所以需要先部署zookeeper
### kafka使用场景
extractor和relay-cluster之间的消息通信

## 部署
建议部署3个以上的节点来保证可用性，下面以3个节点的kafka集群为例
```
#如果没有部署jre，需要执行下面两步操作
sudo apt update
sudo apt install openjdk-9-jre-headless -y

sudo mkdir /opt/loopring
sudo chown -R ubuntu:ubuntu /opt/loopring

cd /opt/loopring
wget http://apache.mirrors.lucidnetworks.net/kafka/0.11.0.2/kafka_2.12-0.11.0.2.tgz
tar xzf kafka_2.12-0.11.0.2.tgz
cd kafka_2.12-0.11.0.2/
sudo mkdir -p /opt/loopring/data/kafka-logs
sudo chown -R ubuntu:ubuntu /opt/loopring/data/kafka-logs
```
修改config/server.properties配置项

`vim config/server.properties`
```
#不同节点设置不同的id
broker.id=0
# 修改ip地址为当前节点的内网ip地址
listeners=PLAINTEXT://x.x.x.x:9092

log.dirs=/opt/loopring/data/kafka-logs
offsets.topic.replication.factor=3
min.insync.replicas=1
transaction.state.log.replication.factor=2
log.flush.interval.messages=300
log.flush.interval.ms=300
log.flush.scheduler.interval.ms=300
log.flush.start.offset.checkpoint.interval.ms=2000
log.retention.hours=168
#设置正确的zookeeper配置，如果已经设置了host可以用下面的配置，否则直接指定ip
zookeeper.connect=zoo1:2181,zoo2:2181,zoo3:2181
default.replication.factor=3
```
## 启停
### 启动
`nohup bin/kafka-server-start.sh config/server.properties &`
### 终止
`bin/kafka-server-stop.sh`
## 日志
`/opt/loopring/kafka_2.12-0.11.0.2/logs`
## 安全组
创建名称为`kafka-SecurityGroup`的安全组，配置如下

|类型         | 协议 | 端口范围| 来源     |
|-------------|-----|--------|---------|
| SSH         | TCP | 22     | 0.0.0.0/0|
|自定义 TCP 规则| TCP | 9092   |relayCluster-SecurityGroup|
|自定义 TCP 规则| TCP | 9092   |kafka-SecurityGroup    |
|自定义 TCP 规则| TCP | 9092   |miner-SecurityGroup    |
