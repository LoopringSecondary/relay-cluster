# Deployment
> The relay-cluster related services are deployed in aws, and some functions strongly depend on related services provided by aws. If other cloud service providers are used, some of the functions may be unavailable or some unexpected errors may occur.

> The relay-cluster and its dependent extractor services need to be deployed through the cluster to avoid all points of failure. Although only a single node can be deployed, the normal startup of a single node still needs to rely on the components that make up the cluster, including zookeeper and kafka.

Follow the order of dependencies
### Universal
Before deploying any service, select the appropriate EC2 instance to start and the subsequent services will be deployed on this instance. After the instance is started, you also need to associate custom security groups to allow specific ports to communicate between instances. Specific service deployment instructions will give suggested EC2 and security group configuration

* [EC2 example](https://github.com/Loopring/relay-cluster/wiki/%E5%90%AF%E5%8A%A8aws-EC2%E5%AE%9E%E4%BE%8B)

* [Aws security group](https://github.com/Loopring/relay-cluster/wiki/%E9%85%8D%E7%BD%AEaws%E5%AE%89%E5%85%A8%E7%BB%84)
### Storage and Communication
* ethnode

Relay-cluster implements eth network access by interacting with eth node
* [mysql](https://github.com/Loopring/relay-cluster/wiki/%E9%83%A8%E7%BD%B2mysql)

This is the main backend store of the relay-cluster, the storage contains orders and transactions

* redis

Mainly used to increase the access speed of the system, or store non-critical data
* [zookeeper](https://github.com/Loopring/relay-cluster/wiki/%E9%83%A8%E7%BD%B2zookeeper)

Used for system configuration management and kafka metadata storage
* [kafka](https://github.com/Loopring/relay-cluster/wiki/%E9%83%A8%E7%BD%B2kafka%E9%9B%86%E7%BE%A4)

Kafka implements asynchronous communication between services to facilitate system decoupling and expansion
### Service
* [Access CodeDeploy](https://github.com/Loopring/relay-cluster/wiki/%E6%8E%A5%E5%85%A5CodeDeloy)

At present, the relevant components of the relay are deployed through aws CodeDeploy+github to facilitate rapid iteration.

* [relay-cluster](https://github.com/Loopring/relay-cluster/wiki/%E9%83%A8%E7%BD%B2relay-cluster)

This is the core component of the backend service. It provides external jsonRpc interface. It is an access service for the Loopring wallet and transaction system, and also provides a motan-rpc interface to the miner.

* [extractor](https://github.com/Loopring/relay-cluster/wiki/%E9%83%A8%E7%BD%B2extractor)

The service analyzes the eth network transaction and synchronizes the result with the relay-cluster via kafka

* [miner](https://github.com/Loopring/relay-cluster/wiki/%E9%83%A8%E7%BD%B2miner)

Miner used to match transactions
### Web access
Currently accessing external web requests via the aws ALB as the pre-service of the relay-cluster

Deploy ALB

### Auxiliary management system
* [kafka-manager](https://github.com/Loopring/relay-cluster/wiki/%E9%83%A8%E7%BD%B2kafka-manager)

This is an open-source kafka cluster browsing and management system

* [node-zookeeper-browser](https://github.com/Loopring/relay-cluster/wiki/%E9%83%A8%E7%BD%B2node-zk-browser)

Provides a web interface to assist in viewing and editing kafka for easy system administration

### Monitoring and alerting
* cloudwatch

Cloudwatch can report indicators and configure related rules for alarms
* sns

Sns is an aws notification service. You can access the SNS API for direct system notifications. The notifications include an SMS and an email. It is convenient to insert this notification service into the key business logic of the system, so that the operation and maintenance personnel can find the fault point in time.
<!--stackedit_data:
eyJoaXN0b3J5IjpbLTE0NTcwNjYzMDcsLTU2MTcyNzQxOSwxNT
A4MzUyNDE0XX0=
-->