# Deploy miner

## Initialization environment

### Start EC2 instance
Start the EC2 instance and add support for CodeDeploy during the startup process, and refer to: [Start EC2 instance](new_ec2.md)

### Configure the security group
Apply the `miner-SecurityGroup` security group. If the security group has not been created, create it first, please refer to the [aws security group](security_group.md). 

## Deployment of the configuration file

The basic miner configuration is implemented through static configuration files. Therefore, you need to configure the configuration file locally and upload all the servers that need to be deployed. This task is only necessary during the first time you deploy it, and the subsequent static configuration files will be used. Start the service [to be optimized].

### Create a configuration file
* miner.toml

Make the following necessary modifications based on `Loopring/miner/config/miner.toml`
```
    output_paths = ["/var/log/miner/zap.log"]
    error_output_paths = ["/var/log/miner/err.log"]
...
[mysql]
    hostname = "xx.xx.xx.xx"
    port = "3306"
    user = "xxx"
    password = "xxx"
...
[redis]
    host = "xx.xx.xx.xx"
    port = "6379"
#The following is the ip node's intranet IP address
[accessor]
    raw_urls = ["http://xx.xx.xx.xx:8545", "http://xx.xx.xx.xx:8545"]
#The following is the eth main network contract configuration, if it is not the current main network, please contact open-source personnel for the latest updated test configuration
[loopring_accessor.address]
    "v1.5" = "0x8d8812b72d1e4ffCeC158D25f56748b7d67c1e78"
...
[miner]
    ....
    feeReceipt = "0x111111111111111111111111111111"
    [[miner.normal_miners]]
        address = "0x111111111111111111111111111111"
...
[keystore]
    keydir = "/opt/loopring/miner/config/keystore"
...
[market_cap]
        base_url = "https://api.coinmarketcap.com/v2/ticker/?convert=%s&start=%d&limit=%d"
        currency = "CNY"
...
[market_util]
    token_file = "/opt/loopring/miner/config/tokens.json"
    old_version_weth_address = "0x88699e7fee2da0462981a08a15a3b940304cc516"
...
[data_source]
    type = "motan"
    [data_source.motan_client]
        client_id="miner-client"
        conf_file="/opt/loopring/miner/config/motan_client.yaml"

#zk intranet ip address
[zk_lock]
    zk_servers = "xx.xx.xx.xx:2181,xx.xx.xx.xx:2181,xx.xx.xx.xx:2181"
...
#kafka intranet ip address
[kafka]
    brokers = ["xx.xx.xx.xx:9092","xx.xx.xx.xx:9092","xx.xx.xx.xx:9092"]

[cloudwatch]
    enabled = false
    region = ""
```

> If `cloudwatch` segment's config `enabled` is set to true, please refer to: [deploy credentials file](new_ec2.md#deploy-credentials-file) to deploy the authentication file. For the value of the region, please refer to: [aws doc](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html)

* motan_client.yaml

Make the following necessary modifications based on `Loopring/miner/config/motan_client.yaml`
```
log_dir: "/var/log/miner"
...
#Set the zookeeper intranet IP address
  zk-registry:
    protocol: zookeeper
    host: xx.xx.xx.xx,xx.xx.xx.xx,xx.xx.xx.xx
    port: 2181
```
* tokens.json

According to you needs, make the necessary modifications based on: [tokens.json](tokens_main.md)

### Configure EC2 instances
* Deployment of the configuration file

Execute script on EC2 instance
```
sudo mkdir -p /opt/loopring/miner
sudo chown -R ubuntu:ubuntu /opt/loopring
cd /opt/loopring/miner 
mkdir bin/ config/ src/
```
Upload local configuration file
```
scp -i xx.pem miner.toml ubuntu@x.x.x.x:/opt/loopring/miner/config
scp -i xx.pem motan_client.yaml ubuntu@x.x.x.x:/opt/loopring/miner/config
scp -i xx.pem tokens.json ubuntu@x.x.x.x:/opt/loopring/miner/config
```
* Deploy the keystore

Copy the eth address corresponding to the miner's fee to the keystore file to the directory `/opt/loopring/miner/config/keystore`

### Deploy the deamontools configuration

Unlike the other two services, since the miner startup script contains local parameters, it is not possible to place each overlaying deployment in an automatic startup script. You need to manually configure the startup script before your first deployment.

Create the temp directory by executing the following script on the EC2 instance
```
mkdir -p /tmp/svc/log
```
Modify svc/run based on `Loopring/miner/bin/svc/run`
```
#Modify the unlocks to accept the address for the miner's fee. The password is the corresponding password for the address. The address here should be the same as the keystore address configured above.
exec setuidgid ubuntu $WORK_DIR/bin/miner --unlocks=0x1111111111111111111111111111 --passwords xxxx --config $WORK_DIR/config/miner.toml 2>&1
```
Upload the configuration script
```
scp -i xx.pem svc/run ubuntu@x.x.x.x:/tmp/svc
scp -i xx.pem svc/log/run ubuntu@x.x.x.x:/tmp/svc/log
```
Deploy the configuration file
```
sudo mkdir -p /etc/service/miner
sudo cp -rf /tmp/svc/* /etc/service/miner
sudo chmod -R 755 /etc/service/miner
rm -rf /tmp/svc
```

## Deployment
The configuration is via CodeDeploy, for detailed step-by-step instructions, reference: [Access CodeDeloy](codedeploy.md)

## Service log

## Miner service log
`/var/log/miner/zap.log`

## Motan-rpc log
`/var/log/miner/miner.INFO`

## stdout
`/var/log/svc/miner/current`

## Start and Termination
Deploy via CodeDeploy your service will by supervised by daemontools, in case it down accidentally, daemontools will restart it automatically. This also result that, you can't kill the process by simplly kill it.

### Start up
`sudo svc -u /etc/service/miner`

### Termination
`sudo svc -d /etc/service/miner`