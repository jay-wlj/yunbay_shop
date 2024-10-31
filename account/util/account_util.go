package util

import (
	"fmt"
	"net/url"

	"github.com/jie123108/glog"
)

type zJPassword struct {
	ZJPassword string `json:"zjpassword"`
}

func AuthUserZJPassword(token string, zjpassword string) (err error) {
	uri := "/v1/YBAccount/zfauth"
	v := zJPassword{ZJPassword: zjpassword}

	headers := map[string]string{"X-YF-Token": token}
	err = post_info(uri, "account", headers, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("UserZJPasswordAuth fail! err=", err)
		return
	}
	return
}

type codeParms struct {
	Code string `json:"code"`
}

func AuthSmsCode(token string, code string) (err error) {
	uri := "/man/sms/code/check"
	v := codeParms{Code: code}
	headers := map[string]string{"X-YF-Token": token}
	err = post_info(uri, "account", headers, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("UserZJPasswordAuth fail! err=", err)
		return
	}
	return
}

type smspasswordParms struct {
	Code       string `json:"code"`
	ZJPassword string `json:"zjpassword"`
}

// 验证短信码及支付密码
func AuthSmsPasswrod(token string, code, password string) (err error) {
	uri := "/man/YBAccount/smspwd/check"
	v := smspasswordParms{Code: code, ZJPassword: password}
	headers := map[string]string{"X-YF-Token": token}
	err = post_info(uri, "account", headers, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("UserZJPasswordAuth fail! err=", err)
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

func UserInfoGet(token string) (v *Account, err error) {
	uri := "/v1/YBAccount/userinfo/get"
	headers := map[string]string{"X-YF-Token": token}
	var m Account
	err = get_info(uri, "account", headers, "", &m, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("UserInfoGet fail! err=", err)
		return
	}
	v = &m
	return
}

func GetYoubuyAccount(cc, tel string) (v *Account, err error) {
	uri := "/man/account/userinfo/search" + fmt.Sprintf("?cc=%v&tel=%v", url.QueryEscape(cc), url.QueryEscape(tel))
	var m Account
	err = get_info(uri, "youbuy_account", nil, "", &m, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("UserInfoGet fail! err=", err)
		return
	}
	v = &m
	return
}
