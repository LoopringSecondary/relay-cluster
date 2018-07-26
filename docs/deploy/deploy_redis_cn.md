# 部署redis集群

redis是relay-cluster后端服务的缓存，并提供非关键业务数据的存储

## 选择redis实例类型

> 目前redis为非cluster方式部署，后续会升级为cluster模式

测试场景下仅需自建redis单实例即可，简便快捷

生产场景下推荐采用aws的ElastiCache，其包含集群功能，方便进行弹性伸缩，并且提供更丰富的监控及管理功能，适合线上环境使用


## 创建ElastiCache实例（生产场境）

从服务列表查找`ElastiCache`找到入口

### 创建参数组

因为默认参数组无法进行修改，所以建议创建新的参数组

选择【缓存参数组】，点击【创建缓存参数组】，【系列】选择`redis3.2`，名称使用`loopring-relay-params`，点击【创建】，在参数组列表中找到新建的`loopring-relay-params`，然后进行编辑，修改以下配置项
```
cluster-enabled no
slow-log-slower-than 1000
slowlog-max-len 1000
```
参数具体含义可参考[aws doc](https://docs.aws.amazon.com/zh_cn/AmazonElastiCache/latest/red-ug/ParameterGroups.Redis.html)

### 创建ElastiCache实例
选择【Redis】功能，点击【启动缓存集群】

* 集群引擎

选择Redis

* Redis 设置

【名称】输入类似relay-cache，miner-cache，【引擎版本兼容性】选择3.2.10，端口默认6379，参数组选择新建的`loopring-relay-params`，选择合适的节点类型，副本数选择建议至少为1，

* 高级 Redis 设置

如果副本数不小于1，则勾选【具有自动故障转移功能的多可用区】，【VPC ID】默认，【子网】选择和依赖redis服务相同的子网，然后选择另外一个子网用来部署副本

【首选可用区】，选择【选择可用区】，然后指定主副本和其他副本的分布

* 安全性

【安全组】选择名称为`redis-SecurityGroup`的安全组，如果还没有创建，请参考[配置aws安全组](security_group_cn.md)关于`redis-SecurityGroup`部分的说明进行配置后再进行关联

* 备份

建议启用备份并设置备份窗口

* 维护

建议设置维护窗口，避免业务高峰执行变更操作

最后点击【创建】来创建redis集群

### 创建Redis单实例（测试场境）
申请1台EC2实例，参考启动aws [EC2实例](new_ec2_cn.md)，并且关联`redis-securityGroup`安全组

> 测试场景以简便快捷为主，因此mysql和redis可部署到同一台实例，再同时关联`mysql-securityGroup`和`redis-securityGroup`这两个安全组即可


执行以下命令部署redis实例
```
sudo apt update
sudo apt -y install redis-server
```

修改配置项，并禁用redis bind

`sudo vim /etc/redis/redis.conf`

```
cluster-enabled no
slowlog-log-slower-than 1000
slowlog-max-len 1000
# bind 127.0.0.1
```

最后重启redis即可 `sudo systemctl restart redis`

#### 启停

* ##### 启动

`sudo systemctl start redis`

* ##### 终止

`sudo systemctl stop redis`

* ##### 重启

`sudo systemctl restart redis`


## 连接redis

默认端口 6379
