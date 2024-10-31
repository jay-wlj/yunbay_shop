package dao

import (
	//"fmt"
	//"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	//"YBAsset/util/cache"
	"github.com/jay-wlj/gobaselib/cache"
	"yunbay/ybeos/conf"
	//"github.com/jinzhu/gorm"
    //_ "github.com/jinzhu/gorm/dialects/postgres"
)


var g_rediscache *cache.RedisCache
var g_apicache *cache.RedisCache

func GetDefaultCache() (*cache.RedisCache, error) {
	if g_rediscache == nil {
		var err error
		g_rediscache, err = cache.NewRedisCacheFromCfg(conf.Config.CommonRedis.Addr, conf.Config.CommonRedis.Password, conf.Config.CommonRedis.Timeout, conf.Config.CommonRedis.DBIndex)
		if err != nil {
			glog.Error("NewRedisCacheFromCfg failed! err:", err, " cfg:", conf.Config.CommonRedis)
			return nil, err
		}
	}
	return g_rediscache, nil
}



func GetApiCache() (*cache.RedisCache, error) {
	if g_apicache == nil {
		var err error
		g_apicache, err = cache.NewRedisCacheFromCfg(conf.Config.ApiRedis.Addr, conf.Config.ApiRedis.Password, conf.Config.ApiRedis.Timeout, conf.Config.ApiRedis.DBIndex)
		if err != nil {
			glog.Error("NewRedisCacheFromCfg failed! err:", err, " cfg:", conf.Config.ApiRedis)
			return nil, err
		}
	}
	return g_apicache, nil
}