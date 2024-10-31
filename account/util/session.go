package util

import (
	"github.com/jay-wlj/gobaselib/cache"

	"github.com/jie123108/glog"
)

var session_redis *cache.RedisCache = nil

func Token_exist(token string) (exist bool) {
	exist = false
	if session_redis == nil {
		var err error
		if session_redis, err = cache.GetReader("session"); err != nil {
			glog.Error("session session_redis is nil!")
		}
		return
	}

	key := "tk:" + token
	val, err := session_redis.Get(key)
	exist = true
	if err != nil || val == "" {
		exist = false
	}
	return
}
