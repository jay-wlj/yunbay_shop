package dao

import (
	"yunbay/account/common"
	"yunbay/account/db"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func GetAccountById(id int64) (vp *common.Account, err error) {
	var v common.Account
	if err = db.GetDB().Find(&v, "user_id=?", id).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetAccountByTel fail! err=", err)
		return
	}
	vp = &v
	return
}

func GetAccountByIds(ids []int64) (vs []common.Account, err error) {
	vs = []common.Account{}
	if len(ids) > 0 {
		if err = db.GetDB().Preload("Cert").Find(&vs, "user_id in(?)", ids).Error; err != nil {
			glog.Error("GetAccountByIds fail! err=", err)
			return
		}
	}
	return
}

func GetAccountByTel(Cc, Tel string) (vp *common.Account, err error) {
	var v common.Account
	if err = db.GetDB().Find(&v, "cc=? and tel=?", Cc, Tel).Error; err != nil {
		glog.Error("GetAccountByTel fail! err=", err)
		return
	}
	vp = &v
	err = nil
	return
}

// 检测用户名是否已经被使用
func CheckUsernameExist(user_name string) (exist bool, err error) {
	var count int64
	if err = db.GetDB().Model(&common.Account{}).Where("user_name=?", user_name).Count(&count).Error; err != nil {
		glog.Error("CheckUsernameExist fail! err=", err)
		return
	}
	exist = count > 0
	return
}

// 获取帐号信息根据用户名
func GetAccountByUsername(user_name string) (vp *common.Account, err error) {
	var v common.Account
	if err = db.GetDB().Find(&v, "user_name=?", user_name).Error; err != nil {
		glog.Error("GetAccountByUsername fail! err=", err)
		return
	}
	vp = &v
	return
}
