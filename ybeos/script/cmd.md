
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
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" http://127.0.0.1:2008/man/transaction/push -d '{"order_id":51887,"amount":1, "memo":"ha"}'
```
