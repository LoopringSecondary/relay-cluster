# 部署ethnode
部署官方的PPA(geth)版本，作为接入eth网络的入口

## 启动新实例
建议部署两台以上eth节点，避免因单台故障、节点延时或运维不当等造成的服务不可用

因需要部署eth全节点，相对比较耗费资源，所以推荐最低配置为8核_32g内存_300GB SSD以上的实例，参考[EC2实例](new_ec2_cn.md)


ethnode实例需要关联`ethnode-SecurityGroup`安全组
> 若未创建该安全组，请务必参考[aws安全组](security_group_cn.md)关于`ethnode-SecurityGroup`安全组的说明，创建后再关联

## 部署
```
sudo apt-get -y install software-properties-common
sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get -y install ethereum
sudo mkdir -p /data/ethereum
sudo chown -R ubuntu:ubuntu /data/ethereum
```

## 启动

# 启动脚本请参考如下配置，x.x.x.x为本机内网ip地址

```
sudo geth --datadir /data/ethereum --fast --cache=1024 --rpc --rpcaddr x.x.x.x --rpcport 8545 --rpccorsdomain *
```
