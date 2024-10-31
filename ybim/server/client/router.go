package client

import (
	"yunbay/ybim/common"
)

func InitClientRouter(ver string, routerinfos map[string]common.RouterInfo) {
	//评论相关操作
	routers := []common.RouterInfo{

		{common.HTTP_GET, "/user/token", true, true, true, IMGetToken},

		{common.HTTP_GET, "/ws", false, true, false, IMWebsocket},
		{common.HTTP_GET, "/web", false, false, false, IMWebsocketWeb},
		// {common.HTTP_POST, "/device/contract", true, true, true, UpdateContract},
	}
	common.Routeraddlist(ver, routerinfos, routers)
}
