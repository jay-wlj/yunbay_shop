package task

import (
	"yunbay/ybasset/common"

	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"math"
	"time"
	"yunbay/ybcron/conf"
	"yunbay/ybcron/util"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 每天凌晨0点2分执行收益金分配脚本
func Ybt_Rebat() {
	fmt.Println("Ybt_Rebat begin")
	now := time.Now()
	db := db.GetTxDB(nil)
	if err := rebat_ybt(db); err != nil {
		s := fmt.Sprintf("Ybt_Rebat end fail! err=%v", err)
		glog.Error(s)
		db.Rollback()
		MainSend(s)
		util.SendDingTextTalk(s, []string{"15818717950"})
		return
	}
	db.Commit()
	fmt.Println("Ybt_Rebat end success! tick=", time.Since(now).String())
}

type YbtUnlockType struct {
	Consume     float64 // 用户消费奖励
	Seller      float64 // 用户商家奖励
	Invite      float64 // 用户邀请奖励
	Activity    float64 // 活动奖励
	AirUnlock   float64 // 空投释放奖励
	AirDrop     float64 // 空投奖励
	Project     float64 // 项目奖励
	Rebat       float64 // 贡献值
	SellerRebat float64 // 商家贡献值
}

type rebatDatas struct {
	UserUnlock    map[int64]YbtUnlockType // 用户ybt解锁信息
	AirUnlock     float64                 // 当日空投释放
	AirDrop       float64                 // 当日空投奖励
	Activity      float64                 // 当日活动释放
	ProjectUnlock float64                 // 当日项目释放
	Consumer      int                     // 当日消费人数
	OrderYbts     []common.Ordereward     // 当日订单探矿记录
}

type userUnlock struct {
	Id       int64
	UserId   int64
	Activity float64
}

// 平台当日分发及销毁的ybt
func rebat_ybt(db *db.PsqlDB) (err error) {
	nTime := time.Now()
	yester_day := nTime.AddDate(0, 0, -1).Format("2006-01-02")
	now := time.Now().Unix()
	// 获取昨日资金池交易额,利润
	day_amount, day_profit, pool, err := GetAssetPoolAmountDay(yester_day, db.DB)

	// 判断昨日kt分红是否已经释放
	var v common.YBAssetDetail
	if v, err = GetYBAssetDetail(db.DB, yester_day); err != nil {
		glog.Error("GetYBAssetDetail fail! err=", err)
		return
	}
	// 昨日ybt已发放 不用更新
	if v.YbtStatus > 0 {
		glog.Error("rebat_ybt no need update kt! v.YbtStatus:", v.YbtStatus)
		return
	}

	if err == gorm.ErrRecordNotFound {
		v.Period = GetLastPeriod()
	}
	v.Mining, v.Difficult, err = GetMiningYbt(day_profit)
	if err != nil {
		glog.Error("GetCanIssueYbt  fail! err", err)
		return
	}

	v.Amount = day_amount
	v.Profit = day_profit
	v.Date = yester_day
	v.CreateTime = now
	v.UpdateTime = now
	// 平台ybt分红额度记录(昨日的平台分红记录)
	// v = common.YBAssetDetail{Amount:day_amount, Profit:day_profit, Date:yester_day, IssueYbt:v.IssueYbt, Period:v.Period, Difficult:v.Difficult, CreateTime:now, UpdateTime:now}
	if err = db.Save(&v).Error; err != nil {
		glog.Error("YBAssetDetail  fail! err", err.Error())
		return
	}

	if base.IsEqual(v.Mining, base.FLOAT_MIN) {
		glog.Error("issue_ybt is 0!")
		return
	}
	// 分配用户ybt返利记录
	ret, err := AllocUserYbt(db, v.Profit, v.Mining, pool)
	if err != nil {
		glog.Error("AllocUserYbt  fail! err", err)
		return
	}

	// 添加订单探矿记录
	for _, v := range ret.OrderYbts {
		db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (order_id) DO UPDATE SET buyer_userid=%v,seller_userid=%v,recommender_userid=%v,recommender2_userid=%v,yunbay_userid=%v,ybt=%v, buyer_ybt=%v,seller_ybt=%v,recommender_ybt=%v,recommender2_ybt=%v,yunbay_ybt=%v,update_time=%v",
			v.BuyerUserId, v.SellerUserId, v.ReUserId, v.Re2UserId, v.YunbayUserId, v.Ybt, v.BuyerYbt, v.SellerYbt, v.ReYbt, v.Re2Ybt, v.YunbayYbt, now))
		if err = db.Create(&v).Error; err != nil {
			glog.Error("UserAssetDetail create fail! err", err)
			return
		}
	}
	db.DB = db.Set("gorm:insert_option", "")
	// 获取昨日的用户活动奖励
	var myud map[int64]common.YbtUnlockDetail
	if myud, err = GetYbtUnlockDetail(db, yester_day); err != nil {
		glog.Error("GetUserActivityReward fail! err", err)
		return
	}
	// 更新昨日空投奖励
	// if v.AirDrop, err = GetAirDrop(db.DB, yester_day); err != nil {
	// 	glog.Error("GetAirDrop fail! err", err)
	// 	return
	// }
	// 更新昨日所有锁定的ybt
	if v.FreezeYbt, err = GetFreezedYbt(db.DB); err != nil {
		glog.Error("GetAirDrop fail! err", err)
		return
	}

	// 更新昨日空投释放和项目释放
	v.AirDrop = ret.AirDrop
	v.AirUnlock = ret.AirUnlock
	v.Project = ret.ProjectUnlock
	v.Activity = ret.Activity
	v.IssueYbt = v.Mining + v.Activity + v.AirUnlock + v.Project // 昨日总释放 = 挖矿释放+活动释放+空投+项目方释放
	v.Miners = ret.Consumer                                      // 消费人数即挖矿人数

	// 保存昨日平台数据
	if err = db.Save(&v).Error; err != nil {
		glog.Error("YBAssetDetail  fail! err", err.Error())
		return
	}
	// 更新平台总交易及总释放信息
	if err = update_yunbay_asset(db.DB, yester_day); err != nil {
		glog.Error("update_yunbay_asset  fail! err", err.Error())
		return
	}
	total_unlock := v.IssueYbt
	// 生成用户挖矿明细
	var us []common.YbtUnlockDetail
	if us, err = MakeAndUpdateYbtUnlockDetail(ret.UserUnlock, day_profit, total_unlock, myud); err != nil {
		glog.Error("MakeAndUpdateYbtUnlockDetail  fail! err", err.Error())
		return
	}

	// 保存记录
	for _, v := range us {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("rebat_ybt fail! err=", err)
			return
		}
	}

	// 更新平台昨日可分红的ybt
	total_issue, total_bonus_ybt, err := GetAssetBonusAmountDay(yester_day, db.DB)
	if err = db.Model(&common.YBAssetDetail{}).Where("date=?", yester_day).Updates(map[string]interface{}{"bonus_ybt": total_bonus_ybt, "update_time": now}).Error; err != nil {
		glog.Error("YBAssetDetail fail! err=", err)
		return
	}

	if total_bonus_ybt > 0 {
		v.Perynbay = (1 / total_bonus_ybt) * v.Profit
		if err = db.Model(&v).Updates(map[string]interface{}{"perynbay": v.Perynbay}).Error; err != nil {
			glog.Error("YBAssetDetail  fail! err", err.Error())
			return
		}
		if err = db.Model(&common.TradeFlow{}).Where("date=?", yester_day).Updates(map[string]interface{}{"perynbay": v.Perynbay}).Error; err != nil {
			glog.Error("TradeFlow update fail! err=?", err)
			err = nil //直接返回nil
		}
	}

	// 更新今日探矿难度
	if err = updateDiffcultByTotalIssue(db, total_issue); err != nil {
		glog.Error("YBAssetDetail  fail! err", err.Error())
		return
	}
	// today := time.Now().Format("2006-01-02")
	// t := common.YBAssetDetail{Difficult:getdifficultbyrelease(total_issue), Date:today, CreateTime:now, UpdateTime:now}
	// db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (date) DO UPDATE SET difficult=%v, update_time=%v", t.Difficult, t.UpdateTime))
	// if err = db.Save(&t).Error; err != nil {
	// 	glog.Error("YBAssetDetail  fail! err", err.Error())
	// 	return
	// }

	util.SendDingTextTalk(fmt.Sprintf("----ybt释放计划:%v----\r\n今日交易额:%v\r\n今日营收:%v\r\n挖矿难度:%v\r\n挖矿释放:%.4f\r\n空投释放:%v\r\n共释放ybt:%.4f\r\n挖矿人数:%v", yester_day, v.Amount, v.Profit, v.Difficult, v.Mining, v.AirUnlock, v.IssueYbt, v.Miners), nil)
	return
}

