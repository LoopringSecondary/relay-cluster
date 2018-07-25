# Start EC2 Instance

## Start a new instance

> If you don't have an aws account, please register it first

Log in to the aws console [http://aws.amazon.com](http://aws.amazon.com) and select [Instance-Instance-Startup Instance]

[Step 1: Select an Amazon System Image (AMI) page], select the ubuntu instance type because all subsequent deployments are based on ubuntu instances. It is recommended to use the generic instance`Ubuntu Server 16.04 LTS (HVM), SSD Volume Type`, and then click [Select]

[Step 2: Select an example type], select the appropriate type on the page. If you do an experiment, you can select a free instance.

[Step 3: Configure Example Details page], select Default VPC in Network. If there are multiple instances, it is recommended that you create them in groups, and then select different subnets for each group to avoid the single aws room failing, which would make the service unavailable. For IAM roles, you do not need to set them and configure them as needed. Other choices are the default
> If you need to support deployment of services through CodeDeploy in this instance, you must configure additional IAM roles and initialization scripts, refer to [Access CodedDploy](#support-codedeploy)

[Step 4: Add Storage page], the disk size is recommended to be greater than 20G. [Volume Type] is [Default Universal SSD]

[Step 5: Add Label], skip and follow-up can be manually configured

[Step 6: Configure Security Group], select [Select an existing security group] and select the security group whose name is [launch-wizard-1]. That is, only the configuration of ssh port 22 is enabled. Follow-up is performed as needed. Adjustments, click [review and start]

[Step 7: Verify Instance Startup], confirm that the relevant configuration is correct and click [Start]. In the Startup popup box, you will be prompted to select the login private key. If not, select Create new one, otherwise select the existing one. If you create a key pair, download the new key and save it. All subsequent ssh operations depend on the bound xx.pem private key file. Click Start Instance.

Click [View Example] to enter EC2 instance list page, you can see the startup status of the instance

## Login instance
Modify xx.pem permissions`chmod 400 xx.pem`

For instances where the instance-instance-instance state is running, execute `ssh -i xx.pem ubuntu@ipv4_add` to log in to the instance. Here ipv4_add is the IPv4 public IP of the instance.

## Support CodeDeploy

> `CodeDeployEc2InstanceProfile`, please refer[CodeDeploy](codedeploy.md)

IAM select `CodeDeployEc2InstanceProfile`, enter the text for advanced details
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

## Deploy aws sdk credentials file
Cloudwatch and SNS(Simple Notification Service) are now used in backend services by aws sdk, which need credentials file for initialization.

If you enabled any of these two services in your service configuration, you will need deploy the credentials file descibed below, otherwise just ignore following operation.

### Create IAM user
Refer [aws doc](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html)

Open IAM console, Click [Add User]

[Step 1], enter user name`sdkUser`, [Access type] select [Programmatic access]

[Step 2], select [Attach existing policies directly]

	In case cloudwatch is enabled, please add those three policies: `CloudWatchAgentServerPolicy`, `AmazonAPIGatewayPushToCloudWatchLogs`, `CloudWatchActionsEC2Access`.

	In case SNS is enabled, please add `AmazonSNSFullAccess` policy.

[Step 3], click [Create user].

[Step 4], record [Access key ID] and [Secret access key] for usage at below.

### Deploy credentials file

Deploy path in every EC2 instance is `~/.aws/credentials`.

Add [Access key ID] and [Secret access key] to this file as below format
```
[default]
aws_access_key_id = xxxx
aws_secret_access_key = xxxx
```