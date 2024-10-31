package client

import (
	"yunbay/common"
)

func InitRouter(ver string, routerinfos map[string]common.RouterInfo) {

	routers := []common.RouterInfo{

		{common.HTTP_GET, "/index", true, false, Index}, //

	}
	common.Routeraddlist(ver, routerinfos, routers)
}
