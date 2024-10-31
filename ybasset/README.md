# ybasset YunBay商城资产服务

# 介绍

* ybasset为YunBay商城资产服务。主要包含用户及平台有关资产、充提、支付、收益发放等功能
* 此服务线上已独立成单个服务，通过域名asset.yunbay.com进行访问。线上代码目录为：/opt/ybasset

## 技术栈

- 使用go gin框架开发，具体的框架说明这里不累述。可查看官方文档：https://github.com/gin-gonic/gin
- 使用redis及postgres数据库
- 使用gorm中间件
- 使用go nsq消息服务解耦,异步削峰
- docker集成


# 环境配置：

* 1，centos 7.x
* 2，supervisor进程监控




