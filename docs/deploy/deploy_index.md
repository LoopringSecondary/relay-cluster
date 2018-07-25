# Deployment
> The relay-cluster related services are deployed in aws, and some functions strongly depend on related services provided by aws. If other cloud service providers are used, some of the functions may be unavailable or some unexpected errors may occur.

> The relay-cluster and its dependent extractor services need to be deployed through the cluster to avoid all points of failure. Although only a single node can be deployed, the normal startup of a single node still needs to rely on the components that make up the cluster, including zookeeper and kafka.

Follow the order of dependencies

## Universal
Before deploying any service, select the appropriate EC2 instance to start and the subsequent services will be deployed on this instance. After the instance is started, you also need to apply customized security groups to allow specific ports to communicate between instances.

* [Start EC2 instance](new_ec2.md)

Start EC2 instance by ec2 wizard, you can customize parameters to satisfy the requirement based on the module to be deployed on this instance.

* [Aws security group](security_group.md)

Security group will refuse connection request on not permitted port, thus will improve instance security. We will describe all needed configurations in this doc.

## Storage and Communication
* [ethnode](deploy_geth.md)

Relay-cluster will communicate with go-ethereum nodes to access eth network.

* [mysql](deploy_mysql.md)

This is the main backend store of the relay-cluster.

* [redis](deploy_redis.md)

Mainly used to increase the request speed for API, or store non-critical data.

* [zookeeper](deploy_zookeeper.md)

Used for system configuration management and kafka metadata storage.

* [kafka](deploy_kafka.md)

Kafka implements asynchronous communication between services to facilitate system decoupling and expansion

## Service
* [Access CodeDeploy](codedeploy.md)

At present, the relevant components of the relay-cluster are deployed through aws CodeDeploy, you will need configure CodeDeploy before any service deployment.

* [relay-cluster](deploy_relay_cluster.md)

This is the core component of the backend service. It provides external jsonRpc interface. It is an access service for the Loopring wallet and transaction system, and also provides a motan-rpc interface to the miner.

* [extractor](deploy_extractor.md)

The service analyzes the eth network transaction and synchronizes the result with the relay-cluster via kafka

* [miner](deploy_miner.md)

Used to match orders as rings, and submit them to eth network.

## Web access
Web access is supported by aws ALB, which work as a proxy for relay-cluster rpc API and ethnode API.

* [ALB](deploy_alb.md)

## Auxiliary management system(optional)
* [kafka-manager](deploy_kafka_manager.md)

This is an open-source kafka cluster browsing and management system

* [node-zookeeper-browser](deploy_zk_browser.md)

Provides a web interface to assist in viewing and editing kafka for easy system administration

## Monitoring and alarm(optional)
* [cloudwatch](cloudwatch.md)

Cloudwatch can report metrics and configure related rules for alarms
* [SNS](sns.md)

Sns is aws notification service. You can access the SNS service by API for trigger notifications, notification type includes email, SMS and so on. It is convenient to enable this abilibity for key logic monitor or system failure alarm.