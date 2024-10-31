package client

import (
	"yunbay/ybim/conf"
	"yunbay/ybim/server/share"
	"yunbay/ybim/util"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

// 获取用户im token
func IMGetToken(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	if !conf.Config.IMEanble {
		yf.JSON_Fail(c, yf.DATA_NOT_SUPPORT)
		return
	}
	v, err := share.GetUserIMToken(user_id)
	if err != nil {
		glog.Error("IMCreateAccount fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"im_token": v.Token, "accid": v.ImId})
}
