## 订单广播
在Loopring的设计里，我们通过订单广播来共享订单，从而提高撮合的效率，增加用户体验。在初始设计时，我们尝试采用IPFS进行广播，但是由于IPFS没有验证也无法进行有效地控制，如果一个topic被攻击只能更换topic，并且所有的接入方都需要同步更换，代价较高。现在我们尝试使用Matrix作为广播的渠道。

### Matrix介绍
Matrix是一个开源分布式的消息协议，可以使不同的服务器的用户进行通信，通信记录存放在用户选择的服务器上。详细内容可以参考[官网](https://matrix.org/)

### 加入已有的广播渠道
可以通过以下步骤加入已有的广播渠道，

1、创建一个账号，

 * 选择任意Matrix客户端，
 * 更改homeserver到想要加入的渠道（如果该服务器已经加入联盟，也可选择官方matrix.org注册）
 * 根据聊天室的设置，直接加入或者等待邀请再加入聊天室
 
2、更改配置文件

```
[gateway]
    is_broadcast = true
    max_broadcast_time = 3
    #订单广播配置
    [[gateway.matrix_pub_options]]
        rooms = [ "!RoJQgzCfBKHQznReRT:localhost"]  #聊天室的roomid
        [gateway.matrix_pub_options.MatrixClientOptions]
            hs_url = "http://13.112.62.24:8008" #服务器地址
            user = "broadcast1"   #用户名
            password = "broadcast1"  #密码
            access_token = "MDAxN2xvY2F0aW9uIGxvY2FsaG9zdAowMDEzaWRlbnRpZmllciBrZXkKMDAxMGNpZCBnZW4gPSAxCjAwMjhjaWQgdXNlcl9pZCA9IEBicm9hZGNhc3QyOmxvY2FsaG9zdAowMDE2Y2lkIHR5cGUgPSBhY2Nlc3MKMDAyMWNpZCBub25jZSA9IEtqS0JVNjtXSSY9NnJURWoKMDAyZnNpZ25hdHVyZSC7rg8ODsNmK_SXxe3gxHt7m6RmtHkJd1RCMdoUCE5iuAo" #token，也可不填
    #接受订单广播的配置
    [[gateway.matrix_sub_options]]
        rooms = [ "!RoJQgzCfBKHQznReRT:localhost"]
        cache_from = true
        cache_ttl = 86400
        [gateway.matrix_sub_options.MatrixClientOptions]
            hs_url = "http://13.112.62.24:8008"
            user = "broadcast1"
            password = "broadcast1"
            access_token = "MDAxN2xvY2F0aW9uIGxvY2FsaG9zdAowMDEzaWRlbnRpZmllciBrZXkKMDAxMGNpZCBnZW4gPSAxCjAwMjhjaWQgdXNlcl9pZCA9IEBicm9hZGNhc3QyOmxvY2FsaG9zdAowMDE2Y2lkIHR5cGUgPSBhY2Nlc3MKMDAyMWNpZCBub25jZSA9IEtqS0JVNjtXSSY9NnJURWoKMDAyZnNpZ25hdHVyZSC7rg8ODsNmK_SXxe3gxHt7m6RmtHkJd1RCMdoUCE5iuAo"

```

todo:路印官方广播渠道

### 管理广播渠道
由于使用了Matrix中的room概念进行广播，那么你可以自行创建一个room作为广播的渠道
你也可以对接入方做各种控制，如邀请、禁止、剔除等，这些操作也可以通过Matrix客户端完成
