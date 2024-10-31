package dao

import (
	"fmt"
	"strconv"
	"time"
	"yunbay/ybgoods/common"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

	"yunbay/ybgoods/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

const (
	HIGHLIST_EXPIRE_TIME  time.Duration = 1 * time.Hour
	PREFIX_GOODS_INFO     string        = "goods_info"
	PREFIX_GOODS_LIST     string        = "goods_list:"
	PREFIX_GOODS_SELFLIST string        = "goods_self_list:"
	PREFIX_GOODS_HIGHLIST string        = "goods_high_list:"
)

func init() {
	cache.MakeHCacheQuery(&GoodsListCacheQuery)
	cache.MakeHCacheQuery(&GoodsSelfListCacheQuery)
	cache.MakeHCacheQuery(&GoodsHighListCacheQuery)
	cache.MakeHCacheQuery(&GoodsInfoCacheQuery)
}

var GoodsListCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int, string, int, int) ([]interface{}, error), int, string, int, int) ([]interface{}, error, string)

var GoodsSelfListCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int64, int, int, int) ([]common.Product, error), int64, int, int, int) ([]common.Product, error, string)

var GoodsHighListCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int, int, int, []int64) (ListInfo, error), int, int, int, []int64) (ListInfo, error, string)

var GoodsInfoCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int64) (common.Product, error), int64) (common.Product, error, string)

// 刷新商品缓存
func RefreshGoods(id int64) (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	// 得在删除缓存前获取，否则在删除后获取的缓存会因当前数据还未提交造成商品信息异常
	var v common.Product
	if v, err = GetGoodsInfo(id); err != nil {
		glog.Error("RefreshGoods fail! err=", err)
	}

	_, err = ch.Del(PREFIX_GOODS_INFO, fmt.Sprintf("%v", id))

	// 删除人气最高商品列表缓存
	ch.Del(PREFIX_GOODS_HIGHLIST + strconv.Itoa(v.PublishArea))
	return
}

func get_goods_info(id int64) (v common.Product, err error) {
	err = db.GetDB().Find(&v, "id=?", id).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}

	return
}
func GetGoodsInfo(id int64) (v common.Product, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	v, err, _ = GoodsInfoCacheQuery(ch, PREFIX_GOODS_INFO, strconv.FormatInt(id, 10), 0, get_goods_info, id)

	return
}

func get_self_goods_list(user_id int64, publish_area int, page, page_size int) (rs []common.Product, err error) {
	vs := []common.Product{}
	db := db.GetDB().ListPage(page, page_size).Where("status=? and publish_area=?", common.STATUS_OK, publish_area)
	db = db.Where("user_id=?", user_id)

	if err = db.Order("sold desc, update_time desc, id desc").Select("id,category_id,type,images[1:1],title,rebat,price,stock,sold,publish_area,extinfo,discount").Find(&vs).Error; err != nil {
		glog.Error("get_self_goods_list fail! err=", err)
		return
	}

	return
}

func GetSelfGoodsList(user_id int64, publish_area, page, page_size int) (total int, vs []common.Product, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetSelfGoodsList fail! err=", err)
		return
	}

	failed := fmt.Sprintf("%v-%v", page, page_size)

	key := PREFIX_GOODS_SELFLIST + fmt.Sprintf("%v-%v", user_id, publish_area)
	vs, err, _ = GoodsSelfListCacheQuery(ch, key, failed, EXPIRE_TIME, get_self_goods_list, user_id, publish_area, page, page_size)

	return
}

func get_goods_highlist(publish_area, page, page_size int, category_ids []int64) (v ListInfo, err error) {

	db := db.GetDB()
	db.DB = db.Where("status=? and publish_area=? and is_hid=0", common.STATUS_OK, publish_area)

	if len(category_ids) > 0 {
		db.DB = db.Where("category_id in(?)", category_ids)
	}

	if err = db.Model(&common.Product{}).Count(&v.Total).Error; err != nil {
		glog.Error("get_goods_highlist fail! err=", err)
		return
	}

	if err = db.ListPage(page, page_size).Order("sold desc, update_time desc, id desc").Select("id,category_id,type,images[1:1],title,rebat,price,stock,sold,publish_area,extinfo,discount").Find(&v.List).Error; err != nil {
		glog.Error("get_goods_highlist fail! err=", err)
		return
	}
	v.ListEnded = base.IsListEnded(page, page_size, len(v.List), v.Total)
	return
}

func GetHighGoodsList(publish_area, page, page_size int, category_ids []int64) (v ListInfo, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetHighGoodsList fail! err=", err)
		return
	}

	failed := fmt.Sprintf("%v-%v", page, page_size)
	if len(category_ids) > 0 {
		category_ids = base.UniqueInt64Slice(category_ids)
		str_ids := base.Int64SliceToString(category_ids, ",")
		// 取str_ids的md5前5位
		failed += "-" + util.Sha1hex([]byte(str_ids))[:5]
	}

	key := PREFIX_GOODS_HIGHLIST + strconv.Itoa(publish_area)
	//failed := fmt.Sprintf("%v-%v", page, page_size)
	v, err, _ = GoodsHighListCacheQuery(ch, key, failed, HIGHLIST_EXPIRE_TIME, get_goods_highlist, publish_area, page, page_size, category_ids)

	return
}
