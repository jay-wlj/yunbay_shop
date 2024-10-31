package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"strings"
	"time"

	"github.com/jie123108/glog"
)

const (
	XMSSMS                = "https://Request.ucpaas.com/sms-partner/access/b03eu1/sendsms"
	XMSTIMESMS            = "https://Request.ucpaas.com/sms-partner/access/b03eu1/timer_send_sms"
	XMSSMSIN              = "https://Request.ucpaas.com/sms-partner/access/b03eu1/sendsms"
	XMSTIMESMSIN          = "https://Request.ucpaas.com/sms-partner/access/b03eu1/timer_send_sms"
	CHINACLIENTID         = "b03eu1"
	CHINAPASSWORD         = "c7073769"
	INTERNATIONALCLIENTID = "b03eu1"
	INTERNATIONALPASSWORD = "c7073769"
	// INTERNATIONALCLIENTID		=	"b00in0"
	// INTERNATIONALPASSWORD		=	"9d5ce5e8"
	SMSTYPE              = "4"
	COMPRESSTYPE         = "0"
	APPKEY               = "20180503193938000"
	TPUSERCodeTable      = `yunbay_user_code_info`
	TPUSERImageCodeTable = `yunbay_user_imagecode_info`
	//手机号
	regExpChina = "^(1[3-9])\\d{9}$"
	regExpXg    = "^(5|6|8|9)\\d{7}$"
)

type GetSMSUpReq struct {
	Clientid string `json:"clientid"`
	Password string `json:"password"`
	Mobile   string `json:"mobile"`
	Smstype  string `json:"smstype"`
	Content  string `json:"content"`
	Sendtime string `json:"sendtime"`
	Extend   string `json:"extend"`
	Uid      string `json:"uid"`
}

type GetSMSUpRes struct {
	TotalFee int            `json:"total_fee"`
	Data     []GetSMSUpData `json:"data"`
}

type GetSMSUpData struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Fee    int    `json:"fee"`
	Mobile string `json:"mobile"`
	Sid    string `json:"sid"`
	Uid    string `json:"uid"`
}

//MD5加密
func MD5(data string) string {
	h := md5.New()
	h.Write([]byte(data)) //需要加密的字符串
	return hex.EncodeToString(h.Sum(nil))
}

//获取uuid第二个值，这个用来四位验证码
func IdeaPlus(data string) string {
	as := strings.Split(data, "+")
	if len(as) > 1 {
		return strings.Split(data, "+")[1]
	}
	return ""
}

func SendUCPassSms(cc, tel, code string) (err error) {
	var in GetSMSUpReq

	in.Smstype = SMSTYPE
	//code := RandomSample("0123456789", 6)

	glog.Info("[yunbay_sms] ", cc, "-", tel, " dataCode in:", code)

	in.Sendtime = time.Now().String()
	in.Extend = ""
	in.Uid = ""

	foreign := cc != "" && (cc != "+86" && cc != "86")
	if foreign { // 海外手机号
		in.Clientid = INTERNATIONALCLIENTID
		in.Password = MD5(INTERNATIONALPASSWORD)
		in.Mobile = "00" + IdeaPlus(cc) + tel
		in.Content = "[yunbay]Your verification code:(" + code + "), the code is valid within three minutes, please do not tell anyone, including customer service, to prevent account theft."
	} else {
		in.Clientid = CHINACLIENTID
		in.Password = MD5(CHINAPASSWORD)
		in.Mobile = tel
		in.Content = "【yunbay】验证码" + code + "，该验证码3分钟内有效，请勿告知任何人包括客服，防止账号被盗。"
	}

	req_byte, err := json.Marshal(in)
	if err != nil {
		return
	}

	uri := XMSSMS
	// if foreign {
	// 	uri = XMSSMSIN
	// }
	err = sendSms(uri, req_byte)
	return
}

func SendSms(cc, tel, content string) (err error) {
	var in GetSMSUpReq

	in.Smstype = SMSTYPE
	//code := RandomSample("0123456789", 6)

	glog.Info("[yunbay_sms] ", cc, "-", tel, " dataContent in:", content)

	in.Sendtime = time.Now().String()
	in.Extend = ""
	in.Uid = ""

	foreign := cc != "" && (cc != "+86" && cc != "86")
	if foreign { // 海外手机号
		in.Clientid = INTERNATIONALCLIENTID
		in.Password = MD5(INTERNATIONALPASSWORD)
		in.Mobile = "00" + IdeaPlus(cc) + tel
		in.Content = "【yunbay】" + content
	} else {
		in.Clientid = CHINACLIENTID
		in.Password = MD5(CHINAPASSWORD)
		in.Mobile = tel
		in.Content = "【yunbay】" + content
	}

	req_byte, err := json.Marshal(in)
	if err != nil {
		return
	}

	uri := XMSSMS
	// if foreign {
	// 	uri = XMSSMSIN
	// }
	err = sendSms(uri, req_byte)

	return
}

func sendSms(uri string, content []byte) (err error) {
	headers := make(map[string]string)
	res := base.HttpPost(uri, []byte(content), headers, EXPIRE_RES_INFO)
	if res.StatusCode != 200 {
		glog.Error("request [", res.ReqDebug, "] failed! err:", res.Error)
		err = res.Error
		if err == nil {
			err = fmt.Errorf("http-error: %d", res.StatusCode)
		}
		return
	}
	var ret GetSMSUpRes
	err = json.Unmarshal(res.RawBody, &ret)
	if err != nil {
		glog.Error("[yunbay_sms] GetSMSChina Unmarshal error:", err, " responseData:", string(res.RawBody))
		return
	}
	if len(ret.Data) > 0 && ret.Data[0].Code != 0 {
		glog.Error("[yunbay_sms] GetSMSChina Unmarshal error:", err, " responseData:", string(res.RawBody))
		err = fmt.Errorf("send sms fail!")
	}
	return
}
