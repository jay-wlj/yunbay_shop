package man

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {
	//评论相关操作
	routers = []yf.RouterInfo{

		// 订单相关接口
		{yf.HTTP_POST, "/order/pay", true, false, Orders_Pay},
		{yf.HTTP_GET, "/order/info", true, false, Orders_Info},
		{yf.HTTP_GET, "/order/list", true, false, Orders_List},
		{yf.HTTP_POST, "/order/cancel", true, false, Orders_Cancel},
		{yf.HTTP_POST, "/order/finish", true, false, Orders_Finish},
		{yf.HTTP_GET, "/order/status/query", true, false, Orders_StatusQuery},
		{yf.HTTP_POST, "/order/auto_deliver", true, false, Orders_AutoDeiver},
		{yf.HTTP_POST, "/order/pay/lotterys", true, false, Orders_CreateByLotterys},

		// 欧飞充值查询
		{yf.HTTP_GET, "/order/of/query", true, false, OfOrderQuery},
		// {yf.HTTP_POST, "/device/contract", true, true, true, UpdateContract},

		// 订单折扣相关接口
		{yf.HTTP_POST, "/order/rebat/update", true, false, Orders_RebatUpdate},

		// 订单表格导出
		{yf.HTTP_GET, "/order/report", false, false, Orders_Report},
	}
	return
}
