package man

import (
	"yunbay/ybeos/common"
)

func InitManagerRouter(ver string, routerinfos map[string]common.RouterInfo) {

	routers := []common.RouterInfo{
		// alipay
		//{common.HTTP_POST, "/transaction/push", true, false, false, Transaction_Push},     // 发送一笔交易
		{common.HTTP_POST, "/transaction/push", true, false, false, Transaction_Push}, // 发送一笔交易
	}
	common.Routeraddlist(ver, routerinfos, routers)
}
