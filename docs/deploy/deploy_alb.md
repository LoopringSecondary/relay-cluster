# Deploy Aws LoadBalancer

ALB (Application Load Balancer) is a load balancer provided by aws, and ALB has a relatively high availability.

By mapping a set of servers to a target group and then associating ALBs, we can configure the mapping between the request url and the backend service to forward the request to the correct backend interface.

## Configure the target group
The target group aggregates the port of a group of servers into a group. This group serves as the request-forwarding destination of the ALB.

### Creating the first target group
[EC2-Load Balance-Target Group], click [Create Target Group]

* Basic Information

Enter the target group name as `relayClusterGroup`, the protocol is HTTP, port 8083, for Target Type select the instance, and for VPC select the default.

* Operational status health check

For [protocol] select HTTP, path default `/`

By default, the health check port is the same as the port of the target group. If you need to modify the health check port, [Health Check - Advanced Health Check - Port], select [Overwrite] to enter the new port number.

Click [Create]

### Other target groups
Refer to the relayCluster target group above to create another target group

| Target Group Name  	 | Protocol |   Port | Operational status health check|
|------------------------|-----|--------|---------|
| relayClusterGroup      | HTTP | 8083  |Default 8083    |
| relayClusterWebSockets | HTTP | 8087  |Coverage is 8083 |
| ethGroup               | HTTP | 8545  |Default 8545 |

### Configuration target

Select a target group, select the following [Target], click Edit, in the pop-up dialog box below the [instance] section, select the correct instance, which is the one that has deployed the service, click [Add to Registered], Select [Save].

In the [Target] tab, you can see the status of the newly added instance. If the target group is not associated with the ALB, it should be displayed as [unused].

The correspondence between the target group and the instance deployment service is as follows:


| Target Group Name         		 | Instance deployment service type |
|------------------------|---------------|
| relayClusterGroup      | relay-cluster |
| relayClusterWebSockets | relay-cluster |
| ethGroup               | ethnode |

## Configure ALB

### Create ALB
[EC2-Load Balance - Load Balancer], click [Create Load Balancer], select [Application Load Balancer - Create]

* Step 1: Configure the Load Balancer

For [Name] enter relayCluster, for [Mode] select [Internet]

For [listener], add HTTP, select the default 80 port. If you apply for a certificate at a certificate authority, add an additional HTTPS listener

For [Available Area] select all the available areas where you have deployed a relay-cluster, and select at least one area

* Step 2: Configure security settings

If you have already applied for the https certificate in the certificate authority, please configure the certificate here. Select the security policy "ELBSecurityPolicy-2016-08"

* Step 3: Configure Security Groups

Select [Select an existing security group], select the `alb-SecurityGroup` security group. If you have not created the security group, please refer to: [Aws security group](security_group.md). For the description of the `alb-SecurityGroup` security group, create it after association.

* Step 4: Configure Routing

For [Target Group] select [Existing Target Group], select `relayClusterGroup`

* Step 5: Registration goals

Simple confirmation

* Step 6: Review

Click [Create]

### Configure the route
Select the [relayCluster] load balancer you just created, and click [View/Edit Rule] in the [listener] tab

Click [Edit] and [Add] icon on the page, and [Add Condition] select [Path is ..], add the following rules one by one, and [Save]

| Path         | Forward to |
|-------------|----------|
| /rpc/*      | relayClusterGroup |
| /rpc/v2/*   | relayClusterGroup |
| /eth        | ethGroup |
| /socket.io/*| relayClusterGroup |

## Confirm deployment status
[EC2-Load Balance-Target Group], click on the [Target] tab of the three target groups created previously, and confirm that [Status] is listed as [healthy]. If [unhealthy] is displayed, move the mouse to the next exclamation mark icon, check the prompt reason, and then solve.

## Connect ALB
After the ALB configuration is complete, the backend service can be requested through the ALB.

Click on the created `relayCluster` load balancer. In the [description] tab, find the value corresponding to [DNS name]. This domain name can be used to access the related interfaces of the relay-cluster and ethnode by adding the previously configured route url.

## Check ALB access log
Sometimes, you need to analyze the problem through the ALB access log. By default, the log function is disabled. You need to configure the log recording function.

For more specific operations, you can refer to: [aws doc](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-access-logs.html)