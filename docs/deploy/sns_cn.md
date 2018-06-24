# 接入SNS(Simple Notification Service)

SNS提供了消息通知的功能，通知信息会发送到主题，然后消息会以短信，邮件等方式推动给订阅者

## 创建主题
【SNS-主题】，点击【创建新主题】即可创建主题，我们需要创建的主题是`RelayPerformance` `RelayNotification`
第一个主题用来配置相关告警，第二个主题用来发送关键的业务推送

## 创建订阅
点击主题的ARN连接，进入主题详情页，点击【创建订阅】

| 订阅类型 | 协议 | 格式 |
|-------|-----|--------|
| 短信  | SMS | +areacode+phonenum  |
| 邮件  | Email | email  |

> 默认短信条数会有限制，如果需要扩容，需要通过工单系统提交申请

## 服务配置中使用主题
在服务中，可以配置主题的ARN来指定SNS API所要推送的主题，通常会配置主题`RelayNotification`的ARN