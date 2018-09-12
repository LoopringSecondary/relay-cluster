# Loopring Extractor Docker 中文文档

loopring开发团队提供loopring/extractor镜像,最新版本是v1.5.1<br>

## 部署
* 获取docker镜像
```bash
docker pull loopring/extractor
```
* 创建log&config目录
```bash
mkdir your_log_path your_config_path
```
* 配置extractor.toml文件，[参考](https://github.com/Loopring/relay-cluster/wiki/%E9%83%A8%E7%BD%B2extractor#%E9%83%A8%E7%BD%B2%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)
* telnet测试mysql,redis,zk,kafka,ethereum rpc相关端口能否连接

* 运行
运行时需要挂载logs&config目录,并指定config文件
```bash
docker run --name extractor -idt -v your_log_path:/opt/loopring/extractor/logs -v your_config_path:/opt/loopring/extractor/config loopring/extractor:latest --config=/opt/loopring/extractor/config/extractor.toml /bin/bash
```

## 历史版本

| 版本号         | 描述         |
|--------------|------------|
| v1.5.0| release初始版本|
| v1.5.1| 支持relay钱包用户自定义代币, 变更aws配置|


## 注意事项:
使用桥接网络本地调试时, 需要保障extractor依赖的相关服务能正常访问:
* mysql: 允许容器访问宿主mysql,登录mysql-cli
```bash
GRANT ALL PRIVILEGES ON *.* TO 'username'@'%' IDENTIFIED BY 'yourpassword' WITH GRANT OPTION;
```

* redis: 允许容器访问宿主redis，找到redis.conf注释掉
```bash
# bind 127.0.0.1
```
另外，关闭保护模式, 在redis.conf中设置:
```bash
protected-mode no
```
或者登录redis-cli，运行
```bash
CONFIG SET protected-mode no
```
或者使用密码访问redis

* ethereum-node: 允许容器访问宿主eth节点:<br>
运行geth等客户端时:<br>
--rpc(打开rpc服务)<br>
--rpcaddr(指定rpc监听地址, 默认localhost)

* zookeeper-kafka: 允许容器访问zk-kafka相关端口
找到server.properties文件,修改ip地址
```bash
zookeeper.connect=your_ip_addr:2181
advertised.listeners=PLAINTEXT://your_ip_addr:9092
```
