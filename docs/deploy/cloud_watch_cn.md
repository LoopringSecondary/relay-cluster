cloudwatch提供了完整的指标收集，图形化展示，基于规则报警的功能

# 收集指标

## 默认指标
cloudwatch 会自动收集包括 EC2，ALB，ElastiCache，RDS 等组件的基础指标，相关指标可以在相关服务所在的管理系统查看，这些指标是免费的。

这些默认指标是免费的，但是可能是会有一定的缺失，或者只是粗粒度的，可以根据实际需要配置需要的额外标准指标，更细粒度的指标，或者自定义上传指标

## 收集额外指标
对于RDS，ElastiCache可以到相关管理界面配置更丰富的指标上报

对于EC2，默认指标缺失非常关键的三个指标 memory_usage disk_usage swap_usage，  需要部署watchcloud agent来上报

### 创建CloudWatch代理所需的IAM角色
可以参考[aws doc](https://docs.aws.amazon.com/zh_cn/AmazonCloudWatch/latest/monitoring/create-iam-roles-for-cloudwatch-agent.html)

登录 AWS 管理控制台 并通过以下网址打开 IAM 控制台 https://console.aws.amazon.com/iam/。

在左侧的导航窗格中，选择角色，然后选择创建角色。

对于 Choose the service that will use this role，选择 EC2 Allows EC2 instances to call AWS services on your behalf。选择 Next: Permissions。

在策略列表中，选中 CloudWatchAgentServerPolicy 旁边的复选框。如有必要，请使用搜索框查找策略。

如果您要使用 SSM 安装或配置 CloudWatch 代理，请选中 AmazonEC2RoleforSSM 旁边的复选框。如有必要，请使用搜索框查找策略。如果要仅通过命令行启动和配置代理，则不需要此策略。

选择 Next: Review。

确认 CloudWatchAgentServerPolicy 和 (可选) AmazonEC2RoleforSSM 是否显示在 Policies 旁边。在角色名称中，键入角色的名称，例如，CloudWatchAgentServerRole。(可选) 为角色提供说明，然后选择创建角色。

将立即创建该角色。

### 配置收集指标的角色
【EC2-实例-实例-操作】，选择【操作-实力设置-附加/替换IAM角色】，将刚刚创建的CloudWatchAgentServerRole附加到实例

如果实例已经附加了别的角色，则需要修改原角色，并附加上策略 CloudWatchAgentServerPolicy 即可

### 部署cloudwatch 代理
参考 [aws doc](https://docs.aws.amazon.com/zh_cn/AmazonCloudWatch/latest/monitoring/Install-CloudWatch-Agent.html)

* 安装代理

通过下面脚本安装代理
```
wget https://s3.amazonaws.com/amazoncloudwatch-agent/linux/amd64/latest/AmazonCloudWatchAgent.zip
unzip AmazonCloudWatchAgent.zip
sudo ./install.sh
```

* 生成代理配置文件

可以在[cloudwatch_agent_config](cloudwatch_agent_config.md)，基础上修改

或者参考[aws doc](https://docs.aws.amazon.com/zh_cn/AmazonCloudWatch/latest/monitoring/create-cloudwatch-agent-configuration-file-wizard.html) 生成

将配置文件copy到实例 /opt/aws/cloudwatch_agent_config.json

* 启动代理

执行脚本

`sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a fetch-config -m ec2 -c file:/opt/aws/cloudwatch_agent_config.json -s`

* 确认状态

如果第一个脚本提示running说明成功启动，否则可以结合stop命令和上面的启动脚本来重启
```
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -m ec2 -a status
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -m ec2 -a stop
```

# 使用指标

## 查看指标
进入cloudwatch控制台，选择【指标】功能，能看到分为下面两种。选择相关空间、维度可以图形化的方式查看指标

### 自定义命名空间

* CWAgent

是通过cloudwatch 代理收集的指标

* LoopringDefine

以及通过cloudwatch api 收集的指标。如果在relay-cluster等服务的配置中打开了自定义指标上报开关，则可以看到该指标空间

### aws命名空间
是aws自定义的标准指标

## 定义告警
> 定义告警前，需要预先定义告警信息推送的主题，如何定义主题，请参考[sns](sns_cn.md)

可以在EC2控制台的 实例，负载均衡器 及 目标组 界面中进行告警的配置，也可以在cloudwatch的指标功能定义查找全量的指标并定义相关告警

通常需要对cpu, memused, disk used, swap, network, disk io 等基础指标，以及 目标组存活节点，http 响应时间等进行报警规则的设置
