Node-zk-browser provides functions for traversing and modifying zookeeper data for subsequent viewing and modification of zk-based configurations

## Configuration Ecosystem
### Deployment dependencies
```
sudo add-apt-repository ppa:fkrull/deadsnakes
sudo apt-get update
sudo apt-get install python2.6 python2.6-dev -y
sudo update-alternatives --install /usr/bin/python python /usr/bin/python2.6 2
sudo apt-get install gcc g++ libffi-dev libkrb5-dev  libsasl2-dev libsasl2-modules-gssapi-mit libssl-dev libxml2-dev libxslt-dev make libldap2-dev python-dev python-setuptools libgmp3-dev npm
```

### Deploy the node
```
cd /opt/loopring
wget https://github.com/nodejs/node-v0.x-archive/archive/v0.12.7.tar.gz
tar xzf v0.12.7.tar.gz
cd node-v0.x-archive-0.12.7/
./configure
make
sudo make install
```

### Deploy node-zk-browser
```
cd /opt/loopring
git clone https://github.com/killme2008/node-zk-browser.git
cd node-zk-browser
npm install -d
```

## Start and Termination
### Start up
Edit the startup script to set up the connected zk node

`vim start.sh`

Modify the following configuration items for the correct ip and port, separated by commas for multiple zk nodes
```
export ZK_HOST="xx.xx.xx.xx:2181"
```
Final Start up step
```
./start.sh

```

### Termination
```
pkill -f "node ./app.js"
```

## Logs
`/opt/loopring/node-zk-browser/logs`


## Web interface
### Open port
[EC2/Network and Security/Security Group] Create a new security group named zookeeperBrowser-SecurityGroup and add rules to the stack.

```
protocol TCP
Port range 3000
Source ssh login suffix for client ip address[/32]
```
[EC2/Instance/Instance] Select the node to deploy the browser, [Operation/Networking/Change Security Group], add the new security group zookeeperBrowser-SecurityGroup


### Access
[EC2/Instance/Instance] Find [IPv4 Public IP], Browser Accesses x.x.x.x:3000

If you need to edit, click on [SignIn] to log in, then type in your user name and password to view the configuration file `/opt/loopring/node-zk-browser/user.json`
<!--stackedit_data:
eyJoaXN0b3J5IjpbMTc0NjI0MzgwOCwxNzc0Mjk2OTddfQ==
-->