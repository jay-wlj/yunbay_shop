package main

import (
	"flag"
	_ "net/http/pprof"
	"yunbay/ybcron/conf"
	"yunbay/ybcron/task"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
	"github.com/robfig/cron"
)

func main() {
	var configfilename string

	flag.StringVar(&configfilename, "config", "./conf/config.yml", "ini config filename")

	flag.Parse()
	defer glog.Flush()
	// glog.Errorf("################### Build Time: %s ###################", base.BuildTime)

	_, err := conf.LoadConfig(configfilename)
	if err != nil {
		return
	}

	db.InitPsqlDb(conf.Config.PsqlUrl["asset"], conf.Config.Debug) // 默认db
	db.InitPsqlDb(conf.Config.PsqlUrl["api"], conf.Config.Debug)
	cache.InitRedis(conf.Config.Redis)

	c := cron.New()

	// 快照
	if v, ok := conf.Config.Crons["snapshot"]; ok {
		c.AddFunc(v, task.SnapShot)
	}

	// 计算昨日ybt分红数
	if v, ok := conf.Config.Crons["rebat"]; ok {
		c.AddFunc(v, task.Ybt_Rebat)
	}

	// 计算昨日ybt分红
	if v, ok := conf.Config.Crons["rebat_kt"]; ok {
		c.AddFunc(v, task.YBAsset_KtRebat)
	}

	// 释放ybtkt
	if v, ok := conf.Config.Crons["release_ybt_kt"]; ok {
		c.AddFunc(v, task.YBAsset_ReleaseYbtKt)
	}

	// 刷新平台交易数据
	if v, ok := conf.Config.Crons["tradflow"]; ok {
		c.AddFunc(v, task.YBAsset_Statistics)
	}

	// 订单自动取消及自动确认
	if v, ok := conf.Config.Crons["auto_cancel_orders"]; ok {
		c.AddFunc(v, task.Orders_AutoCancelCheck)
	}
	if v, ok := conf.Config.Crons["auto_finish_orders"]; ok {
		c.AddFunc(v, task.Orders_AutoFinishCheck)
	}

	// 定时提交提币申请
	if v, ok := conf.Config.Crons["chain_draw"]; ok {
		c.AddFunc(v, task.Chain_WithDraw)
	}

	if v, ok := conf.Config.Crons["of_orders_query"]; ok {
		c.AddFunc(v, task.OfOrdersQuery)
	}
	// 定时查询提币状态
	if v, ok := conf.Config.Crons["chain_draw_query"]; ok {
		c.AddFunc(v, task.Chain_WithDrawQuery)
	}

	// 计算用户昨日ybt奖励
	//c.AddFunc(conf.Config.Crons["user_rebat"], task.UserAsset_Rebat)

	// 空投用户ybt奖励
	if v, ok := conf.Config.Crons["reward_ybt"]; ok {
		c.AddFunc(v, task.Ybt_Reward)
	}

	// 回收空投过期的ybt奖励
	if v, ok := conf.Config.Crons["recovery_ybt"]; ok {
		c.AddFunc(v, task.YBT_Recover)
	}

	// 定时注册im用户
	if v, ok := conf.Config.Crons["imcreate"]; ok {
		c.AddFunc(v, task.IM_AutoRegisterCheck)
	}

	// 定时快照第三方持有ybt用户
	if v, ok := conf.Config.Crons["snap_thirdaccount"]; ok {
		c.AddFunc(v, task.SnapYunexAccount)
	}

	// 每日一次自动对帐处理
	if v, ok := conf.Config.Crons["day_check"]; ok {
		c.AddFunc(v, task.DayCheck)
	}

	if v, ok := conf.Config.Crons["currency_update"]; ok {
		c.AddFunc(v, task.Currency_Update)
	}

	c.Start()
	select {}
}
