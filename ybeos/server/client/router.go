package client

import (
	"yunbay/ybeos/common"
)

func InitRouter(ver string, routerinfos map[string]common.RouterInfo) {
	//评论相关操作
	routers := []common.RouterInfo{
		// rmg充值渠道相关接口
		// 支付宝相关接口

	}
	common.Routeraddlist(ver, routerinfos, routers)
}