// 更新当日挖矿难度
func updateDiffcultByTotalIssue(db *db.PsqlDB, total_issue float64) (err error) {
	today := time.Now().Format("2006-01-02")
	now := time.Now().Unix()
	t := common.YBAssetDetail{Difficult: getdifficultbyrelease(total_issue), Date: today, CreateTime: now, UpdateTime: now}
	db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (date) DO UPDATE SET difficult=%v, update_time=%v", t.Difficult, t.UpdateTime))
	if err = db.Save(&t).Error; err != nil {
		glog.Error("YBAssetDetail  fail! err", err.Error())
		return
	}

	cache, er := cache.GetWriter("pub")
	if er != nil {
		glog.Error("updateDiffcultByTotalIssue fail! err=", er)
		return
	}
	if cache != nil {
		cache.HDel("diffcult", today) // 删除缓存
	}
	return
}

// 可挖出的ybt
func GetMiningYbt(profit float64) (issue_ybt, difficule float64, err error) {
	// 获取当期挖矿难度
	difficule = getdifficult()

	// 获取昨天总挖矿产出
	issue_ybt = profit / difficule

	// 判断累计挖出的ybt不能大于总量
	ybt, err1 := GetYbt()
	if err1 != nil {
		glog.Error("GetYbt err=", err)
		err = err1
		return
	}
	if ybt.LockMinepool < issue_ybt {
		glog.Error("GetMiningYbt LockMinepool < issue_ybt", " LockMinepool:", ybt.LockMinepool, " issue_ybt:", issue_ybt)
		issue_ybt = ybt.LockMinepool
	}
	return
}

