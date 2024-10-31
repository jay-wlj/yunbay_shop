
# 生成token
```
curl http://account.nicefilm.com:810/account/man/token?user_id=380
```

### 重新加载缓存数据
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: ybgoods" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 62361670a0b60c852fcc1e69189c233e" 'http://127.0.0.1:96/man/cache/reload' -d ''

