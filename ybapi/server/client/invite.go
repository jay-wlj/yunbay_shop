package client

import (
	"yunbay/ybapi/common"
	"yunbay/ybapi/util"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	//"time"
)

func Beinvite(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	v := common.Invite{FromInviteIds: pq.Int64Array{}}
	if err := db.GetDB().Where("user_id=? and type=1", user_id).Find(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Beinvite fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"recommend_userids": v.FromInviteIds})
}

func Invite_List(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	vs := []common.Invite{}

	if err := db.GetDB().ListPage(page, page_size).Where("user_id=? and type=0", user_id).Order("create_time desc").Find(&vs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Invite_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	list_ended := true
	if page_size == len(vs) {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
}

func Invite_Count(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}

	// 获取邀请人数
	var total int64 = 0
	if err := db.GetDB().Model(&common.Invite{}).Where("user_id=? and type=0", user_id).Count(&total).Error; err != nil {
		glog.Error("Invite_Count fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"total": total})
}
