package client

import (
	"yunbay/ybim/server/share"
	"yunbay/ybim/util"

	"github.com/gin-gonic/gin"
)

// 获取用户im token
func IMWebsocket(c *gin.Context) {
	// token, _ := base.CheckQueryStringField(c, "token")
	// user_id, _, _, err := yf.TokenCheck(token)
	// if err != nil {
	// 	c.JSON(401, gin.H{"ok": false, "reason": yf.ERR_TOKEN_INVALID})
	// 	return
	// }
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}

	platform, _ := util.GetPlatformVersionByContext(c)
	keys := make(map[string]interface{})
	keys["platform"] = platform
	keys["user_id"] = user_id
	// websocket 链接
	share.GetWsMgr().HandleRequestWithKeys(c.Writer, c.Request, keys)
}

func IMWebsocketWeb(c *gin.Context) {
	// websocket 链接
	keys := make(map[string]interface{})
	keys["platform"] = "web"
	share.GetWsMgr().HandleRequestWithKeys(c.Writer, c.Request, keys)
}
