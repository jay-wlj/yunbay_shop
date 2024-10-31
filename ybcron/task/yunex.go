package task

import (
	"yunbay/ybasset/common"
	"yunbay/ybcron/conf"
	"yunbay/ybcron/util"
	"fmt"
	"github.com/jay-wlj/gobaselib/db"
	"time"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 快照云网所持ybt用户
func SnapYunexAccount() {
	//fmt.Println("Orders_AutoCancelCheck begin")
	db, err := db.InitPsqlDb(conf.Config.PsqlUrl["asset"], conf.Config.Debug)
	if err != nil {
		fmt.Println("SnapYunexAccount end fail! err=", err)
	}
	db = db.Begin()
	if err := snapYunexAccount(db); err != nil {
		glog.Error("SnapYunexAccount end fail! err=", err)
		db.Rollback()
		return
	}
	db.Commit()

	//fmt.Println("Orders_AutoCancelCheck end success!")
}

// 快照云网所持ybt用户
func snapYunexAccount(db *gorm.DB) (err error) {
	now := time.Now().Unix()
	yester_day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	third, ok := conf.Config.ThirdAccount["yunex"]
	if !ok {
		s := fmt.Sprintf("third plat:yunex not define!")
		glog.Error("snapYunexAccount fail! err=", s)
		err = fmt.Errorf(s)
		return
	}
	var page int = 1
	var total_ybt float64 = 0
	total, vs, err := util.SnapYunexYbtAccount(page, 100, yester_day)
	if err != nil {
		glog.Error("snapYunexAccount fail! err=", err)
		return
	}

	for {
		if int64(len(vs)) >= total {
			break
		}
		page += 1
		_, s, e := util.SnapYunexYbtAccount(page, 100, yester_day)
		if e != nil {
			glog.Error("snapYunexAccount fail! err=", err)
			break
		}
		if len(s) > 0 {
			vs = append(vs, s...)
		} else {
			str := fmt.Sprintln("snapYunexAccount fail! len(s)=0 start=", page, " total=", total)
			glog.Error(str)
			err = fmt.Errorf(str)
			break
		}
	}
	if int64(len(vs)) != total {
		str := fmt.Sprintln("snapYunexAccount fail! len(vs)=", len(vs), " total=", total)
		glog.Error(str)
		err = fmt.Errorf(str)
		return
	}

	for _, v := range vs {
		m := common.ThirdBonus{Tid: third.UserId, Uid: v.UserId, Ybt: v.Total, Date: yester_day, CreateTime: now, UpdateTime: now}
		if err = db.Save(&m).Error; err != nil {
			glog.Error("snapYunexAccount save fail! err=", err)
			return
		}
	}

	// 生成第三方平台用户资产信息
	if third.BonusId > 0 {
		bs := common.KtBonusDetail{UserAsset: common.UserAsset{UserId: third.BonusId, TotalYbt: total_ybt, NormalYbt: total_ybt}, ThirdBonus: 1, BonusYbt: total_ybt, Date: yester_day}
		if err = db.Save(&bs).Error; err != nil {
			glog.Error("snapYunexAccount save fail! err=", err)
			return
		}
	} else {
		glog.Error("not define Yunexbousid")
		err = fmt.Errorf("not define Yunexbousid")
	}

	return
}
