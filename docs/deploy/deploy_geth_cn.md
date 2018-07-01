# 部署ethnode
选择部署官方的go-ethereum(geth)版本的eth服务，作为接入eth网络的入口

## 启动新实例
建议部署两台以上eth节点，避免单台故障、节点延时、运维不当等导致的服务不可用

由于需要部署geth全节点，会相当的耗费资源，所以应选择8核_32g_300GB及以上配置的实例，可以参考[EC2实例](new_ec2_cn.md)

实例需关联`ethnode-SecurityGroup`安全组。若未创建该安全组，请务必参考[aws安全组](security_group_cn.md)关于`ethnode-SecurityGroup`安全组的说明，创建后再关联

## 部署
具体部署操作，请参考eth官方文档[go-ethereum for ubuntu](https://github.com/ethereum/go-ethereum/wiki/Installation-Instructions-for-Ubuntu)

## 启动
```
sudo mkdir -p /data/ethereum
sudo chown -R ubuntu:ubuntu /data/ethereum
#启动脚本请参考如下配置，ip为本机内网地址
sudo geth --datadir /data/ethereum --fast --cache=1024 --rpc --rpcaddr x.x.x.x --rpcport 8545 --rpccorsdomain * &
```
