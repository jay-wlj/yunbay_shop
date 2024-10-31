package dao

import (
	"yunbay/ybasset/common"
	"fmt"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"time"

	"github.com/jie123108/glog"
)

const (
	PREFIX_WITHDRAW_TOTAL string = "wd_t"
)

func init() {
	cache.MakeCacheQuery(&TotalWithdrawCacheQuery)
	cache.MakeHCacheQuery(&QueryUserDayWithdraw)
}

type TotalWithDraw struct {
	TxType int     `json:"tx_type"`
	Count  int64   `json:"count"`
	Amount float64 `json:"amount"`
	Fee    float64 `json:"fee"`
}

var TotalWithdrawCacheQuery func(
	*cache.RedisCache, string, time.Duration, func() (TotalWithDraw, error)) (TotalWithDraw, error, string)

func GetTotalWithdraw() (v TotalWithDraw, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisAsset)
	if err != nil {
		glog.Error("cache.GetWriter(common.RedisAsset) fail! err=", err)
		return
	}

	expiretime := time.Duration(24 * time.Hour)

	cache_key := PREFIX_WITHDRAW_TOTAL
	v, err, _ = TotalWithdrawCacheQuery(ch, cache_key, expiretime, get_total)
	if err != nil {
		glog.Error("GetTotalWithdraw fail! err=", err)
		return
	}

	return
}

func get_total() (v TotalWithDraw, err error) {
	db := db.GetDB().Model(&common.WithdrawFlow{})
	if err = db.Count(&v.Count).Error; err != nil {
		glog.Error("get_total fail! err=", err)
		return
	}

	if err = db.Where("status=?", common.TX_STATUS_SUCCESS).Select("sum(amount) as amount, sum(fee) as fee").Scan(&v).Error; err != nil {
		glog.Error("get_total fail! err=", err)
		return
	}
	return
}

func RefleshTotalWidthDraw() (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisAsset)
	if err != nil {
		glog.Error("cache.GetWriter(common.RedisAsset) fail! err=", err)
		return
	}
	ch.Del(PREFIX_WITHDRAW_TOTAL)
	return
}

// 获取某个用户当日已提的币种
var QueryUserDayWithdraw func(
	*cache.RedisCache, string, string, time.Duration, func(int64, int) (float64, error), int64, int) (float64, error, string)

type amountSt struct {
	Amount float64
}

func get_userdaywithdraw(user_id int64, tx_type int) (amount float64, err error) {
	// 查询当天已提币的余额
	today := time.Now().Format("2006-01-02")
	var v amountSt
	db := db.GetDB()
	if err = db.Model(&common.WithdrawFlow{}).Where("user_id=? and date=? and tx_type=? and status not in(?)", user_id, today, tx_type, []int{common.TX_STATUS_NOTPASS, common.TX_STATUS_FAILED}).Select("sum(amount+fee) as amount").Scan(&v).Error; err != nil {
		glog.Error("GetUserWithdarwAvialbe fail! err=", err)
		return
	}
	amount = v.Amount
	return
}

// 获取某个用户当日已提的币种
func GetUserDayWithdraw(user_id int64, txType int) (amount float64, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisAsset)
	if err != nil {
		glog.Error("cache.GetWriter(common.RedisAsset) fail! err=", err)
		return
	}

	expiretime := time.Duration(24 * time.Hour)

	key := fmt.Sprintf("withdraw:%v:", time.Now().Format("2006-01-02"))
	filed := fmt.Sprintf("%v_%v", user_id, common.GetCurrencyName(txType))
	amount, err, _ = QueryUserDayWithdraw(ch, key, filed, expiretime, get_userdaywithdraw, user_id, txType)
	if err != nil {
		glog.Error("GetTotalWithdraw fail! err=", err)
		return
	}
	return
}

// 刷新某个用户的已提币数量
func RefleshUserDayWidthDraw(user_id int64, txType int) (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisAsset)
	if err != nil {
		glog.Error("cache.GetWriter(common.RedisAsset) fail! err=", err)
		return
	}
	key := fmt.Sprintf("withdraw:%v:", time.Now().Format("2006-01-02"))
	filed := fmt.Sprintf("%v_%v", user_id, common.GetCurrencyName(txType))
	ch.HDel(key, filed)
	return
}
