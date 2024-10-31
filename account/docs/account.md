# 登录服务
----

# 签名AppKey
* AppId: account
* AppKey： 16c86816ab0cfa1493da230cdf356476
* [签名方法](./Sign.md)

# 协议说明

请求及响应体均采用json格式。响应格式为ok_json。
请求成功返回200,失败返回4xx,5xx并返回相应的JSON。
帐号服务，采用HTTPS进行通讯。

# 域名及测试环境

* account.betterchian.io

### ok_json格式说明：

ok_json主要有三个字段，ok, reason, data。其中：
* ok 为true表示服务成功。ok为false表示服务失败。
* reason中是失败原因(当ok=false时)，或为空(ok=true时)。
* data 为接口返回的数据，具体接口会不一样。可为空。
* 示例：

```
{"ok": true,
"reason": "失败原因",
"data": {}}
```

以下接口中，响应数据只列出data部分，ok_json部分不再列出。

* 公共reason值：
    * ERR_ARGS_INVALID 请求参数错误：缺少字段，或字段值不对。
    * ERR_SERVER_ERROR 服务器内部错误：如访问数据库出错了
    * ERR_SIGN_ERROR 签名错误。
    * ERR_TOKEN_INVALID Access-Token非法。
    * ERR_TOKEN_EXPIRED token已经过期


