package util

import (
	"github.com/jie123108/glog"
)


type zJPassword struct {
	ZJPassword string `json:"zjpassword"`
}
func AuthUserZJPassword(token, zjpassword string)(err error) {
	uri := "/v1/account/zfauth"
	v := zJPassword{ZJPassword:zjpassword}
	headers := map[string]string{"X-YF-Token":token}
	err = post_info(uri, "account", headers, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("UserZJPasswordAuth fail! err=", err)
		return
	}
	return
}

type userTypeSt struct {
	UserId int64 `json:"user_id"`
	Status int `json:"status"`
}
func SetUserType(user_id int64, status int)(err error) {
	uri := "/man/account/usertype"
	v := userTypeSt{UserId:user_id, Status:status}
	err = post_info(uri, "account", nil, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("SetUserType fail! err=", err)
		return
	}
	return
}