// 分配挖出的ybt
func AllocUserYbt(db *db.PsqlDB, day_profit, issue_ybt float64, pool []common.YBAssetPool) (ret rebatDatas, err error) {
	now := time.Now().Unix()
	yester_day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var yunbay_amount float64 = 0 // 项目所得总ybt

	munlocks := make(map[int64]float64)           // 用户获取的总ybt挖矿释放奖励
	mrebat := make(map[int64]float64)             // 用户贡献值
	msellerebat := make(map[int64]float64)        // 商家贡献值
	if base.IsEqual(day_profit, base.FLOAT_MIN) { // 平台昨日总利润等于0，平台所发放的ybt全归平台所有
		yunbay_amount = issue_ybt
	} else { // 平台昨日总利润大于0
		// 获取昨日已发货的订单id列表
		var mOrderIds map[int64]bool
		var order_ids []int64
		if order_ids, mOrderIds, err = getOrderIdsByDateStatus(yester_day, common.ORDER_STATUS_SHIPPED); err != nil {
			glog.Error("AllocUserYbt fail! getOrderIdsByDateStatus err=", err)
		}
		var validCount int = 0
		for _, v := range pool {
			rebat_ybt := (v.RebatAmount / day_profit) * issue_ybt // 每笔交易的挖矿额度= 每笔交易产生的贡献值/当日全平台总利润（总贡献值） * 当日产生的代币数

			// 获取用户的推荐人
			var recommend_userids []int64
			recommend_userids, err = util.GetRecommenders(v.PayerUserId) //man.GetUserBeInvite(v.PayerUserId)
			if err != nil {
				glog.Error("GetUserBeInvite  fail! err", err)
				return
			}
			rebat := conf.Config.Rebat
			// 订单挖矿分配
			or := common.Ordereward{OrderId: v.OrderId, Ybt: rebat_ybt, BuyerUserId: v.PayerUserId, BuyerYbt: rebat_ybt * rebat.BuyerRebatPercent, SellerUserId: v.SellerUserId, SellerYbt: rebat_ybt * rebat.SellerRebatPercent, Date: yester_day, CreateTime: now, UpdateTime: now}
			munlocks[or.BuyerUserId] += or.BuyerYbt // 买家
			//munlocks[or.SellerUserId] += or.SellerYbt		// 商家
			// 对商家已发货的订单才1:1释放给商家
			if mOrderIds[v.OrderId] {
				mOrderIds[v.OrderId] = false // 为下面判断是否不在昨日订单中
				validCount += 1

				or.Valid = common.STATUS_OK               // 标记该笔订单已发货 是有效的订单
				munlocks[or.SellerUserId] += or.SellerYbt // 商家参与1:1释放解锁
				//glog.Error("munlocks[", or.SellerUserId,"]", "=", or.SellerYbt)
			}

			mrebat[or.BuyerUserId] += v.RebatAmount
			msellerebat[or.SellerUserId] += v.RebatAmount
			if len(recommend_userids) > 0 {
				or.ReUserId = recommend_userids[0]
				//or.ReYbt = rebat_ybt * rebat.RecommendRebatPercent
				or.ReYbt = rebat_ybt - or.BuyerYbt - or.SellerYbt
				munlocks[or.ReUserId] += or.ReYbt // 经济人
			}
			or.YunbayUserId = 0
			if or.ReYbt == 0 { // 没有经纪人的奖励转入系统帐号
				or.YunbayYbt = rebat_ybt - or.BuyerYbt - or.SellerYbt - or.ReYbt - or.Re2Ybt
			}

			ret.OrderYbts = append(ret.OrderYbts, or)
			yunbay_amount += or.YunbayYbt // 项目所得ybt
			// if or.SellerUserId == 0 {		// 商家是平台方
			// 	yunbay_amount += or.SellerYbt
			// }
		}

		if validCount != len(order_ids) {
			// 剩余的订单id在昨天以前的订单挖矿中
			oIds := []int64{}
			for k, v := range mOrderIds {
				if v { // 剔除掉上面已生效的订单
					oIds = append(oIds, k)
				}
			}
			if len(oIds) == len(order_ids)-validCount && len(oIds) > 0 {
				var orders []common.Ordereward
				if err = db.Model(&common.Ordereward{}).Find(&orders, "valid=0 and order_id in(?)", oIds).Error; err != nil {
					glog.Error("AllocUserYbt fail! find ordereward order_ids=", oIds)
					return
				}
				for _, v := range orders {
					munlocks[v.SellerUserId] += v.SellerYbt // 前几天的订单今天发货也要1:1释放ybt
				}
			} else {
				glog.Error("AllocUserYbt len(oIds)=", len(oIds), "len(order_ids)-validCount=", len(order_ids)-validCount, "oIds:", oIds, " order_ids:", order_ids)
			}
		}
		// 更新昨天及以前所有订单挖矿的有效状态
		if err = db.Model(&common.Ordereward{}).Where("valid=0 and date<? and order_id in(?)", yester_day, order_ids).Updates(map[string]interface{}{"valid": common.STATUS_OK, "update_time": now}).Error; err != nil {
			glog.Error("AllocUserYbt update ordereward valid fail! err=", err)
			return
		}

	}

	// 解冻用户的空投冻结ybt
	var mair_unlock map[int64]float64
	var mair_drop map[int64]float64
	//glog.Error("munlocks[", munlocks)
	mair_unlock, err = GetUnLockUserAsset(munlocks)
	//glog.Error("mair_unlock[", mair_unlock)
	if err != nil {
		glog.Error("GetUnLockUserAsset  fail! err", err)
		return
	}
	// 获取用户空投记录
	if mair_drop, ret.AirDrop, err = GetUserAirDrop(db.DB, yester_day); err != nil {
		glog.Error("GetUserAirDrop  fail! err", err)
		return
	}

	var air_unlock float64 = 0
	//	计算挖矿释放总量
	for _, v := range mair_unlock {
		air_unlock += v
	}
	ret.AirUnlock = air_unlock

	// 预发行项目方释放量 = 挖矿释放量
	project_amount := issue_ybt
	// 判断是否超出初始发行
	var ybt common.Ybt
	ybt, err = GetYbt()
	if err != nil {
		glog.Error("GetYbt fail! err", err)
		return
	}
	// 释放的项目方ybt不能超过冻结的
	if ybt.LockProject < project_amount {
		project_amount = ybt.LockProject
	}
	ret.ProjectUnlock = project_amount

	// 项目的云贝奖励冻结部分处理
	//ret.AssetLocks = append(ret.AssetLocks, AllocYunbayYbt(yunbay_amount)...)
	// 添加资产记录
	mconsume, mseller, minvite := GetAssetDetailByOrdereward(ret.OrderYbts, yunbay_amount)
	ret.Consumer = len(mconsume) // 消费人数

	// 获取用户活动释放记录
	var mactivtys map[int64]float64
	if mactivtys, ret.Activity, err = GetUserActivity(db.DB, yester_day); err != nil {
		glog.Error("GetUserActivity  fail! err", err)
		return
	}
	// 获取项目释放的奖励
	var mprojects map[int64]float64
	mprojects = getprojectybt(project_amount)
	ret.UserUnlock = MakeUserUnlockTypes(mconsume, minvite, mseller, mair_drop, mair_unlock, mactivtys, mprojects, mrebat, msellerebat)
	return
}

