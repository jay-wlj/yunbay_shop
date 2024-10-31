package task

import (
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"math"
	"yunbay/ybasset/common"
	"yunbay/ybcron/conf"
	"yunbay/ybcron/util"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func init() {

}

// 调用区块接口进行提币操作
func DayCheck() {
	var err error

	db := db.GetDB()
	err = day_asset_check_yunex(db.DB)
	if err != nil {
		s := fmt.Sprintf("DayCheck fail! err=%v", err.Error())
		util.SendDingTextTalk(s, []string{"15818717950"})
		return
	}

	err = day_asset_check_hotcoin(db.DB)
	if err != nil {
		s := fmt.Sprintf("DayCheck fail! err=%v", err.Error())
		util.SendDingTextTalk(s, []string{"15818717950"})
		return
	}

}

type usAmount struct {
	Type   int     `json:"type"`
	Amount float64 `json:"amount"`
}

// yunex互转平台资产对帐
func day_asset_check_yunex(db *gorm.DB) (err error) {

	// 获取yunbay在yunex注册的帐号余额信息
	mAsset, er1 := util.GetYunExBalance()
	if er1 != nil {
		err = er1
		glog.Error("day_asset_check fail! GetYunExBalance err=", err)
		return
	}
	// 获取yunbay提出到yunex的帐号余额
	var us []usAmount
	if err = db.Model(&common.UserAssetDetail{}).Where("user_id=? and transaction_type=?", conf.Config.ThirdAccount["yunex"].WithDrawId, common.YBT_TRANSACTION_RECHARGE).Group("type").Select("type, sum(amount) as amount").Scan(&us).Error; err != nil {
		glog.Error("day_asset_check fail! err=", err)
		return
	}

	// 获取yunbay在yunex中的总资产
	total_amount := make(map[string]float64)

	for i := common.CURRENCY_YBT; i < common.CURRENCY_UNKNOW; i++ {
		txType := common.GetCurrencyName(i)
		if total_str, ok := conf.Config.ThirdAccount["yunex"].Ext[txType]; ok {
			if total_amount[txType], err = base.StringToFloat64(total_str); err != nil {
				glog.Error("day_asset_check fail! err=", err)
				return
			}
		}
	}

	// 效验kt和ybt的资产是否OK
	for _, v := range us {
		txType := common.GetCurrencyName(v.Type)
		if !base.IsEqual(v.Amount+mAsset[txType], total_amount[txType]) {
			err = fmt.Errorf("yunex account %v not equal! total_amount=%v yunex balance=%v withdraw account amount=%v", txType, total_amount[txType], mAsset[txType], v.Amount)
			return
		}
	}
	return
}

// hotcoin互转平台资产对帐
func day_asset_check_hotcoin(db *gorm.DB) (err error) {

	// 获取yunbay在yunex注册的帐号余额信息
	hotcoin_kt, er1 := util.HotCoinBalance()
	if er1 != nil {
		err = er1
		glog.Error("day_asset_check fail! GetYunExBalance err=", err)
		return
	}
	// 获取yunbay提出到hotcoin的帐号余额
	var pick amountSt
	if err = db.Model(&common.UserAssetDetail{}).Where("user_id=? and transaction_type=? and type=1", conf.Config.ThirdAccount["hotcoin"].UserId, common.YBT_TRANSACTION_PICKUP).Select("sum(amount) as amount").Scan(&pick).Error; err != nil {
		glog.Error("day_asset_check fail! err=", err)
		return
	}

	// 获取yunbay提到到hotcoin的金额
	var with amountSt
	if err = db.Model(&common.UserAssetDetail{}).Where("user_id=? and transaction_type=? and type=1", conf.Config.ThirdAccount["hotcoin"].WithDrawId, common.YBT_TRANSACTION_RECHARGE).Select("sum(amount) as amount").Scan(&with).Error; err != nil {
		glog.Error("day_asset_check fail! err=", err)
		return
	}

	// 在hotcoin中剩余资产计算
	var total_kt float64 = 0
	if total_kt, err = base.StringToFloat64(conf.Config.ThirdAccount["hotcoin"].Ext["kt"]); err != nil {
		glog.Error("day_asset_check fail! err=", err)
		return
	}
	if !base.IsEqual(with.Amount-math.Abs(pick.Amount), total_kt-hotcoin_kt) {
		err = fmt.Errorf("hotcoin account kt not equal! yunbya diff=%v hotcoin diff=%v", with.Amount-math.Abs(pick.Amount), total_kt-hotcoin_kt)
		return
	}

	return
}
