# Access SNS(Simple Notification Service)

SNS provides a message notification function, where notification information will be sent to the theme, then the message will be promoted to subscribers by SMS, email, etc.

## Create a theme
[SNS-Theme], click [Create New Theme] to create the theme, the theme we need to create is `RelayPerformance` `RelayNotification`

The first topic is used to configure related notifications, and the second topic is used to send key business push notifications

## Create a subscription
Click the theme ARN connection, enter the theme details page, click [create subscription]

| Subscription type | Protocol | Format |
|-------|-----|--------|
| SMS  | SMS | +areacode+phonenum  |
| Email  | Email | email  |

> The number of default SMS messages will be limited. If you need to expand the capacity, you need to submit an application through the work order system.

## Use themes in service configuration
In the service, the ARN theme can be configured to specify the theme to be pushed out by the SNS API. Usually, the ARN theme `RelayNotification` is configured.