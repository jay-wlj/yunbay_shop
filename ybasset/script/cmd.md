
# 生成token
```
curl 'http://account.yunbay.com:92/man/account/token?user_id=380'
```

### 发送一个消息到nsq服务中
```
curl -H"Content-Type: application/json" 'http://127.0.0.1:4151/pub?topic=mqurl' \
-d '{"delay":"6s", "method":"POST","appkey":"ybapi","uri":"/man/business/amount/update/m","headers":null,"data":[{"user_id":55221,"amount":1,"rebat":0.99}],"timeout":0,"maxtrys":3}'
```

### 创建指定用户钱包接口 
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" http://127.0.0.1:95/man/wallet/address/info -d '{"user_id":0, "chain":0}'
```

### 创建指定用户钱包接口 
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" http://127.0.0.1:95/man/currency/rmbratio/set -d '{"from_type":"yBt", "ratio":4}'
```

### rmb充值通知接口 
```
curl -H"Content-Type: application/json" -H"X-Yf-Maner: system" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" http://127.0.0.1:95/man/rmb/recharge/notify -d '{"recharge_id":2}'
```



### 释放某天ybt和kt接口 
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Maner: system" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" 'http://127.0.0.1:95/man/ybt/reward/check' -d '{"date":"2018-10-25"}'	

curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Maner: system" 'http://127.0.0.1:95/man/kt/reward/check'  -d '{"date":"2018-11-04"}'

curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Maner: system" 'http://127.0.0.1:95/man/kt/reward/seller/check'  -d '{"date":"2018-10-24"}'
```

### 后台提币帐号接口(只能提系统帐号)
```
curl -H"Content-Type: application/json" -H"X-Yf-Maner: system" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" http://127.0.0.1:95/man/user/wallet/draw -d '{"user_id":3, "amount":10, "to_user_id":51887,  "tx_type":3, "comment":"从回购转出10个nsetT到YBT专区KT帐户"}'
```

update yunbay_asset set total_perynbay=(select sum(perynbay) from yunbay_asset_detail where date<='2018-11-07') where date='2018-11-07';


### 交易流水查询接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Maner: system" -H"X-Yf-AppId: ybasset" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" 'http://127.0.0.1:95/man/tradeflow/list?country=1'
```

### 内盘充值snet接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Maner: system" -H"X-Sign: ybasset" -H"X-Ts: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" http://127.0.0.1:95/v1/wallet/snet/recharge -d '{"platform":"miner", "order_id":"456456423sdfsdf", "user_id":51884, "amount":20000}'
```

### 链上充值snet接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Maner: system" -H"X-Sign: ybasset" -H"X-Ts: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" http://127.0.0.1:95/man/wallet/recharge/callback -d '[{"symbol":"snet", "tx_hash":"Adsfidjfi", "user_id":"24", "coin_address":"0xbedde7f3340cb3aaba48f2ea48fa792dc4a633d7","amount":100000}]'
```


### 国内版提现处理接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Country: 1" -H"X-Yf-Maner: system" -H"X-Sign: ybasset" -H"X-Ts: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" http://127.0.0.1:95/man/wallet/draw/set -d '{"id":16026080, "success":false, "reason":"金额太大"}'
```

### 挖矿难度接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Country: 1" -H"X-Yf-Maner: system" -H"X-Sign: ybasset" -H"X-Ts: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e"  'http://127.0.0.1:95/v1/ybasset/difficult'
```

### 获取平台每日及累积数据接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Country: 1" -H"X-Yf-Maner: system" -H"X-Sign: ybasset" -H"X-Ts: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e"  'http://127.0.0.1:95/man/ybasset/list'
```

### 获取平台扫码信息接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Country: 1" -H"X-Yf-Maner: system" -H"X-Sign: ybasset" -H"X-Ts: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e"  'http://127.0.0.1:95/v1/qrcode/query?code_str=0x37140d2699c13095573dbaea7fb48981c44e62c2"}'
```

### 平台用户内部转帐接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Country: 1" -H"X-Yf-Maner: system" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNJMCCIMY8yPLWJWQwuj1KARdusu.8d7" http://127.0.0.1:95/v1/user/wallet/transfer -d '{"user_id":51869, "type":1, "amount":500, "zjpassword":"c2ab38482d05fa55b181e0df92adb0bf06543a4a"}'
```

### 平台用户内部转帐接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Country: 1" -H"X-Yf-Maner: system" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNJMCCIMY8yPLWJWQwuj1KARdusu.8d7" 'http://127.0.0.1:95/man/currency/ratio?from=2&to=4'
```

### 代金券充值接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Country: 1" -H"X-Yf-Maner: system" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNJMCCIMY8yPLWJWQwuj1KARdusu.8d7" 'http://127.0.0.1:95/man/voucher/recharge' -d '{"user_id":51887,"type":4,"amount":10000,"title":"USD 1000 代金券"}'
```

### 代金券消费接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Country: 1" -H"X-Yf-Maner: system" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNJMCCIMY8yPLWJWQwuj1KARdusu.8d7" 'http://127.0.0.1:95/v1/user/voucher/pay' -d '{"voucher_id":2,"pay_amount":4000,"user_id":51849,"zjpassword":"2d50652466e0392a302ef3a8791f35781cae31ea"}'
```

### 代金券消费记录查询接口
```
curl -H"Content-Type: application/json" -H"X-Yf-Country: 1" -H"X-Yf-Maner: system" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" -H"X-Yf-Token: T1gkMkaNJMCCIMY8yPLWJWQwuj1KARdusu.8d7" 'http://127.0.0.1:95/v1/user/voucher/record/list'
```

### 清掉用户资产信息 
```
\c ybasset ybasset
begin;
delete from asset_lock;
delete from user_asset_detail where type<>1 and transaction_type<>0;

delete from kt_bonus_detail;
delete from ybt_unlock_detail;
delete from bonus_ybt_detail;
delete from bonus_kt_detail;
delete from yunbay_asset_detail;
delete from yunbay_asset;
delete from ybt_flow;
delete from ybt_day_flow;
update reward_record set status=0;
delete from ordereward;
delete from recharge_flow where tx_type <>1;
delete from withdraw_flow;
delete from yunbay_asset_pool;
commit;


```


