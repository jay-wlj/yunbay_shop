package share

import (
	"fmt"

	"github.com/jay-wlj/gobaselib/cache"

	//"strconv"

	"time"
	"yunbay/account/util"

	"github.com/jie123108/glog"
)

var session_redis *cache.RedisCache
var token_timeout time.Duration

func InitSessionRedis(timeout string) (err error) {
	session_redis, err = cache.GetWriter("session")
	if err != nil {
		return
	}

	token_timeout, err = time.ParseDuration(timeout)
	if err != nil {
		glog.Error("invalid Config[session_redis::token_timeout]: ", timeout)
		return
	}

	return
}

// 保存 token->user_id
func token_save(token string, user_id int64, token_timeout time.Duration) (err error) {
	user_save(user_id, token, token_timeout)

	key := "tk:" + token
	err = session_redis.Set(key, user_id, token_timeout)
	return
}

// 保存 user_id->token
func user_save(user_id int64, token string, token_timeout time.Duration) (err error) {
	key := "uk:" + fmt.Sprintf("%v", user_id)
	err = session_redis.Set(key, token, token_timeout)
	return
}

func TokenDelByUserid(user_id int64) (err error) {
	key := "uk:" + fmt.Sprintf("%v", user_id)
	var token string
	token, err = session_redis.Get(key)
	if token != "" {
		err = Token_delete(token, user_id)
	}
	_, err = session_redis.Del(key)
	return
}

// 登陆token
func Login_internal(user_id int64, user_type int16) (token string, err error) {
	expires := time.Now().Unix() + int64(token_timeout.Seconds())
	token = util.TokenEncrypt(user_type, user_id, expires)
	err = token_save(token, user_id, token_timeout)
	return
}

// 删除登录token信息
func Token_delete(token string, user_id int64) (err error) {
	key := "tk:" + token
	_, err = session_redis.Del(key)
	return
}
