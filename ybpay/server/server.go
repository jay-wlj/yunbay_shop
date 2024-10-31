package server

import (
	"yunbay/ybpay/conf"
	"yunbay/ybpay/server/share"

	"yunbay/ybpay/server/client"
	"yunbay/ybpay/server/man"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/jie123108/glog"
)

func StartServer(config *conf.ApiConfig) (err error) {
	//r := Load(config)

	db.InitPsqlDb(conf.Config.Server.PSQLUrl, true /*conf.Config.Server.Debug*/)
	cache.InitRedis(conf.Config.Redis)

	share.InitBankId(config.Server.Ext["bank_cfg"].(string)) // 加载银行bid信息

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
		glog.Errorf("gracehttp.Serve  start error:%v ", err)
	}

	return err
}
