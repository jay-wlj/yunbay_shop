package man

import (
	"yunbay/ybim/common"
)

func InitManagerRouter(ver string, routerinfos map[string]common.RouterInfo) {
	//评论相关操作
	routers := []common.RouterInfo{
		{common.HTTP_POST, "/user/register", true, false, true, RegisterIMUser},
		{common.HTTP_POST, "/user/info/update", true, false, true, UpdateIMInfo},
		{common.HTTP_GET, "/user/info/update/all", true, false, false, UpdateAllIMInfo},
		{common.HTTP_GET, "/user/info", true, false, false, GetIMUserInfo},
		{common.HTTP_POST, "/msg/send", false, false, false, MsgSend},

		{common.HTTP_GET, "/ws", false, false, false, IMWebsocket},
	}
	common.Routeraddlist(ver, routerinfos, routers)
}
