package man

import (
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"time"
	"yunbay/ybim/common"
	"yunbay/ybim/server/share"
	"yunbay/ybim/util"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

type uidSt struct {
	UserId int64 `json:"user_id"`
}

// 注册im 用户信息
func RegisterIMUser(c *gin.Context) {
	var req uidSt
	if ok := util.UnmarshalBodyAndCheck(c, &req); !ok {
		return
	}

	// 直接返回
	yf.JSON_Ok(c, gin.H{})
	return
	v, err := share.GetUserIMToken(req.UserId)
	if err != nil {
		glog.Error("IMCreateAccount fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"user_id": v.UserId, "accid": v.ImId, "im_token": v.Token})
}

// 更新im 用户信息
func UpdateIMInfo(c *gin.Context) {
	var req uidSt
	if ok := util.UnmarshalBodyAndCheck(c, &req); !ok {
		return
	}
	// 直接返回
	yf.JSON_Ok(c, gin.H{})
	return

	err := share.UpdateIMInfo(req.UserId)
	if err != nil {
		glog.Error("UpdateIMInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

// 更新所有im用户的用户信息
func UpdateAllIMInfo(c *gin.Context) {
	start_id, _ := base.CheckQueryInt64Field(c, "start_user_id")
	db := db.GetTxDB(c)
	uids := []uidSt{}

	t := time.Now()
	if err := db.Model(&common.IMToken{}).Select("user_id").Where("user_id>=?", start_id).Order("user_id asc").Scan(&uids).Error; err != nil {
		glog.Error("UpdateAllIMInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	for _, v := range uids {
		if err := share.UpdateIMInfo(v.UserId); err != nil {
			glog.Error("UpdateAllIMInfo fail! err=", err)
			yf.JSON_Fail(c, fmt.Sprintf("user_id=%v", v.UserId))
			return
		}
	}
	yf.JSON_Ok(c, gin.H{"tick": time.Since(t).String()})
}

// 发送消息给用户
func MsgSend(c *gin.Context) {
	var req share.MsgReq
	if ok := util.UnmarshalBodyAndCheck(c, &req); !ok {
		return
	}

	err := req.Send()
	if err != nil {
		glog.Error("MsgSend fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

// 查询im 用户信息
func GetIMUserInfo(c *gin.Context) {
	str_user_ids, _ := base.CheckQueryStringField(c, "user_ids")

	user_ids := base.StringToInt64Slice(str_user_ids, ",")

	ret, err := util.QueryIMUInfo(user_ids)
	if err != nil {
		glog.Error("GetIMUserInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, ret)
}

// 查询单独历史消息
func MsgQuery(c *gin.Context) {

}
