package server

import (
	"yunbay/account/conf"
	"yunbay/account/server/client"
	"yunbay/account/server/share"

	"yunbay/account/db"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jie123108/glog"
)

func StartServer(config *conf.ApiConfig) (err error) {
	r := Load(config)

	db.InitDB(conf.Config.Server.PSQLUrl)
	cache.InitRedis(conf.Config.Redis)

	client.InitImagecodeRedis(conf.Config.Server.Ext["img_expires"].(int))
	share.InitSessionRedis(conf.Config.Server.Ext["token_timeout"].(string))
	share.InitSmsRedis(conf.Config.Server.Ext["sms_timeout"].(string))

	err = r.Run(config.Server.Listen)
	if err != nil {
		glog.Errorf("gracehttp.Serve  start error:%v ", err)
	}

	return err
}
