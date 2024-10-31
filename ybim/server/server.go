package server

import (
	"yunbay/ybim/conf"
	"yunbay/ybim/server/share"

	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
)

func StartServer(config *conf.ApiConfig) (err error) {
	r := Load(config)
	//defer util.WaitGroup_wait()

	db.InitPsqlDb(conf.Config.Server.PSQLUrl, conf.Config.Server.Debug)

	// init redis

	// start timer
	// c := cron.New()
	// //c.AddFunc("0/5 * * * * ?", func() { fmt.Printf("Hit\n") })
	// c.AddFunc("0 1 0 * * ?", timer.YBAsset_Rebat)
	// c.Start()

	err = r.Run(conf.Config.Server.Listen)
	if err != nil {
		glog.Errorf("gracehttp.Serve  start error:%v ", err)
	}
	share.GetWsMgr().Close()

	return err
}
