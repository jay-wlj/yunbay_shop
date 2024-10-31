package util

import (
	"github.com/jie123108/glog"
)

type smsContent struct {
	UserIds []int64 `json:"user_ids"`
	Content string  `jsn:"content"`
}

func SendSms(uids []int64, content string) (fail_ids []int64, err error) {
	uri := "/man/account/sms/send"
	v := smsContent{UserIds: uids, Content: content}

	fail_ids = []int64{}
	err = post_info(uri, "account", nil, v, "fail_ids", &fail_ids, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("SendSms fail! err=", err)
		return
	}
	return
}

type smsTelsContent struct {
	Tels    []TelInfo `json:"tels"`
	Content string    `jsn:"content"`
}

type TelInfo struct {
	Cc  string `json:"cc"`
	Tel string `json:"tel" valid:"Required"`
}

func SendTelsSms(tels []TelInfo, content string) (fail_ids []TelInfo, err error) {
	uri := "/man/account/sms/send_by_tels"
	v := smsTelsContent{Tels: tels, Content: content}

	fail_ids = []TelInfo{}
	err = post_info(uri, "account", nil, v, "fail_ids", &fail_ids, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("SendSms fail! err=", err)
		return
	}
	return
}
