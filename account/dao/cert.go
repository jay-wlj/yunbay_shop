package dao

import (
	"yunbay/account/common"

	"yunbay/account/db"

	"github.com/jie123108/glog"
)

// 检测身份证号是否已存在
func ChecCardIdExist(card_id string) (exist bool, err error) {
	var count int64
	if err = db.GetDB().Model(&common.Cert{}).Where("card_id=?", card_id).Count(&count).Error; err != nil {
		glog.Error("CheckUsernameExist fail! err=", err)
		return
	}
	exist = count > 0
	return
}

// 获取用户的实名信息
func GetCertByUserId(user_id int64) (v *common.Cert, err error) {
	var m common.Cert
	if err = db.GetDB().Find(&m, "user_id=?", user_id).Error; err != nil {
		glog.Error("GetCertByUserId fail! err=", err)
		return
	}
	v = &m
	return
}
