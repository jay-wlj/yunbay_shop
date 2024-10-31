package share

import (
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"time"
	"yunbay/ybim/common"
	"yunbay/ybim/dao"
	"yunbay/ybim/util"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

type userinfoSt struct {
	UserId   int64  `json:"user_id"`
	UserType int16  `json:"user_type"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Birthday string `json:"birthday"`
}

func GetUserIMToken(user_id int64) (v common.IMToken, err error) {
	// 优先从redis里获取
	v.ImId, v.Token, err = dao.GetImToken(user_id)
	if err == nil && v.Token != "" {
		return
	}
	db := db.GetDB()
	err = db.Find(&v, "user_id=?", user_id).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	if err == gorm.ErrRecordNotFound {
		// 获取用户信息
		var u userinfoSt
		if err = db.Raw("select user_id, user_type, username, avatar, birthday from account where user_id=?", user_id).Scan(&u).Error; err != nil {
			glog.Error("get userinfo ", user_id, ", fail! err=", err)
			return
		}
		ex := make(map[string]interface{})
		ex["user_type"] = u.UserType
		accid, token, err1 := util.GetIMToken(user_id, u.Username, u.Avatar, u.Birthday, 0, ex) // 调用云信IM接口获取im token
		if err1 != nil {
			err = err1
			glog.Error("GetUserIMToken fail! user_id=", user_id, " err=", err)
			return
		}
		if accid != base.Int64ToString(user_id) {
			glog.Error("in accid:", user_id, " out accid:", accid)
		}
		now := time.Now().Unix()
		v = common.IMToken{UserId: user_id, ImId: accid, Token: token, CreateTime: now, UpdateTime: now}
		db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (user_id) DO update set token='%v', update_time=%v", token, now))
		if err = db.Save(&v).Error; err != nil {
			glog.Error("IMToken save fail! err=", err)
			return
		}
	}
	if er := dao.SaveImToken(user_id, v.ImId, v.Token); er != nil {
		glog.Error("SaveImToken fail! err=", err)
	}
	return
}

func UpdateIMInfo(user_id int64) (err error) {
	db := db.GetDB()
	var v common.IMToken
	err = db.Find(&v, "user_id=?", user_id).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	// 获取用户信息
	var u userinfoSt
	if err = db.Raw("select user_id, user_type, username, avatar, birthday from account where user_id=?", user_id).Scan(&u).Error; err != nil {
		glog.Error("get userinfo ", user_id, ", fail! err=", err)
		return
	}
	ex := make(map[string]interface{})
	ex["user_type"] = u.UserType
	err = util.UpdateIMUInfo(user_id, u.Username, u.Avatar, u.Birthday, 0, ex)
	if err != nil {
		glog.Error("UpdateIMInfo fail! user_id=", user_id, " err=", err)
		return
	}
	return
}
