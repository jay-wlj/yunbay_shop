package util

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jay-wlj/gobaselib/yf"
	"github.com/jie123108/glog"
)

func GetPlatformVersionByContext(c *gin.Context) (platform string, version string) {
	headers := c.Request.Header
	platform_header := headers["X-Yf-Platform"]
	version_header := headers["X-Yf-Version"]

	if len(platform_header) > 0 {
		platform = strings.ToLower(platform_header[0])
	}
	if len(version_header) > 0 {
		version = version_header[0]
	}
	return
}

func GetRemoteAddress(c *gin.Context) (address string, err error) {
	return GetHeaderString(c, "remote_addr")
}

func GetDevId(c *gin.Context) (devid string, err error) {
	return GetHeaderString(c, "X-Yf-Devid")
}
func GetDevType(c *gin.Context) (devtype string, err error) {
	return GetHeaderString(c, "X-Yf-Devtype")
}

func GetHeaderString(c *gin.Context, key string) (v string, err error) {
	headers := c.Request.Header
	vals := headers[key]

	if len(vals) > 0 {
		v = vals[0]
	}
	return
}

func GetToken(c *gin.Context) (string, error) {
	return GetHeaderString(c, "X-Yf-Token")
}

func GetUid(c *gin.Context) (user_id int64, ok bool) {
	ok = true
	user_id = c.GetInt64("user_id")
	if user_id == 0 {
		ok = false
		glog.Error("user_id not exist! ")
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return

	}
	return
}
func GetCountry(c *gin.Context) (country int) {
	if v, err := GetHeaderString(c, "X-Yf-Country"); err == nil {
		country, _ = strconv.Atoi(v)
	}
	return
}

// 获取第三方标识
func GetThirdId(c *gin.Context) (third_id int) {
	if v, err := GetHeaderString(c, "X-Yf-Third"); err == nil {
		third_id, _ = strconv.Atoi(v)
	}

	return
}

func GetUserType(c *gin.Context) (user_type int16, ok bool) {
	ok = true
	usertype := c.GetInt64("user_type")
	user_type = int16(usertype)
	return
}

// func UnmarshalBodyAndCheck(c *gin.Context, req interface{}) bool {
// 	if err := base.CheckQueryJsonField(c, &req); err != nil {
// 		glog.Info("UnmarshalBodyAndCheck args invalid! err=", err)
// 		yf.JSON_FailEx(c, yf.ERR_ARGS_INVALID, err.Error())
// 		return false
// 	}
// 	return true
// }

func UnmarshalReq(c *gin.Context, req interface{}) bool {
	var err error
	switch c.Request.Method {
	case "GET":
		err = c.ShouldBindQuery(req)
	default:
		err = c.ShouldBindJSON(req)
	}

	if err != nil {
		//if err := base.CheckQueryJsonField(c, &req); err != nil {
		glog.Info("UnmarshalReq args invalid! err=", err)
		yf.JSON_FailEx(c, yf.ERR_ARGS_INVALID, err.Error())
		return false
	}

	return true
}
