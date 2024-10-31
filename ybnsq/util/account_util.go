package util

import (
	"github.com/jie123108/glog"
)

type userTypeSt struct {
	UserId   int64 `json:"user_id"`
	BrokerId int64 `json:"broker_id"`
	Status   int   `json:"status"`
}

func SetUserType(user_id int64, status int) (err error) {
	uri := "/man/account/usertype"
	v := userTypeSt{UserId: user_id, Status: status}
	err = post_info(uri, "account", nil, v, "", nil, "", false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("SetUserType fail! err=", err)
		return
	}
	return
}

func AddUserAsset(user_id, invit_id int64) (err error) {
	uri := "/man/user/asset/add"
	v := userTypeSt{UserId: user_id, BrokerId: invit_id}
	err = post_info(uri, "ybasset", nil, v, "", nil, "", false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("SetUserType fail! err=", err)
		return
	}
	return
}

type UserAdd struct {
	UserId   int64  `json:"user_id"`   //用户ID
	Tel      string `json:"tel"`       // 用户号码
	BrokerId int64  `json:"broker_id"` //推荐用户ID
}

func AddInvite(v UserAdd) (err error) {
	uri := "/man/invite/add"
	err = post_info(uri, "ybapi", nil, v, "", nil, "", false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("AddInvite fail! err=", err)
		return
	}
	return
}

func RegisterIM(user_id int64) (err error) {
	uri := "/man/user/register"
	v := userTypeSt{UserId: user_id}
	err = post_info(uri, "ybim", nil, v, "", nil, "", false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("RegisterIM fail! err=", err)
		return
	}
	return
}
