package task

import (
	"yunbay/ybasset/common"
	"yunbay/ybcron/conf"
	"fmt"
	"github.com/jay-wlj/gobaselib/db"
	"time"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 每小时统计一次平台交易数据
func YBAsset_Statistics() {
	fmt.Println("YBAsset_Statistics begin")
	now := time.Now()
	db := db.GetTxDB(nil)
	if err := statistics(db); err != nil {
		glog.Error("YBAsset_Statistics end fail! err=", err)
		db.Rollback()
		return
	}
	db.Commit()
	fmt.Println("YBAsset_Statistics end success! tick=", time.Since(now).String())
}

type ybtamount struct {
	YBT float64 `json:"ybt"`
}

func statistics(db *db.PsqlDB) (err error) {
	nTime := time.Now()
	today := nTime.Format("2006-01-02")

	// 获取当日资金池交易额,利润
	day_amount, day_profit, _, err := GetAssetPoolAmountDay(today, db.DB)
	if err != nil {
		glog.Error("GetAssetAmountDay fail! err=", err)
		return
	}

	// if err = UpdateAssetLockAmount(today, db.DB); err != nil {
	// 	glog.Error("UpdateAssetBonusAmount fail! err=", err)
	// 	return
	// }
	// 更新昨日所有锁定的ybt
	var freeze_ybt float64 = 0
	var bonus_ybt float64 = 0
	if freeze_ybt, err = GetFreezedYbt(db.DB); err != nil {
		glog.Error("GetAirDrop fail! err", err)
		return
	}

	// 获取累计发放可分红的ybt
	_, bonus_ybt, err = GetAssetBonusAmountDay(today, db.DB)
	if err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetAssetBonusAmountDay  fail! err", err.Error())
		return
	}
	//fmt.Println("yunbay_ybt:", yunbay_ybt)

	//day_issue, err := GetIssueYbtByDay(tomorrow)
	period := GetLastPeriod() + 1
	//difficule := getdifficult(GetLastPeriod()+1)
	// //difficult, period := getDifficultByTime(time.Now())
	// // 获取昨天总挖矿产出
	// var day_issue float64 = 0
	// if difficule > 0 {
	// 	day_issue = day_profit / difficule
	// }

	now := time.Now().Unix()
	v := common.YBAssetDetail{Amount: day_amount, Profit: day_profit, Period: period, BonusYbt: bonus_ybt, FreezeYbt: freeze_ybt, Date: today, CreateTime: now, UpdateTime: now}
	// v.CanBonusYbt = yunbay_ybt + day_issue	// 减去当前交易冻结的 加上当日待发行的ybt
	// if v.CanBonusYbt > 0 {
	// 	v.Perynbay = (1 / v.CanBonusYbt) * v.Profit
	// }

	db_in := db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (date) DO UPDATE SET amount=%v, profit=%v,freeze_ybt=%v,update_time=%v", v.Amount, v.Profit, v.FreezeYbt, v.UpdateTime))
	if err = db_in.Save(&v).Error; err != nil {
		glog.Error("YBAssetDetail save fail! err", err.Error())
		return
	}

	// 更新当日交易流水
	//total_orders, payerd_orders, payers,err1 := GetOrdersCount()
	mOrders, err1 := GetOrdersCountryCount()
	if err1 != nil {
		err = err1
		glog.Error("GetOrdersCount fail! err=", err1)
		return
	}
	mPools, err1 := GetAssetPoolAmountDayByCountry(today, db.DB)
	if err1 != nil {
		glog.Error("GetAssetPoolAmountDayByCountry fail! err=", err1)
		return
	}
	for k, o := range mOrders {
		day_country_amount := float64(0)
		day_country_profilt := float64(0)
		if _, ok := mPools[k]; ok {
			day_country_amount = mPools[k].Amount
			day_country_profilt = mPools[k].Profit
		}

		t := common.TradeFlow{TotalOrders: o.total_orders, PayedOrders: o.pay_orders, TotalPayers: o.payers, TotalAmount: day_country_amount, TotalProfit: day_country_profilt, Perynbay: v.Perynbay, Date: today, Country: k, CreateTime: now, UpdateTime: now}
		db_in = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (date,country) DO UPDATE SET total_orders=%v, payed_orders=%v,total_payers=%v,total_amount=%v, total_profit=%v, perynbay=%v, update_time=%v",
			t.TotalOrders, t.PayedOrders, t.TotalPayers, t.TotalAmount, t.TotalProfit, t.Perynbay, now))
		if err = db_in.Save(&t).Error; err != nil {
			glog.Error("TradeFlow save fail! err", err.Error())
			return
		}
	}

	return
}

type cnt struct {
	Country int   `json:"country"`
	Status  int   `json:"status"`
	Count   int64 `json:"count"`
}

type orderTradeSt struct {
	total_orders int64
	pay_orders   int64
	payers       int64
}

func GetOrdersCountryCount() (m map[int]orderTradeSt, err error) {
	var d *gorm.DB
	d, err = db.InitPsqlDb(conf.Config.PsqlUrl["api"], conf.Config.Debug)
	// 获取订单数及支付人数
	if err != nil {
		glog.Error("InitPsqlDb api fail! err=", err)
		return
	}
	m = make(map[int]orderTradeSt)

	// 获取当日总订单数
	today := time.Now().Format("2006-01-02")
	vs := []cnt{}
	if err = d.Table("orders").Select("country, status, count(*) as count").Group("country, status").Where("date=?", today).Scan(&vs).Error; err != nil {
		glog.Error("get orders count fail! err=", err)
		return
	}
	// 获取当日订单支付人数
	users := []cnt{}
	if err = d.Table("orders").Select("country, count(distinct(user_id)) as count").Where("date=? and status>=?", today, common.ORDER_STATUS_PAYED).Group("country").Scan(&users).Error; err != nil {
		glog.Error("get orders user count fail! err=", err)
		return
	}
	for _, v := range vs {
		r, ok := m[v.Country]
		if !ok {
			r = orderTradeSt{}
		}
		r.total_orders += v.Count
		//total_orders += v.Count
		if v.Status >= common.ORDER_STATUS_PAYED && v.Status != common.ORDER_STATUS_CANCEL {
			//payed_orders += v.Count
			r.pay_orders += v.Count
		}
		m[v.Country] = r
	}
	for _, v := range users {
		r, ok := m[v.Country]
		if !ok {
			r = orderTradeSt{}
		}
		r.payers += v.Count
		m[v.Country] = r
	}
	return
}