func GetAssetDetailByOrdereward(rewards []common.Ordereward, yunbay_amount float64) (mconsume, mseller, minvite map[int64]float64) {
	mconsume = make(map[int64]float64) // 用户昨天消费奖励
	mseller = make(map[int64]float64)  // 用户昨天商家奖励
	minvite = make(map[int64]float64)  // 用户昨天经济人奖励

	// 平台挖矿奖励以经济人方式发放
	if yunbay_amount > 0 {
		minvite[0] = yunbay_amount
	}
	bzjAccountId := conf.Config.SystemAccounts["bzj_account"]
	if bzjAccountId == 0 {
		glog.Error("GetAssetDetailByOrdereward fail! bzjAccountId==0")
		return
	}
	for _, v := range rewards {
		if v.BuyerUserId > 0 {
			mconsume[v.BuyerUserId] += v.BuyerYbt
		}
		if v.ReUserId > 0 {
			minvite[v.ReUserId] += v.ReYbt
		}
		// 去掉商家是平台的
		if v.SellerUserId > 0 {
			// mseller[v.SellerUserId] += v.SellerYbt		注意 商家奖励全部先放在保证金帐号中
			mseller[bzjAccountId] += v.SellerYbt
		}
	}

	return
}

// 分配项目方ybt
func AllocYunbayYbt(amount float64) (vs []common.AssetLock) {
	if amount > 0 {
		now := time.Now().Unix()
		today := time.Now().Format("2006-01-02")
		vs = []common.AssetLock{}
		// 定期冻结项目天数
		yunbayRebat := conf.Config.Rebat.YunbayRebat
		unlockTime := time.Now().Unix() + (yunbayRebat.FixDays * 24 * 3600) // 将冻结天数转成s
		fix_amount := amount * yunbayRebat.Fix
		vs = append(vs, common.AssetLock{UserId: 0, Type: common.CURRENCY_YBT, LockType: common.ASSET_LOCK_FIX, LockAmount: fix_amount, UnlockTime: unlockTime, Date: today, CreateTime: now, UpdateTime: now})

		// 分配固定期限的ybt
		forever_amount := amount * yunbayRebat.Forever
		vs = append(vs, common.AssetLock{UserId: 0, Type: common.CURRENCY_YBT, LockType: common.ASSET_LOCK_FOREVER, LockAmount: forever_amount, Date: today, CreateTime: now, UpdateTime: now})
	}
	return
}

