#!/bin/bash
#服务启动
#注意更改一下 --data-path 所指定的数据存放路径，否则会无法运行。
log_path=/opt/YBNsq/logs
data_path=/opt/YBNsq
echo '删除日志文件'
rm -f $log_path/nsqlookupd.log
rm -f $log_path/nsqd1.log
rm -f $log_path/nsqd2.log
rm -f $log_path/nsqadmin.log

echo '启动nsq服务'
nohup nsqlookupd >$log_path/nsqlookupd.log 2>&1&

echo '启动nsqd服务'
nohup nsqd --lookupd-tcp-address=0.0.0.0:4160 -tcp-address="0.0.0.0:4150"  --data-path=$data_path/nsqd1  >$log_path/nsqd1.log 2>&1&
nohup nsqd --lookupd-tcp-address=0.0.0.0:4160 -tcp-address="0.0.0.0:4152" -http-address="0.0.0.0:4153" --data-path=$data_path/nsqd2 >$log_path/nsqd2.log 2>&1&

echo '启动nsqdadmin服务'
nohup nsqadmin --lookupd-http-address=0.0.0.0:4161 >$log_path/nsqadmin.log 2>&1&

