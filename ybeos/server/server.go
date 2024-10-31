package server

import (
	"yunbay/ybeos/conf"

	"github.com/jie123108/glog"
	// "yunbay/ybeos/dao"
	// "yunbay/ybeos/server/share"
)

func StartServer(config *conf.ApiConfig) (err error) {
	r := Load(config)

	//dao.InitPsqlDb(conf.Config.Server.PSQLUrl, true /*conf.Config.Server.Debug*/)

	err = r.Run(config.Server.Listen)
	if err != nil {
		glog.Errorf("gracehttp.Serve  start error:%v ", err)
	}

	return err
}
