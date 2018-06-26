# zookeeper Usage Scenario
* kafka cluster configuration center
* motan-rpc configuration center
* relay-cluster distributed lock type
* miner load balancing

# Deployment
Zookeeper needs cluster deployment to ensure availability. We recommended deploying more than 3 different nodes.

## Configuration Ecosystem
3 nodes are an example
* Apply for 3 ubuntu instances
* Use the three server's intranet ip addresses to set up the three hosts, zoo1~zoo3, for later configuration

`sudo vim /etc/hosts`
Set as
```
x.x.x.x zoo1
x.x.x.x zoo2
x.x.x.x zoo3
```

* Initialize the zk ecosystem
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
cd /opt/loopring/zookeeper-3.4.10/bin/
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
<!--stackedit_data:
eyJoaXN0b3J5IjpbLTE2OTIzOTczMywxOTM5OTYxMTQ2XX0=
-->
