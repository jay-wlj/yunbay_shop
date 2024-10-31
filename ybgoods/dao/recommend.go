package dao

import (
	//"github.com/dgrijalva/jwt-go/request"
	"fmt"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"time"
	"yunbay/ybgoods/common"

	"github.com/jie123108/glog"
)

const (
	EXPIRE_TIME            time.Duration = 24 * time.Hour
	PREFIX_INDEX_RECOMMEND string        = "index_re"
	PREFIX_RECOMMEND_LIST  string        = "index_list:"
	PREFIX_RECOMMEND       string        = "rel_"
)

type RecommendData interface {
	GetByIds(ids []int64) ([]common.Product, error)
}

var mRecommend map[string]RecommendData

func init() {
	cache.MakeHCacheQuery(&RecommendListCacheQuery)

	mRecommend = make(map[string]RecommendData)

	g := &goods{}
	mRecommend[common.RecommendBestGoods] = g
	mRecommend[common.RecommendNewGoods] = g
	mRecommend["goods_0_0"] = g // 国内版最精商品推荐
	mRecommend["goods_0_1"] = g // 国内版最新商品推荐
	mRecommend["goods_0_2"] = g // 国际版折扣商品推荐
	mRecommend["goods_1_0"] = g // 国际版最精商品推荐
	mRecommend["goods_1_1"] = g // 国际版最新商品推荐
	mRecommend["goods_1_2"] = g // 国际版折扣商品推荐
}

type Recommend struct {
	*db.PsqlDB
}

type ListInfo struct {
	List      []common.Product `json:"list"`
	Total     int              `json:"total"`
	ListEnded bool             `json:"list_ended"`
}

var RecommendListCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(string, int, int) (ListInfo, error), string, int, int) (ListInfo, error, string)

type pIds struct {
	Ids []int64 `json:"ids"`
}

func reset_recommendids(key string, ids []int64) (err error) {
	ch, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		return
	}

	// 推荐id
	relkey := PREFIX_RECOMMEND + key
	_, err = ch.Del(relkey)
	_, err = ch.ZAddI64(relkey, ids)

	// 删除推荐列表
	relist := PREFIX_RECOMMEND_LIST + key
	ch.Del(relist)
	// vals := []redis.Z{}
	// for _, id := range ids {
	// 	vals = append(vals, id)
	// }

	// 更新首页服务列表
	//ch.Del(PREFIX_INDEX_BUSINESS_LIST)
	//_, err = ch.RPush(key, vals...).Result()
	return
}

func get_recommner_list(key string, page, page_size int) (ret ListInfo, err error) {
	ret.List = []common.Product{}

	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("get_recommner_list fail! err=", err)
		return
	}

	// v := common.Recommend{}
	// if err := db.GetDB().Find(&v, "key=?", key).Error; err != nil {
	// 	glog.Error("get_recommner_list fail! err=", err)
	// }
	// total := len(v.Ids)
	reKey := PREFIX_RECOMMEND + key
	var total int64
	if total, err = ch.ZCard(reKey).Result(); err != nil {
		glog.Error("get_recommner_list fail! LLen err=", err)
		return
	}
	var offset int64
	end := total

	// 需要分页显示
	if page > 1 && page_size > 1 {
		offset = (int64(page) - 1) * int64(page_size)
		if offset > total {
			return
		}
		end = offset + int64(page_size)
		if end > total {
			end = total
		}
	}

	if offset > end {
		return
	}

	var ids []int64
	if ids, err = ch.ZRangeI64(reKey, offset, end-1); err != nil {
		glog.Error("get_recommner_list fail! LRangeI64 err=", err)
		return
	}

	// ids := v.Ids[offset:end]

	// 根据推荐类型返回不同推荐内容
	if v, ok := mRecommend[key]; ok {
		ret.List, err = v.GetByIds(ids)
	}
	if end == total {
		ret.ListEnded = true
	}
	return
}

func (t *Recommend) List(key string, page, page_size int) (v ListInfo, err error) {
	v.List = []common.Product{}

	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	expiretime := time.Duration(240 * time.Hour)
	var ishit string

	cache_key := PREFIX_RECOMMEND_LIST + key
	v, err, ishit = RecommendListCacheQuery(ch, cache_key, fmt.Sprintf("%v-%v", page, page_size), expiretime, get_recommner_list, key, page, page_size)

	glog.Infof("Recommend_List key:%v ishit:%v ", cache_key, ishit)
	return
}

// 首页推荐
func (t *Recommend) IndexRecommend(key string) (v interface{}, err error) {
	var l ListInfo
	l, err = t.List(key, 0, 0)
	v = l.List
	return
}

// // // 更新推荐
// func (t *Recommend) Upsert(v common.RecommendIndex) (err error) {
// 	if t.PsqlDB == nil {
// 		t.PsqlDB = db.GetDB()
// 	}
// 	v.ProductIds = base.UniqueInt64Slice(v.ProductIds) // 去重

// 	v.CreateTime = time.Now().Unix()
// 	v.UpdateTime = v.CreateTime
// 	t.DB = t.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (key) DO UPDATE SET ids=array[%v]::bigint[],update_time=%v ",
// 		base.Int64SliceToString(v.ProductIds, ","), v.UpdateTime))

// 	if err = t.Save(&v).Error; err != nil {
// 		glog.Error("ProductRecommend Upsert err:", err)
// 		return
// 	}

// 	key := PREFIX_INDEX_RECOMMEND + strconv.Itoa(v.Type)
// 	// 更新缓存
// 	reset_recommendids(key, v.ProductIds)
// 	return
// }

func (t *Recommend) IsReocmmend(reKey string, id int64) (ok bool, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("get_recommner_list fail! err=", err)
		return
	}
	key := PREFIX_RECOMMEND + reKey
	return ch.ZIsMember(key, fmt.Sprintf("%v", id))
}

func (t *Recommend) GetRcommendIds(reKey string) (ids []int64, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("get_recommner_list fail! err=", err)
		return
	}
	key := PREFIX_RECOMMEND + reKey
	ids, err = ch.ZRangeI64(key, 0, -1)
	return
}
