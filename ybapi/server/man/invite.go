package man

import (
	"fmt"
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybapi/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

type inviteData struct {
	UserId       int64  `json:"user_id" binding:"gt=0"`
	Tel          string `json:"tel"` // 用户号码
	FromInviteId int64  `json:"from_inviteid"`
}

func ManInvite_Add(c *gin.Context) {
	var args inviteData
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	db := db.GetTxDB(c)
	now := time.Now().Unix()
	if args.FromInviteId > 0 {
		v := common.Invite{UserId: args.FromInviteId, Type: 0, InviteUserId: args.UserId, InviteTel: args.Tel}
		if err := db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (user_id,invite_userid) DO update set update_time=%v", now)).Save(&v).Error; err != nil {
			glog.Error("ManInvite_Add fail! err=", err)
			return
		}

		// 添加推荐人入库
		fv := common.Invite{UserId: args.UserId, Type: 1}
		fv.FromInviteIds = append(fv.FromInviteIds, args.FromInviteId)
		var ffv common.Invite
		if err := db.First(&ffv, "invite_userid=?", args.FromInviteId).Error; err == nil {
			if ffv.UserId > 0 {
				fv.FromInviteIds = append(fv.FromInviteIds, ffv.UserId)
			}
		}
		if err := db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (user_id,invite_userid) DO update set recommend_userids=array[%v]::bigint[]", base.Int64SliceToString(fv.FromInviteIds, ","))).Save(&fv).Error; err != nil {
			glog.Error("ManInvite_Add fail! err=", err)
			return
		}
	}

	yf.JSON_Ok(c, gin.H{})
}

func ManInvite_BeInvite(c *gin.Context) {
	user_id, _ := base.CheckQueryInt64Field(c, "user_id")
	if user_id < 1 {
		glog.Error("ManInvite_BeInvite fail! user_id <1")
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	recommend_userids, err := GetUserBeInvite(user_id)
	if err != nil {
		glog.Error("ManInvite_BeInvite fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"recommend_userids": recommend_userids})
}

// 获取一批用户的直接邀请者
func ManInvite_BeInvites(c *gin.Context) {
	str_user_ids, _ := base.CheckQueryStringField(c, "user_ids")
	user_ids := base.StringToInt64Slice(str_user_ids, ",")
	mids, err := GetUserBeInvites(user_ids)
	if err != nil {
		glog.Error("ManInvite_BeInvite fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"inviters": mids})
}

func ManInvite_List(c *gin.Context) {
	user_id, _ := base.CheckQueryIntField(c, "user_id")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	if user_id < 1 {
		glog.Error("ManInvite_BeInvite fail! user_id<1")
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	vs := []common.Invite{}
	if err := db.GetDB().Where("user_id=? and type=0", user_id).Order("create_time desc").Limit(page_size).Offset((page - 1) * page_size).Find(&vs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("ManInvite_BeInvite fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
}

func GetUserBeInvite(user_id int64) (recommend_userids []int64, err error) {
	var v common.Invite
	recommend_userids = []int64{}
	if err = db.GetDB().Where("user_id=? and type=1", user_id).Find(&v).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("ManInvite_BeInvite fail! err=", err)
		return
	}
	err = nil
	recommend_userids = v.FromInviteIds
	return
}

func GetUserBeInvites(user_ids []int64) (mids map[int64][]int64, err error) {
	var vs []common.Invite
	mids = make(map[int64][]int64)
	if err = db.GetDB().Where("user_id in(?) and type=1", user_ids).Find(&vs).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("ManInvite_BeInvite fail! err=", err)
		return
	}
	err = nil
	for _, v := range vs {
		mids[v.UserId] = v.FromInviteIds
	}
	return
}
