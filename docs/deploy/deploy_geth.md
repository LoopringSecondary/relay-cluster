# Deploy ethnode
We chose to deploy the official go-ethereum (geth) version of the eth service as an entrance to the eth network.

## Start a new instance
We recommend you deploy more than two eth nodes to avoid failures or having the service unavailable due to node delay, operation issues, and maintenance reasons.

Due to the need to deploy the full node of the geth, it will consume relatively more resources, and you need to select the instance with the 8core 32g and above configuration. Please also refer to: [EC2 example](new_ec2.md)

The `ethnode-SecurityGroup` security group needs to be connected after the instance is started. If the security group has not been created, please refer to [aws security group](security_group.md) for the description of the `ethnode-SecurityGroup` security group, and then create the connection again.

## Deployment
For the specific deployment operations, refer to eth official documents: [go-ethereum for ubuntu](https://github.com/ethereum/go-ethereum/wiki/Installation-Instructions-for-Ubuntu)

## Start up
```
sudo mdkir -p /data/ethereum
sudo chown -R ubuntu:ubuntu /data/ethereum
#Start the script and please refer to the following configuration, where the ip address is the local network address
geth --datadir /data/ethereum --fast --cache=1024 --rpc --rpcaddr x.x.x.x --rpcport 8545 --rpccorsdomain *
```