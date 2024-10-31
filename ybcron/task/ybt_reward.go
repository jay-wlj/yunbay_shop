package task

import (
	"fmt"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybcron/conf"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
	//"github.com/jinzhu/gorm"
)

// 每隔5分钟执行一次ybt注册空投赠送
func Ybt_Reward() {
	fmt.Println("Ybt_Reward begin")
	now := time.Now()
	db := db.GetTxDB(nil)
	if err := reward_ybt(db); err != nil {
		glog.Error("Ybt_Reward end fail! err=", err)
		db.Rollback()
		return
	}
	db.Commit()
	fmt.Println("Ybt_Reward end success! tick=", time.Since(now).String())
}

// 发放赠送的ybt空投奖励等
func reward_ybt(db *db.PsqlDB) (err error) {
	now := time.Now().Unix()
	today := time.Now().Format("2006-01-02")
	// 获取奖励池可发放量
	ybt, err := GetYbt()
	if err != nil {
		glog.Error("reward_ybt fail! err=", err)
		return
	}
	if base.IsEqual(ybt.NormalReward, base.FLOAT_MIN) {
		glog.Error("ybt normal_reward is zeor, can't reward_ybt")
		return
	}
	// 获取用户的资产记录(含空投总量)
	vs := []common.RewardRecord{}
	if err = db.Order("create_time asc").Find(&vs, "type=? and release_type=? and status=? and lock=?", common.CURRENCY_YBT, common.YBT_REWARD_AIRDROP, common.STATUS_INIT, false).Error; err != nil {
		glog.Error("reward_ybt fail! err=", err)
		return
	}
	if len(vs) == 0 {
		return
	}
	// 计算共需释放空投量
	var total_amount float64 = 0
	var end_user int = len(vs)
	for i, v := range vs {
		total_amount += v.Amount
		if total_amount > ybt.NormalReward {
			end_user = i
			break
		}
	}

	yrf := []common.YbtFlow{}
	// 计算回收空投奖励时间
	unlock_time := time.Now().Unix() + (conf.Config.AirdopCfg.Timeout * 24 * 3600)
	// 从奖励池中发放空投奖励
	// 不足只能给前面的用户空投奖励 后面的就没有了
	for i := 0; i < end_user; i++ {
		// 先划款给用户 再冻结剩余空投部分
		// 添加空投赠送用户资产记录
		v := vs[i]
		ua := common.UserAssetDetail{UserId: v.UserId, Type: common.CURRENCY_YBT, TransactionType: common.YBT_TRANSACTION_AIRDROP, Amount: v.Amount, Date: today}
		if err = db.Save(&ua).Error; err != nil {
			glog.Error("recover_ybt fail! UserAssetDetail err=", err)
			return
		}
		al := common.AssetLock{UserId: v.UserId, Type: common.CURRENCY_YBT, LockType: common.ASSET_LOCK_AIRDROP, LockAmount: v.Amount, Date: today, UnlockTime: unlock_time, CreateTime: now, UpdateTime: now}
		if err = db.Save(&al).Error; err != nil {
			glog.Error("recover_ybt fail! AssetLock err=", err)
			return
		}

		yrf = append(yrf, common.YbtFlow{UserId: ua.UserId, Type: common.YBT_REWARD_AIRDROP, Amount: v.Amount, UserAssetId: ua.Id, Date: today, CreateTime: now, UpdateTime: now})
	}

	// 从奖励池中释放记录
	for _, v := range yrf {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("recover_ybt fail! YbtFlow err=", err)
			return
		}
	}

	ids := []int64{}
	for _, v := range vs {
		ids = append(ids, v.Id)
	}
	// 更新发放状态
	if err = db.Model(&common.RewardRecord{}).Where("id in(?)", ids).Updates(map[string]interface{}{"status": common.STATUS_OK, "update_time": now}).Error; err != nil {
		glog.Error("recover_ybt fail! RewardRecord err=", err)
		return
	}
	return
}
