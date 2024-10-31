package task

import (
	"yunbay/ybcron/conf"
	"yunbay/ybcron/util"
	"fmt"
	"github.com/jay-wlj/gobaselib/db"
	"time"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// im用户帐号自动注册定时检查
func IM_AutoRegisterCheck() {
	fmt.Println("IM_AutoRegisterCheck begin")
	now := time.Now()
	db, err := db.InitPsqlDb(conf.Config.PsqlUrl["account"], conf.Config.Debug)
	if err != nil {
		fmt.Println("IM_AutoRegisterCheck end fail! err=", err)
	}
	db = db.Begin()
	if err := check_im(db); err != nil {
		glog.Error("IM_AutoRegisterCheck end fail! err=", err)
		db.Rollback()
		return
	}
	db.Commit()
	fmt.Println("IM_AutoRegisterCheck end tick=", time.Since(now).String())

}

type uidSt struct {
	UserId int64
}

// im用户帐号自动注册定时检查
func check_im(db *gorm.DB) (err error) {
	//args := conf.Config.Rebat
	// 查找所有未注册im的用户
	vs := []uidSt{}
	if err = db.Raw("select A.user_id from account A left join imtoken B on A.user_id=B.user_id where B.user_id is null limit 100").Scan(&vs).Error; err != nil {
		glog.Error("cancel_orders fail! err=", err)
		return
	}
	if len(vs) > 0 {
		// 注册100个
		if len(vs) > 100 {
			vs = vs[:100]
		}
		for _, v := range vs {
			if err = util.RegisterIMUser(v.UserId); err != nil {
				glog.Error("RegisterIMUser fail! err=", err)
				return
			}
		}
	}

	return
}