# 短信接口
### 发送验证码(后台使用)
* URL: https://ip:2000/v1/sms/send
* 是否需要签名：是
* 是否需要登录：否
* 请求参数(POST参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
cc | string | 是 | 手机国家代码, 若为空,默认为86, 表示中国.
tel | string | 否 | 手机号。
code | string | 否 | 发送的验证码
expires | int | 否 | 验证码过期时长(秒)
	

* 响应：

OK_JSON

* reason值：
	* ERR\_TEL\_INVALID 手机号无效，无法成功发送验证码。
	* ERR\_TEL\_LIMIT: 短信发送超出限制。每天5条。
	* ERR\_TEL\_EXIST 手机号已经注册。`type为0时，但手机号已经存在了。`
	* ERR\_TEL\_NOT\_EXIST 手机号不存在(重设密码，找回密码时) `type为1时，但手机号还未注册。`

* 示例: 

```
curl -H"Content-Type: application/json" \
http://127.0.0.1:2000/v1/sms/send \
-d '{"cc": "86", "tel": "10020003000", "code": "1234"}'
```

### 验证短信验证码(后台使用)
* URL: https://ip:2000/v1/sms/check
* 是否需要签名：是
* 是否需要登录：否
* 请求参数(POST参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
cc | string | 是 | 手机国家代码, 若为空,默认为86, 表示中国.
tel | string | 否 | 手机号。
code | string | 否 | 发送的验证码

* 响应：

OK_JSON

如果未返回任何错误码,表示校验成功.

* reason值：
	* ERR\_CODE\_INVALID 验证码无效

* 示例: 

```
curl -H"Content-Type: application/json" \
http://127.0.0.1:2000/v1/sms/check \
-d '{"cc": "86", "tel": "10020003000", "code": "1234"}'
```

### 查询某一个手机的验证码(测似使用)
* URL: https://ip:2000/v1/sms/get/code
* 是否需要签名：是
* 是否需要登录：否
* 请求参数(POST参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
cc | string | 是 | 手机国家代码, 若为空,默认为86, 表示中国.
tel | string | 否 | 手机号。
key | string | 否 | 查询使用的超级KEY/密码: 4cc78ae51849277c030ea78337614e31


* 响应：

验证码


* 示例: 

```
curl -H"Content-Type: application/json" \
http://127.0.0.1:2000/v1/sms/get/code\?cc\=86\&tel\=10020003006\&key\=4cc78ae51849277c030ea78337614e31
```


# 登录接口设计

### 密码加密(摘要)方法
密码在客户端进行摘要后，再传输到服务器上。摘要算法是：

```
$password_hash = HEX(SHA1($magic "|" + $password))

其中：
$magic值为：0f82393x
$password 是原始密码

$password_hash 是摘要后的密码，将用于传输。
```

### 发送验证码
* URL: https://account.yunbay.com/v1/YBAccount/sms/send
* 是否需要签名：是
* 是否需要登录：否
* 请求参数(POST参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
cc | string | 是 | 手机国家代码, 若为空,默认为86, 表示中国.
tel | string | 否 | 手机号。
devid | string | 是 | 设备唯一ID，比如: android_id.
type | int | 是 | 短信类型，默认是0： 0: 注册新用户，1：重设密码

* 响应：

OK_JSON

* reason值：
	* ERR\_TEL\_INVALID 手机号无效，无法成功发送验证码。
	* ERR\_TEL\_LIMIT: 短信发送超出限制。每天5条。
	* ERR\_TEL\_EXIST 手机号已经注册。`type为0时，但手机号已经存在了。`
	* ERR\_TEL\_NOT\_EXIST 手机号不存在(重设密码，找回密码时) `type为1时，但手机号还未注册。`

```
对于ERR_TEL_LIMIT错误发生时，每天的短信次数，在data.max_sms_per_day字段中给出。
```

* 示例

```
curl -H"Content-Type: application/json" \
http://127.0.0.1:2000/v1/YBAccount/sms/send \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
-d '{"cc": "86", "tel": "13480660725"}'
```

### 验证码校验接口
* URL: https://account.yunbay.com/v1/YBAccount/sms/check
* 是否需要签名：是
* 是否需要登录：否
* 请求参数(POST参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
cc | string | 是 | 手机国家代码, 若为空,默认为86, 表示中国.
tel | string | 否 | 手机号。
code | string | 是 | 短信验证码

* 响应：

OK_JSON

* reason值：
	* ERR_CODE_INVALID 验证码错误

* 示例

```
curl -H"Content-Type: application/json" \
http://127.0.0.1:2000/v1/YBAccount/sms/check \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
-d '{"cc": "86", "tel": "10020003002", "code": "019839"}'
```


### 注册接口(重设密码)

注册帐号同时带登录功能（注册成功后，会同时返回Token）。

* URL: https://account.yunbay.com/v1/YBAccount/reg
* 是否需要签名：是
* 是否需要登录：否
* 请求参数(POST参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
cc | string | 是 | 手机国家代码, 若为空,默认为86, 表示中国.
tel | string | 否 | 手机号。
password | string | 否 | 密码(hash值)
code | string | 否 | 验证码
reset_pwd | bool | 是 | 是否是重设密码。默认为false

* 响应：

OK_JSON

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
user_id | bigint | 否 | 帐号ID
token | string | 否 | 会话令牌。

* reason值：
    * ERR\_CODE\_INVALID 验证码错误。
    * ERR\_CODE\_ERR\_LIMIT 验证码连接错误次数，超过限制。(需要重新获取验证码)
    * ERR\_TEL\_EXIST 手机号已经注册。
    * ERR\_TEL\_NOT\_EXIST 手机号不存在(重设密码，找回密码时)

```
对于ERR_CODE_ERR_LIMIT错误发生时，连续出错次数，在data.max_err_times字段中给出。
```
* 示例(reset)

```
curl -H"Content-Type: application/json" \
http://127.0.0.1:2000/v1/YBAccount/reg \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
-d '{"cc": "86", "tel": "10020003001", "reset_pwd": true, "password": "123456", "code": "513903"}'
```

* 示例(reg)

```
curl -H"Content-Type: application/json" \
http://127.0.0.1:2000/v1/YBAccount/reg \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
-d '{"cc": "86", "tel": "10020003002", "reset_pwd": false, "password": "123456", "code": "402924"}'
```

### 绑定区块链账号

注册帐号后, 自动创建并绑定区块链账号.

* URL: https://account.yunbay.com/v1/YBAccount/bind/chain_account
* 是否需要签名：是
* 是否需要登录：否
* 请求参数(POST参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
id | int64 | 否 | 需要进行绑定的账号.

* 响应：

OK_JSON

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
chain_name | string | 否 | 绑定的账号名.

* reason值：
   

* 示例

```
curl -H"Content-Type: application/json" \
http://127.0.0.1:2000/v1/YBAccount/bind/chain_account \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
-d '{"id": 1}'
```


### 登录接口(密码登录)
* URL: https://account.yunbay.com/v1/YBAccount/login
* 是否需要签名：是
* 是否需要登录：否
* 请求参数(POST参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
cc | string | 是 | 手机国家代码, 若为空,默认为86, 表示中国.
tel | string | 否 | 手机号。
password | string | 否 | 密码(hash值)
devid | string | 否 | 设备ID.(如android_id等)

* 响应：

OK_JSON

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
user_id | bigint | 否 | 帐号ID
token | string | 否 | 会话令牌。

* reason 值：
    * ERR_PASSWORD_ERR 用户名或密码错误。
    * ERR_TEL_NOT_EXIST 手机号不存在.

* 示例(reg)

```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
http://127.0.0.1:2000/v1/YBAccount/login \
-d '{"cc": "86", "tel": "10020003002", "password": "123456", "devid": "513903"}'
```

### 登出接口（注销登录）
* URL: https://account.yunbay.com/v1/YBAccount/logout
* 是否需要签名：是
* 是否需要登录：是
* 请求参数(POST参数)：

无(当POST请求，无任何参数时，Body传：`{}`)

* 响应：

OK_JSON

* 示例(reg)

```
curl -XPOST -H"Content-Type: application/json" \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
-H"X-Yf-Token: T1gkN0dYZLEiUObZnfZBRcdgJphxrw.f0f" \
http://127.0.0.1:2000/v1/YBAccount/logout
```

### 检查用户名，是否被占用
* URL: https://account.yunbay.com/v1/YBAccount/check_username
* 是否需要签名：是
* 是否需要登录：是
* 请求参数(GET参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
username | string | 否 | 用户名

* 响应：

OK_JSON

* 错误码：
	* ERR_USERNAME_EXIST username已经被占用，需要重新设置。

`如果返回ok: true 表示未被占用，如果返回ok: false,并且reason=ERR_USERNAME_EXIST，则表示被占用。`

* 示例(reg)

```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
-H"X-Yf-Token: T1gkN0dYZLEiUObZTqZCIHSKOMbxSP.431" \
http://127.0.0.1:2000/v1/YBAccount/check_username?username=123456
```

### 用户信息设置接口
* URL: https://account.yunbay.com/v1/YBAccount/userinfo/set
* 是否需要签名：是
* 是否需要登录：是
* 请求参数(POST参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
avatar | string | 否 | 头像URL
username | string | 否 | 用户名
sex | int | 否 | 性别: 0：未知, 1: 女. 2：男. 
birthday | uint32 |是| 出生年月日,unix时间戳
motto | string | 是 | 个人签名(座右铭)
auto | bool | 是 | 是否是自动设置，默认为false。

auto的值：当采用第三方登录成功后，自动设置时，需要设置成true

* 响应：

OK_JSON

* 错误码：
	* ERR_USERNAME_EXIST username已经被占用，需要重新设置。
	* ERR_USERNAME_INVALID username 非法, 用户名,只能是字母,数字及下划线组成, 长度必须在3~32位之间.
* 示例

```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
-H"X-Yf-Token: T1gkN2dYZIRHULOcuLZGZSXaBfJmOE.dbb" \
http://127.0.0.1:2000/v1/YBAccount/userinfo/set \
-d '{"avatar":"头像URL","username":"lxj", "sex": 0,
"birthday": 234242432, "motto": "lksadfa中国"}'
```	



### 用户信息获取接口
* URL: https://account.yunbay.com/v1/YBAccount/userinfo/get
* 是否需要签名：是
* 是否需要登录：是
* 请求参数(GET参数)：无

* 响应：

OK_JSON

* UserInfo(后面协议会引用到)

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
user_id |bigint | 否 | 用户ID
tel | string | 是 | 手机号。
avatar | string | 是 | 头像URL
username | string | 是 | 用户名
sex | int | 是 | 性别: 0: 女，1：男 
birthday | uint32 |是| 出生年月日,格式为：20160304
motto | string | 是 | 个人签名(座右铭)
weixin_id | string | 是 | 微信帐号ID(第三方) 
qq_id | string | 是 | QQ帐号ID(第三方)
weibo_id | string | 是 | 微博ID(第三方)

* 示例

```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
-H"X-Yf-Token: T1gkN0dYZLEiUObZTqZCIHSKOMbxSP.431" \
http://127.0.0.1:2000/v1/YBAccount/userinfo/get 
```	


### 他人信息获取接口（匿名获取）
* URL: https://account.yunbay.com/v1/YBAccount/userinfo/get/any
* 是否需要签名：是
* 是否需要登录：否
* 请求参数(GET参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
user_id |bigint | 否 | 用户ID

* 响应：

OK_JSON

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
user_id |bigint | 否 | 用户ID
avatar | string | 是 | 头像URL
username | string | 是 | 用户名
sex | int | 是 | 性别: 0：女. 1：男. 2: 未知
age | int | 是 | 年龄。
motto | string | 是 | 个人签名(座右铭)

* reason值
	* ERR_USER_NOT_EXIST

* 示例

```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
http://127.0.0.1:2000/v1/YBAccount/userinfo/get/any\?user_id\=5
```

### 校验Token信息

* URL: https://account.yunbay.com/v1/YBAccount/token/check
* 是否需要签名：是
* 是否需要登录：是
* 请求参数(GET参数)：无

* 响应：

OK_JSON

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
user_id | bigint | 否 | 帐号ID

* 校验不通过会返回(其它需要校验Token的接口也会返回这些值)：
	* ERR_TOKEN_INVALID 非法的Token
	* ERR_TOKEN_EXPIRED Token已经过期

* 示例

```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
-H"X-Yf-Token: T1gkN0dYZLEiUObZTqZCIHSKOMbxSP.431" \
http://127.0.0.1:2000/v1/YBAccount/token/check 
```

### 获取用户信息(管理后台用)
* URL: http://account.yunbay.com:92/man/YBAccount/userinfo/get
* 是否需要签名：否
* 是否需要登录：否
* 请求参数(GET参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
user_ids | string | 否 | 用户ID列表(以逗号分割)


* 响应：

OK_JSON


字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
users | list(UserInfo) | 否 | 用户信息列表。(不在列表中的用户，表示未查到) 

### 查找用户
* URL: https://account.yunbay.com/v1/YBAccount/search
* 是否需要签名：是
* 是否需要登录：否
* 请求参数(GET参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
cc |string | 是 | 根据电话号码查询时需要
tel | string | 是 | 根据电话号码查询时需要
username | string | 是 | 根据用户名查询

* 响应：

OK_JSON

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
users|list<UserInfo> | 是 | 用户列表.


* 示例

```
curl -H"Content-Type: application/json" \
-H"X-Yf-AppId: account" -H"X-Yf-rid: 1" -H"X-Yf-Platform: test" -H"X-Yf-Version: 2.8.1" \
-H"X-Yf-Sign: 69c5c1c89f9f6093559af661bc4e4df1" \
http://127.0.0.1:2000/v1/YBAccount/search?username=xxx
```

### 根据用户名查用户信息（管理后台用）
* URL: http://account.yunbay.com:92/man/YBAccount/userinfo/search
* 是否需要签名：否
* 是否需要登录：否
* 请求参数(GET参数)：

字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
name | string | 否 | 用户名（模糊查询）
type | int | 是 | 0：查询所有用户，1：只查询内部用户。

* 响应：

OK_JSON


字段 | 类型 | 是否可为空 | 说明
--- | --- | ------- | ----
users | list(UserInfo) | 否 | 用户信息列表。(不在列表中的用户，表示未查到) 
