# Configure security group

## AWS Security

> Because the default security group will reject traffic except port 22, we must configure the security group in order to provide services externally.

The aws security group accesses the Ec2 server by setting access rules and intercepting illegal traffic to improve the network security of the server. For details, refer to [aws security group](https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/UserGuide/using-network-security.html)

## Configuration entry
[EC2-Network and Security-Security Group] is the entrance to edit and view security groups

## Default security group
If you are not creating your own security group, select the default security group `launch-wizard-1`. This security group opens ssh port 22 and allows access with ssh.

When starting a new instance, it is recommended to use this default security group and manually configure the applicable security group later.

## Custom security group
When the service deployed by the instance needs additional ports for external access, or needs to access specific ports of other instances, you must create a new custom security group.

### If only access to external instances is needed
If you only need to access the ports of other instances, you do not need to provide ports for external access. You only need to create service groups with no additional configuration rules.

### Allowing external access to this instance
Set the allowable inbound rules, usually we deploy services are provided through the TCP protocol to external services, so the following only describes the inbound TCP settings

*  Allow specific ip

Allow only specific ip access which is usually only used in special circumstances

We can usually allow access to a specific ipv4 address by configuring ʻip+/32` in the source, which is similar to `203.0.113.1/32`

Do not use this for access between services because you will need to manually modify the service once it has been expanded or migrated.

*  Allow a set of instances

This is a common practice that allows a set of instances that are bound to a particular security group, group A, to access a set of instances that the current security group, group B, will bind to.

For example, a security group with the name ethnode-SecurityGroup will already have a group of servers deployed with eth nodes. Here, we need to allow the relay-cluster to access port 8545 of this group of eth nodes. We will create a new relayCluster-SecurityGroup and get the group ID, assuming sg-123456. Then, edit ethnode-SecurityGroup and add new rules to allow the source of port 8545 to be sg-123456. In this case, we associate the relayCluster-SecurityGroup with the node where the relay-cluster is deployed. The eth port 8545 is then opened to the relay-cluster. Following the eth or relay-cluster expansion, the rule will automatically take effect.
> Set the security group as the source of the scene, you need to enter [Group ID] in the source field instead of the group name

* Allow all traffic

For similar http services that need access provided access to the public network, you need to allow all IPs in the public network to access

Only two rules need to be configured at this time

|Types       | Protocol | Port range| Source     |
|------------|-----|--------|---------|
|自定义 TCP 规则| TCP | 80     | 0.0.0.0/0|
|自定义 TCP 规则| TCP | 80     | ::/0    |

## Modifying the instance associated with the security group
[EC2-Instance-Instance-Operation-Networking-Change Security Group]

## Security group overlay
For situations where a single server can deploy multiple services, it is recommended to set a security group for each service and then edit the multiple security groups that the server associates. This can improve the security group configuration flexibility and facilitate naming.

## The relay-cluster involving all security group configurations
The security group is independent of the EC2 instance, but can be created beforehand to establish a relationship between them. When the service is deployed, only the association is required.
* alb-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| HTTP | TCP | 80  | 0.0.0.0/0|
| HTTP | TCP | 80    | ::/0    |
| customized TCP rule | TCP | 443  | 0.0.0.0/0|
| customized TCP rule | TCP | 443    | ::/0    |
* miner-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
* relayCluster-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| customized TCP rule | TCP | 8087    |alb-securityGroup|
| customized TCP rule | TCP | 8083    |alb-securityGroup|
| customized TCP rule | TCP | 8100    |miner-SecurityGroup|
* extractor-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
* mysql-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| customized TCP rule | TCP | 3306    |relayCluster-securityGroup|
| customized TCP rule | TCP | 3306    |miner-SecurityGroup|
| customized TCP rule | TCP | 3306    |extractor-SecurityGroup|
* redis-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| customized TCP rule | TCP | 6379    |relayCluster-securityGroup|
| customized TCP rule | TCP | 6379    |miner-SecurityGroup|
| customized TCP rule | TCP | 6379    |extractor-SecurityGroup|
* ethnode-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| customized TCP rule | TCP | 8545    |alb-SecurityGroup|
| customized TCP rule | TCP | 8545    |relayCluster-securityGroup|
| customized TCP rule | TCP | 8545    |miner-SecurityGroup|
| customized TCP rule | TCP | 8545    |extractor-SecurityGroup|
* kafkaManager-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| customized TCP rule | TCP | 9000    |yourSpecialIp|
* kafka-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| customized TCP rule | TCP | 9092    |relayCluster-securityGroup|
| customized TCP rule | TCP | 9092    |miner-SecurityGroup|
| customized TCP rule | TCP | 9999    |extractor-SecurityGroup|
| customized TCP rule | TCP | 9092    |kafka-SecurityGroup|
| customized TCP rule | TCP | 9999    |kafkaManager-SecurityGroup|
* motanManager-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| customized TCP rule | TCP | 8080    |yourSpecialIp|
* zookeeperBrowser-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| customized TCP rule | TCP | 3000    |yourSpecialIp|
* zookeeper-securityGroup

|Type         | Protocol | Port range| Source     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| customized TCP rule | TCP | 2181    |relayCluster-securityGroup|
| customized TCP rule | TCP | 2181    |miner-SecurityGroup|
| customized TCP rule | TCP | 2181    |extractor-SecurityGroup|
| customized TCP rule | TCP | 2181    |kafka-SecurityGroup|
| customized TCP rule | TCP | 2181    |motanManager-SecurityGroup|
| customized TCP rule | TCP | 2181    |zookeeperBrowser-SecurityGroup|
| customized TCP rule | TCP | 2888    |zookeeper-SecurityGroup|
| customized TCP rule | TCP | 3888    |zookeeper-SecurityGroup|
