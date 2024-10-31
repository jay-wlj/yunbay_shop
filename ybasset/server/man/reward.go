package man

import (
	"fmt"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"time"
	"yunbay/ybasset/common"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/shopspring/decimal"

	//"yunbay/ybasset/conf"
	//"github.com/jinzhu/gorm"
	base "github.com/jay-wlj/gobaselib"
	"yunbay/ybasset/util"
)

type RewarSt struct {
	UserId int64           `json:"user_id" binding:"required"`
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

type activitySt struct {
	Activitys   []RewarSt `json:"activitys" binding:"required"`
	ReleaseType int       `json:"release_type"`
	FixDays     int       `json:"fixdays"`
	Reason      string    `json:"reason"`
}

func Man_YbtActivityReward(c *gin.Context) {
	maner, err := util.GetHeaderString(c, "X-Yf-Maner")
	if maner == "" || err != nil {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var req activitySt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	if req.FixDays < 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	// 释放ybt
	db := db.GetTxDB(c)
	if reason, err := release_activity(db, req, maner); err != nil {
		glog.Error("release_activity fail! err=", err)
		if reason == "" {
			reason = yf.ERR_SERVER_ERROR
		}
		yf.JSON_Fail(c, reason)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

func release_activity(db *db.PsqlDB, req activitySt, maner string) (reason string, err error) {
	var ybt common.Ybt
	if ybt, err = GetYbt(); err != nil {
		glog.Error("Man_YbtActivityReward fail! err=", err)
		return
	}
	var total_amount decimal.Decimal
	for _, v := range req.Activitys {
		if v.Amount.IsNegative() {
			reason = yf.ERR_ARGS_INVALID
			err = fmt.Errorf(reason)
			return
		}
		total_amount = total_amount.Add(v.Amount)
	}
	now := time.Now().Unix()
	today := time.Now().Format("2006-01-02")

	// release_type := common.YBT_REWARD_ACTIVITY
	// switch req.ReleaseType {
	// case "minepool":
	// case "project":
	// 	err = fmt.Errorf("release_activity fail! not surpot type:", req.ReleaseType)
	// 	return
	// default:

	// }
	// 默认从奖励池中释放

	if total_amount.GreaterThan(decimal.NewFromFloat(ybt.NormalReward)) {
		reason = common.ERR_YBT_USER_REWARD_NOT_MORE
		err = fmt.Errorf(reason)
		return
	}
	release_type := common.YBT_REWARD_AIRDROP
	transaction_type := common.YBT_TRANSACTION_AIRDROP
	switch req.ReleaseType {
	case 0:
		release_type = common.YBT_REWARD_AIRDROP
		transaction_type = common.YBT_TRANSACTION_AIRDROP
	case 1:
		release_type = common.YBT_REWARD_ACTIVITY
		transaction_type = common.YBT_TRANSACTION_ACTIVITY
	default:
		reason = yf.ERR_ARGS_INVALID
		err = fmt.Errorf(reason)
		return
	}

	// 记录ybt释放流水
	yfs := []common.YbtFlow{}
	als := []common.AssetLock{}
	grs := []common.RewardRecord{}
	var unlock_time int64 = 0

	if req.FixDays > 0 {
		unlock_time = now + int64(req.FixDays*24*3600) // 将冻结天数转成s
	}

	// 添加用户资产明细
	for _, v := range req.Activitys {
		amount, _ := v.Amount.Float64()
		if v.Amount.IsPositive() {
			u := common.UserAssetDetail{UserId: v.UserId, Type: common.CURRENCY_YBT, TransactionType: transaction_type, Amount: amount, Date: today}
			if err = db.Save(&u).Error; err != nil {
				glog.Error("release_activity UserAssetDetail fail! err=", err)
				return
			}
			if release_type == common.YBT_REWARD_AIRDROP {
				unlock_time = now + int64(req.FixDays*24*3600) // 将冻结天数转成s
				als = append(als, common.AssetLock{UserId: v.UserId, Type: common.CURRENCY_YBT, LockType: common.ASSET_LOCK_AIRDROP, LockAmount: amount, Date: today, CreateTime: now, UpdateTime: now})
			} else if unlock_time > 0 {
				als = append(als, common.AssetLock{UserId: v.UserId, Type: common.CURRENCY_YBT, LockType: common.ASSET_LOCK_FIX, LockAmount: amount, UnlockTime: unlock_time, Date: today, CreateTime: now, UpdateTime: now})
			}
			grs = append(grs, common.RewardRecord{UserId: v.UserId, Type: common.CURRENCY_YBT, ReleaseType: req.ReleaseType, Fixdays: req.FixDays, Amount: amount, Reason: req.Reason, Maner: maner, Date: today, Status: common.STATUS_OK, CreateTime: now, UpdateTime: now})
			yfs = append(yfs, common.YbtFlow{UserId: v.UserId, Type: release_type, Amount: amount, UserAssetId: u.Id, Date: today, CreateTime: now, UpdateTime: now})
		}
	}
	// 冻结相应ybt
	for _, v := range als {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("release_activity AssetLock fail! err=", err)
			return
		}
	}
	// 生成ybt释放流水
	for _, v := range yfs {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("release_activity YbtFlow fail! err=", err)
			return
		}
	}
	// 生成赠送记录
	for _, v := range grs {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("release_activity RewardRecord fail! err=", err)
			return
		}
	}
	return
}

func Man_GiftRewardList(c *gin.Context) {

	id, _ := base.CheckQueryInt64DefaultField(c, "id", 0)
	date, _ := base.CheckQueryStringField(c, "date")
	user_id, _ := base.CheckQueryInt64DefaultField(c, "user_id", -1)
	release_type, _ := base.CheckQueryInt64DefaultField(c, "release_type", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")

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
	if begin_date != "" {
		db.DB = db.Where("date>=?", begin_date)
	}
	if end_date != "" {
		db.DB = db.Where("date<=?", end_date)
	}
	if release_type > -1 {
		db.DB = db.Where("release_type=?", release_type)
	}
	var total int = 0
	if err := db.Model(&common.RewardRecord{}).Count(&total).Error; err != nil {
		glog.Error("Man_GiftRewardList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	db.DB = db.ListPage(page, page_size)
	vs := []common.RewardRecord{}
	if err := db.Order("id desc").Find(&vs).Error; err != nil {
		glog.Error("Man_RewardRecordList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": vs, "list_page": base.IsListEnded(page, page_size, len(vs), total), "total": total})
}

// type rewardSt struct {
// 	Rewards []RewarSt `json:"rewards"`
// 	ReleaseType string `json:"release_type"`
// }

// // kt赠送奖励
// func Man_KtGiftReward(c *gin.Context) {
// 	maner, err := util.GetHeaderString(c, "X-Yf-Maner")
// 	if maner == "" || err != nil {
// 		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
// 		return
// 	}
// 	var req rewardSt
// 	if ok := util.UnmarshalReq(c, &req); !ok {
// 		return
// 	}

// 	// 释放ybt
// 	db := db.GetTxDB(c)
// 	if reason, err := gift_kt(db, req, maner); err != nil {
// 		glog.Error("release_activity fail! err=", err)
// 		if reason == "" {
// 			reason = yf.ERR_SERVER_ERROR
// 		}
// 		yf.JSON_Fail(c, reason)
// 		return
// 	}

// 	yf.JSON_Ok(c, gin.H{})
// }

// // 从平台帐户转出kt赠送给其它人
// func gift_kt(db *db.PsqlDB, req rewardSt, maner string)(reason string, err error) {
// 	var total_amount float64 = 0
// 	for _, v := range req.Rewards {
// 		if v.Amount < 0 {
// 			reason = yf.ERR_ARGS_INVALID
// 			err = fmt.Errorf(reason)
// 			return
// 		}
// 		total_amount += v.Amount
// 	}
// 	// 获取平台帐户kt资产
// 	var yb common.UserAsset
// 	if err = db.Find(&yb, "user_id=", 0).Error; err != nil {
// 		glog.Error("gift_kt fail! find user_asset err=", err)
// 		return
// 	}
// 	if yb.NormalKt < total_amount {
// 		glog.Error("gift_kt fail! find user_asset err=", err)
// 		reason = common.ERR_KT_NOT_MORE
// 		err = fmt.Errorf(reason)
// 		return
// 	}
// 	now := time.Now().Unix()
// 	today := time.Now().Format("2006-01-02")
// 	vs := []common.UserAssetDetail{}
// 	for _, v := range req.Rewards {
// 		vs = append(vs, common.UserAssetDetail{UserId:v.UserId, Type:common.CURRENCY_KT, TransactionType:KT_TRANSACTION_GIFT, Amount:v.Amount, Date:today, CreateTime:now, UpdateTime:now})
// 	}
// 	// 从项目方用户中扣除相应的kt
// 	vs = append(vs, common.UserAssetDetail{UserId:0, Type:common.CURRENCY_KT, TransactionType:KT_TRANSACTION_GIFT, Amount:-v.Amount, Date:today, CreateTime:now, UpdateTime:now})

// 	// 生成用户资产明细
// 	for _, v := range vs {
// 		if err = db.Create(&v).Error; err != nil {
// 			glog.Error("gift_kt fail! UserAssetDetail create fail! err=", err)
// 			return
// 		}
// 	}
// 	return
// }

type useridSt struct {
	UserId int64 `json:"user_id"`
}

// 解锁空投奖励状态
func Man_UnlockReward(c *gin.Context) {
	var req useridSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		glog.Error("Man_UnlockReward fail! args invalid!", req)
		return
	}
	db := db.GetTxDB(c)
	if err := db.Model(&common.RewardRecord{}).Where("invite_id=?", req.UserId).Update(map[string]interface{}{"lock": false, "update_time": time.Now().Unix()}).Error; err != nil {
		glog.Error("Man_UnlockReward fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}
