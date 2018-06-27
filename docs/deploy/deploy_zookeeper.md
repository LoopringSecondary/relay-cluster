# Deploy zookeeper cluster

zookeeper Usage Scenario
* kafka cluster configuration center
* motan-rpc configuration center
* relay-cluster distributed lock type
* miner load balancing

Zookeeper needs cluster deployment to ensure availability. We recommended deploying more than 3 different nodes.

## Deployment
Take 3 nodes as example

### Start EC2 instance and configure security group
Start 3 EC2 instance, refer [Start EC2 instance](new_ec2.md)

Apply security group named `zookeeper-SecurityGroup` for each instance. If the security group hasn't been created, please create it first, refer to: [aws security group](security_group.md) 

### Deployment
Use the three server's intranet ip addresses to set up the three hosts, `zoo1~zoo3`, for later configuration

`sudo vim /etc/hosts`
Set as
```
x.x.x.x zoo1
x.x.x.x zoo2
x.x.x.x zoo3
```

* Initialize the zk environment

```
#If you do not deploy jre, you need to perform the following two operations
sudo apt update
sudo apt install openjdk-9-jre-headless -y

sudo mkdir /opt/loopring
sudo chown -R ubuntu:ubuntu /opt/loopring

cd /opt/loopring
wget http://mirrors.ocf.berkeley.edu/apache/zookeeper/zookeeper-3.4.10/zookeeper-3.4.10.tar.gz
tar xzf zookeeper-3.4.10.tar.gz
cd zookeeper-3.4.10/conf
cp zoo_sample.cfg zoo.cfg
mkdir -p /opt/loopring/data/zookeeper
```

* Modify and add the following configuration items:

`vim /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg`
```
dataDir=/opt/loopring/data/zookeeper
server.1=zoo1:2888:3888
server.2=zoo2:2888:3888
server.3=zoo3:2888:3888
```

Initialize myid, where the value of "n" on the three servers is 1, 2, and 3, respectively, simultaneously with zoo.conf above.

`echo "n" > /opt/loopring/data/zookeeper/myid`

## Start and Termination

### Start up
```
cd /opt/loopring/zookeeper-3.4.10/bin
./zkServer.sh start
```

Confirm that the service starts normally

```
tail -f zookeeper.out
telnet localhost 2181
```

### Termination
```
cd /opt/loopring/zookeeper-3.4.10/bin/
./zkServer.sh stop
```

## Logs
`/opt/loopring/zookeeper-3.4.10/bin/zookeeper.out`