package task

import (
	"github.com/jay-wlj/gobaselib/db"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybcron/conf"
	"yunbay/ybcron/util"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 调用区块接口进行提币操作
func SnapShot() {
	now := time.Now()
	db := db.GetTxDB(nil)
	err := snap(db.DB)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			glog.Error("SnapShot fail! err=", err)
		}
		db.Rollback()
		return
	}
	db.Commit()
	glog.Info("SnapShot success! tick=", time.Since(now).String())
}

func snap(db *gorm.DB) (err error) {
	// 更新昨天可分红的ybt
	day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	// 快照昨日用户资产表
	err = snap_user_asset(db, day)
	if err != nil {
		glog.Error("snap_user_asset err=", err)
		return
	}
	// 更新昨日平台资产信息
	err = update_yunbay_asset(db, day)
	if err != nil {
		glog.Error("update_yunbay_asset err=", err)
		return
	}

	// err = UpdateAssetLockAmount(day, db)
	// if err != nil {
	// 	glog.Error("snap err=", err)
	// 	return
	// }

	return
}

// 快照昨日用户资产记录
func snap_user_asset(db *gorm.DB, date string) (err error) {
	now := time.Now().Unix()
	vs := []common.KtBonusDetail{}

	yester_day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	db = db.Model(&common.UserAsset{})
	user_ids, err1 := GetBonusUserIds(yester_day)
	if err1 == nil {
		db = db.Where("total_kt>0 or total_ybt-freeze_ybt>0 or user_id in(?)", user_ids) // 优化 只备份正常持有的ybt及kt及特定用户的资产
	} else {
		glog.Error("snap_user_asset fail! err1=", err1)
	}
	rows, err1 := db.Rows() // (*sql.Rows, error)
	//rows, err1 := db.Model(&common.UserAsset{}).Rows() // (*sql.Rows, error)
	if err1 != nil {
		err = err1
		glog.Error("snap_user_asset fail! err=", err)
		return
	}
	for rows.Next() {
		var v common.UserAsset
		if err = db.ScanRows(rows, &v); err != nil {
			glog.Error("snap_user_asset fail! err=", err)
			return
		}
		v.Id = 0
		v.CreateTime = now
		v.UpdateTime = now
		vs = append(vs, common.KtBonusDetail{UserAsset: v, Date: date})
	}

	for _, v := range vs {
		//db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (user_id, date) DO UPDATE SET total_ybt=%v,normal_ybt=%v,lock_ybt=%v,freeze_ybt=%v,total_kt=%v,normal_kt=%v,lock_kt=%v,seller_ybt=%v,recommender_ybt=%v,recommender2_ybt=%v,yunbay_ybt=%v,update_time=%v",
		if err = db.Save(&v).Error; err != nil {
			glog.Error("snap_user_asset fail! err=", err)
			return
		}
	}
	return
}

// 获取昨日可分红的用户id列表
func GetBonusUserIds(date string) (ids []int64, err error) {
	ids = []int64{}
	db := db.GetDB()

	// 获取昨日的交易流水记录
	var ops []common.YBAssetPool
	if err = db.Find(&ops, "currency_type=? and date=?", common.CURRENCY_KT, date).Error; err != nil {
		glog.Error("GetBonusUserIds fail! err= ", err)
		return
	}
	mbuy := make(map[int64]bool)
	mid := make(map[int64]bool)
	for _, v := range ops {
		mbuy[v.PayerUserId] = true
		mid[v.PayerUserId] = true
		mid[v.SellerUserId] = true
	}

	// 查询消费者的直接邀请人
	uids := []int64{}
	for k := range mbuy {
		uids = append(uids, k)
	}
	var mInvites map[int64][]int64
	if mInvites, err = util.GetInvitersByIds(uids); err != nil {
		glog.Error("GetInvitersByIds fail! err=", err)
		return
	}
	for _, v := range mInvites {
		if len(v) > 0 {
			mid[v[0]] = true
		}
	}

	// 获取以前的订单还没有发放商家奖励的用户
	var ors []common.Ordereward
	if err = db.Find(&ors, "date<=? and seller_status=0", date).Error; err != nil {
		glog.Error("GetBonusUserIds fail! err=", err)
		return
	}
	for _, v := range ors {
		mid[v.SellerUserId] = true
	}

	// 获取活动奖励释放的用户
	var ma map[int64]float64
	if ma, _, err = GetUserActivity(db.DB, date); err != nil {
		glog.Error("GetBonusUserIds fail! GetUserActivity err=", err)
		return
	}
	for k := range ma {
		mid[k] = true
	}

	// // 获取昨日空投的用户
	// if ma, _, err = GetUserAirDrop(db.DB, date); err != nil {
	// 	glog.Error("GetBonusUserIds fail! GetUserAirDrop err=", err)
	// 	return
	// }
	// for k := range ma {
	// 	mid[k] = true
	// }

	// 获取项目用户id
	for _, v := range conf.Config.ProjectYbtAllot {
		mid[v.UserId] = true
		for _, c := range v.Users {
			mid[c.UserId] = true
		}
	}
	// 获取系统帐号id
	for _, v := range conf.Config.SystemAccounts {
		mid[v] = true
	}

	for k := range mid {
		ids = append(ids, k)
	}

	return
}

type yunbayassetSt struct {
	Amount       float64
	Profit       float64
	IssueYbt     float64
	DestoryedYbt float64
	Mining       float64
	AirDrop      float64
	AirUnlock    float64
	Activity     float64
	Project      float64
}

// 更新平台总交易及总释放信息
func update_yunbay_asset(db *gorm.DB, date string) (err error) {
	var v yunbayassetSt
	if err = db.Model(&common.YBAssetDetail{}).Where("date<=?", date).Select("sum(amount) as amount, sum(profit) as profit, sum(issue_ybt) as issue_ybt,sum(destoryed_ybt) as destoryed_ybt, sum(mining) as mining, sum(air_drop) as air_drop, sum(air_unlock) as air_unlock, sum(activity) as activity, sum(project) as project").Scan(&v).Error; err != nil {
		glog.Error("update_yunbay_asset fail! err=", err)
		return
	}
	now := time.Now().Unix()
	if err = db.Model(&common.YBAsset{}).Where("date=?", date).Updates(map[string]interface{}{"total_kt": v.Amount, "total_kt_profit": v.Profit, "total_issue_ybt": v.IssueYbt, "total_destroyed_ybt": v.DestoryedYbt,
		"total_mining": v.Mining, "total_air_drop": v.AirDrop, "total_air_unlock": v.AirUnlock, "total_activity": v.Activity, "total_project": v.Project, "update_time": now}).Error; err != nil {
		glog.Error("update_yunbay_asset fail! err=", err)
		return
	}

	return
}
