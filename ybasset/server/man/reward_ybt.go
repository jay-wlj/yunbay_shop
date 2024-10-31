package man

import (
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"
	"yunbay/ybasset/dao"
	"yunbay/ybasset/util"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

func Man_YbtRewardList(c *gin.Context) {
	id, _ := base.CheckQueryInt64DefaultField(c, "id", 0)
	date, _ := base.CheckQueryStringField(c, "date")
	user_id, _ := base.CheckQueryInt64DefaultField(c, "user_id", -1)
	check_status, _ := base.CheckQueryIntDefaultField(c, "status", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	db := db.GetDB()
	if id > 0 {
		db.DB = db.Where("id=?", id)
	}
	if user_id > -1 {
		db.DB = db.Where("user_id=?", user_id)
	}
	if date != "" {
		db.DB = db.Where("date=?", date)
	}
	if check_status > -1 {
		db.DB = db.Where("check_status=?", check_status)
	}
	var total int = 0
	if err := db.Model(&common.YbtUnlockDetail{}).Count(&total).Error; err != nil {
		glog.Error("BonusOrders_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	db.DB = db.ListPage(page, page_size)
	vs := []common.YbtUnlockDetail{}
	if err := db.Order("id desc").Find(&vs).Error; err != nil {
		glog.Error("Man_YbtRewardList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": vs, "list_page": base.IsListEnded(page, page_size, len(vs), total), "total": total})
}

type datecheck struct {
	Date string `json:"date" binding:"required"`
}

func Man_YbtRewardCheck(c *gin.Context) {
	checker_name, err := util.GetHeaderString(c, "X-Yf-Maner")
	if checker_name == "" || err != nil {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var args datecheck
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	// 释放ybt
	now := time.Now()
	db := db.GetTxDB(c)
	if err := release_ybt(db, args.Date); err != nil {
		s := fmt.Sprintf("Man_YbtRewardCheck fail! err=%v", err)
		glog.Error(s)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		util.PublishMsg(common.MQMail{Receiver: []string{"305898636@qq.com"}, Subject: "Yunbay Error", Content: s})
		util.SendDingTextTalk(s, []string{"15818717950"})
		return
	}
	util.SendDingTextTalk(fmt.Sprintf("ybt:%v 释放完毕 共耗时:%v", args.Date, time.Since(now).String()), nil)
	glog.Infof("Man_YbtRewardCheck ok! tick=%v", time.Since(now).String())
	yf.JSON_Ok(c, gin.H{})
}

// 释放当天ybt
func release_ybt(db *db.PsqlDB, date string) (err error) {
	today := time.Now().Format("2006-01-02")
	yester_day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	now := time.Now().Unix()

	// nDate := time.Now()
	// if nDate, err = time.Parse("2006-01-02", date); err != nil {
	// 	glog.Error("release_ybt fail! err=", err)
	// 	return
	// }
	// if nDate.After(time.Now().AddDate(0, 0, -1)) {
	// 	glog.Error("release_ybt fail! the args.date=", date)
	// 	err = fmt.Errorf("ERR_ARGS_INVALID")
	// 	return
	// } else
	if yester_day != date {
		glog.Info("the args.date is ", date, " and yester day is ", yester_day)
		yester_day = date // 说明操作的不是昨天发放记录
	}

	// 获取昨日平台交易记录
	db_asset := db.Model(&common.YBAssetDetail{}).Where("date=? and ybt_status=?", date, common.STATUS_INIT).Updates(map[string]interface{}{"ybt_status": common.STATUS_OK, "update_time": now})
	if err = db_asset.Error; err != nil {
		glog.Error("release_ybt fail! YBAssetDetail update err=", err)
		return
	}
	if db_asset.RowsAffected == 0 {
		glog.Error("release_ybt has released! return")
		return
	}

	vs := []common.YbtUnlockDetail{}
	if err = db.Find(&vs, "date=? and check_status=?", date, common.STATUS_INIT).Error; err != nil {
		glog.Error("release_ybt fail! err=", err)
		return
	}
	// if len(vs) == 0 {
	// 	return
	// }

	// 将状态置为已发放
	if err = db.Model(&common.YbtUnlockDetail{}).Where("date=? and check_status=?", date, common.STATUS_INIT).Updates(map[string]interface{}{"check_status": common.STATUS_OK, "update_time": now}).Error; err != nil {
		glog.Error("release_ybt faiL! err=", err)
		return
	}

	// 先删除当日ybt返利记录
	db.Delete(common.UserAssetDetail{}, "date=? and type=? and transaction_type=?", today, common.CURRENCY_YBT, common.YBT_TRANSACTION_CONSUME)
	us := []common.UserAssetDetail{}
	al := []common.AssetLock{}
	for _, v := range vs {
		if v.Consume > 0 {
			// 生成用户消费资产记录
			us = append(us, common.UserAssetDetail{UserId: v.UserId, Type: common.CURRENCY_YBT, TransactionType: common.YBT_TRANSACTION_CONSUME, Amount: v.Consume, Date: today})
		}

		if v.Sale > 0 {
			// 生成用户商家奖励记录
			us = append(us, common.UserAssetDetail{UserId: v.UserId, Type: common.CURRENCY_YBT, TransactionType: common.YBT_TRANSACTION_SELLER, Amount: v.Sale, Date: today})
		}
		if v.Invite > 0 {
			// 生成用户邀请奖励记录
			us = append(us, common.UserAssetDetail{UserId: v.UserId, Type: common.CURRENCY_YBT, TransactionType: common.YBT_TRANSACTION_INVITE, Amount: v.Invite, Date: today})
		}
		if v.Project > 0 {
			// 项目方奖励释放
			us = append(us, common.UserAssetDetail{UserId: v.UserId, Type: common.CURRENCY_YBT, TransactionType: common.YBT_TRANSACTION_PROJECT, Amount: v.Project, Date: today})
			var as []common.AssetLock
			as, err = lock_user_project(v.UserId, v.Project)
			if err != nil {
				glog.Error("lock_user_project faiL! err=", err)
				return
			}
			al = append(al, as...)
		}
		// 释放空投奖励
		if v.AirUnlock > 0 {
			al = append(al, common.AssetLock{UserId: v.UserId, Type: common.CURRENCY_YBT, LockType: common.ASSET_LOCK_AIRDROP, LockAmount: -v.AirUnlock, Date: today, CreateTime: now, UpdateTime: now})
		}
	}

	// 先产生ybt释放流水
	yfs := []common.YbtFlow{}
	for _, v := range us {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("release_ybt fail! user_asset_detail add err=", err)
			return
		}
		// 算作ybt昨日发放流水
		flow_type := common.YBT_REWARD_MING
		if v.TransactionType == common.YBT_TRANSACTION_PROJECT {
			flow_type = common.YBT_REWARD_PROJECT
		}
		yfs = append(yfs, common.YbtFlow{Type: flow_type, UserId: v.UserId, Amount: v.Amount, UserAssetId: v.Id, Date: yester_day, CreateTime: now, UpdateTime: now})
	}
	// 解锁空投奖励
	for _, v := range al {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("release_ybt fail! AssetLock add err=", err)
			return
		}
	}

	// 释放ybt
	for _, v := range yfs {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("release_ybt fail! YbtFlow add err=", err)
			return
		}
	}

	// // 项目方释放
	// if err = release_project(db, yb.Project); err != nil {
	// 	glog.Error("release_project fail! YbtFlow add err=", err)
	// 	return
	// }

	// 添加用户每日分红ybt记录
	if err = saveUserYbtDetail(db, yester_day); err != nil {
		glog.Error("release_project fail! saveUserYbtDetail err=", err)
		return
	}

	// 更新ybt_day_flow 空投释放数量
	var yb common.YBAssetDetail
	if err = db.Find(&yb, "date=?", yester_day).Error; err != nil {
		glog.Error("release_ybt fail! YBAssetDetail find err=", err)
		return
	}

	// 更新资产锁定的量
	if err = db.Model(&common.AssetLock{}).Where("type=? and lock_type in(?)", common.CURRENCY_YBT, []int{common.ASSET_LOCK_FIX, common.ASSET_LOCK_FOREVER}).Select("sum(lock_amount) as lock_ybt").Scan(&yb).Error; err != nil {
		glog.Error("release_ybt fail! UserAsset err=", err)
		return
	}

	if err = db.Model(&common.YbtDayFlow{}).Where("date=?", yester_day).Updates(map[string]interface{}{"unlock_reward": yb.AirUnlock + yb.Activity, "update_time": now}).Error; err != nil {
		glog.Error("release_project fail! YbtDayFlow err=", err)
		return
	}
	if err = db.Model(&common.YBAssetDetail{}).Where("date=?", yester_day).Updates(map[string]interface{}{"lock_ybt": yb.LockYbt, "update_time": now}).Error; err != nil {
		glog.Error("release_project fail! YbtDayFlow err=", err)
		return
	}

	dao.RefrenshUserBonus(yester_day)
	dao.RefrenshYBAssetDetailCache()
	return
}

func get_projectallot_by_userid(user_id int64) (ret *conf.YbtAllot) {
	allot := conf.Config.ProjectYbtAllot
	for _, v := range allot {
		if v.UserId == user_id {
			ret = &v
			return
		}
		for _, m := range v.Users {
			if m.UserId == user_id {
				ret = &v
				return
			}
		}
	}
	return nil
}

// 项目方奖励释放的用户锁定记录
func lock_user_project(user_id int64, amount float64) (as []common.AssetLock, err error) {
	today := time.Now().Format("2006-01-02")
	now := time.Now().Unix()
	as = []common.AssetLock{}
	v := get_projectallot_by_userid(user_id)
	if v == nil {
		err = fmt.Errorf("user_id is not project users! user_id=%v", user_id)
		return
	}
	var foreverAmount float64
	if v.Forever > 0 { // 永久冻结
		foreverAmount = amount * v.Forever
		if base.IsEqual(v.Forever, 1) {
			foreverAmount = amount
		}
		as = append(as, common.AssetLock{UserId: user_id, Type: common.CURRENCY_YBT, LockType: common.ASSET_LOCK_FOREVER, LockAmount: foreverAmount, Date: today, CreateTime: now, UpdateTime: now})
	}
	if v.Fix > 0 { // 固定期限冻结
		unlockTime := time.Now().Unix() + (v.FixDays * 24 * 3600) // 将冻结天数转成s
		lockAmount := amount * v.Fix
		if base.IsEqual(v.Fix+v.Forever, 1) {
			lockAmount = amount - foreverAmount
		}
		as = append(as, common.AssetLock{UserId: user_id, Type: common.CURRENCY_YBT, LockType: common.ASSET_LOCK_FIX, LockAmount: lockAmount, UnlockTime: unlockTime, Date: today, CreateTime: now, UpdateTime: now})
	}
	return
}

// // 分配项目方ybt
// func release_project(db *db.PsqlDB, user_id int64, amount float64) (err error) {
// 	// 判断是否超出初始发行
// 	var ybt common.Ybt
// 	ybt, err = GetYbt()
// 	if err != nil {
// 		glog.Error("GetYbt fail! err", err)
// 		return
// 	}
// 	// 释放的项目方ybt不能超过冻结的
// 	if ybt.LockProject < amount {
// 		amount = ybt.LockProject
// 	}
// 	if base.IsEqual(amount, base.FLOAT_MIN) {
// 		glog.Error("release_project no ybt released")
// 		return
// 	}
// 	now := time.Now().Unix()
// 	// today := time.Now().Format("2006-01-02")
// 	yester_day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

// 	// v := common.UserAssetDetail{UserId:0, Type:common.CURRENCY_YBT, TransactionType:common.YBT_TRANSACTION_PROJECT, Amount:amount, Date:today, CreateTime:now, UpdateTime:now}
// 	// if err = db.Save(&v).Error; err != nil {
// 	// 	glog.Error("release_ybt fail! UserAssetDetail add err=", err)
// 	// 	return
// 	// }
// 	yf := common.YbtFlow{Type:common.YBT_REWARD_PROJECT, UserId:0, Amount:amount, UserAssetId:0, Date:yester_day, CreateTime:now, UpdateTime:now}
// 	if err = db.Save(&yf).Error; err != nil {
// 		glog.Error("release_ybt fail! YbtFlow add err=", err)
// 		return
// 	}

// 	// 战略投资人发放处理
// 	vs, as := getprojectybt(amount)

// 	for _, v := range vs {
// 		if err = db.Save(&v).Error; err != nil {
// 			glog.Error("release_ybt fail! UserAssetDetail add err=", err)
// 			return
// 		}
// 	}
// 	for _, v := range as {
// 		if err = db.Save(&v).Error; err != nil {
// 			glog.Error("release_ybt fail! AssetLock add err=", err)
// 			return
// 		}
// 	}
// 	return
// }

// // 计算战略投资ybt奖励
// func getprojectybt(amount float64)(vs []common.UserAssetDetail, as []common.AssetLock) {
// 	today := time.Now().Format("2006-01-02")
// 	now := time.Now().Unix()
// 	allot := conf.Config.ProjectYbtAllot
// 	for _, v := range allot {
// 		if v.Percent <= 0 {
// 			continue
// 		}
// 		m := amount * v.Percent
// 		if 0 == len(v.Users) {	// user_id和users同时存在的话 优先用users
// 			v.Users = append(v.Users, conf.UserAllot{UserId:v.UserId, Percent:1.0})
// 		}

// 		for _, u := range v.Users {
// 			mt := m * u.Percent
// 			if base.IsEqual(u.Percent, 1.0) {
// 				mt = m
// 			} else if base.IsEqual(u.Percent, 0) || u.Percent < 0{
// 				continue
// 			}

// 			vs = append(vs, common.UserAssetDetail{UserId:u.UserId, Type:common.CURRENCY_YBT, TransactionType:common.YBT_TRANSACTION_PROJECT, Amount:mt, Date:today, CreateTime:now, UpdateTime:now})
// 			if v.Forever > 0 {	// 永久冻结
// 				as = append(as, common.AssetLock{UserId:u.UserId, Type:common.CURRENCY_YBT, LockType:common.ASSET_LOCK_FOREVER, LockAmount:mt*v.Forever, Date:today, CreateTime:now, UpdateTime:now})
// 			}
// 			if v.Fix > 0 {		// 固定期限冻结
// 				unlockTime := time.Now().Unix()+ (v.FixDays*24*3600)	// 将冻结天数转成s
// 				as = append(as, common.AssetLock{UserId:u.UserId, Type:common.CURRENCY_YBT, LockType:common.ASSET_LOCK_FIX, LockAmount:mt*v.Fix, UnlockTime:unlockTime, Date:today, CreateTime:now, UpdateTime:now})
// 			}
// 		}
// 	}
// 	return
// }

// 计算用户当日的ybt奖励
func saveUserYbtDetail(db *db.PsqlDB, yesterday string) (err error) {
	uids := []uIds{}
	now := time.Now().Unix()
	if err = db.Model(&common.UserAsset{}).Select("user_id").Group("user_id").Scan(&uids).Error; err != nil {
		glog.Error("get all user_ids fail! err=", err)
		return
	}
	vs := []common.YbtUnlockDetail{}

	// 统计昨天的ybt分红，其交易流水是今天的
	if err = db.Find(&vs, "date=? and (mining>0 or activity>0 or air_drop>0)", yesterday).Error; err != nil {
		glog.Error("SaveUserYbtDetail fail! err=", err)
		return
	}
	mys := make(map[int64]*common.YbtUnlockDetail)
	for i, v := range vs {
		mys[v.UserId] = &vs[i]
	}
	var user_id int64
	for _, v := range uids {
		user_id = v.UserId
		infos, total := getYbtAmount(mys[user_id])
		r := common.BonusYbtDetail{UserId: user_id, Infos: base.StructToMap(infos), TotalYbt: total, Date: yesterday, CreateTime: now, UpdateTime: now}
		if err = db.Save(&r).Error; err != nil {
			glog.Error("SaveUserYbtDetail fail! err=", err)
			return
		}
	}

	return
}

func getYbtAmount(v *common.YbtUnlockDetail) (ret common.YbtBonusTypeAmount, total float64) {
	if v != nil {
		ret.Consume = v.Consume
		ret.Seller = v.Sale
		ret.Invite = v.Invite
		ret.Activity = v.Activity
		ret.AirDrop = v.AirDrop
		total += (v.Mining + v.Activity + v.AirDrop)
	}
	return
}

// 获取累积已发行的ybt等
func GetYbt() (v common.Ybt, err error) {
	if err = db.GetDB().First(&v).Error; err != nil {
		glog.Error("GetYbt fail! err=", err)
		return
	}
	return
}
