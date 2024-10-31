# yborder YunBay商城订单服务

# 介绍

* 此服务主要包含用户的商品订单信息
* 此服务线上已独立成单个服务，通过域名yborder.yunbay.com进行访问。线上代码目录为：/opt/yborder

## 技术栈

- 使用go gin框架开发，具体的框架说明这里不累述。可查看官方文档：https://github.com/gin-gonic/gin
- 使用redis及postgres数据库
- 使用gorm中间件
- 使用go nsq消息服务解耦,异步削峰
- docker集成


# 环境配置：

* 1，centos 7.x
* 2，supervisor进程监控




