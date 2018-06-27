# Access CodeDeploy

CodeDeloy is a code deployment tool provided by aws that adds support for codeDeploy deployment by adding configuration files and scripts to the project. CodeDeloy downloads github project source code to the target server, and then parses the project configuration file, and executes the corresponding script in different stages of deployment to achieve code compilation and service startup, verification and other operations.

## Configuration dependencies

### Create IAM object
You need to give CodeDeploy deployment permissions by creating related IAM objects and binding EC2 instances, IAM console entry [https://console.aws.amazon.com/iam/](https://console.aws.amazon.com/iam/)

#### Create a new IAM user and bind the policy

For [IAM-User] select [Add User], for reference: [aws doc](https://docs.aws.amazon.com/codedeploy/latest/userguide/getting-started-provision-user.html)

Step 1 - Enter the user name `CodeDeployer`, select the access type  [programmatic access], click [Next: permissions]

Step 2 - Select [Directly Attach Existing Policy], select [Create Strategy], click the [JSON] tab page, paste the JSON text in the following code block, click [Review Policy], enter the name of the new policy, CodeDeployPolicy, and click [Create Policy]. Once completed, go back to the [Step 2] page, click [Refresh], and select the newly created "CodeDeployPolicy" policy, click [Next: Audit]

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

Step 3. Click Create User to complete the creation of the CodeDeployer user

Step 4, confirm and close

#### Creating a Role

For [IAM-Role] select [Create Role] for reference: [aws doc](https://docs.aws.amazon.com/codedeploy/latest/userguide/getting-started-create-service-role.html)

Step 1 -  Select [AWS Products], select [Service CodeDeploy], select [CodeDeploy], and click [Next: Permissions]

Step 2. Confirm that [AWSCodeDeployRole] appears and click [Next: Audit]

Step 3. Enter the role name `CodeDeployServiceRole` and click on [Create Role]

#### Create EC2 binding policy

[IAM-Strategy], select [Create Strategy], Reference: [aws doc](https://docs.aws.amazon.com/codedeploy/latest/userguide/getting-started-provision-user.html)

Step 1 - Select [JSON], paste the following text, and click [Review policy]
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
Step 2 - Enter `CodeDeployEc2Permission` as the name and click [Create Policy]

[IAM-Role], select [Create Role]

Step 1 - Select [AWS Products], Service Selection [EC2], Case Selection [EC2], and click [Next: Permissions]

Step 2 - In the search box, enter the `CodeDeployEc2Permission` you just created, tick the policy, and click [Next: Audit]

Step 3 - For the role name, enter `CodeDeployEc2InstanceProfile`, select [create role]


### Configure EC2 instances

> Only need operate once for one EC2 instance

> It recommended to configure the EC2 instance when starting it. Reference: [Support CodeDeploy](new_ec2.md#support-codedeploy), and if you have a similar configuration, you do not need to perform the following two operations.

#### Configuration role
Modify the existing instance of the IAM role, [EC2-instance-instance], select an EC2 instance that needs to be deployed through CodeDeploy, [Operation - Instance Settings - Attach/Replace IAM Role], [IAM Role] select the previously created `CodeDeployEc2InstanceProfile' `, select [Application]

#### Deploy and start codedeploy-agent
ssh login EC2 instance, execute the following script

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
Verify that the agent executes normally after you execute the script `sudo service codedeploy-agent status`. If the response is `running`, it means it is normal, otherwise, restart through the following script, and refer to: [aws doc](https://docs.aws.amazon.com/codedeploy/latest/userguide/codedeploy-agent-operations-install-ubuntu.html)

```
sudo service codedeploy-agent stop
sudo service codedeploy-agent start
```

#### Set the label
Instances with the same tag set can be filtered, which is necessary in CodeDeploy deployments.

The specific operation is, [EC2-instance-instance], select a single instance, and in the [tab] tab on the details page below, you can edit the label as needed.

For example, here is the labeling needed to deploy the relay-cluster instance [relay-cluster/test]

## Configure and deploy applications

### Configure CodeDeploy Application

> Only need operate once for one application

#### Create an application
Open the CodeDeploy console, and select [Start Now]

The following uses the relay-cluster as an example to demonstrate selecting [Custom Deployment - Skip Exercise]

* Create an application

For [Application Name] enter `relay-cluster`, for [Calculation Platform] select `EC2/Local`, and for [Deploy Group Name] enter `myRelayClusterGroup`

* Deployment type

Select [in-place deployment]


* Environment configuration

Select [Amazon Ec2 Instance] and add the previously configured label [relay-cluster/test] in [Label Group] to automatically filter the matched instances.
* Deployment configuration

Select CodeDeployDefault.OneAtATime

* Service role

Select the CodeDeployServiceRole created earlier

After other defaults, click [Create Application]

### Deploy the application

Select [Deploy] in the pull-down menu in the upper left corner of the console and select [Create Deployment]

* Configure basic information

For [Application] Select `relay-cluster`, for [deployment group] select `myRelayClusterGroup`, for [repository type] select Github, for [deployment configuration] select CodeDeployDefault.OneAtATime
> If you are deploying an application for the first time, you need to select CodeDeployDefault.AllAtOnce

* Connect to GitHub

For deploying an application for the first time, you need to connect your github account, by entering the github email account in the input box, and then jumping to github for authentication.

* Repository name

Enter `Loopring/relay-cluster`

* Submit id

Select the branch to be deployed in github, usually master, then select commits, find the latest commitid, paste it into the input box

* Start deployment

Click [Deploy] button to start deployment

### Check the deployment status
If you go back to the [deployment] list, you can see all the deployments that are currently underway. You can click to view the deployment details, and click [View All Instances] to see which nodes are currently deployed and at which stage.

* The latest event

This is the phase identifier of the deployment. The specific definition can be referenced: [aws doc](https://docs.aws.amazon.com/codedeploy/latest/userguide/reference-appspec-file-structure-hooks.html)
* Failure log

For failures or errors, click on the "View events" log column to view

#### Solve deployment issues
* Common problems

Reference documents: [aws doc](https://docs.aws.amazon.com/codedeploy/latest/userguide/troubleshooting-ec2-instances.html)

* Check the log file

`/opt/codedeploy-agent/deployment-root/deployment-logs`

* Confirm that the aws agent status is normal

Execute the script `sudo service codedeploy-agent status`. If the prompt is not running, try executing the following:
```
sudo service codedeploy-agent stop
sudo service codedeploy-agent start
```

If it cannot be restored, you need to restart the service, kill related processes, then execute: `sudo service codedeploy-agent start`

* Confirm codedeploy remote connection is open

Run `ps axu|grep CodeDeployPlugin` to find the process id xxx

Check whether TCP connection is established by execute `sudo lsof -p xxx |grep TCP`, it will be ok if return status is `ESTABLISHED`, otherwise, you need restart the agent.
