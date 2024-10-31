package dao

import (
	"github.com/jie123108/glog"
	"github.com/jay-wlj/gobaselib/cache"
	"yunbay/ybim/conf"
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

func GetIMCache() (*cache.RedisCache, error) {
	if g_apicache == nil {
		var err error
		g_apicache, err = cache.NewRedisCacheFromCfg(conf.Config.IMRedis.Addr, conf.Config.IMRedis.Password, conf.Config.IMRedis.Timeout, conf.Config.IMRedis.DBIndex)
		if err != nil {
			glog.Error("NewRedisCacheFromCfg failed! err:", err, " cfg:", conf.Config.IMRedis)
			return nil, err
		}
	}
	return g_apicache, nil
}
