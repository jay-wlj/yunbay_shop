package task

import (
	"github.com/jay-wlj/gobaselib/db"
	"time"

	//"fmt"
	"yunbay/ybasset/common"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
	//base "github.com/jay-wlj/gobaselib"
)

// 获取平台累计可分红的ybt
func GetAssetBonusAmountDay(date string, db *gorm.DB) (total_issue, amount float64, err error) {
	v := common.YBAsset{}
	if err = db.Where("date=?", date).Find(&v).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("YBAsset  fail! err", err.Error())
		return
	}
	total_issue = v.TotalIssuedYbt
	amount = v.TotalIssuedYbt - v.TotalDestroyedYbt
	// // 需减去所有的空投锁定的ybt
	// vd := common.YBAssetDetail{}
	// if err = db.Where("date=?", date).Find(&v).Error; err != nil&&err!=gorm.ErrRecordNotFound {
	// 	glog.Error("YBAssetDetail fail! err", err.Error())
	// 	return
	// }
	// amount -= vd.FreezeYbt		// 减去空投的ybt，不享受分红
	// if amount < 0 {
	// 	glog.Error("GetAssetBonusAmountDay amount < 0 amount:",amount)
	// }
	// amount += vd.AirUnlock		// 加上当天空投释放的ybt，享受分红
	// freeze_ybt = vd.FreezeYbt
	return
}

// 获取当日资金池交易额,利润
func GetAssetPoolAmountDay(date string, d *gorm.DB) (amount, profit float64, pool []common.YBAssetPool, err error) {
	if d == nil {
		d = db.GetDB().DB
	}
	pool = []common.YBAssetPool{}
	if err = d.Where("currency_type=? and date=?", common.CURRENCY_KT, date).Find(&pool).Error; err != nil {
		glog.Error("GetAssetPoolAmountDay fail! err=", err)
		return
	}

	for _, v := range pool {
		amount += v.PayAmount
		profit += v.RebatAmount
	}
	return
}

type poolSt struct {
	Country int     `json:"country"`
	Amount  float64 `json:"amount"`
	Profit  float64 `json:"prifit"`
}

// 获取国际国内的当日资金池交易额,利润
func GetAssetPoolAmountDayByCountry(date string, d *gorm.DB) (m map[int]poolSt, err error) {
	if d == nil {
		d = db.GetDB().DB
	}
	vs := []poolSt{}
	if err = d.Model(&common.YBAssetPool{}).Where("currency_type=? and date=?", common.CURRENCY_KT, date).Select("country, sum(pay_amount) as amount, sum(rebat_amount) as profit").Group("country").Scan(&vs).Error; err != nil {
		glog.Error("GetAssetPoolAmountDay fail! err=", err)
		return
	}
	m = make(map[int]poolSt)
	for _, v := range vs {
		m[v.Country] = v
	}
	return
}

type UserBonus struct {
	UserId int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

// 获取所有用户可分红的ybt总额
func GetUserBonusAmountDay(db *gorm.DB, date string) (ms map[int64]float64, err error) {
	ms = make(map[int64]float64)
	//sql := "select * from (select user_id, total_ybt-lock_ybt_fanli as amount, lock_ybt_fanli, date, row_number() over(partition by user_id order by date desc) row_id from user_asset) t where  t.row_id=1"

	rows, err1 := db.Model(&common.KtBonusDetail{}).Where("date=? and check_status=?", date, common.STATUS_INIT).Select("user_id, total_ybt-freeze_ybt as amount").Rows() // (*sql.Rows, error)
	if err1 != nil {
		glog.Error("rebat_kt fail! err:", err)
		err = err1
		return
	}
	defer rows.Close()
	for rows.Next() {
		var v UserBonus
		db.ScanRows(rows, &v)
		ms[v.UserId] = v.Amount
	}
	return
}

type userunlock struct {
	UserId    int64
	Mining    float64
	Airunlock float64
	Project   float64
}

// 获取挖矿及空投释放
func GetUserUnlockYbts(db *gorm.DB, date string) (mining, mair_unlock, mproject map[int64]float64, err error) {
	mining = make(map[int64]float64)
	mair_unlock = make(map[int64]float64)
	mproject = make(map[int64]float64)
	vs := []common.YbtUnlockDetail{}
	if err = db.Find(&vs, "date=? and (mining>0 or air_unlock>0 or project>0)", date).Error; err != nil {
		glog.Error("GetUserUnlockYbts fail! err=", err)
		return
	}
	for _, v := range vs {
		if v.Mining > 0 {
			mining[v.UserId] = v.Mining
		}
		if v.AirUnlock > 0 {
			mair_unlock[v.UserId] = v.AirUnlock
		}
		if v.Project > 0 {
			mproject[v.UserId] = v.Project
		}
	}
	return
}

// // 获取kt分红记录
// func GetKtBonusDetail(db *gorm.DB, date string)(vs []common.KtBonusDetail, err error) {
// 	vs = []common.KtBonusDetail{}
// 	if err := db.Find(&vs, "date=? and status=?", date, common.STATUS_INIT).Error; err != nil {
// 		glog.Error("GetKtBonusDetail fail! err=", err)
// 		return
// 	}
// 	return
// }

// 获取所有用户交易结冻的ybt总额
func GetFreezedYbt(db *gorm.DB) (amount float64, err error) {
	//sql := "select sum(total_ybt) as total_ybt, sum(lock_ybt_fanli) as fanli_amount from (select user_id, total_ybt, lock_ybt_fanli, date, row_number() over(partition by user_id order by date desc) row_id from user_asset) t where t.total_ybt>0 and t.row_id=1"
	sql := "select sum(freeze_ybt) as amount from user_asset"
	var v amountSt
	err = db.Raw(sql).Scan(&v).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetUserFanliAmount fail! err:", err)
		return
	}
	amount = v.Amount
	err = nil
	return
}

