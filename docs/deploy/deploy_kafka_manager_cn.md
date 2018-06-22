kafka-manager是yahoo开源的kafka管理工具，可以用来查看集群内的节点和topic状态，以及一些性能指标

## 部署
```
sudo mkdir /opt/loopring
sudo chown -R ubuntu:ubuntu /opt/loopring
cd /opt/loopring
sudo apt update
sudo apt install openjdk-8-jdk-headless -y
wget https://downloads.lightbend.com/scala/2.12.6/scala-2.12.6.deb
sudo dpkg -i scala-2.12.6.deb 
wget https://dl.bintray.com/sbt/debian/sbt-1.1.5.deb
sudo dpkg -i sbt-1.1.5.deb
git clone https://github.com/yahoo/kafka-manager.git
cd kafka-manager
sbt clean dist
mv ./target/universal/kafka-manager-1.3.3.17.zip ./
unzip kafka-manager-1.3.3.17.zip
cd kafka-manager-1.3.3.17
```
修改conf/application.conf

`vim conf/application.conf`
```
#如果设置了zookeeper的hosts，可以如下配置，否则使用zookeeper 内网ip地址
kafka-manager.zkhosts="zoo1:2181,zoo2:2181,zoo3:2181"
#根据实际情况配置下面认证选项
basicAuthentication.enabled=true
basicAuthentication.username="admin"
basicAuthentication.password="admin"
```
## 启停

### 启动
`nohup ./bin/kafka-manager &`

### 终止
`pkill -f play.core.server.ProdServerStart`

## 日志
/opt/loopring/kafka-manager-1.3.3.17/nohup.out

## 访问
浏览器访问 `http://外网ip:9000`

## 安全组
创建名称为`kafka-manager-SecurityGroup`的安全组，配置如下

|类型         | 协议 | 端口范围| 来源     |
|-------------|-----|--------|---------|
| SSH         | TCP | 22     | 0.0.0.0/0|
| 自定义 TCP 规则| TCP | 9000   | 访问端出口ip/32 |