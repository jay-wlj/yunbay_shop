package server

import (
	"yunbay/ybsearch/conf"

	"github.com/jie123108/glog"
)

func StartServer(config *conf.ApiConfig) (err error) {
	r := Load(config)

	//db.InitPsqlDb(conf.Config.Server.PSQLUrl, conf.Config.Server.Debug)
	//cache.InitRedis(conf.Config.Redis)

	err = r.Run(config.Server.Listen)
	if err != nil {
		glog.Errorf("gracehttp.Serve  start error:%v ", err)
	}

	return err
}
