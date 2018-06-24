# 接入CodeDeploy

CodeDeloy是aws提供的代码部署工具，通过在项目中添加配置文件和脚本可以添加对CodeDeploy部署的支持。CodeDeloy通过下载github的项目源码到目标服务器，然后解析项目配置文件，并在部署的不同阶段执行相应的脚本来实现代码编译和服务的启动、验证等操作


## 配置依赖

### 创建IAM对象

> 同一个aws账号，本操作只需要执行一次

需要通过创建相关IAM对象并绑定EC2实例来赋予CodeDeploy部署的权限，IAM控制台入口 [https://console.aws.amazon.com/iam/](https://console.aws.amazon.com/iam/)

#### 新建IAM用户并绑定策略
【IAM-用户】选择【添加用户】，可参考[aws doc](https://docs.aws.amazon.com/zh_cn/codedeploy/latest/userguide/getting-started-provision-user.html)

步骤1，输入用户名`CodeDeployer`，访问类型勾选【编程访问】，点击【下一步：权限】

步骤2，选择【直接附加现有策略】，选择【创建策略】，点击【JSON】tab页，粘贴下面代码块中的JSON文本，点击【Review Plicy】，输入新建策略的名称`CodeDeployPolicy`，点击【Create Policy】完成创建。回到【步骤2】的页面，点击【刷新】，并选择新创建的`CodeDeployPolicy`策略，点击【下一步：审核】

```
{
  "Version": "2012-10-17",
  "Statement" : [
    {
      "Effect" : "Allow",
      "Action" : [
        "autoscaling:*",
        "codedeploy:*",
        "ec2:*",
        "lambda:*",
        "elasticloadbalancing:*",
        "iam:AddRoleToInstanceProfile",
        "iam:CreateInstanceProfile",
        "iam:CreateRole",
        "iam:DeleteInstanceProfile",
        "iam:DeleteRole",
        "iam:DeleteRolePolicy",
        "iam:GetInstanceProfile",
        "iam:GetRole",
        "iam:GetRolePolicy",
        "iam:ListInstanceProfilesForRole",
        "iam:ListRolePolicies",
        "iam:ListRoles",
        "iam:PassRole",
        "iam:PutRolePolicy",
        "iam:RemoveRoleFromInstanceProfile", 
        "s3:*"
      ],
      "Resource" : "*"
    }    
  ]
}
```

步骤3，点击【创建用户】，完成CodeDeployer用户的创建

步骤4，确认并关闭

#### 创建角色
【IAM-角色】选择【创建角色】，可参考[aws doc](https://docs.aws.amazon.com/zh_cn/codedeploy/latest/userguide/getting-started-create-service-role.html)

步骤1，选择【AWS产品】，服务选择【CodeDeploy】，案例选择【CodeDeploy】，点击【下一步：权限】

步骤2，确认出现【AWSCodeDeployRole】，点击【下一步：审核】

步骤3，输入角色名称`CodeDeployServiceRole`，点击【创建角色】

#### 创建EC2绑定策略
【IAM-策略】，选择【创建策略】，参考[aws doc](https://docs.aws.amazon.com/zh_cn/codedeploy/latest/userguide/getting-started-provision-user.html)

步骤1，选择【JSON】，粘贴下面文本，点击【Review policy】
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "s3:Get*",
                "s3:List*"
            ],
            "Effect": "Allow",
            "Resource": "*"
        }
    ]
}
```
步骤2，名称输入`CodeDeployEc2Permission`，点击【Create Policy】

【IAM-角色】，选择【创建角色】

步骤1，选择【AWS 产品】，服务选择【EC2】，案例选择【EC2】，点击【下一步：权限】

步骤2，搜索框输入刚才创建的`CodeDeployEc2Permission`，并勾选该策略，点击【下一步：审核】

步骤3，角色名，输入`CodeDeployEc2InstanceProfile`，选择【创建角色】


### 配置EC2实例

> 同一个EC2实例，本操作只需要执行一次

> 推荐在启动EC2实例的时候进行相关配置，参考[支持CodeDeploy](https://github.com/Loopring/relay-cluster/wiki/%E5%90%AF%E5%8A%A8aws-EC2%E5%AE%9E%E4%BE%8B#%E6%94%AF%E6%8C%81codedeploy)，如果已经进行相关配置，则不需要进行下面两个部分的操作

#### 配置角色
修改现有实例的IAM角色，【EC2-实例-实例】，选择一个需要通过CodeDeploy进行部署的EC2实例，【操作-实例设置-附加/替换IAM角色】，【IAM角色】选择前面创建的`CodeDeployEc2InstanceProfile`，选择【应用】

#### 部署并启动codedeploy-agent
ssh登录EC2实例，执行下面脚本
```
#!/bin/bash
sudo apt-get -y update
sudo apt-get -y install ruby
sudo apt-get -y install wget
cd /home/ubuntu
wget https://aws-codedeploy-ap-northeast-1.s3.amazonaws.com/latest/install
chmod +x ./install
sudo ./install auto
```
验证agent正常执行，执行脚本`sudo service codedeploy-agent status`，如果响应结果是`running`说明正常，否则可以
通过以下脚本进行重启，可参考[aws doc](https://docs.aws.amazon.com/zh_cn/codedeploy/latest/userguide/codedeploy-agent-operations-install-ubuntu.html)
```
sudo service codedeploy-agent stop
sudo service codedeploy-agent start
```

#### 设置标签
设置了相同标签的实例可以进行筛选，这在CodeDeploy的部署中是必须的。

具体操作为，【EC2-实例-实例】，选择单个实例，在下方详情页的【标签】tab中，可以根据需要标签的编辑

比如这里为部署的relay-cluster实例都添加标签【relay-cluster/test】

## 配置和部署应用

### 配置CodeDeploy应用

> 同一个应用，本操作只需要执行一次

#### 创建应用程序
打开CodeDeploy控制台，选择【立即开始使用】

下面以relay-cluster为例进行演示，选择【自定义部署-跳过演练】

* 创建应用程序

【应用程序名称】输入`relay-cluster`，【计算平台】选择`EC2/本地`，【部署组名称】输入`myRelayClusterGroup`

* 部署类型

选择【就地部署】

* 环境配置

选择【Amazon Ec2实例】,在【标签组】中添加之前配置的标签【relay-cluster/test】，会自动将符合条件的实例筛选出来
* 部署配置

选择CodeDeployDefault.OneAtATime

* 服务角色

选择之前创建的CodeDeployServiceRole

其他默认后，点击【创建应用程序】

### 部署应用
选择控制台左上角下拉菜单的【部署】，选择【创建部署】

* 配置基本信息

【应用程序】选择`relay-cluster`，【部署组】`myRelayClusterGroup`，【存储库类型】选择Github，【部署配置】选择CodeDeployDefault.OneAtATime
> 如果是首次部署，需要选择CodeDeployDefault.AllAtOnce

* 连接到 GitHub

初次部署应用，需要关联github账号，再输入框中输入github的邮箱账号，会跳转到github进行鉴权

* 存储库名称

输入`Loopring/relay-cluster`

* 提交id

在github中选择需要部署的分支，通常是master，然后选择commits，找到最新的commitid，粘贴到输入框中

* 启动部署
点击【部署】按钮，启动部署

#### 查看部署状态
回到【部署】列表，可以看到当前正在进行的所有部署，可以点击查看部署详情，点击【查看所有实例】能够查看当前那些节点在部署，以及进行到那个阶段

* 最新事件

是部署的阶段标识，具体定义可以参考[aws doc](https://docs.aws.amazon.com/zh_cn/codedeploy/latest/userguide/reference-appspec-file-structure-hooks.html)
* 失败日志

对于失败的情况，可以点击【查看事件】的日志列进行查看

#### 解决部署问题
* 常见问题

参考文档 [aws doc](https://docs.aws.amazon.com/zh_cn/codedeploy/latest/userguide/troubleshooting-ec2-instances.html)

* 查看日志文件

`/opt/codedeploy-agent/deployment-root/deployment-logs`
* 确认aws agent状态正常

执行脚本`sudo service codedeploy-agent status` ，如果提示非running，请尝试执行
```
sudo service codedeploy-agent stop
sudo service codedeploy-agent start
```
如果无法恢复正常，需要重启服务

kill 相关进程 然后执行 `sudo service codedeploy-agent start`

* 确认codedeploy远程连接打开

执行 `ps axu|grep CodeDeployPlugin`，找到进程id xxx

确认与远程服务端口连接打开，执行`sudo lsof -p xxx |grep TCP`，状态如果是 ESTABLISHED 说明正常，如果是CLOSE_WAIT说明断开连接，需要重启