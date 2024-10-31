#Yunbay商城后台系统

# 项目简介
该目录包含yunbay绝大部分服务项目(ybproduct为php写的服务，没有包括在本目录下)，以下文件夹作一一介绍:
- account: 帐号服务
- common: 公共目录
- conf: 公共配置目录
- eosio: eos节点操作库
- psql: 数据库结构设计目录
- utils: 公共工具目录
- ybapi: api服务
- ybasset: 资产服务
- ybcron: 定时程序
- ybeos: eos交易服务
- ybgoods: 商品服务
- ybim: im服务
- ybnsq: nsq消息处理程序
- yborder: 订单服务
- ybpay: 第三方支付服务
- ybsearch: 搜索服务

另还有php的服务项目不在yunbay目录中，单独列出如下:
- ybproduct 商品服务(app及管理后台调用)

# 环境部署

每个项目都分成三种环境：

- 开发环境：	供开发人员自用

- 测试环境：	暴露出一个端口号，部署在内网，通过IP和端口号或host访问

- 生产环境：	部署在线上，谷歌云

### hosts配置
```
172.17.6.140 account.yunbay.com
172.17.6.140 asset.yunbay.com
172.17.6.140 api.yunbay.com
172.17.6.140 product.yunbay.com
172.17.6.140 im.yunbay.com
172.17.6.140 m.yunbay.com
172.17.6.140 pay.yunbay.com
172.17.6.140 goods.yunbay.com
172.17.6.140 search.yunbay.com
```

## 端口使用情况

这些是部署在测试/生产环境时所开启的端口号

使用nginx负载均衡

| 服务 					| 开发端口	| 测试端口	| 生产端口	|
| --------				| :-----:	| :-----:	| :-----:	|
| account        		| 92		| 92		| 92		|
| ybapi      			| 90 		| 90		| 90		|
| ybasset    			| 95		| 95		| 95		|
| ybgoods				| 96		| 96		| 96		|
| yborder    			| 91		| 91		| 91		|
| ybeos             	| 97		| 97		| 2007		|
| ybim              	| 98		| 98		| 2004		|
| ybpay             	| 94		| 94		| 2006		|
| ybsearch  			| 93		| 93		| 9001		|
| upload        		| 2000		| 2000		| 2000		|

# 项目部署
```
1、安装postgres,redis,supervisor,nginx服务

2、部署nsq分布式实时消息服务,如nsqd,nsqlookup,nsqadmin目前均在supervisor中监控

3、部署以下所有服务到/opt下,并配置supervisor监控各服务程序

4、配置nginx各域名服务及负载均衡,配置文件目录:/etc/nginx/sites/*.conf
   
5、在supervisorctl中更新监控配置，重启、暂停等服务
    如:supervisorctl update 更新监控配置
       supervisorctl restart account   重启帐号服务

6、supervisor监控服务的日志在 /tmp/supervisor/*.log 中可查看相应服务的输出日志
   同时各服务的glog详细日志可在 /tmp/logs/ 中查看

```

