package client

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {
	//评论相关操作
	routers = []yf.RouterInfo{
		// rmg充值渠道相关接口
		// 支付宝相关接口
		{yf.HTTP_POST, "/alipay/trade/pay", true, true, Alipay_pay},         // 用户
		{yf.HTTP_POST, "/alipay/trade/notify", false, false, Alipay_Notify}, // 订单回调通知(ali回调)
		//{common.HTTP_POST, "/alipay/trade/app_notify", false, false, true, Alipay_AppNotify},		// 订单回调通知(app回调)

		// 微信支付相关接口
		{yf.HTTP_POST, "/weixin/trade/pay", true, true, WeixinPay},                      // 生成微信签名串
		{yf.HTTP_POST, "/weixin/trade/notify", false, false, WeixinNotify},              // 微信交易回调
		{yf.HTTP_POST, "/weixin/trade/refund", false, false, WeixinNotify},              // 微信退款接口
		{yf.HTTP_POST, "/weixin/trade/refund/notify", false, false, WeixinRefundNotify}, // 微信退款回调接口

		{yf.HTTP_GET, "/trade/query", true, true, Trade_Query}, // 交易状态查询
		{yf.HTTP_GET, "/bank/query", true, false, Bank_Query},  // 银行卡bid查询
	}
	//common.Routeraddlist(ver, routerinfos, routers)
	return
}
