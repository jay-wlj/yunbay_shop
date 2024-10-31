
# 生成token
```
curl http://account.nicefilm.com:810/account/man/token?user_id=380
```
# 播单
----



### 获取指定的用户id的推荐人
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" 'http://127.0.0.1:90/man/invite/beinvites?user_ids=51888,51896,51911'
```


### 获取指定的用户id的推荐人
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" 'http://127.0.0.1:90/man/order/rebat/update' -d '{"order_id":1604785, "rebat":0.22, "tx_hash":"62361670a0b60c852fcc1e69189c233e"}'

### 获取指定商品售出列表
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" 'http://127.0.0.1:90/v1/order/list_by_product?product_id=10000016'
```

### 获取订单折扣信息
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" 'http://127.0.0.1:90/v1/order/rebat/info?order_id=1605088'
```



### 获取订单折扣信息
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" 'http://127.0.0.1:90/man/order/report?begin_date=2019-08-01&end_date=2019-08-31'
```