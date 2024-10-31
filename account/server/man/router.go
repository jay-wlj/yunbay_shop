package man

import (
	"yunbay/account/common"
)

func InitManagerRouter(ver string, routerinfos map[string]common.RouterInfo) {

	account := &ManAccountController{}
	routers := []common.RouterInfo{
		// 管理后台相关接口
		{common.HTTP_POST, "/account/logout", true, false, account.Logout},                  // 退出登录
		{common.HTTP_POST, "/account/sms/send", true, false, account.SendSms},               // 发送短信码
		{common.HTTP_POST, "/account/sms/send_by_tels", true, false, account.SendSmsByTels}, // 发送短信码
		{common.HTTP_GET, "/account/token/check", false, true, account.TokenCheck},          // 检测token
		{common.HTTP_POST, "/account/usertype", true, false, account.SetUserType},           // 设置用户类型
		{common.HTTP_GET, "/account/userinfo/get", true, false, account.UserinfoGet},        // 获取用户信息
		{common.HTTP_GET, "/account/userinfo/search", true, false, account.UserinfoSearch},  // 搜索用户信息
		{common.HTTP_POST, "/account/sms/code/check", true, false, account.CodeCheck},       // 验证码效验

		{common.HTTP_POST, "/account/smspwd/check", true, true, account.SmsPwdCheck},               // 验证资金密码及手机验证码
		{common.HTTP_POST, "/account/userinfo/auth/check", true, false, account.UserinfoAuthCheck}, // 更新用户实名

		{common.HTTP_GET, "/account/userinfo/auth/list", true, false, account.UserinfoAuthList}, // 获取用户实名认证列表

		{common.HTTP_GET, "/account/userinfo/list", true, false, account.List},               // 获取用户列表
		{common.HTTP_GET, "/third/youbuy/account", true, false, account.Third_YoubuyAccount}, // 获取优买会帐号id
	}
	common.Routeraddlist(ver, routerinfos, routers)
}
