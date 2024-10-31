package dao

import (
	//"strconv"
	"strconv"
	"time"
	"yunbay/ybgoods/common"

	"fmt"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jie123108/glog"
)

const (
	PREFIX_CATEGORY_LIST string = "category_list"
	PREFIX_CATEGORY_ID   string = "category_id:"
)

func init() {
	cache.MakeHCacheQuery(&CategoryListCacheQuery)
}

var CategoryListCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int64, bool) ([]interface{}, error), int64, bool) ([]interface{}, error, string)

// 刷新商品缓存
func RefreshCategory(parent_id int64) (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	_, err = ch.HDel(PREFIX_CATEGORY_LIST, fmt.Sprintf("%v-%v", parent_id, true), fmt.Sprintf("%v-%v", parent_id, false))

	return
}

func get_category_list(parent_id int64, child bool) (rs []interface{}, err error) {
	vs := []common.ProductCategory{}
	if err = db.GetDB().Order("sort desc, id asc").Find(&vs, "parent_id=?", parent_id).Error; err != nil {
		glog.Error("get_category_list fail! err=", err)
		return
	}

	// 获取二级分类商品
	if child {
		for i, v := range vs {
			cs := []common.ProductCategory{}
			if v.ParentId == 0 {
				if err = db.GetDB().Order("sort desc, id asc").Find(&cs, "parent_id=?", v.Id).Error; err != nil {
					glog.Error("get_category_list fail! err=", err)
					return
				}
			}
			vs[i].Children = cs
		}
	}

	rs = base.FilterStruct(vs, true, "title", "id","picture", "children").([]interface{})
	return
}

// 获取商户服务列表信息
func GetCategoryList(parent_id int64, child bool) (vs []interface{}, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	vs, err, _ = CategoryListCacheQuery(ch, PREFIX_CATEGORY_LIST, fmt.Sprintf("%v-%v", parent_id, child), EXPIRE_TIME, get_category_list, parent_id, child)

	return
}

// 重新加载分类缓存
func ReloadCategoryId() (err error) {
	db := db.GetDB()
	vs := []common.ProductCategory{}
	if err = db.Select("id, parent_id").Find(&vs).Error; err != nil {
		glog.Error("GoodsRedisReset fail! err=", err)
		return
	}

	mids := make(map[int64][]int64)
	for _, v := range vs {
		mids[v.ParentId] = append(mids[v.ParentId], v.Id)
	}

	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	for k, v := range mids {
		key := PREFIX_CATEGORY_ID + strconv.FormatInt(k, 10)
		ch.Del(key)
		ch.SAddInt64(key, v)
	}
	return
}

func AddCategoryId(parent_id, id int64) (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}
	key := PREFIX_CATEGORY_ID + strconv.FormatInt(parent_id, 10)
	_, err = ch.SAdd(key, id).Result()
	if err == nil {
		RefreshCategory(parent_id)
	}
	return
}

func RemoveCategoryId(id, parent_id int64) (err error) {
	// var v common.ProductCategory
	// db := db.GetDB()
	// if err = db.Select("parent_id").Find(&v, "id=?", id).Error; err != nil {
	// 	if err != gorm.ErrRecordNotFound {
	// 		glog.Error("RemoveCategoryId fail! err=", err)
	// 	}
	// 	return
	// }
	// parent_id := v.ParentId

	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	key := PREFIX_CATEGORY_ID + strconv.FormatInt(parent_id, 10)
	_, err = ch.SRem(key, id).Result()
	if err == nil {
		RefreshCategory(parent_id)
	}
	return
}

// 获取商品二级分类id
func GetGoodsCategoryIds(id int64) (ids []int64, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}
	key := PREFIX_CATEGORY_ID + strconv.FormatInt(id, 10)
	if !ch.Exists(key).Val() {
		// 读取子分类
		vs := []common.ProductCategory{}
		if err = db.GetDB().Find(&vs, "parent_id=?", id).Error; err != nil {
			glog.Error("GetCategoryIds fail! err=", err)
			return
		}
		for _, v := range vs {
			ids = append(ids, v.Id)
		}
		ch.SAddInt64(key, ids)
	}
	if ids, err = ch.SMembersInt64(key); err != nil {
		glog.Error("GetCategoryIds fail! err=", err)
		return
	}
	// 获取二级分类
	second_ids := []int64{}
	if 0 == id {
		for _, id := range ids {
			vs := []int64{}
			if vs, err = GetGoodsCategoryIds(id); err != nil {
				glog.Error("GetCategoryIds fail! err=", err)
				return
			}
			second_ids = append(second_ids, vs...)
		}
		ids = second_ids
	}

	return
}

func GetCategoryGoods(category_id int64, page, page_size int) (v ListInfo, err error) {
	ids, e := GetGoodsCategoryIds(category_id)
	if err = e; err != nil {
		glog.Error("GetCategoryGoods fail! err=", err)
		return
	}

	return GetHighGoodsList(common.PUBLISH_AREA_KT, page, page_size, ids)
}
