package man

import (
	"yunbay/common"
)

func InitRouter(ver string, routerinfos map[string]common.RouterInfo) {
	routers := []common.RouterInfo{
		{common.HTTP_POST, "/hid", true, false, Hid},       // 屏B商品
		{common.HTTP_POST, "/status", true, false, Status}, // 上下架商品
	}
	common.Routeraddlist(ver, routerinfos, routers)
}
