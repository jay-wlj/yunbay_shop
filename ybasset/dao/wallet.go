package dao

import (
	"yunbay/ybasset/common"
	"fmt"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"time"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

const (
	PREFIX_USER_WALLET string = "uw_address"
)

func init() {
	cache.MakeHCacheQuery(&QueryUserWalletAddress)
}

var QueryUserWalletAddress func(
	*cache.RedisCache, string, string, time.Duration, func(int64) (string, error), int64) (string, error, string)

func get_user_wallet_address(user_id int64) (address string, err error) {
	var v common.UserWallet

	db := db.GetDB()
	// 获取昨日ybt
	if err = db.First(&v, "user_id=?", user_id).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("get_user_wallet_address failed! err=", err)
		return
	}
	address = v.BindAddress
	return
}

// 获取昨日的kt收益金及ybt奖励
func GetUserWalletAddress(user_id int64) (address string, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisAsset)
	if err != nil {
		glog.Error("cache.GetWriter(common.RedisAsset) fail! err=", err)
		return
	}

	expiretime := time.Duration(30 * 24 * time.Hour)

	cache_key := PREFIX_USER_WALLET
	field := fmt.Sprintf("%v", user_id)
	address, err, _ = QueryUserWalletAddress(ch, cache_key, field, expiretime, get_user_wallet_address, user_id)
	if err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetUserWalletAddress fail! err=", err)
		return
	}

	return
}

// Nmh
func RefleshUserWalletAddress(user_id int64) (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisAsset)
	if err != nil {
		glog.Error("cache.GetWriter(common.RedisAsset) fail! err=", err)
		return
	}
	cache_key := PREFIX_USER_WALLET
	field := fmt.Sprintf("%v", user_id)
	ch.HDel(cache_key, field)
	if err != nil {
		glog.Error("RefleshUserWalletAddress fail! err=", err)
		return
	}

	return
}
