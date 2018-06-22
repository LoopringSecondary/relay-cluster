node-zk-browser提供遍历和修改zookeeper数据的功能，方便后续查看和修改基于zk的配置

## 配置环境
### 部署依赖
```
sudo add-apt-repository ppa:fkrull/deadsnakes
sudo apt-get update
sudo apt-get install python2.6 python2.6-dev -y
sudo update-alternatives --install /usr/bin/python python /usr/bin/python2.6 2
sudo apt-get install gcc g++ libffi-dev libkrb5-dev  libsasl2-dev libsasl2-modules-gssapi-mit libssl-dev libxml2-dev libxslt-dev make libldap2-dev python-dev python-setuptools libgmp3-dev npm
```

### 部署node
```
cd /opt/loopring
wget https://github.com/nodejs/node-v0.x-archive/archive/v0.12.7.tar.gz
tar xzf v0.12.7.tar.gz
cd node-v0.x-archive-0.12.7/
./configure
make
sudo make install
```

### 部署node-zk-browser
```
cd /opt/loopring
git clone https://github.com/killme2008/node-zk-browser.git
cd node-zk-browser
npm install -d
```

## 启停
### 启动
编辑启动脚本，设置连接的zk节点

`vim start.sh`

修改下面的配置项为正确的ip和端口，多个zk节点使用逗号分隔
```
export ZK_HOST="xx.xx.xx.xx:2181"
```
启动
```
./start.sh

```

### 终止
```
pkill -f "node ./app.js"
```

## 日志
`/opt/loopring/node-zk-browser/logs`


## web界面
### 打开端口
【EC2/网络与安全/安全组】新建名称为zookeeperBrowser-SecurityGroup 的安全组，入栈添加规则

```
协议 TCP
端口范围 3000
来源 ssh登录的client ip地址添加后缀[/32]
```
【EC2/实例/实例】选择部署browser的节点，【操作/联网/更改安全组】，附加新建的安全组 zookeeperBrowser-SecurityGroup


### 访问
【EC2/实例/实例】找到【IPv4 公有 IP】，浏览器访问x.x.x.x:3000

如果需要编辑，则点击【SignIn】登录，用户名口令查看配置文件 `/opt/loopring/node-zk-browser/user.json`