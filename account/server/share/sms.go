package share

import (
	"yunbay/account/common"

	"time"
	"yunbay/account/models"
	"yunbay/account/util"

	"github.com/jay-wlj/gobaselib/cache"

	"github.com/jie123108/glog"
)

var sms_redis *cache.RedisCache

var default_sms_timeout time.Duration

func init() {
	default_sms_timeout = 5 * time.Minute
}

func InitSmsRedis(timeout string) (err error) {
	default_sms_timeout, err = time.ParseDuration(timeout)
	if err != nil {
		glog.Error("invalid Config[session_redis::token_timeout]: ", timeout)
		return
	}
	sms_redis, err = cache.GetWriter("sms")
	return
}

func saveCode(tel, code string, expires time.Duration) (err error) {
	key := "cd:" + tel
	err = sms_redis.Set(key, code, expires)
	return
}

func getCode(tel string) (code string, err error) {
	key := "cd:" + tel
	code, err = sms_redis.Get(key)
	return
}

// 发送短信验证码
func SendSmsCode(cc, tel string, code string) (ok bool, reason string) {
	t := models.TelInfo{Cc: cc, Tel: tel}
	tel_full := t.FullTel()
	// 如果没有传入过期时间, 默认为5分钟.
	expires := default_sms_timeout

	glog.Info("----------- Send Sms Code, tel:", tel_full, ", Code:", code)
	// 使用运营商短信服务发送短信
	// msg := fmt.Sprintf("你的验证码是: %s", code)
	// err := util.SmsSend(tel.FullTel(), msg)
	err := util.SendUCPassSms(t.Cc, t.Tel, code)
	if err != nil {
		glog.Error("SmsSend(", tel_full, ",", code, ") failed! err:", err)
		return false, common.ERR_SERVER_ERROR
	}

	// 保存验证码
	err = saveCode(tel_full, code, expires)
	if err != nil {
		glog.Error("saveTel(", tel_full, ",", code, ",", expires, ") failed! err:", err)
		return false, common.ERR_SERVER_ERROR
	}
	return true, ""
}

// 检测短信验证码是否正确
func CheckSmsCode(tel_full string, req_code string) (ok bool, reason string) {
	glog.Info("----------- Check Sms Code, tel:", tel_full, ", Code:", req_code)
	code, err := getCode(tel_full)
	if err != nil {
		if err == cache.ErrNotExist {
			glog.Error("GetCode(", tel_full, ") failed! err: code not exist!")
			return false, common.ERR_CODE_INVALID
		} else {
			glog.Error("GetCode(", tel_full, ") failed! err:", err)
			return false, common.ERR_SERVER_ERROR
		}
	}

	if code != req_code {
		glog.Error("----------- Check Sms Code, tel:", tel_full, ", Code:", req_code, " failed! server code: ", code)
		return false, common.ERR_CODE_INVALID
	}
	return true, ""
}
