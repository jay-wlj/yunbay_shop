package util

import (
	"github.com/gin-gonic/gin"
	"github.com/jay-wlj/gobaselib/yf"
	"github.com/jie123108/glog"
)

func GetPlatformVersionByContext(c *gin.Context) (platform string, version string) {
	headers := c.Request.Header
	platform_header := headers["X-Yf-Platform"]
	version_header := headers["X-Yf-Version"]

	if len(platform_header) > 0 {
		platform = platform_header[0]
	}
	if len(version_header) > 0 {
		version = version_header[0]
	}
	return
}

func GetRemoteAddress(c *gin.Context) (address string, err error) {
	headers := c.Request.Header
	userids := headers["remote_addr"]
	if len(userids) > 0 {
		address = userids[0]
	}
	return
}

func GetDevId(c *gin.Context) (devid string, err error) {
	headers := c.Request.Header
	Devid := headers["X-Yf-Devid"]

	if len(Devid) > 0 {
		devid = Devid[0]
	}
	return
}
func GetDevType(c *gin.Context) (devtype string, err error) {
	headers := c.Request.Header
	Devtype := headers["X-Yf-Devtype"]

	if len(Devtype) > 0 {
		devtype = Devtype[0]
	}
	return
}

func GetHeaderString(c *gin.Context, key string) (v string, err error) {
	headers := c.Request.Header
	vals := headers[key]

	if len(vals) > 0 {
		v = vals[0]
	}
	return
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

func GetUserType(c *gin.Context) (user_type int16, ok bool) {
	ok = true
	usertype := c.GetInt64("user_type")
	user_type = int16(usertype)
	return
}

func UnmarshalBodyAndCheck(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		//if err := base.CheckQueryJsonField(c, &req); err != nil {
		glog.Info("UnmarshalBodyAndCheck args invalid! err=", err)
		yf.JSON_FailEx(c, yf.ERR_ARGS_INVALID, err.Error())
		return false
	}

	// if err := Valid(req); err != nil {
	// 	glog.Info("UnmarshalBodyAndCheck args invalid! err=", err)
	// 	yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
	// 	return false
	// }

	return true
}