// 获取空投冻结的用户资产信息
func ListUserFreezeYbtAsset(user_ids []int64) (m map[int64]common.UserAsset, err error) {
	m = make(map[int64]common.UserAsset)

	if len(user_ids) == 0 {
		return
	}
	vs := []common.UserAsset{}
	db := db.GetDB()
	if err = db.Model(&common.UserAsset{}).Where("freeze_ybt>0").Find(&vs, "user_id in(?)", user_ids).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("ListUserFanliAsset failed! err=", err)
		return
	}
	for _, v := range vs {
		m[v.UserId] = v
	}

	return
}

// 获取消费解冻记录
func GetUnLockUserAsset(munlocks map[int64]float64) (mair_unlock map[int64]float64, err error) {
	user_ids := []int64{}
	for k, _ := range munlocks {
		user_ids = append(user_ids, k)
	}
	ms, err := ListUserFreezeYbtAsset(user_ids)

	mair_unlock = make(map[int64]float64)
	var m common.UserAsset
	var ok bool = false
	for k, v := range munlocks {
		if k > 0 {
			if m, ok = ms[k]; !ok {
				continue
			}
			if m.FreezeYbt > 0 {
				lock_amount := math.Abs(v)
				if lock_amount > math.Abs(m.FreezeYbt) { // 解冻的金额不能大于总可解冻金额
					lock_amount = math.Abs(m.FreezeYbt)
				}
				mair_unlock[k] = lock_amount
			}
		}
	}

	return
}

