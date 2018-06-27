# Deployment
> The relay-cluster related services are deployed in aws, and some functions strongly depend on related services provided by aws. If other cloud service providers are used, some of the functions may be unavailable or some unexpected errors may occur.

> The relay-cluster and its dependent extractor services need to be deployed through the cluster to avoid all points of failure. Although only a single node can be deployed, the normal startup of a single node still needs to rely on the components that make up the cluster, including zookeeper and kafka.

Follow the order of dependencies
### Universal
Before deploying any service, select the appropriate EC2 instance to start and the subsequent services will be deployed on this instance. After the instance is started, you also need to associate custom security groups to allow specific ports to communicate between instances. Specific service deployment instructions will give suggested EC2 and security group configuration

* [EC2 example](new_ec2.md)

* [Aws security group](security_group.md)
### Storage and Communication
* ethnode

Relay-cluster implements eth network access by interacting with eth node
* [mysql](deploy_mysql.md)

This is the main backend store of the relay-cluster, the storage contains orders and transactions

* [redis](deploy_redis.md)

Mainly used to increase the access speed of the system, or store non-critical data
* [zookeeper](deploy_zookeeper.md)

Used for system configuration management and kafka metadata storage
* [kafka](deploy_kafka.md)

Kafka implements asynchronous communication between services to facilitate system decoupling and expansion

### Service
* [Access CodeDeploy](codedeploy.md)

At present, the relevant components of the relay are deployed through aws CodeDeploy+github to facilitate rapid iteration.

* [relay-cluster](deploy_relay_cluster.md)

This is the core component of the backend service. It provides external jsonRpc interface. It is an access service for the Loopring wallet and transaction system, and also provides a motan-rpc interface to the miner.

* [extractor](deploy_extractor.md)

The service analyzes the eth network transaction and synchronizes the result with the relay-cluster via kafka

* [miner](deploy_miner.md)

Miner used to match transactions

### Web access
Currently accessing external web requests via the aws ALB as the pre-service of the relay-cluster

[Deploy ALB](deploy_alb.md)

### Auxiliary management system
* [kafka-manager](deploy_kafka_manager.md)

This is an open-source kafka cluster browsing and management system

* [node-zookeeper-browser](deploy_zk_browser.md)

Provides a web interface to assist in viewing and editing kafka for easy system administration

### Monitoring and alerting
* [cloudwatch](cloudwatch.md)

Cloudwatch can report indicators and configure related rules for alarms
* [SNS](sns.md)

Sns is an aws notification service. You can access the SNS API for direct system notifications. The notifications include an SMS and an email. It is convenient to insert this notification service into the key business logic of the system, so that the operation and maintenance personnel can find the fault point in time.