package dao

import (
	"fmt"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"strconv"
	"time"
	"yunbay/ybgoods/common"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

const (
	PREFIX_GOODS_ATTR_KEY        string = "goods_attr_key"
	PREFIX_GOODS_ATTR_KEY_NAME   string = "goods_attr_key_name"
	PREFIX_GOODS_ATTR_VALUE_NAME string = "goods_attr_value_name"
)

func init() {
	cache.MakeHCacheQuery(&GoodsAttrInfo)
}

var GoodsAttrInfo func(
	*cache.RedisCache, string, string, time.Duration, func(int64) (common.ProductAttrKey, error), int64) (common.ProductAttrKey, error, string)

// 刷新商品缓存
func RefreshAttrKey(id int64) (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	_, err = ch.Del(PREFIX_GOODS_ATTR_KEY, strconv.FormatInt(id, 10))

	return
}

func get_attr_info(id int64) (v common.ProductAttrKey, err error) {
	db := db.GetDB()
	err = db.Find(&v, "id=?", id).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	// 获取关联的商品规格值
	err = db.Model(v).Related(&v.Values).Error
	if err != nil {
		glog.Error("GoodsInfo fail! err=", err)
		return
	}

	return
}

// 获取属性key及其所有值
func GetGoodsAttrInfo(id int64) (v common.ProductAttrKey, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	v, err, _ = GoodsAttrInfo(ch, PREFIX_GOODS_ATTR_KEY, strconv.FormatInt(id, 10), 0, get_attr_info, id)

	return
}

// 查询或创建属性key
func SetGetAttrKey(category int64, name string) (id int64, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}
	field := fmt.Sprintf("%v-%v", category, name)
	id, err = ch.HGetI(PREFIX_GOODS_ATTR_KEY_NAME, field)
	if id == 0 {
		var v common.ProductAttrKey
		db := db.GetDB()
		if err = db.Select("id").First(&v, "category_id=? and name=?", category, name).Error; err != nil && err != gorm.ErrRecordNotFound {
			glog.Error("SetGetAttrKey fail! err=", err)
			return
		}
		// 没有此属性key，则创建
		if err == gorm.ErrRecordNotFound {
			v.CategoryId = category
			v.Name = name
			if err = db.Save(&v).Error; err != nil {
				glog.Error("SetGetAttrKey fail! err=", err)
				return
			}
		}
		id = v.Id
		ch.HSet(PREFIX_GOODS_ATTR_KEY_NAME, field, id, 0) // 设置缓存
		ch.HDel(PREFIX_GOODS_ATTR_KEY, strconv.FormatInt(id, 10))
	}
	return
}

// 查询或创建属性值
func SetGetAttrValue(attr_id int64, value string) (id int64, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}
	field := fmt.Sprintf("%v-%v", attr_id, value)
	id, err = ch.HGetI(PREFIX_GOODS_ATTR_VALUE_NAME, field)
	if id == 0 {

		var v common.ProductAttrValue
		db := db.GetDB()
		if err = db.Select("id").First(&v, "product_attr_key_id=? and value=?", attr_id, value).Error; err != nil && err != gorm.ErrRecordNotFound {
			glog.Error("SetGetAttrValue fail! err=", err)
			return
		}
		// 属性key中没有此属性值，则创建
		if err == gorm.ErrRecordNotFound {
			v.ProductAttrKeyId = attr_id
			v.Value = value
			if err = db.Save(&v).Error; err != nil {
				glog.Error("SetGetAttrValue fail! err=", err)
				return
			}
		}
		id = v.Id
		ch.HSet(PREFIX_GOODS_ATTR_VALUE_NAME, field, id, 0)            // 设置缓存
		ch.HDel(PREFIX_GOODS_ATTR_KEY, strconv.FormatInt(attr_id, 10)) // 删除属性key缓存
	}
	return
}