func MakeUserUnlockTypes(mconsume, minvite, mseller, mair_drop, mair_unlock, mactivity, mprojects, mrebat, msellerebat map[int64]float64) (ret map[int64]YbtUnlockType) {
	ret = make(map[int64]YbtUnlockType)
	for k, v := range mconsume {
		if f, ok := ret[k]; ok {
			f.Consume += v
			ret[k] = f
		} else {
			f := YbtUnlockType{Consume: v}
			ret[k] = f
		}
	}
	for k, v := range mseller {
		if f, ok := ret[k]; ok {
			f.Seller += v
			ret[k] = f
		} else {
			f := YbtUnlockType{Seller: v}
			ret[k] = f
		}
	}
	for k, v := range minvite {
		if f, ok := ret[k]; ok {
			f.Invite += v
			ret[k] = f
		} else {
			f := YbtUnlockType{Invite: v}
			ret[k] = f
		}
	}
	for k, v := range mair_drop {
		if f, ok := ret[k]; ok {
			f.AirDrop += v
			ret[k] = f
		} else {
			f := YbtUnlockType{AirDrop: v}
			ret[k] = f
		}
	}
	for k, v := range mair_unlock {
		if f, ok := ret[k]; ok {
			f.AirUnlock += v
			ret[k] = f
		} else {
			f := YbtUnlockType{AirUnlock: v}
			ret[k] = f
		}
	}
	for k, v := range mactivity {
		if f, ok := ret[k]; ok {
			f.Activity += v
			ret[k] = f
		} else {
			f := YbtUnlockType{Activity: v}
			ret[k] = f
		}
	}
	for k, v := range mprojects {
		if f, ok := ret[k]; ok {
			f.Project += v
			ret[k] = f
		} else {
			f := YbtUnlockType{Project: v}
			ret[k] = f
		}
	}
	for k, v := range mrebat {
		if f, ok := ret[k]; ok {
			f.Rebat += v
			ret[k] = f
		} else {
			f := YbtUnlockType{Rebat: v}
			ret[k] = f
		}
	}
	for k, v := range msellerebat {
		if f, ok := ret[k]; ok {
			f.SellerRebat += v
			ret[k] = f
		} else {
			f := YbtUnlockType{SellerRebat: v}
			ret[k] = f
		}
	}
	return
}

