package client

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {
	//评论相关操作
	routers = []yf.RouterInfo{

		// 购物车相关接口
		{yf.HTTP_POST, "/cart/add", true, true, Cart_Add},
		{yf.HTTP_POST, "/cart/update", true, true, Cart_Update},
		{yf.HTTP_POST, "/cart/del", true, true, Cart_Del},
		{yf.HTTP_GET, "/cart/list", true, true, Cart_List},

		// 订单相关接口
		{yf.HTTP_GET, "/order/rebat/info", true, false, Order_RebatInfo},
		{yf.HTTP_GET, "/order/info", true, true, Orders_Info},
		{yf.HTTP_GET, "/order/card", true, true, Orders_Card},
		{yf.HTTP_GET, "/order/list", true, true, Orders_List},
		{yf.HTTP_GET, "/order/list_by_product", true, false, Orders_ListByProduct},

		{yf.HTTP_GET, "/order/count", true, true, Orders_Count},
		{yf.HTTP_GET, "/order/seller/count", true, true, Orders_SellerCount},
		{yf.HTTP_GET, "/order/aftersale/list", true, true, Orders_AfterSaleList},
		{yf.HTTP_GET, "/order/seller/aftersale/list", true, true, Orders_SellerAfterSaleList},

		//{yf.HTTP_POST, "/order/shipped", true, true, true, Orders_Shipped},
		{yf.HTTP_POST, "/order/prepay", true, true, Orders_PrePay},             // 创建订单
		{yf.HTTP_POST, "/order/pay", true, true, Order_Pay},                    // 支付订单
		{yf.HTTP_POST, "/order/del", true, true, Order_Del},                    // 删除订单
		{yf.HTTP_POST, "/order/cancel", true, true, Order_Cancel},              // 取消订单
		{yf.HTTP_POST, "/order/seller/cancel", true, true, Order_SellerCancel}, // 商家取消订单
		//{yf.HTTP_POST, "/order/refund", true, true, true, Order_Refund},

		{yf.HTTP_POST, "/order/finish", true, true, Order_Finish},                      // 完成订单
		{yf.HTTP_POST, "/order/shipped", true, true, Orders_Shipped},                   // 发货接口
		{yf.HTTP_POST, "/order/aftersale", true, true, Order_AfterSale},                // 售后(已弃用)
		{yf.HTTP_GET, "/order/seller/list", true, true, Order_SellerList},              // 商家订单列表
		{yf.HTTP_GET, "/seller/order/list/report", true, true, Order_SellerSearchList}, // 商家订单导出

		{yf.HTTP_GET, "/of/telcheck", true, false, TelCheck},        // 欧飞话费充值号码检测
		{yf.HTTP_POST, "/of/ret/notify", false, false, OfRetNotify}, // 欧飞充值结果回调
	}
	return
}
