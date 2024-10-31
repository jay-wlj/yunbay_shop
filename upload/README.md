# upload YunBay商城文件上传服务

* 本项目采用Go gin框架，具体的框架说明这里不累述。可查看官方文档：https://github.com/gin-gonic/gin
* go 为1.13.4版
* 此服务线上已独立成单个服务，通过域名upload.yunbay.com进行访问。线上代码目录为：/opt/upload

## 技术栈

- 使用go gin框架开发
- 使用redis及postgres数据库
- 使用gorm中间件
- docker集成



# 环境配置：

* 1，go 1.13.X  postgres 10.x.x git
* 2，git clone https://e.coding.net/yunbayshop/yunbay_services.git
* 3，cd yunbay_services/upload && go build 
* 4, 使用谷歌云存储服务，需要配置环境变量，添加谷歌存储验证文件 GOOGLE_APPLICATION_CREDENTIALS=/opt/upload/conf/store.json