// 获取当日ybt释放记录
func GetYbtUnlockDetail(db *db.PsqlDB, date string) (ms map[int64]common.YbtUnlockDetail, err error) {
	vs := []common.YbtUnlockDetail{}
	if err = db.Find(&vs, "date=?", date).Error; err != nil {
		glog.Error("GetUserActivityReward fail! err=", err)
		return
	}
	ms = make(map[int64]common.YbtUnlockDetail)
	for _, v := range vs {
		ms[v.UserId] = v
	}
	return
}

func MakeAndUpdateYbtUnlockDetail(ms map[int64]YbtUnlockType, day_profit, total_unlock float64, myud map[int64]common.YbtUnlockDetail) (vs []common.YbtUnlockDetail, err error) {
	vs = []common.YbtUnlockDetail{}
	for k, v := range ms {
		var u common.YbtUnlockDetail
		if _, ok := myud[k]; ok {
			u = myud[k]
			if u.CheckStatus == common.STATUS_OK {
				err = fmt.Errorf("ybtunlock status is common.STATUS_OK? date=%v id=%v", u.Date, u.Id)
				glog.Error(err.Error())
				return
			}
		} else {
			yester_day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			now := time.Now().Unix()
			u = common.YbtUnlockDetail{UserId: k, Date: yester_day, CreateTime: now, UpdateTime: now}
		}

		u.Consume = v.Consume
		u.Sale = v.Seller
		u.Invite = v.Invite
		u.AirUnlock = v.AirUnlock
		u.Rebat = v.Rebat
		u.AirDrop = v.AirDrop
		u.Activity = v.Activity
		u.SaleRebat = v.SellerRebat
		u.Project = v.Project
		u.Mining = v.Consume + v.Seller + v.Invite
		u.TotalUnlock = u.Mining + u.Activity + u.AirUnlock + u.Project
		// if k == 0 {	// 只有项目用户才有项目释放
		// 	u.Project = project
		// 	u.TotalUnlock += u.Project
		// }
		if total_unlock > 0 {
			u.YbtPercent = float32(u.TotalUnlock / total_unlock)
		}
		if day_profit > 0 {
			u.RebatPercent = float32(u.Rebat / day_profit)
			u.SalePercent = float32(u.SaleRebat / day_profit)
		}

		vs = append(vs, u)
	}
	return
}

// 计算项目ybt奖励
func getprojectybt(amount float64) (mp map[int64]float64) {
	// today := time.Now().Format("2006-01-02")
	// now := time.Now().Unix()
	mp = make(map[int64]float64)
	if amount <= 0 {
		return
	}
	allot := conf.Config.ProjectYbtAllot
	for _, v := range allot {
		if v.Percent <= 0 {
			continue
		}
		m := amount * v.Percent
		if 0 == len(v.Users) { // user_id和users同时存在的话 优先用users
			v.Users = append(v.Users, conf.UserAllot{UserId: v.UserId, Percent: 1.0})
		}

		for _, u := range v.Users {
			mt := m * u.Percent
			if base.IsEqual(u.Percent, 1.0) {
				mt = m
			} else if base.IsEqual(u.Percent, 0) || u.Percent < 0 {
				continue
			}
			mp[u.UserId] = mt
			// vs = append(vs, common.UserAssetDetail{UserId:u.UserId, Type:common.CURRENCY_YBT, TransactionType:common.YBT_TRANSACTION_PROJECT, Amount:mt, Date:today, CreateTime:now, UpdateTime:now})
			// if v.Forever > 0 {	// 永久冻结
			// 	as = append(as, common.AssetLock{UserId:u.UserId, Type:common.CURRENCY_YBT, LockType:common.ASSET_LOCK_FOREVER, LockAmount:mt*v.Forever, Date:today, CreateTime:now, UpdateTime:now})
			// }
			// if v.Fix > 0 {		// 固定期限冻结
			// 	unlockTime := time.Now().Unix()+ (v.FixDays*24*3600)	// 将冻结天数转成s
			// 	as = append(as, common.AssetLock{UserId:u.UserId, Type:common.CURRENCY_YBT, LockType:common.ASSET_LOCK_FIX, LockAmount:mt*v.Fix, UnlockTime:unlockTime, Date:today, CreateTime:now, UpdateTime:now})
			// }
		}
	}
	return
}
