# Deploy kafka

> kafka relies on zookeeper, so you need to deploy zookeeper first

Kafka is used for Message communication between extractor and relay-cluster

Kafka needs cluster deployment to ensure availability. We recommended deploying more than 3 different nodes.

## Deployment
take 3 node as example

### Start EC2 instance and configure security group
Start 3 EC2 instance, refer [Start EC2 instance](new_ec2.md)

Apply security group named `kafka-SecurityGroup` for each instance. If the security group hasn't been created, please create it first, refer to: [aws security group](security_group.md)

### Deployment

We recommend you deploy more than three nodes to ensure availability. The following takes a three-node kafka cluster as an example.
```
#If you do not deploy jre, you need to perform the following two operations
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
Modify the config/server.properties configuration item

`vim config/server.properties`
```
#Different nodes set different ids
broker.id=0
#Modify the ip address to the current node's intranet IP address
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
#Set the correct zookeeper configuration, if you have set the host already, you can use the following configuration, otherwise you must directly specify the ip
zookeeper.connect=zoo1:2181,zoo2:2181,zoo3:2181
default.replication.factor=3
```

## Start and Termination
### Start up
`nohup bin/kafka-server-start.sh config/server.properties &`

### Termination
`bin/kafka-server-stop.sh`

## Logs
`/opt/loopring/kafka_2.12-0.11.0.2/logs`
