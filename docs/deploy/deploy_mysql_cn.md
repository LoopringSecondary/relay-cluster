# 部署mysql集群

mysql是relay-cluster后端服务的主要存储

## 创建MySQL实例
生产环境下推荐aws的RDS，优点是包含更丰富的监控及管理功能，扩展也更加方便，而测试环境仅需自建一个mysql单实例即可

> relay和miner都会用到mysql数据库，生产环境建议创建不同的数据库实例，避免相互影响

### 创建RDS实例（生产环境）
从服务列表查找`RDS`找到入口，然后选择【立即开始使用】

步骤1，【选择引擎】，点选【MySql】，点击【下一步】

步骤2，根据自己的场景，选择【生产-MySQL】，点击【下一步】

步骤3
* 实例规格

【数据库引擎版本】选择5.7以上的新版本，【数据库实例类型】建议选择【db.m4.xlarge】及以上版本主机，【多区域可用部署】会提供更高的可用性，根据需要确认是否选择。【存储类型】选择【通用型(SSD)】，分配存储空间建议50G以上

* 设置

【数据库实例标识符】输入类似loopring-relay，relay-miner，并相应输入【主用户名】，以及对应的密码

步骤4
* 网络与安全

【Virtual Private Cloud】选择默认，【子网组】选择default，【公开可用性】根据实际情况选择，【可用区】选择和后续会访问该MySQL库的服务所在的分区，VPC安全组，选择`mysql-securityGroup`，如果还没有创建请参考[配置aws安全组](security_group_cn.md)关于`mysql-securityGroup`创建完成后，在回来选择

* 数据库选项

【数据库名称】输入类似relay，miner的名称，【端口】使用默认的3306，【数据库参数组】和【选项组】选择默认，【IAM 数据库身份验证】根据实际情况选择，【加密】选择【启用加密】
> 【数据库参数组】建议新建，可以参考后面的，因为自建参数组可以修改，而默认的是不能修改的。通过修改参数组的参数可以很方便的对数据库进行配置，而多数配置是不需要对数据进行重启的

* 备份

建议备份，并且设置合适的开始时间，需要注意的是这里是UTC时间，北京时间需要+8个小时

* 监控

建议启动增强监控

* 日志导出

建议勾选【错误日志】，【常规】，【慢查询日志】

* 维护

禁用【自动次要版本升级】，和备份类似选择合适的维护窗口，选择启动数据库实例

### 创建MySQL单实例（测试环境）
参考[启动aws EC2实例](new_ec2_cn.md)，启动实例，并且关联`mysql-securityGroup`安全组

> 测试环境下mysql和redis可部署到同一台实例，再关联`mysql-securityGroup`和`redis-securityGroup`两个安全组即可


执行以下命令部署Mysql实例
```
sudo apt update
sudo apt -y install mysql-server
```
根据界面提示输入root用户口令

创建relay db，开启root远程访问
```
mysql --host=localhost --port=3306 --user=root -p
CREATE DATABASE relay;
use mysql;
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' IDENTIFIED BY '填该账号的密码' WITH GRANT OPTION;
FLUSH PRIVILEGES;
```

取消mysql ip绑定

`sudo vim /etc/mysql/mysql.conf.d/mysqld.cnf`

注释这句 `bind-address= 127.0.0.1`

重启mysql

`sudo systemctl restart mysql`

## 连接数据库

记录前面创建db时指定的用户口令，在相关配置文件中配置即可，同时也可通过上面的命令行工具访问数据库实例
