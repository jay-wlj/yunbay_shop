# ybsearch YunBay商城搜索服务

# 介绍

* 此服务主要包含商品名称类型等搜索功能
* 此服务线上已独立成单个服务，通过域名ybsearch.yunbay.com进行访问。线上代码目录为：/opt/ybsearch

## 技术栈

- 使用go gin框架开发，具体的框架说明这里不累述。可查看官方文档：https://github.com/gin-gonic/gin
- 使用redis及postgres数据库
- 使用gorm中间件
- 使用go nsq消息服务解耦,异步削峰
- 使用sphinx搜索服务
- docker集成


# 环境配置：

* 1，centos 7.x
* 2，supervisor进程监控
* sphinx 3.1.1




# 索引更新策略

* 全量索引：每天凌晨4：40分更新一次
* 增量索引：每分钟更新一次（比如新增商品、修改商品、商品的上/下架、商品的屏蔽）
* 对商品进行下架和屏蔽处理，如果是昨天以前的数据，实时在搜索中无法找到， 如果是今天的数据，最晚一分钟后在搜索中无法找到
* 搜索结果默认综合排序（权重）：标题 > 类别 > 特色介绍

# docs目录文件：

* productAll.sh : 全量索引运行脚本
* productDelta.sh : 增量索引运行脚本
* sphinx.conf : 索引配置，线上此配置在：/usr/local/sphinx/etc/sphinx.conf


# 下面是常用的sphinx命令：

    ./searchd -c /usr/local/sphinx/etc/sphinx.conf 启动搜索服务
    ./searchd -c /usr/local/sphinx/etc/sphinx.conf --stop 停止搜索服务
    ./searchd -c /usr/local/sphinx/etc/sphinx.conf -h 查看帮助
    ./indexer -c /usr/local/sphinx/etc/sphinx.conf --all 重建所有索引（服务未启动时）
    ./indexer -c /usr/local/sphinx/etc/sphinx.conf i_product 重建单个索引
    ./indexer -c /usr/local/sphinx/etc/sphinx.conf --all --rotate 平滑重建所有索引（服务已启动时）
    ./indexer -c /usr/local/sphinx/etc/sphinx.conf --merge main delta --rotate 合并索引
