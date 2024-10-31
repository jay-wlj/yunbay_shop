package client

import (
	"yunbay/account/common"
)

func InitRouter(ver string, routerinfos map[string]common.RouterInfo) {
	//评论相关操作

	account := &Account{}
	imgcode := &ImgageCode{}
	routers := []common.RouterInfo{
		// 图片验证码
		{common.HTTP_GET, "/imgcode/get", false, false, imgcode.GetImgCode},    // 交易状态查询
		{common.HTTP_GET, "/imgcode/check", true, false, imgcode.CheckImgCode}, // 银行卡bid查询

		// 短信验证码
		{common.HTTP_POST, "/account/sms/send", true, false, SmsCodeSend},    // 发送验证码
		{common.HTTP_POST, "/account/sms/usersend", true, true, SmsUsersend}, // 登录状态下发送验证码
		{common.HTTP_POST, "/account/sms/check", true, false, SmsCodeCheck},  // 效验验证码

		// 登录相关接口
		{common.HTTP_POST, "/account/reg", true, false, account.Reg}, // 注册
		{common.HTTP_POST, "/account/login", true, false, account},   // 登录
		{common.HTTP_POST, "/account/logout", true, true, account},   // 退出

		{common.HTTP_POST, "/account/reset/login_pwd", true, true, account.ResetLoginPwd}, // 重设登录密码
		{common.HTTP_POST, "/account/reset/zjpwd", true, true, account.ZJReset},           // 重设资金密码
		{common.HTTP_GET, "/account/token/check", true, true, account.TokenCheck},         // token效验
		{common.HTTP_POST, "/account/zfauth", true, true, account.ZFAuth},                 // 效验支付密码

		{common.HTTP_GET, "/account/check_username", true, false, account.CheckUsername},               // 检测用户名
		{common.HTTP_POST, "/account/userinfo/set/username", true, false, account.UserinfoSetUsername}, // 设置用户名

		{common.HTTP_POST, "/account/userinfo/auth", true, true, account.UserinfoAuth},  // 实名
		{common.HTTP_POST, "/account/userinfo/set", true, true, account.UserinfoSet},    // 设置用户信息
		{common.HTTP_GET, "/account/userinfo/get", true, true, account.UserinfoGet},     // 获取用户信息
		{common.HTTP_GET, "/account/userinfo/other", true, true, account.UserinfoOther}, // 获取用户信息根据用户id
		{common.HTTP_GET, "/account/login/record", true, true, account.LoginRecord},     // 获取历史登录记录

		{common.HTTP_GET, "/third/youbuy/account", true, true, account.Third_YoubuyAccount}, // 获取优买会帐号id

	}
	common.Routeraddlist(ver, routerinfos, routers)
}
