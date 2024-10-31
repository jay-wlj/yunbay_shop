package client

import (
	"yunbay/account/common"
	"yunbay/account/dao"
	"yunbay/account/models"
	"yunbay/account/server/share"
	"yunbay/account/util"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 无登录状态发送验证码
func SmsCodeSend(c *gin.Context) {
	var req models.AccountSMSSendReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	if !req.Valid() {
		yf.JSON_Fail(c, common.ERR_TEL_INVALID)
		return
	}
	if req.Type == 0 || req.Type == 1 {
		if req.ImgKey == "" || req.ImgCode == "" { //注册时图片验证码必填
			glog.Error("reg.ImgCode is empty!")
			yf.JSON_Fail(c, common.ERR_CODE_INVALID)
			return
		}
		if ok := ImgCodeCheck(req.ImgKey, req.ImgCode); !ok {
			glog.Error("reg.ImgCode check fail!")
			yf.JSON_Fail(c, common.ERR_IMGCODE_INVALID)
			return
		}
	}

	code := util.RandomSample("0123456789", 6)
	account, err := dao.GetAccountByTel(req.Cc, req.Tel)
	if err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetAccountByTel(", req.Cc, ",", req.Tel, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}

	if req.Type == 1 { //重设登陆密码
		// 重设密码，但是手机号并不存在。(需要提示注册)
		if account == nil {
			glog.Error("tel [", req.FullTel(), "] is not exist")
			yf.JSON_Fail(c, common.ERR_TEL_NOT_EXIST)
			return
		}
	} else if req.Type == 2 { //重设资金密码
		// 重设资金密码，但是手机号并不存在。(需要提示注册)
		if account == nil {
			glog.Error("tel [", req.FullTel(), "] is not exist")
			yf.JSON_Fail(c, common.ERR_TEL_NOT_EXIST)
			return
		}
	} else if req.Type == 0 {
		// 注册，但是手机号已经存在。
		if account != nil {
			glog.Error("tel [", req.FullTel(), "] is exist account: ", account)
			yf.JSON_Fail(c, common.ERR_TEL_EXIST)
			return
		}
	}

	//ok, reason := SendSmsCode(tel_full, code, expires)
	ok, reason := share.SendSmsCode(req.Cc, req.Tel, code)
	if !ok {
		yf.JSON_Fail(c, reason)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

// 用户发送验证码
func SmsUsersend(c *gin.Context) {
	user_info := c.MustGet("user_info").(util.TokenInfo)

	account, err := dao.GetAccountById(user_info.UserId)
	if err != nil {
		glog.Error("GetAccountById(", user_info.UserId, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	if account == nil {
		yf.JSON_Fail(c, common.ERR_USER_NOT_EXIST)
		return
	}
	// TODO: 短信频率检测, 限制

	//ok, reason := SendSmsCode(tel_full, code, expires)
	code := util.RandomSample("0123456789", 6)
	ok, reason := share.SendSmsCode(account.Cc, account.Tel, code)
	if !ok {
		yf.JSON_Fail(c, reason)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

// 检测验证码
func SmsCodeCheck(c *gin.Context) {
	var req models.AccountSMSCheckReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	tel_full := req.FullTel()
	ok, reason := share.CheckSmsCode(tel_full, req.Code)
	if !ok {
		yf.JSON_Fail(c, reason)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}
