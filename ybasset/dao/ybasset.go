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
	PREFIX_YBASSET_DETAIL string = "ybd"
)

func init() {
	cache.MakeHCacheQuery(&YBAssetDetailCacheQuery)
}

type AssetDetailSt struct {
	common.YBAssetDetail
	TotalIssuedYbt float64 `json:"total_issue_ybt" gorm:"column:total_issue_ybt"`
}

var YBAssetDetailCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int, int) ([]AssetDetailSt, error), int, int) ([]AssetDetailSt, error, string)

func get_ybasset_detail(page, page_size int) (vs []AssetDetailSt, err error) {
	today := time.Now().Format("2006-01-02")
	vs = []AssetDetailSt{}

	if err = db.GetDB().ListPage(page, page_size).Where("yunbay_asset_detail.date <> ?", today).Model(&common.YBAssetDetail{}).Joins("JOIN yunbay_asset ON yunbay_asset.date = yunbay_asset_detail.date").Order("yunbay_asset_detail.date desc").Select("yunbay_asset_detail.*, yunbay_asset.*").Scan(&vs).Error; err != nil {
		glog.Error("get_ybasset_detail fail! err=", err)
		return
	}

	return
}

// 获取昨日的kt收益金及ybt奖励
func ListYBAssetDetail(page, page_size int) (ret []AssetDetailSt, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisAsset)
	if err != nil {
		glog.Error("cache.GetWriter(common.RedisAsset) fail! err=", err)
		return
	}

	expiretime := time.Duration(24 * time.Hour)

	cache_key := PREFIX_YBASSET_DETAIL
	filed := fmt.Sprintf("%v-%v", page, page_size)
	ret, err, _ = YBAssetDetailCacheQuery(ch, cache_key, filed, expiretime, get_ybasset_detail, page, page_size)
	if err != nil {
		glog.Error("ListYBAssetDetail fail! err=", err)
		return
	}

	return
}

func RefrenshYBAssetDetailCache() (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisAsset)
	if err != nil {
		glog.Error("cache.GetWriter(common.RedisAsset) fail! err=", err)
		return
	}
	cache_key := PREFIX_YBASSET_DETAIL
	_, err = ch.Del(cache_key)
	return
}
