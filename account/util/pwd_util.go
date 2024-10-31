package util

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/jie123108/glog"
)

var tab string
var password_key string

func init() {
	tab = "abcdefghijklmnopqrstuvwxyz0123456789"
	password_key = "a8a90084d92e46dc866db6900257be6c"
}

func random_slat() string {
	return RandomSample(tab, 8)
}

func HmacSha1(password, slat string) []byte {
	mac := hmac.New(sha1.New, []byte(slat))
	// enc_pwd := fmt.Sprintf("%x", mac.Sum([]byte(password)))
	enc_pwd := mac.Sum([]byte(password))
	return enc_pwd
}

// local function tohex(str)
//     return (str:gsub('.', function (c)
//         return string.format('%02x', string.byte(c))
//     end))
// end

// function _M.hash_password2(password)
//     local pwd = tohex(ngx.sha1_bin("fe5a113f3|" .. password))
//     return pwd
// end

func HashPassword(password string) string {
	h := sha1.New()
	io.WriteString(h, password)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func MakePwd(password string) string {
	slat := random_slat()
	enc_pwd := HmacSha1(password, slat)
	return base64.StdEncoding.EncodeToString([]byte(slat + string(enc_pwd)))
}

// 检测密码是否正确
// pwd为原始密码(即用户输入的密码)
// enc_pwd为已经编码后的密码(即保存到数据库里面的密码)
func CheckPwd(pwd, enc_pwd string) bool {
	dec_str, err := base64.StdEncoding.DecodeString(enc_pwd)
	if err != nil || len(dec_str) <= 8 {
		glog.Error("decrypt password [", enc_pwd, "] failed! err:", err)
		return false
	}
	slat := string(dec_str[0:8])
	enc_str := string(dec_str[8:])
	return string(HmacSha1(pwd, slat)) == enc_str
}
