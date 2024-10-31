package man

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {

	routers = []yf.RouterInfo{
		// alipay
		{yf.HTTP_POST, "/alipay/trade/pay", true, false, Alipay_pay},                     // 用户
		{yf.HTTP_POST, "/alipay/trade/refund", true, false, Alipay_Refund},               // 退款
		{yf.HTTP_POST, "/alipay/trade/trasfer", true, false, Alipay_Transfer},            // 支付宝转帐
		{yf.HTTP_POST, "/alipay/trade/trasfer/query", true, false, Alipay_TransferQuery}, // 支付宝转帐

		{yf.HTTP_POST, "/weixin/trade/refund", true, false, Weixin_Refund}, // 退款

		{yf.HTTP_POST, "/trade/close", true, false, Trade_Close}, // 关闭交易
	}
	//common.Routeraddlist(ver, routerinfos, routers)

	return
}
