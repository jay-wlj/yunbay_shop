
# 生成token
```
curl 'http://account.yunbay.com:92/man/account/token?user_id=380'
```

### 发送一个消息到nsq服务中
```
curl -H"Content-Type: application/json" 'http://127.0.0.1:4151/pub?topic=mqurl' \
-d '{"delay":5000000,"method":"POST","appkey":"ybapi","uri":"/man/business/amount/update","headers":null,"data":[{"user_id":55221,"amount":1,"rebat":"0.99"}],"timeout":0,"maxtrys":0}'
```


### 请求一个支付宝支付链接
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" http://127.0.0.1:2006/man/alipay/trade/pay -d '{"user_id":51887,"amount":0.5,"order_ids":[1604374], "subject":"新订单有"}'
```


### 支付宝支付异步通知
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1"  http://127.0.0.1:2006/v1/alipay/trade/notify -d '{"out_trade_no":"5","trade_no":"zfbv2","buyer_logon_id":"sd3e","trade_status":"TRADE_SUCCESS", "total_amount":"200"}'
```
curl -d "out_trade_no=5&trade_no=zfbv2" "http://127.0.0.1:2006/v1/alipay/trade/notify"



### 请求一个支付链接
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" http://127.0.0.1:2006/man/alipay/trade/pay -d '{"user_id":51887,"amount":1,"order_ids":[1,2], "subject":"每一个订单"}'
```

### 请求一个支付宝转帐
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" http://127.0.0.1:2006/man/alipay/trade/trasfer -d '{"out_biz_no":"1","payee_type":"ALIPAY_LOGONID","payee_account":"305898636@qq.com", "amount":"1.1", "remark":"yunbay转帐"}'
```


### 查询支付宝转帐状态
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" http://127.0.0.1:2006/man/alipay/trade/trasfer/query -d '{"out_biz_no":"51887"}'
```


### 根据银行卡前6,7位查询所属银行
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" 'http://127.0.0.1:2006/v1/bank/query?card_id=62122615'
```


### 关闭交易订单
```
curl -v -X POST     -H'X-YF-SIGN: 62361670a0b60c852fcc1e69189c233e' -H'Host: 172.17.6.140' -H'X-YF-AppId: ybpay' -H'X-YF-rid: 1' -H'X-YF-Platform: man' -H'X-YF-Version: 1.0.1' -H'Content-Type: application/json' 'http://172.17.6.140:2006/man/trade/close' -d '{}'
```


### 请求一个微信支付链接
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" http://127.0.0.1:2006/man/alipay/trade/pay -d '{"user_id":51887,"amount":1,"order_ids":[1,2], "subject":"每一个订单"}'
```