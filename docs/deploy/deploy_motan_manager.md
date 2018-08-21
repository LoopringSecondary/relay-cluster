# Deploy motan-manager

Motan-manager is part of the open-source component of weibo motan-rpc. It can be used to view the rpc service in zookeeper, where the motan-rpc is located, and can perform simple management operations.

## Applying for an EC2 instance and connecting a security group
To apply for an EC2 server, reference: [Start EC2 instance](new_ec2.md)

Apply the `motanManger-SecurityGroup` security group. If the security group is not created, create it first, please refer to: [aws security group](security_group.md)

## Deployment
```
#Deploy mysql and write down the username and password
sudo apt install mysql-server -y
sudo apt install maven -y
sudo apt install openjdk-8-jdk-headless -y
sudo mkdir -p /opt/loopring/
sudo chown -R ubuntu:ubuntu /opt/loopring/
cd /opt/loopring/
git clone https://github.com/weibocom/motan.git
cd motan
mvn install -DskipTests
cd motan-manager
```

Modify the initialization sql `vim src/main/resources/motan-manager.sql` and add a script to create the motan_manager db

```
create database motan_manager;
use motan_manager;
```

Modify the configuration file, `vim src/main/resources/application.properties`
```
jdbc_url=jdbc:mysql://127.0.0.1:3306/motan-manager?useUnicode=true&characterEncoding=UTF-8
#Set up the correct database user
jdbc_username=xxx
jdbc_password=xxx
#Configure the zookeper address corresponding to motan-rpc
registry.url=127.0.0.1:2181
```

Initialize the motan_manager db

`mysql --host=localhost --port=3306 --user=xxx -p < src/main/resources/motan-manager.sql`

Build the jar package

`mvn package`

## Start and Termination

### Start up
`nohup java -jar target/motan-manager.jar &`

### Termination
`pkill -f "motan-manager"`

## Logs
`/opt/loopring/motan/motan-manager/nohup.out`

## Access
Browser access `http://extranetip:8080`
