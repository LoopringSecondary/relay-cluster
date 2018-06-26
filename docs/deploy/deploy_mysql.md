Mysql is the main source of the relay-cluster backend service
## Create MySQL instance
You can choose the aws RDS or self-built mysql, it is recommended to use the RDS, because it will contain richer monitoring and management functions, and the extension is more convenient for the test scenario. However, you can build a MySQL single instance, if you so choose.
### Create an RDS instance
Find `RDS` from the service list to locate the entry, and then select 'Start Now'

Step 1 - [Select Engine], click [MySql], click [Next]

Step 2 - according to your own scenario, select [Production - MySQL], click [Next]

Step 3
* Instance specifications

For [Database Engine Version] select a new version higher than 5.7. [Database Instance Type] is recommended to select [db.m4.xlarge] and later versions of the host. [Multiple Zone Available Deployment] will provide higher availability, and confirm what to select according to your needs. For [Storage Type] select [General Purpose (SSD)] to allocate more than 50GB of storage space.

* Settings

For [Database Instance Identifier] enter a similar loopring-relay, relay-miner, then enter the [main username], and the corresponding password.

Step 4
* Network and Security

For [Virtual Private Cloud] select default, for [Subnet Group] select default, for [Public Availability] select according to the situation, for [Available Area] select the partition where the service that will access the MySQL library resides, for VPC security group, select `mysql -securityGroup`, if not yet created, please refer to: [Configure the aws security group](https://github.com/Loopring/relay-cluster/wiki/%E9%85%8D%E7%BD%AEaws%E5%AE%89%E5%85%A8%E7%BB%84), and after creating, come back to choose the `mysql-securityGroup`.

* Database options

For [Database Name] enter a name similar to the relay name and miner name, [Port] uses the default 3306, [Database Parameter Group] and [Option Group] select the default, for [IAM Database Authentication] choose according to the actual situation, for [Encryption] select [Enabled Encryption]
> For [Database Parameter Group] we recommend you create a new one. This is because the self-built parameter group can be modified, and the default one cannot be modified. The database can be easily configured by modifying the parameters of the parameter group. Most configurations do not require data to be restarted. You can refer to the following ones.

* Back up

We recommend you perform a back up and set a suitable start time. Note that this is UTC time, and Beijing time requires +8 hours.

* Monitor

We recommend you start enhanced monitoring.

* Log export

Suggest to check [Error Log], [General], [Slow Query Log]

* Maintenance

Disable [Automatic Minor Version Upgrade], a similar process to backing up, select the proper maintenance window, and select start the database instance
### Create a stand-alone MySQL instance
Reference [启动aws EC2实例](https://github.com/Loopring/relay-cluster/wiki/%E5%90%AF%E5%8A%A8aws-EC2%E5%AE%9E%E4%BE%8B), start the instance, and connect the `mysql-securityGroup` security group

Execute the following script to deploy a Mysql instance
```
sudo apt install mysql-server
```
Enter the corresponding password for the root user as prompted and confirm again

Create a relay db
```
mysql --host=localhost --port=3306 --user=root -p
CREATE DATABASE relay;
```
## Connect to the database
> Both the relay and miner will use the mysql database. We recommend you create different database instances to avoid mutual influence.

Record the password of the user name specified in the previous creation of db, which can be configured in the relevant configuration file. Also, you access the database instance through the 1000 command-line tools.
<!--stackedit_data:
eyJoaXN0b3J5IjpbMTMyMDc1NzM2MywxODM1OTA2NjA3XX0=
-->