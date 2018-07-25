# Loopring Relay-cluster Docker

The loopring development team provides the loopring/relay-cluster image. The latest version is v1.5.0

## Run
get the latest docker image
``` 
docker pull loopring/relay-cluster
```
create log&config dir
```bash
mkdir your_log_path your_config_path
```
config relay.toml，[reference](https://github.com/Loopring/relay-cluster/wiki/%E9%83%A8%E7%BD%B2relay-cluster#%E9%83%A8%E7%BD%B2%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)

before deployment, perform telnet tests according to the ports related to the configuration files mysql, redis, kafka, and zk to ensure that these dependencies can be accessed normally.

mount the log dir and config dir, and run
```bash
docker run --name relay -idt -v your_log_path:/opt/loopring/relay/logs -v your_config_path:/opt/loopring/relay/config loopring/relay-cluster:latest --config=/opt/loopring/relay/config/relay.toml /bin/bash
```

## History version

| version         | desc         |
|--------------|------------|
| v1.5.0| the first release version|
