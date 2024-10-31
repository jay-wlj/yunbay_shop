
# 生成token
```
curl 'http://account.yunbay.com:92/man/account/token?user_id=380'
```

### 获取用户IM token接口 
```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" \
-H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" \
http://127.0.0.1:2004/v1/user/token
```

### 注册用户IM帐号 
```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" \
'http://127.0.0.1:2004/man/user/register' \
-d '{"user_id":0}'
```

### 更新IM用户信息
```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" \
'http://127.0.0.1:2004/man/user/info/update' 
-d '{"user_id":51869}'
```

### 更新所有IM用户信息
```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" \
'http://127.0.0.1:2004/man/user/info/update/all?start_user_id=0'
```

### 获取IM用户信息
```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" \
'http://127.0.0.1:2004/man/user/info?user_ids=51869'
```

### 发送消息 
```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" \
http://127.0.0.1:2004/man/msg/send \
-d '{"type":0, "to":[], "id":1, "action":"lotterys_notify","data":{"content":"你是谁?", "lotterys_id":1, "status":2}}'
```

### 建立链接 
```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: ybapi" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" \
-H"X-Yf-Token: T1g0MkaNJMCCILP87fKDVRQ+JRh2uoCnjp.b0a" \
-H"Connection: Upgrade" \
-H"Upgrade: websocket" \
-H"Host: im.yunbay.com" \
-H"Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
-H"Sec-WebSocket-Version: 13" \
http://172.17.10.80:2004/v1/ws?token=T1g0MkaNJMCCILP87fKDVRQ+JRh2uoCnjp.b0a
```