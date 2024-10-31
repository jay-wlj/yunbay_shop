package middleware

import (
	"yunbay/account/util"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

var check_urls map[string]bool

func token_check_internal(ctx *gin.Context, token string, warnlog bool) (reason string, code int) {
	if token == "" {
		return yf.ERR_TOKEN_INVALID, 401
	}
	user_type, user_id, expire_time, err := util.TokenDecrypt(token)
	if err != nil {
		info := fmt.Sprintf("TokenDecrypt(%s) failed! err:%v", token, err)
		if warnlog {
			glog.Debug(info)
		} else {
			glog.Error(info)
		}
		return yf.ERR_TOKEN_INVALID, 401
	}
	now := time.Now().Unix()
	if expire_time < now {
		info := fmt.Sprintf("token(%s) is expired at: %d", token, expire_time)
		if warnlog {
			glog.Debug(info)
		} else {
			glog.Error(info)
		}
		return yf.ERR_TOKEN_INVALID, 401
	}

	// TODO: 从数据库(redis)校验token合法性.
	if !util.Token_exist(token) {
		glog.Debug("token(%s) is not found in session redis")
		return yf.ERR_TOKEN_INVALID, 401
	}
	tokenInfo := util.TokenInfo{Token: token, UserType: user_type, UserId: user_id, ExpireTime: expire_time}
	ctx.Set("user_info", tokenInfo)
	ctx.Set("user_id", user_id)
	return
}

func SetNeedTokenCheckUrls(m map[string]bool) {
	check_urls = m
}
func TokenCheckFilter(c *gin.Context) {
	uri := base.GetUri(c)

	if !check_urls[uri] || c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}

	token := c.Request.Header.Get("X-Yf-Token")
	reason, code := token_check_internal(c, token, false)
	if reason != "" {
		// 如果是登出, token失效时,也返回成功.
		jso := gin.H{"ok": false, "reason": reason, "data": gin.H{}}

		if c.Request.URL.Path == "/v1/YBAccount/logout" {
			code = 200
			jso = gin.H{"ok": true, "reason": "", "data": gin.H{}}
		}
		c.JSON(code, jso)
		c.Abort()
		return
	}

	c.Next()
}
