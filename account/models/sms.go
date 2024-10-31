package models

import (
	"regexp"
)

type TelInfo struct {
	Cc  string `json:"cc"`
	Tel string `json:"tel" valid:"Required"`
}

func (v *TelInfo) FullTel() string {
	return v.Cc + "-" + v.Tel
}

func (v *TelInfo) Valid() (valid bool) {
	switch v.Cc {
	case "+86", "86", "":
		valid, _ = regexp.MatchString("^1[3-9]\\d{9}$", v.Tel)
	case "+852":
		valid, _ = regexp.MatchString("^(5|6|8|9)\\d{7}$", v.Tel)
	default:
		valid, _ = regexp.MatchString("^[0-9]{4,11}$", v.Tel)
	}
	return
}

type SMSSendReq struct {
	TelInfo
	Code    string `json:"code" valid:"Required"`
	Expires int    `json:"expires"`
}

type SMSCheckReq struct {
	TelInfo
	Code string `json:"code" valid:"Required"`
}
