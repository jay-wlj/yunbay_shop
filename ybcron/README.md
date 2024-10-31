# ybcron YunBay商城定时程序

# 介绍

* ybcron为YunBay商城系统的定时程序。例如，定时资产快照，刷新交易数据，订单支付超时，汇率刷新，ybt和kt发放等定时操作

## 技术栈

- 使用github.com/robfig/cron定时库开发
- 使用redis及postgres数据库
- 使用gorm中间件
- 使用go nsq消息服务解耦,异步削峰
- docker集成


# 环境配置：

* 1，centos 7.x
* 2，supervisor进程监控




