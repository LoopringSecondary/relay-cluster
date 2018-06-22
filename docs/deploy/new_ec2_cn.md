> 如果还没有注册aws账户，请先注册账户
### 启动新实例
登录 aws控制台 [http://aws.amazon.com](http://aws.amazon.com)，选择【实例-实例-启动实例】

在【步骤 1: 选择一个 Amazon 系统映像(AMI)】页面，选择ubuntu实例类型，因为后续所有部署都是基于ubuntu实例进行操作的。建议使用通用实例`Ubuntu Server 16.04 LTS (HVM), SSD Volume Type`，然后点击【选择】

在【步骤 2: 选择一个实例类型】页面选择合适的类型，如果做实验，可以选择免费的实例。

在【步骤 3: 配置实例详细信息】页面，【网络】选择默认VPC即可。如果有多个实例，建议分批创建，然后每批选择不同的【子网】，可以避免单aws机房故障导致服务不可用。对于IAM角色，可以不设置，后续根据需要进行配置。其他选择默认即可
> 如果需要支持通过CodeDeploy在该实例部署服务，这里需要额外配置IAM角色和初始化脚本，参考[支持codedeploy](https://github.com/Loopring/relay-cluster/wiki/%E5%90%AF%E5%8A%A8aws-EC2%E5%AE%9E%E4%BE%8B#%E6%94%AF%E6%8C%81codedeploy)

在【步骤 4: 添加存储】页面，磁盘大小建议选择大于20G，【卷类型】为【默认通用SSD】

在【步骤 5: 添加标签】，跳过，后续可以手动配置

在【步骤 6: 配置安全组】，选择【选择一个现有的安全组】，选择名称为【launch-wizard-1】的安全组，也就是只开通ssh端口22的配置，后续根据需要在进行调整，点击【审核和启动】

在【步骤 7: 核查实例启动】，确认相关配置正确，点击【启动】，在启动弹出框中，会提示选择登录私钥，如果没有则选择创建新的，否则选择现有的即可。如果创建秘钥对，下载新秘钥然后进行保存，后续所有ssh操作都依赖这个绑定的xx.pem私钥文件，点击【启动实例】

点击【查看实例】进入EC2实例列表页面，能看到实例的启动状态

### 登录实例
修改xx.pem 的权限 `chmod 400 xx.pem`

对于【实例-实例-实例状态】为running的实例，执行`ssh -i xx.pem ubuntu@ipv4_add` 来登录实例，这里ipv4_add是该实例的【IPv4 公有 IP】

### 支持CodeDeploy
IAM角色选择`CodeDeployEc2InstanceProfile`，高级详细信息输入一下文本
```
#!/bin/bash
apt-get -y update
apt-get -y install ruby
apt-get -y install wget
cd /home/ubuntu
wget https://aws-codedeploy-ap-northeast-1.s3.amazonaws.com/latest/install
chmod +x ./install
./install auto
service codedeploy-agent stop
service codedeploy-agent start
```