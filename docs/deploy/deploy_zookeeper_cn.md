# 部署zookeeper集群

zookeeper使用场景
* kafka集群配置中心
* motan-rpc配置中心
* relay-cluster分布式锁
* miner负载均衡

zookeeper需要进行集群部署来保证可用性，建议部署3个以上的奇数节点。

## 部署
3个节点为例

### 申请EC2实例并关联安全组
申请3台EC2服务器，参考[EC2实例](new_ec2_cn.md)
>测试环境只需申请1台EC2服务器

关联`zookeeper-SecurityGroup`安全组。
> 如果未创建该安全组，请参考[aws安全组](security_group_cn.md)关于`zookeeper-SecurityGroup`安全组的说明，创建后再关联

### 生产环境部署

```
#如果没有部署jre，需要先部署

sudo apt update
sudo apt -y install openjdk-8-jre-headless

sudo mkdir /opt/loopring
sudo chown -R ubuntu:ubuntu /opt/loopring

cd /opt/loopring
wget http://mirrors.ocf.berkeley.edu/apache/zookeeper/zookeeper-3.4.10/zookeeper-3.4.10.tar.gz
tar xzf zookeeper-3.4.10.tar.gz
cd zookeeper-3.4.10/conf
cp zoo_sample.cfg zoo.cfg
sudo mkdir -p /opt/loopring/data/zookeeper
```

修改以下配置项，依次填入三台zookeeper服务器的内网ip

`vim /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg`

```
dataDir=/opt/loopring/data/zookeeper
server.1=xx.xx.xx.xx:2888:3888
server.2=xx.xx.xx.xx:2888:3888
server.3=xx.xx.xx.xx:2888:3888
```

初始化myid，这里"n"在三台服务器的取值依次为1，2，3，和上面zoo.conf一致，每台服务器仅执行一次自身对应取值的命令

```
echo "n" > /opt/loopring/data/zookeeper/myid
```

### 测试环境部署

> 为了简便，测试环境采用单台实例部署zookeeper伪集群

安装过程参考生产环境的步骤，再执行以下脚本

```
cd zookeeper-3.4.10/conf

cp zoo_sample.cfg zoo.cfg
cp zoo_sample.cfg zoo2.cfg
cp zoo_sample.cfg zoo3.cfg

sudo mkdir -p /opt/loopring/data/zookeeper
sudo mkdir -p /opt/loopring/data/zookeeper2
sudo mkdir -p /opt/loopring/data/zookeeper3
```

修改zoo.cfg

`vim /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg`

```
dataDir=/opt/loopring/data/zookeeper
server.1=127.0.0.1:2888:3888
server.2=127.0.0.1:2889:3889
server.3=127.0.0.1:2890:3890
```

修改zoo2.cfg

`vim /opt/loopring/zookeeper-3.4.10/conf/zoo2.cfg`

```
dataDir=/opt/loopring/data/zookeeper2
server.1=127.0.0.1:2888:3888
server.2=127.0.0.1:2889:3889
server.3=127.0.0.1:2890:3890
```

修改zoo3.cfg

`vim /opt/loopring/zookeeper-3.4.10/conf/zoo3.cfg`

```
dataDir=/opt/loopring/data/zookeeper3
server.1=127.0.0.1:2888:3888
server.2=127.0.0.1:2889:3889
server.3=127.0.0.1:2890:3890
```

继续执行以下命令
```
echo "1" > /opt/loopring/data/zookeeper/myid
echo "2" > /opt/loopring/data/zookeeper2/myid
echo "3" > /opt/loopring/data/zookeeper3/myid
```
#### 测试环境启停
##### 启动
```
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh start /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh start /opt/loopring/zookeeper-3.4.10/conf/zoo2.cfg
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh start /opt/loopring/zookeeper-3.4.10/conf/zoo3.cfg
```
##### 停止
```
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh stop /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh stop /opt/loopring/zookeeper-3.4.10/conf/zoo2.cfg
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh stop /opt/loopring/zookeeper-3.4.10/conf/zoo3.cfg
```
##### 日志
`/opt/loopring/zookeeper-3.4.10/bin/zookeeper.out`

## 生产环境启停

### 启动
```
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh start /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg
```
确认服务正常启动
```
tail -f zookeeper.out
telnet localhost 2181
```

### 终止
```
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh start /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg
```

## 日志
`/opt/loopring/zookeeper-3.4.10/bin/zookeeper.out`