// 获取当日空投奖励
func GetAirDrop(db *gorm.DB, date string) (amount float64, err error) {
	var v amountSt
	if err = db.Model(&common.AssetLock{}).Where("type=? and lock_type=? and date=? and lock_amount>0", common.CURRENCY_YBT, common.ASSET_LOCK_AIRDROP, date).Select("sum(lock_amount) as amount").Scan(&v).Error; err != nil {
		glog.Error("GetAirDrop fail! err=", err)
		return
	}
	amount = v.Amount
	return
}

type userAmountSt struct {
	UserId int64
	Amount float64
}

// 获取当日用户空投奖励
func GetUserAirDrop(db *gorm.DB, date string) (ms map[int64]float64, amount float64, err error) {
	var vs []userAmountSt
	if err = db.Model(&common.AssetLock{}).Where("type=? and lock_type=? and date=? and lock_amount>0", common.CURRENCY_YBT, common.ASSET_LOCK_AIRDROP, date).Group("user_id").Select("user_id, sum(lock_amount) as amount").Scan(&vs).Error; err != nil {
		glog.Error("GetAirDrop fail! err=", err)
		return
	}
	ms = make(map[int64]float64)
	for _, v := range vs {
		ms[v.UserId] = v.Amount
		amount += v.Amount
	}
	return
}

// 获取当日ybt活动释放记录
func GetUserActivity(db *gorm.DB, date string) (ms map[int64]float64, amount float64, err error) {
	var vs []userAmountSt
	if err = db.Model(&common.UserAssetDetail{}).Where("type=? and transaction_type=? and date=?", common.CURRENCY_YBT, common.YBT_TRANSACTION_ACTIVITY, date).Group("user_id").Select("user_id, sum(amount) as amount").Scan(&vs).Error; err != nil {
		glog.Error("GetUserActivity fail! err=", err)
		return
	}
	amount = 0
	ms = make(map[int64]float64)
	for _, v := range vs {
		ms[v.UserId] = v.Amount
		amount += v.Amount
	}
	return
}

// 获取最新的周期数
func GetLastPeriod() (period int64) {
	var v common.YBAssetDetail
	if err := db.GetDB().Order("period desc, date desc").Limit(1).Find(&v).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetLastPeriod fail! err=", err)
		_, period = getDifficultByTime(time.Now())
		err = nil
		return
	}
	period = v.Period
	// if period == 0 {
	// 	period = 1
	// }
	return
}

// 获取当前日期的周期数
func GetPeriodByDate(date string) (period int64) {
	var v common.YBAssetDetail
	if err := db.GetDB().Find(&v, "date=?", date).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetLastPeriod fail! err=", err)
		_, period = getDifficultByTime(time.Now())
		err = nil
		return
	}
	period = v.Period
	if period == 0 {
		period = 1
	}
	return
}

// 获取某日平台资产明细
func GetYBAssetDetail(db *gorm.DB, date string) (v common.YBAssetDetail, err error) {
	if err = db.Find(&v, "date=?", date).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetYBAssetDetail fial! err=", err)
		return
	}
	err = nil
	return
}

// 获取累积已发行的ybt等
func GetLastYBAsset() (v common.YBAsset, err error) {
	if err = db.GetDB().Order("date desc").Limit(1).Find(&v).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetLastYBAsset fail! err=", err)
		return
	}
	err = nil
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
