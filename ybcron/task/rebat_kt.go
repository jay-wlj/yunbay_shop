package task

import (
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybcron/conf"
	"yunbay/ybcron/util"

	"github.com/jie123108/glog"
	//"github.com/jinzhu/gorm"
)

// 每天凌晨0点2分执行收益金分配脚本
func YBAsset_KtRebat() {
	now := time.Now()
	fmt.Println("YBAsset_KtRebat begin")
	db := db.GetTxDB(nil)
	if err := rebat_kt(db); err != nil {
		s := fmt.Sprintf("YBAsset_KtRebat end fail! err=%v", err)
		glog.Error(s)
		db.Rollback()
		MainSend(s)
		util.SendDingTextTalk(s, []string{"15818717950"})
		return
	}
	db.Commit()
	fmt.Println("YBAsset_KtRebat end success! tick=", time.Since(now).String())
}

// 平台当日分发及销毁的ybt
func rebat_kt(db *db.PsqlDB) (err error) {
	//args := conf.Config.Rebat
	// nTime := time.Now()
	// nTime = nTime.AddDate(0, 0, -1)
	yester_day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 判断昨日kt分红是否已经释放
	var yb common.YBAssetDetail
	if yb, err = GetYBAssetDetail(db.DB, yester_day); err != nil {
		glog.Error("GetYBAssetDetail fail! err=", err)
		return
	}
	// 昨日kt分红已发放 不用更新
	if yb.KtStatus > 0 {
		glog.Error("rebat_kt no need update kt! yb.KtStatus:", yb.KtStatus)
		return
	}

	// 获取昨日资金池交易额,利润
	_, day_profit, _, err := GetAssetPoolAmountDay(yester_day, db.DB)
	if err != nil {
		glog.Error("GetAssetPoolAmountDay fail! err=", err)
		return
	}

	now := time.Now().Unix()
	// 获取平台累计已发行可分红的YBT
	total_bonus_ybt := yb.BonusYbt
	// total_bonus_ybt, _, err := GetAssetBonusAmountDay(yester_day, db.DB)
	// if  err != nil {
	// 	glog.Error("GetAssetBonusAmountDay  fail! err", err.Error())
	// 	return
	// }

	if total_bonus_ybt <= 0 {
		glog.Error("day_profit < 0 || total_bonus_ybt < 0")
		return
	}
	// 获取平台累积持有的可分红的ybt
	var mUserBonus map[int64]float64
	var mings map[int64]float64
	var mair_unlock map[int64]float64
	var mproject map[int64]float64

	if mUserBonus, err = GetUserBonusAmountDay(db.DB, yester_day); err != nil {
		glog.Error("GetUserBonusAmount fail! err:", err)
		return
	}
	if mings, mair_unlock, mproject, err = GetUserUnlockYbts(db.DB, yester_day); err != nil {
		glog.Error("errGetUserUnlockYbts fail! err:", err)
		return
	}
	// 将释放的用户id添加到
	for k, _ := range mings {
		if _, ok := mUserBonus[k]; !ok {
			mUserBonus[k] += 0
		}
	}
	for k, _ := range mair_unlock {
		if _, ok := mUserBonus[k]; !ok {
			mUserBonus[k] += 0
		}
	}
	for k, _ := range mproject {
		if _, ok := mUserBonus[k]; !ok {
			mUserBonus[k] += 0
		}
	}
	// vs := []common.UserAssetDetail{}
	// bks := []common.BonusKtDetail{}		// 记录用户每天的可分红的ybt及收益
	ks := []common.KtBonusDetail{}

	// 平台有收益才会有用户的kt收益资产记录
	var bonus_kt float64 = 0
	var bonus_ybt float64 = 0

	// 获取用户的挖矿及空投释放
	// var ms map[int64]unlockSt
	// if ms, err = GetYbtUnlock(db.DB, yester_day); err != nil {
	// 	glog.Error("KtBonusDetail GetYbtUnlock fail! err=", err)
	// 	return
	// }

	glog.Info("rebat_kt total_bonus_ybt:", total_bonus_ybt)
	for k, v := range mUserBonus {
		bonus_amount := v + mings[k] + mair_unlock[k] + mproject[k]
		bonus_ybt += bonus_amount
		if bonus_amount > base.FLOAT_MIN {
			if bonus_amount > total_bonus_ybt {
				glog.Error("用户所持的有ybt不可能 bonus_amount:", bonus_amount, " total_bonus_ybt:", total_bonus_ybt)
			}
			k := common.KtBonusDetail{UserAsset: common.UserAsset{UserId: k, UpdateTime: now}, Date: yester_day, Mining: mings[k], Project: mproject[k], BonusYbt: bonus_amount, AirUnlock: mair_unlock[k], CheckStatus: common.STATUS_INIT}
			k.BonusPercent = bonus_amount / total_bonus_ybt
			if day_profit > base.FLOAT_MIN {
				k.KtBonus = k.BonusPercent * day_profit // (昨日用户持有的YBT/昨日累计平台已发行可分红的YBT)*昨日收益金KT
			}

			bonus_kt += k.KtBonus
			ks = append(ks, k)
		}
	}
	if !base.IsEqual(total_bonus_ybt, bonus_ybt) {
		//glog.Error("mUserBonus:", mUserBonus, " mings:", mings, " airunlock:", mair_unlock)
		s := fmt.Sprintf("total_bonus_ybt != bonus_ybt fail! total_bonus_ybt:%v bonus_ybt:%v diff_ybt:%v", total_bonus_ybt, bonus_ybt, total_bonus_ybt-bonus_ybt)
		glog.Error(s)
		// err = fmt.Errorf(s)
		// return
	}

	if !base.IsEqual(bonus_kt, day_profit) {
		s := fmt.Sprintf("bonus_kt:%v != day_profit:%v", bonus_kt, day_profit)
		glog.Error(s)

		// 多余的钱存入指定帐户中
		if day_profit > bonus_kt {
			last_kt := day_profit - bonus_kt
			last_bonusid := conf.Config.SystemAccounts["last_bonus_account"]
			if last_bonusid == 0 {
				last_bonusid = 10
				glog.Error("rebat_kt last_bonusid=0, then change last_bonusid=", 10)
			}
			ks = append(ks, common.KtBonusDetail{UserAsset: common.UserAsset{UserId: last_bonusid, CreateTime: now, UpdateTime: now}, Date: yester_day, KtBonus: last_kt, BonusPercent: last_kt / total_bonus_ybt, CheckStatus: common.STATUS_INIT})
			glog.Debug("day_profit-bonus_kt=", last_kt, " then turn it into account ", last_bonusid)
		}
		// err = fmt.Errorf(s)
		// return
	}

	// 更新参与分红的人数
	if err = db.Model(&yb).Updates(map[string]interface{}{"bonusers": len(ks), "update_time": now}).Error; err != nil {
		glog.Error("KtBonusDetail update bonusers fail! err=", err)
		return
	}

	// 批量添加数据 保存每天每个用户的kt收益明细
	for _, v := range ks {
		db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (user_id, date) DO update set mining=%v, air_unlock=%v, project=%v, bonus_ybt=%v, bonus_percent=%v, kt_bonus=%v, update_time=%v", v.Mining, v.AirUnlock, v.Project, v.BonusYbt, v.BonusPercent, v.KtBonus, now))
		if err = db.Save(&v).Error; err != nil {
			glog.Error("KtBonusDetail create fail! err=", err)
			return
			err = nil
		}
	}
	util.SendDingTextTalk(fmt.Sprintf("----kt收益:%v----\r\n今日营收:%v\r\n总可分红ybt:%.4f\r\n现可分红ybt:%.4f\r\n剩余收益:%.4f\r\n参与收益人数:%v", yester_day, yb.Profit, total_bonus_ybt, bonus_ybt, day_profit-bonus_kt, len(ks)), nil)
	return
}

// type unlockSt struct {
// 	UserId int64
// 	AirUnlock float64
// 	Mining float64
// }
// func GetYbtUnlock(db *gorm.DB, date)(ms map[int64]unlockSt, err error) {
// 	vs := []unlockSt{}
// 	if err = db.Model(&common.YbtUnlockDetail{}).Where("date=? and (air_unlock>0 or mining>0)").Select("user_id, air_unlock, mining").Scan(&vs).Error; err != nil {
// 		glog.Error("GetYbtUnlock fail! err=", err)
// 		return
// 	}
// 	ms = make(map[int64]unlockSt)
// 	for _, v := range vs {
// 		ms[v.UserId] = v
// 	}
// 	return
// }
