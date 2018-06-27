# Deploying a redis cluster

Redis is the cache of the relay-cluster backend service and provides storage of non-critical business data

## Select the redis instance type

> Currently, redis is deployed in non-cluster mode and will be upgraded to cluster mode in the future.

For ElastiCache you can choose to use aws or self-built redis, it is recommended to use redis

ElastiCache includes clustering functions to facilitate elastic scaling and provide richer monitoring, and management capabilities for online environment usage

Self-built single instance redis is faster and more suitable for test scenarios

## Create an ElastiCache instance

Finding `ElastiCache` from the list of services to find the entrance

### Create parameter group

Since the default parameter group cannot be modified, we recommend you create a new parameter group

Select [Cache Parameter Group], click [Create Cache Parameter Group], for [Series] select `redis3.2`, for the name use `loopring-relay-params`, click [Create], and find the new `loopring-relay-params` in the parameter group list. Then edit and modify the following configuration items:
```
cluster-enabled no
slow-log-slower-than 1000
slowlog-max-len 1000
```
The specific meaning of the parameters can be found at: [aws doc](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/ParameterGroups.Redis.html)

### Create an ElastiCache instance
Select [Redis] function, then click [Start Cache Cluster]

* Cluster engine

Select Redis

* Redis settings

For [Name] enter something similar to the relay-cache or miner-cache, for [Engine version compatibility] select 3.2.10, the default port is 6379, for the parameter group select the new `loopring-relay-params`, lastly, select the appropriate node type, select the number of copies, we recommend it to be at least 1.

* Advanced Redis settings

If the number of copies is not less than 1, check [Multi-AZ with automatic failover], for [VPC ID] choose the default, for [Subnet] what you select should be the same subnet as the redis service, and then select another subnet to use for the Deployment copy.

[Preferred Availability Zone], select [Select Availability Zone], and then specify the distribution of primary replicas and other replicas.

* safety

For [Security Group] select the security group named `redis-SecurityGroup`. If it has not yet been created, please refer to [configure aws security group](security_group.md) to configure the description of `redis-SecurityGroup` section before connecting.

* Back up

It is recommended to enable backup and set the backup window

* Maintenance

It is recommended to set a maintenance window to avoid business peak change operations

Finally click [Create] to create a redis cluster

### Create a standalone Redis instance
Refer to; [Start aws EC2 instance](new_ec2.md), start the instance, and connect the `redis-securityGroup` security group

Execute the following script to deploy the redis instance

`sudo apt install redis-server`

* start up

`sudo systemctl start redis`

* Termination

`sudo systemctl stop redis`

## Connect redis

The default port is 6379