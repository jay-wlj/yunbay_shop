package dao

import (
	"yunbay/yborder/common"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"time"

	"github.com/jie123108/glog"
)

const (
	PREFIX_BANNER_LIST string = "banner_app"
)

type Banner struct {
	*db.PsqlDB
}

func init() {
	cache.MakeCacheQuery(&BannerListCacheQuery)
}

var BannerListCacheQuery func(
	*cache.RedisCache, string, time.Duration, func(int) (common.Banner, error), int) (common.Banner, error, string)

func get_banner(position int) (v common.Banner, err error) {
	if err := db.GetDB().Where("position=?", position).Find(&v).Error; err != nil {
		glog.Error("get_banner_list fail! err=", err)
	}
	return
}

func (t *Banner) Get(position int) (results common.Banner, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	expiretime := time.Duration(300 * time.Minute)
	var ishit string
	results, err, ishit = BannerListCacheQuery(ch, PREFIX_BANNER_LIST, expiretime, get_banner, position)

	glog.Infof("Banner_List key:%v ishit:%v ", PREFIX_BANNER_LIST, ishit)
	return
}

func (t *Banner) Upsert(args []common.Banner) (err error) {
	if t.PsqlDB == nil {
		t.PsqlDB = db.GetDB()
	}
	// 先删除banner 再添加
	if err = t.Delete(&common.Banner{}).Error; err != nil {
		glog.Error("banner delete err:", err)
		return
	}

	for _, v := range args {
		if err = t.Create(&v).Error; err != nil {
			glog.Error("banner Upsert err:", err)
			return
		}
	}

	// 删除缓存
	if ch, err := cache.GetWriter(common.RedisPub); err == nil {
		ch.Del(PREFIX_BANNER_LIST)
	} else {
		glog.Error("GetDefaultCache fail! err=", err)
	}

	return
}
