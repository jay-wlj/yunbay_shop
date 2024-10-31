package dao

import (
	//"github.com/dgrijalva/jwt-go/request"
	"yunbay/ybapi/common"
	"encoding/json"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"time"

	"github.com/jie123108/glog"
)

const (
	//EXPIRE_TIME time.Duration = 24 *time.Hour
	PREFIX_INDEX_RECOMMEND string = "index_re:"
	PREFIX_RECOMMEND_LIST  string = "index_rel:"
)

func init() {
	cache.MakeCacheQuery(&RecommendListCacheQuery)
}

type ProductRecommend struct {
	*db.PsqlDB
}

type prol struct {
	List  []common.Product
	Total int
}

var RecommendListCacheQuery func(
	*cache.RedisCache, string, time.Duration, func(string, int, int, int) (prol, error), string, int, int, int) (prol, error, string)

type pIds struct {
	Ids []int64 `json:"ids"`
}

func del_cache_key(key string) {
	ch, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		return
	}
	ch.Del(key)
}

func get_recommner_list(sql_where string, _type, page, page_size int) (results prol, err error) {
	results.List = []common.Product{}

	v := common.ProductRecommend{Type: _type}
	if err := db.GetDB().Find(&v).Error; err != nil {
		glog.Error("get_recommner_list fail! err=", err)
	}
	total := len(v.ProductIds)
	offset := (page - 1) * page_size
	if offset >= total {
		return
	}
	end := offset + page_size
	if end >= total {
		end = total
	}
	// TODO
	// ids := v.ProductIds[offset:end] // 获取某页数据
	// if results.List, _, err = Product_List(ids, true); err != nil {
	// 	return
	// }
	results.Total = total
	return
}

func (t *ProductRecommend) List(_type, page, page_size int) (results []common.Product, total int, err error) {
	results = []common.Product{}

	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	expiretime := time.Duration(24 * time.Hour)
	sql_where := ""
	var ishit string
	var l prol
	cache_key := PREFIX_RECOMMEND_LIST + fmt.Sprintf("%v-%v-%v", _type, page, page_size)
	l, err, ishit = RecommendListCacheQuery(ch, cache_key, expiretime, get_recommner_list, sql_where, _type, page, page_size)
	results = l.List
	total = l.Total
	glog.Infof("Recommend_List key:%v ishit:%v ", cache_key, ishit)
	return
}

func get_index() (vs []common.ProductRecommend, err error, cached string) {
	var ch *cache.RedisCache
	ch, err = cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err:", err)
		return
	}
	cached = "miss"
	vs = []common.ProductRecommend{}
	var str string
	str, err = ch.Get(PREFIX_INDEX_RECOMMEND)
	if str != "" {
		err = json.Unmarshal([]byte(str), &vs)
		if err == nil {
			cached = "hit"
			return
		}
	}

	// 获取所有类型信息
	if err = db.GetDB().Find(&vs).Error; err != nil {
		return
	}
	// 获取不同类型首页展示量10条商品
	for _, v := range vs {
		count := len(v.ProductIds)
		if count > 10 {
			count = 10
		}
		// TODO
		// ids := v.ProductIds[:count]
		// if vs[i].Products, _, err = Product_List(ids, true); err != nil {
		// 	return
		// }
		//fmt.Println(v.Products)
	}

	body, _ := json.Marshal(vs)
	ch.Set(PREFIX_INDEX_RECOMMEND, string(body), EXPIRE_TIME)
	return
}

// 首页推荐
func (t *ProductRecommend) IndexRecommend() (vs []common.ProductRecommend, err error) {
	vs, err, _ = get_index()
	return
}

func (t *ProductRecommend) Upsert(v common.ProductRecommend) (err error) {
	if t.PsqlDB == nil {
		t.PsqlDB = db.GetDB()
	}
	v.ProductIds = base.UniqueInt64Slice(v.ProductIds) // 去重
	// 新推荐的商品id会一直添加在前面
	//t.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (type) DO UPDATE SET name=%v,img=%v,descimg=%v,product_ids=anyarray_uniq(array_cat(array[%v]::bigint[],product_ids::bigint[])),update_time=%v ",
	t.DB = t.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (type) DO UPDATE SET name='%v',img='%v',descimg='%v',product_ids=array[%v]::bigint[],update_time=%v ",
		v.Name, v.Img, v.DescImg, base.Int64SliceToString(v.ProductIds, ","), v.UpdateTime))

	if err = t.Save(&v).Error; err != nil {
		glog.Error("ProductRecommend Upsert err:", err)
		return
	}

	// 删除缓存
	del_cache_key(PREFIX_INDEX_RECOMMEND)

	return
}
