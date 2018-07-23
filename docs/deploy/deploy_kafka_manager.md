# Deploy kafka-manager

Kafka-manager is a yahoo open-source kafka management tool that can be used to view the status of nodes and topics in the cluster, as well as some performance indicators

## Start EC2 instance and configure security group
Start 1 EC2 instance, refer [Start EC2 instance](new_ec2.md)

Apply security group named `kafkaManager-SecurityGroup` for each instance. If the security group hasn't been created, please create it first, refer to: [aws security group](security_group.md) 

## Deployment
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
modify conf/application.conf

`vim conf/application.conf`
```
#If the ZooKeeper host is set, you can use zookeeper intranet IP address or configure it as follows
kafka-manager.zkhosts="zoo1:2181,zoo2:2181,zoo3:2181"
#Configure the following authentication options according to the situation
basicAuthentication.enabled=true
basicAuthentication.username="admin"
basicAuthentication.password="admin"
```
## Start and Termination
### Start
`nohup ./bin/kafka-manager &`

### Termination
`pkill -f play.core.server.ProdServerStart`

## Logs
`/opt/loopring/kafka-manager-1.3.3.17/nohup.out`

## Access
Access the browser `http://Extranetip:9000`
