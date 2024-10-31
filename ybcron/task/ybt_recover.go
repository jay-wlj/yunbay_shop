package task

import (
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"

	//"yunbay/ybcron/conf"
	"fmt"
	"math"
	"time"
	"yunbay/ybasset/common"

	//base "github.com/jay-wlj/gobaselib"
	"github.com/jinzhu/gorm"
)

// 每天凌晨3点0分执行ybt空投回收
func YBT_Recover() {
	fmt.Println("YBAsset_KtRebat begin")
	now := time.Now()
	db := db.GetTxDB(nil)
	if err := recover_ybt(db); err != nil {
		glog.Error("YBT_Recover end fail! err=", err)
		db.Rollback()
		return
	}
	db.Commit()
	fmt.Println("YBT_Recover end success! tick=", time.Since(now).String())
}

type amountSt struct {
	Amount float64
}

// 平台当日分发及销毁的ybt
func recover_ybt(db *db.PsqlDB) (err error) {
	now := time.Now().Unix()
	today := time.Now().Format("2006-01-02")
	// 获取用户的资产记录(含空投总量)
	vs := []common.UserAsset{}
	if err = db.Find(&vs, "freeze_ybt>0").Error; err != nil {
		glog.Error("recover_ybt fail! err=", err)
		return
	}
	if len(vs) == 0 {
		return
	}
	db.DB = db.Model(&common.AssetLock{})
	mRecoverys := make(map[int64]float64)
	// 查询空投用户的ybt过期数量
	for _, v := range vs {
		// 获取过期ybt量
		var expired_ybt amountSt
		if err = db.Where("user_id=? and type=? and lock_type=? and lock_amount>0 and unlock_time<?", v.UserId, common.CURRENCY_YBT, common.ASSET_LOCK_AIRDROP, now).Select("sum(lock_amount) as amount").Scan(&expired_ybt).Error; err != nil {
			glog.Error("recover_ybt fail! err=", err)
			return
		}
		// 获取已释放空投ybt量
		var total_reward_ybt amountSt
		if err = db.Where("user_id=? and type=? and lock_type=? and lock_amount<0", v.UserId, common.CURRENCY_YBT, common.ASSET_LOCK_AIRDROP).Select("sum(lock_amount) as amount").Scan(&total_reward_ybt).Error; err != nil {
			glog.Error("recover_ybt fail! err=", err)
			return
		}
		// 计算应回收空投量
		recovery_ybt := expired_ybt.Amount - math.Abs(total_reward_ybt.Amount)
		if recovery_ybt > 0 {
			// 回收空投过期的部分ybt
			mRecoverys[v.UserId] = recovery_ybt
		}
	}
	if len(mRecoverys) == 0 {
		return
	}
	// 回收空投过期的ybt
	var total_air_recover float64 = 0
	yrf := []common.YbtFlow{}
	for k, v := range mRecoverys {
		// 添加空投回收用户资产记录
		// 先解冻剩余空投部分 再划款给空投池
		al := common.AssetLock{UserId: k, Type: common.CURRENCY_YBT, LockType: common.ASSET_LOCK_AIRDROP, LockAmount: -v, Date: today, CreateTime: now, UpdateTime: now}
		if err = db.Save(&al).Error; err != nil {
			glog.Error("recover_ybt fail! AssetLock err=", err)
			return
		}
		ua := common.UserAssetDetail{UserId: k, Type: common.CURRENCY_YBT, TransactionType: common.YBT_TRANSACTION_AIRDROP, Amount: -v, Date: today}
		if err = db.Save(&ua).Error; err != nil {
			glog.Error("recover_ybt fail! UserAssetDetail err=", err)
			return
		}
		yrf = append(yrf, common.YbtFlow{UserId: ua.UserId, Type: common.YBT_REWARD_AIRDROP, Amount: -v, UserAssetId: ua.Id, Date: today, CreateTime: now, UpdateTime: now})
		total_air_recover += v
	}

	// 回收记录
	for _, v := range yrf {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("recover_ybt fail! YbtFlow err=", err)
			return
		}
	}

	// 增量添加修改当天ybt空投回收
	if err = db.Model(&common.YBAssetDetail{}).Where("date=?", today).Updates(map[string]interface{}{"air_recover": gorm.Expr("air_recover + ?", total_air_recover), "update_time": now}).Error; err != nil {
		glog.Error("recover_ybt fail! YBAssetDetail err=", err)
		return
	}
	return
}

// 解锁定期冻结的资产
func YBT_UnlockFixYbt() {
	fmt.Println("YBT_UnlockFixYbt begin")
	now := time.Now()
	db := db.GetTxDB(nil)
	if err := unlock_fixybt(db); err != nil {
		glog.Error("YBT_Recover end fail! err=", err)
		db.Rollback()
		return
	}
	db.Commit()
	fmt.Println("YBT_UnlockFixYbt end success! tick=", time.Since(now).String())
}

func unlock_fixybt(db *db.PsqlDB) (err error) {
	now := time.Now().Unix()

	if err = db.Model(&common.AssetLock{}).Where("lock_type=? and lock_amount>0 and unlock_time<?", common.ASSET_LOCK_FIX, now).Updates(map[string]interface{}{"lock_amount": 0, "update_time": now}).Error; err != nil {
		glog.Error("unlock_fixybt fail! err=", err)
		return
	}
	//vs := []common.AssetLock{}
	// if err = db.Find(&vs, "lock_type=? and unlock_time<?", common.ASSET_LOCK_FIX, now).Error; err != nil {
	// 	glog.Error("unlock_fixybt fail! err=", err)
	// 	return
	// }
	// today := time.Now().Format("2006-01-02")
	// // 解锁定期冻结资产
	// for i, v := range vs {
	// 	v.Id = 0
	// 	v.LockAmount = -vs[i].LockAmount // 解锁
	// 	v.Date = today
	// 	v.CreateTime = now
	// 	v.UpdateTime = now

	// 	if err = db.Save(&v).Error; err != nil {
	// 		glog.Error("unlock_fixybt fail! err=", err)
	// 		return
	// 	}
	// }
	return
}
