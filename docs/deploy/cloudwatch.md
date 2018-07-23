# Access cloudwatch
Cloudwatch provides metrics collection, graphical display, and rule-based alarm functions

## Collecting metrics

### Default indicators
Cloudwatch will automatically collect basic metrics, using EC2, ALB, ElastiCache, RDS, and other components. These metrics can be viewed in the management system where the relevant service is located.

The use of these default metrics is free, but coarse-grained metrics may lack features. However, it can be configured according to your extra needs of additional standard metrics, fine-grained metrics, or custom uploaded metrics.

### Collect additional metrics
For RDS, ElastiCache can configure richer metrics to similar management interfaces

For EC2, the default metrics are missing these three key metrics: memory_usage, disk_usage, swap_usage. You need to deploy the watchcloud agent to record these metrics.

#### IAM Roles Required to Create a CloudWatch Agent
You can refer to: [aws doc](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/create-iam-roles-for-cloudwatch-agent.html)

Sign in to the AWS Management Console and open the IAM console at: https://console.aws.amazon.com/iam/.

In the navigation pane on the left, select a role and select [Create Role].

For [Choose the service that will use this role], select [EC2 Allows EC2 instances to call AWS services on your behalf]. select [Next: Permissions].

In the list of policies, check the box next to CloudWatchAgentServerPolicy. If necessary, use the search box to find the strategy.

Select [Next: Review].

Verify that CloudWatchAgentServerPolicy is displayed next to Policies. In Role Name, type a name for the role, for example, CloudWatchAgentServerRole. Provide a description for the role and select Create Role (Optional).

The role will be created immediately.

#### Configure the role of collecting metrics
[EC2-Instance-Instance-Operation], select [Operation-Strength Setting-Add/Replace IAM Role] to attach the CloudWatchAgentServerRole you just created to the instance.

If the instance already has other roles attached, you need to modify the original role and attach the policy CloudWatchAgentServerPolicy.

#### Deploy cloudwatch proxy
Reference: [aws doc](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Install-CloudWatch-Agent.html)

* Install agent

Install the agent through the script below
```
wget https://s3.amazonaws.com/amazoncloudwatch-agent/linux/amd64/latest/AmazonCloudWatchAgent.zip
unzip AmazonCloudWatchAgent.zip
sudo ./install.sh
```

* Generate proxy configuration file

Can be modified based on [cloudwatch_agent_config](cloudwatch_agent_config.md)

Or you can reference: [aws doc](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/create-cloudwatch-agent-configuration-file-wizard.html) Generate

Generate a copy of the configuration file to the instance /opt/aws/cloudwatch_agent_config.json

* Start the agent

Execute script

`sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a fetch-config -m ec2 -c file:/opt/aws/cloudwatch_agent_config.json -s`

* Confirm status

If the first script prompts 'running', this indicates a successful start, otherwise it can be combined with the stop command and the above startup script in order to restart
```
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -m ec2 -a status
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -m ec2 -a stop
```

## Using the metrics

### View metrics
Enter the cloudwatch console, select the "metrics" function, you can see the following two categories. Select relevant space, dimensions can be viewed graphically

#### aws namespace
Include aws custom default standard metrics

#### Custom namespace

* CWAgent

This is an additional standard metric collected through the cloudwatch proxy

* LoopringDefine

This includes metrics collected through the cloudwatch api. If you open a custom metrics recording switch while configuring services, such as a relay-cluster, you will see the space for the metrics.

### Defining notifications
> Before defining a notification, you need to understand the topic of push notification information. Want to know how to define this topic? Look here: [sns](sns.md)

You can configure notifications on the EC2 console instance, load balancer, and target group interfaces. You can also use the cloudwatch's indicator function to find the full number of indicators and configure related notifications.

It is usually necessary to configure notification rules for basic metrics such as cpu, memused, disk used, swap, network, disk io, surviving nodes of the target group, and http response time.