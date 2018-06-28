# Deploy relay-cluster

## Initialization environment

### Start EC2 instance
Start the EC2 instance and add support for CodeDeploy during the startup process. [Start EC2 instance](new_ec2.md)

### Configure the security group
Apply a security group called `relayCluster-SecurityGroup` with each instance. If it hasn't been created, create it first, please refer to: [aws security group](security_group.md) 

### Deployment configuration file
The basic configuration of the current relay is implemented through a static configuration file. Therefore, you need to configure the file locally, and upload all the servers that need to be deployed. However, this task is only necessary during the first deployment, and the subsequent configuration will use this static configuration file to start the service [to be optimized].

#### Create a configuration file

* relay.toml

Make the following necessary modifications based on `Loopring/relay-cluster/config/relay.toml`
```
    output_paths = ["/var/log/relay/zap.log"]
    error_output_paths = ["/var/log/relay/err.log"]
...
[redis]
    host = "xx.xx.xx.xx"
    port = "6379"
...
[mysql]
    hostname = "xx.xx.xx.xx"
    port = "3306"
    user = "xxx"
    password = "xxx"
#The following is the ip node's intranet IP address
[accessor]
    raw_urls = ["http://xx.xx.xx.xx:8545", "http://xx.xx.xx.xx:8545"]
#The following is the eth main network contract configuration, if it is not the current main network, contact open-source personnel for the latest test configuration
[loopring_protocol]
        implAbi = "[{\"constant\":true,\"inputs\":[],\"name\":\"MARGIN_SPLIT_PERCENTAGE_BASE\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ringIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"RATE_RATIO_SCALE\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lrcTokenAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"tokenRegistryAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"delegateAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"orderOwner\",\"type\":\"address\"},{\"name\":\"token1\",\"type\":\"address\"},{\"name\":\"token2\",\"type\":\"address\"}],\"name\":\"getTradingPairCutoffs\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token1\",\"type\":\"address\"},{\"name\":\"token2\",\"type\":\"address\"},{\"name\":\"cutoff\",\"type\":\"uint256\"}],\"name\":\"cancelAllOrdersByTradingPair\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addresses\",\"type\":\"address[5]\"},{\"name\":\"orderValues\",\"type\":\"uint256[6]\"},{\"name\":\"buyNoMoreThanAmountB\",\"type\":\"bool\"},{\"name\":\"marginSplitPercentage\",\"type\":\"uint8\"},{\"name\":\"v\",\"type\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"cancelOrder\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_RING_SIZE\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"cutoff\",\"type\":\"uint256\"}],\"name\":\"cancelAllOrders\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"rateRatioCVSThreshold\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addressList\",\"type\":\"address[4][]\"},{\"name\":\"uintArgsList\",\"type\":\"uint256[6][]\"},{\"name\":\"uint8ArgsList\",\"type\":\"uint8[1][]\"},{\"name\":\"buyNoMoreThanAmountBList\",\"type\":\"bool[]\"},{\"name\":\"vList\",\"type\":\"uint8[]\"},{\"name\":\"rList\",\"type\":\"bytes32[]\"},{\"name\":\"sList\",\"type\":\"bytes32[]\"},{\"name\":\"feeRecipient\",\"type\":\"address\"},{\"name\":\"feeSelections\",\"type\":\"uint16\"}],\"name\":\"submitRing\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"walletSplitPercentage\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_ringIndex\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"_ringHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_miner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_feeRecipient\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_orderInfoList\",\"type\":\"bytes32[]\"}],\"name\":\"RingMined\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_orderHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_amountCancelled\",\"type\":\"uint256\"}],\"name\":\"OrderCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_address\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_cutoff\",\"type\":\"uint256\"}],\"name\":\"AllOrdersCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_address\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_token1\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_token2\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_cutoff\",\"type\":\"uint256\"}],\"name\":\"OrdersCancelled\",\"type\":\"event\"}]"
        delegateAbi = "[{\"constant\":true,\"inputs\":[{\"name\":\"owners\",\"type\":\"address[]\"},{\"name\":\"tradingPairs\",\"type\":\"bytes20[]\"},{\"name\":\"validSince\",\"type\":\"uint256[]\"}],\"name\":\"checkCutoffsBatch\",\"otputs\":[],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"resume\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"max\",\"type\":\"uint256\"}],\"name\":\"getLatestAuthorizedAddresses\",\"outputs\":[{\"name\":\"addresses\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"orderHash\",\"type\":\"bytes32\"},{\"name\":\"cancelOrFillAmount\",\"type\":\"uint256\"}],\"name\":\"addCancelledOrFilled\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"cancelled\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"kill\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"lrcTokenAddress\",\"type\":\"address\"},{\"name\":\"miner\",\"type\":\"address\"},{\"name\":\"feeRecipient\",\"type\":\"address\"},{\"name\":\"walletSplitPercentage\",\"type\":\"uint8\"},{\"name\":\"batch\",\"type\":\"bytes32[]\"}],\"name\":\"batchTransferToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"authorizeAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"tokenPair\",\"type\":\"bytes20\"},{\"name\":\"t\",\"type\":\"uint256\"}],\"name\":\"setTradingPairCutoffs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"claimOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"cancelledOrFilled\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"suspended\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"batch\",\"type\":\"bytes32[]\"}],\"name\":\"batchAddCancelledOrFilled\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"bytes20\"}],\"name\":\"tradingPairCutoffs\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"orderHash\",\"type\":\"bytes32\"},{\"name\":\"cancelAmount\",\"type\":\"uint256\"}],\"name\":\"addCancelled\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addressInfos\",\"outputs\":[{\"name\":\"previous\",\"type\":\"address\"},{\"name\":\"index\",\"type\":\"uint32\"},{\"name\":\"authorized\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isAddressAuthorized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"cutoffs\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"suspend\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"deauthorizeAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"t\",\"type\":\"uint256\"}],\"name\":\"setCutoffs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"number\",\"type\":\"uint32\"}],\"name\":\"AddressAuthorized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"number\",\"type\":\"uint32\"}],\"name\":\"AddressDeauthorized\",\"type\":\"event\"}]"
        [loopring_protocol.address]
         "v1.5" = "0x8d8812b72d1e4ffCeC158D25f56748b7d67c1e78"
...
[market_cap]
        base_url = "https://api.coinmarketcap.com/v2/ticker/?convert=%s&start=%d&limit=%d"
        currency = "CNY"
[market]
    token_file = "/opt/loopring/relay/config/tokens.json"
    old_version_weth_address = "0x2956356cd2a2bf3202f771f50d3d14a367b48070"
#Zk intranet IP address
[zk_lock]
    zk_servers = "xx.xx.xx.xx:2181,xx.xx.xx.xx:2181,xx.xx.xx.xx:2181"
#If you have enabled the function to send SNS messages, you need to configure it. For details, see the Monitoring and Alarming section.
[sns]
    enabled = false
    region = ""
    sns_topic_arn = "arn:aws:sns:region-code:xxxx:RelayNotification"
#Kafka intranet IP address
[kafka]
    brokers = ["xx.xx.xx.xx:9092","xx.xx.xx.xx:9092","xx.xx.xx.xx:9092"]
[motan_server]
    conf_file = "/opt/loopring/relay/config/motan_server.yaml"
[my_token]
    app_id = "xxx"
    app_secret = "xxxx"
    base_url = "https://open.api.mytoken.io/"

[cloudwatch]
    enabled = false
    region = ""
```

