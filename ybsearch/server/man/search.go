package man

import (
	"yunbay/utils"
	"yunbay/ybsearch/server/share"

	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

/**
 * 在搜索索引中隐藏商品
 */
type ids struct {
	Id     uint64 `json:"id"`
	Status int    `json:"status"`
}

func Hid(c *gin.Context) {
	var req ids
	if ok := utils.UnmarshalBodyAndCheck(c, &req); !ok {
		return
	}
	if err := share.GetSphinx().Hid(req.Id, req.Status); err != nil {
		glog.Error("Hid fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

func Status(c *gin.Context) {
	var req ids
	if ok := utils.UnmarshalBodyAndCheck(c, &req); !ok {
		return
	}
	if err := share.GetSphinx().Hid(req.Id, req.Status); err != nil {
		glog.Error("Hid fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}
