package server

import (
	"yunbay/ybgoods/conf"
	"yunbay/ybgoods/server/client"
	"yunbay/ybgoods/server/man"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"github.com/jie123108/glog"
)

func StartServer(config *conf.ApiConfig) (err error) {

	db.InitPsqlDb(conf.Config.Server.PSQLUrl, conf.Config.Server.Debug)

	cache.InitRedis(conf.Config.Redis)

	mgr := yf.NewServer()
	mgr.AddRouter(func() (string, []yf.RouterInfo) {
		return "/v1", client.InitRouter()
	})
	mgr.AddRouter(func() (string, []yf.RouterInfo) {
		return "/man", man.InitRouter()
	})

	s_cfg := config.Server
	err = mgr.Start(&yf.Config{Addr: s_cfg.Listen, Debug: s_cfg.Debug, CheckSign: s_cfg.CheckSign, AppKeys: config.AppKeys, AuthServer: config.Servers["account"]})
	//err = r.Run(config.Server.Listen)
	if err != nil {
		glog.Errorf("gracehttp.Serve start error:%v ", err)
	}

	return err
}
