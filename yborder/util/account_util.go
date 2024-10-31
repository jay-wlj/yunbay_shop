package util

import (
	"fmt"

	base "github.com/jay-wlj/gobaselib"

	"github.com/jie123108/glog"
)

type zJPassword struct {
	ZJPassword string `json:"zjpassword"`
}

func AuthUserZJPassword(token, zjpassword string) (err error) {
	uri := "/v1/account/zfauth"
	v := zJPassword{ZJPassword: zjpassword}
	headers := map[string]string{"X-YF-Token": token}
	err = post_info(uri, "account", headers, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("UserZJPasswordAuth fail! err=", err)
		return
	}
	return
}

type userTypeSt struct {
	UserId int64 `json:"user_id"`
	Status int   `json:"status"`
}

func SetUserType(user_id int64, status int) (err error) {
	uri := "/man/account/usertype"
	v := userTypeSt{UserId: user_id, Status: status}
	err = post_info(uri, "account", nil, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("SetUserType fail! err=", err)
		return
	}
	return
}

type Account struct {
	UserId     int64  `json:"user_id"`
	Cc         string `json:"cc"`
	Tel        string `json:"tel"`
	UserType   int16  `json:"user_type"`
	Platform   string `json:"platform"`
	Version    string `json:"version"`
	Username   string `json:"username"`
	Avatar     string `json:"avatar"`
	CertStatus int    `json:"cert_status"`
	CreateTime int64  `json:"create_time"`
}

func GetUserInfoByIds(ids []int64) (m map[int64]interface{}, err error) {
	uri := fmt.Sprintf("/man/account/userinfo/get?user_ids=%v", base.Int64SliceToString(ids, ","))
	m = make(map[int64]interface{})

	vs := []Account{}
	err = get_info(uri, "account", "list", &vs, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("GetUserInfoByIds fail! err=", err)
		return
	}
	for i, v := range vs {
		m[v.UserId] = vs[i]
	}
	return
}

type ThirdAccount struct {
	UserId  int64 `json:"user_id"`
	ThirdId int64 `json:"third_id"`
	//ThirdAccount map[string]interface{} `json:"third_account"`
	CreateTime int64 `json:"create_time"`
}

func GetYoubuyAccount(user_id int64) (m *ThirdAccount, err error) {
	uri := fmt.Sprintf("/man/third/youbuy/account?user_id=%v", user_id)

	var v ThirdAccount
	err = get_info(uri, "account", "", &v, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("GetYoubuyAccount fail! err=", err)
		return
	}
	m = &v
	return
}