> If `cloudwatch` or `sns` segments' config `enabled` is set to true, please refer to: [deploy credentials file](new_ec2.md#deploy-credentials-file) to deploy the authentication file. For the value of the region, please refer to: [aws doc](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html)

* motan_server.yaml

Make the following necessary modifications based on `Loopring/relay-cluster/config/motan_server.yaml`
```
log_dir: "/var/log/relay"
...
#Set the zookeeper intranet IP address
  zk-registry:
    protocol: zookeeper
    host: xx.xx.xx.xx,xx.xx.xx.xx,xx.xx.xx.xx
    port: 2181
```
* tokens.json

In [tokens.json](tokens_main.md), make the necessary cutting, based on your actual needs

#### Configure EC2 instances
Execute script on EC2 instance
```
sudo mkdir /opt/loopring/relay
sudo chown -R ubuntu:ubuntu /opt/loopring/relay
cd /opt/loopring/relay 
mkdir bin/ config/ src/
```
Upload the local configuration file
```
scp -i xx.pem relay.toml ubuntu@x.x.x.x:/opt/loopring/relay/config
scp -i xx.pem motan_server.yaml ubuntu@x.x.x.x:/opt/loopring/relay/config
scp -i xx.pem tokens.json ubuntu@x.x.x.x:/opt/loopring/relay/config
```

## Deployment
Configuration via CodeDeploy, here is a detailed step-by-step reference: [Access CodeDeloy](codedeploy.md)

## Service log
### relay-cluster service log
`/var/log/relay/zap.log`

### motan-rpc log
`/var/log/relay/relay.INFO`

### stdout
`/var/log/svc/relay/current`

## Start and Termination
Deploying through CodeDeploy will add daemontools support for the service. This means the service will start automatically if it terminates unexpectedly, so it cannot be manually terminated.

### Start up
`sudo svc -u /etc/service/relay`

### Termination
`sudo svc -d /etc/service/relay`
