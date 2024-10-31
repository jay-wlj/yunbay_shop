
# 生成token
```
curl http://account.nicefilm.com:810/YBAccount/man/token?user_id=380
```
# 播单
----



### 后台指定用户退出登录
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" 'https://account.yunbay.com/man/YBAccount/logout' -d '{"user_id":117619}'
```

### 后台发送短信通知接口
```
curl -H"Content-Type: application/json" -H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" -H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" -H"X-Yf-Token: T1gkMkaNIeCCINb5aIe29RQ2gYujcP4Nz9.bf6" 'https://account.yunbay.com/man/YBAccount/sms/send' -d '{"user_id":51887, "content":"你好啊, jayden"}'
```
