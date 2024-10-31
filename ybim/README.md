# ybim YunBay商城im服务

# 介绍

* 此服务提供与各端进行webosket通信，并提供后端接口供调用，进行下发消息到客户端
* 此服务线上已独立成单个服务，通过域名im.yunbay.com进行访问。线上代码目录为：/opt/ybim

## 技术栈

- 使用go gin框架开发，具体的框架说明这里不累述。可查看官方文档：https://github.com/gin-gonic/gin
- 使用redis及postgres数据库
- 使用gorm中间件
- 使用go nsq消息服务解耦,异步削峰
- 网易云信IM(已废弃),类似的还有融云IM,极光等
- docker集成


# 环境配置：

* 1，centos 7.x
* 2，supervisor进程监控




