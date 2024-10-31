package man

import (
	"yunbay/ybim/server/share"

	"github.com/gin-gonic/gin"
)

// 获取用户im token
func IMWebsocket(c *gin.Context) {
	//user_id, _ := base.CheckQueryInt64Field(c, "user_id")
	// if err != nil {
	// 	c.JSON(401, gin.H{"ok": false, "reason": yf.ERR_TOKEN_INVALID})
	// 	return
	// }
	// user_id, ok := util.GetUid(c)
	// if !ok {
	// 	return
	// }

	// m["user_id"] = user_id
	keys := make(map[string]interface{})
	keys["platform"] = "man"

	// websocket 链接
	share.GetWsMgr().HandleRequestWithKeys(c.Writer, c.Request, keys)
}
