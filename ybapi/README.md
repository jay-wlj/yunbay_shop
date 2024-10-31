# ybapi YunBay商城api服务

* YunBay商城api服务。本项目采用Go gin框架，具体的框架说明这里不累述。可查看官方文档：https://github.com/gin-gonic/gin
* go 为1.13.4版
* 此服务线上已独立成单个服务，通过域名ybapi.yunbay.com进行访问。线上代码目录为：/opt/ybapi

## 技术栈

- 使用go gin框架开发
- 使用redis及postgres数据库
- 使用gorm中间件
- 使用go nsq消息服务解耦,异步削峰
- docker集成




# 环境配置：

* 1，go 1.13.X  postgres 10.x.x git
* 2，git clone https://e.coding.net/yunbayshop/yunbay_services.git
* 3，cd yunbay_services/ybapi && go build    



