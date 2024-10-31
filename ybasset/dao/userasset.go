package dao

import (
	"yunbay/ybasset/common"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"time"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

const (
	PREFIX_USER_BONUS string = "ub:"
)

func init() {
	cache.MakeHCacheQuery(&UserBonusCacheQuery)
}

type BonusInfo struct {
	Kt            float64 `json:"kt_bonus"`
	Ybt           float64 `json:"ybt_bonus"`
	Status        bool    `json:"status"`
	TotalYbtBonus float64 `json:"total_ybt_bonus"`
	TotalKtBonus  float64 `json:"total_kt_bonus"`
	Date          string  `json:"date"`
}

var UserBonusCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(string, int64) (BonusInfo, error), string, int64) (BonusInfo, error, string)

func UserAssetIsLocked(user_id int64) (lock bool, err error) {
	var userlock common.UserAsset
	if err = db.GetDB().Last(&userlock, "user_id=?", user_id).Error; err != nil && err != gorm.ErrRecordNotFound {
	}
	err = nil
	lock = (userlock.Status != 0)
	return
}

func GetUserAsset(user_id int64, d *gorm.DB) (v common.UserAsset, err error) {
	v = common.UserAsset{UserId: user_id}
	if d == nil {
		d = db.GetDB().DB
	}
	if err = d.Last(&v, "user_id=?", user_id).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetUserAsset fail! err=", err)
		return
	}
	err = nil
	return
}

func ListUserAsset(user_ids []int64) (m map[int64]common.UserAsset, err error) {
	m = make(map[int64]common.UserAsset)

	if len(user_ids) == 0 {
		return
	}
	db := db.GetDB()
	sql := fmt.Sprintf("select * from where user_id in(%v)", base.Int64SliceToString(user_ids, ","))
	rows, err1 := db.Raw(sql).Rows() // (*sql.Rows, error)
	if err1 != nil {
		glog.Error("rebat_kt fail! err:", err)
		err = err1
		return
	}
	defer rows.Close()
	for rows.Next() {
		var v common.UserAsset
		if err = db.ScanRows(rows, &v); err != nil {
			return
		}
		m[v.UserId] = v
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

type bonusAmount struct {
	Amount float64
}

func get_user_bonus(date string, user_id int64) (results BonusInfo, err error) {
	var ybt common.BonusYbtDetail
	var kt common.BonusKtDetail
	results.Status = true
	db := db.GetDB()
	// 获取昨日ybt
	if err = db.Find(&ybt, "date=? and user_id=?", date, user_id).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("UserAssetDetail_BonusInfo failed! err=", err)
		return
	}
	if err == gorm.ErrRecordNotFound {
		results.Status = false
	}
	// 获取昨日kt
	if err = db.Find(&kt, "date=? and user_id=?", date, user_id).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("UserAssetDetail_BonusInfo failed! err=", err)
		return
	}
	if err == gorm.ErrRecordNotFound {
		results.Status = false
	}
	// 获取累计ybt
	if err = db.Model(&common.BonusYbtDetail{}).Select("sum(total_ybt) as total_ybt_bonus").Where("user_id=?", user_id).Scan(&results).Error; err != nil {
		glog.Error("UserAssetDetail_BonusInfo failed! err=", err)
		return
	}
	// 获取累计kt
	if err = db.Model(&common.BonusKtDetail{}).Select("sum(kt) as total_kt_bonus").Where("user_id=?", user_id).Scan(&results).Error; err != nil {
		glog.Error("UserAssetDetail_BonusInfo failed! err=", err)
		return
	}

	err = nil
	results.Kt = kt.Kt
	results.Ybt = ybt.TotalYbt
	results.Date = date
	return
}

// 获取昨日的kt收益金及ybt奖励
func GetYesterDayBonus(date string, user_id int64) (ret BonusInfo, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisAsset)
	if err != nil {
		glog.Error("cache.GetWriter(common.RedisAsset) fail! err=", err)
		return
	}

	expiretime := time.Duration(24 * time.Hour)

	cache_key := PREFIX_USER_BONUS + date
	field := fmt.Sprintf("%v", user_id)
	ret, err, _ = UserBonusCacheQuery(ch, cache_key, field, expiretime, get_user_bonus, date, user_id)
	if err != nil {
		glog.Error("UserBonusCacheQuery fail! err=", err)
		return
	}

	return
}

func RefrenshUserBonus(date string) (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisAsset)
	if err != nil {
		glog.Error("cache.GetWriter(common.RedisAsset) fail! err=", err)
		return
	}
	cache_key := PREFIX_USER_BONUS + date
	_, err = ch.Del(cache_key)
	return
}