# 谷歌云线上程序配置说明
```
各节点服务程序部署位置及说明(内网已实现docker stack集群部署，不再需要supervisor监控)
谷歌云服务
app1:35.240.147.53  云贝商城后台系统主机
app2:35.240.198.149 开放(龙网)商城后台系统及云贝商城搜索服务所在主机
web:35.198.224.25  所有web端(云贝商城+龙网商城)主机
chain:35.240.208.98 以太坊节点主机

(1)、app1节点 云贝商城后台系统
/opt/supervisor 监控程序目录
/opt/account 云贝商城帐号服务
/opt/conf 各服务公共配置文件
/opt/configs ybproduct搜索服务公共配置文件
/opt/libs ybproduct搜索服务公共库
/opt/tgrobot 机器人配置目录(已废弃)
/opt/upload 上传服务
/opt/ybapi api服务
/opt/ybasset 资产服务
/opt/ybcron 定时程序
/opt/ybeos eos节点服务
/opt/ybgoods 商品服务
/opt/ybim im服务
/opt/ybnsq nsq消息处理程序
/opt/yborder 订单服务
/opt/ybpay 第三方支付(支付宝微信)服务
/opt/ybproduct php写的商品及管理后台api服务
/opt/app 各服务程序(有时连不上线上主机，通过git方式提交下载更新各服务)


(2)、app2节点 开放(龙网)商城后台系统
/opt/supervisor 监控程序目录
/opt/account 云贝商城帐号服务
/opt/conf 各服务公共配置文件
/opt/ybapi api服务(包括公共接口及订单)
/opt/ybcron 定时程序
/opt/ybgoods 商品服务
/opt/ybnsq nsq消息处理程序
/opt/ybopen 第三方(龙网)对接服务
/opt/youbuy 优买会系统目录(已废弃)

(3)、web节点 云贝商城及龙网商城的web端系统
/home/lxc/yunbay 前端各项目
/opt/ybsearch 云贝搜索服务
/usr/local/sphinx/etc/sphinx.conf  sphinx搜索引擎配置文件

(4)、chain节点 云贝商城及龙网商城的web端系统
/data/eth 以太坊结点目录(郑克明负责部署)



```
# 注意事项
```
1、使用国外mailgun邮件服务http://mailgun.com 自建云贝商城的邮件系统，所有发往server@yunbay.com的邮件统一转发到yunbay2018@gmail.com邮箱,具体设置可在mailgun.com官网帐号更改

2、内网禅道
http://172.17.6.140:8088/zentao/user-login.html
用户名:admin
密码:yunfan_123

3、目前虚拟商品的价格(话费及京东钢蹦)不支持在商户后台直接修改，需要后台开发人员直接修改数据库(目前也没提供接口，考虑不是经常改动，优先级低还未实现)
   1、修改该虚拟商品下的product_sku表里的extinfo字段
   2、刷新该商品缓存 redis-cli -a yunbay_123
      hdel goods_info 商品id
      del goods_high_list:1/0
   3、app访问才能拿到更新的价格商品  

4、目前由于没有用到配置管理系统，涉及资产的充提开关，手续费等配置信息时，需要直接修改相关配置文件重启服务生效。
```

# 当前系统存在的缺陷和发展计划
```
   以下内容全是个人在开发过程结合自己的一些经验而得出的个人意见，仅供参考
   1、目前仅对用户每日ybt挖矿及kt分红数据进行按月分表存储(每张表大约150W行数据),后续其它表数据量增长快时也可能需要分表处理
   2、目前系统对接口访问频率及性能相关日志没有具体的监控图表，但对每个接口的调用都记录了详细的日志，目前是通过日志来了解一些接口相关信息，后期可加入到监控图表中及时显示。
   3、订单抽奖等业务接口后期建议加入限流熔断机制，避免抽奖商品接口的高并发访问发生的问题
   4、后期服务间内部可改用grpc协议通讯，同时对外提供RESTful http接口，往微服务架构(go-kit)迁移。
   5、后期线上整个系统部署可改用docker swarm集群方式，简化部署，还能弹性伸缩等(内网已改用docker stack部署，还在测试中)
   6、目前已实现对比较重要的接口异常进行监控，并实时将错误信息推送到钉钉或邮件，但是在相关接口添加的判断逻辑，后续可以独立出来解耦业务逻辑。
   7、性能优化，对相关接口耗时及QPS等进行监控及优化等，对频繁访问的接口进行缓存筛，对SQL进行调优等，后期建议后续加入prometheus自动监控警报系统
   8、加入分布式配置管理,如zookeeper,etcd等实现。管理后台直接设置，服务无需重启即可动态生效
```


# 以下为项目的后端系统源码地址
```
云贝后台系统(Go):https://e.coding.net/yunbayshop/yunbay_services.git
云贝后台系统(php):https://e.coding.net/yunbayshop/ybproduct.git
云贝后台系统(php):https://e.coding.net/yunbayshop/yblibs.git
开放(龙网)商城后台系统:https://e.coding.net/yunbayshop/ybopen.git
优买会后台系统(已下线):https://e.coding.net/yunbayshop/youbuy_services.git
```