package dao

import (
	"fmt"
	"time"
	"yunbay/ybgoods/common"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
)

const (
	PREFIX_RECOMMEND_INDEX string = "recommend_index"
)

func init() {
	cache.MakeHCacheQuery(&IndexRecommendListCacheQuery)
}

var IndexRecommendListCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int, int) ([]common.RecommendIndex, error), int, int) ([]common.RecommendIndex, error, string)

// 刷新商品缓存
func RefreshIndexRecommend(v common.RecommendIndex) (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("RefreshIndexRecommend fail! err=", err)
		return
	}

	fileds := []string{fmt.Sprintf("%v-%v", v.Country, v.Type), fmt.Sprintf("%v-%v", v.Country, -1)}
	_, err = ch.HDel(PREFIX_RECOMMEND_INDEX, fileds...)

	// 更新推荐商品id缓存
	reset_recommendids(fmt.Sprintf("goods_%v_%v", v.Country, v.Type), v.ProductIds)

	return
}

func get_index_recommend_list(country, _type int) (vs []common.RecommendIndex, err error) {
	vs = []common.RecommendIndex{}
	db := db.GetDB()
	if _type > -1 {
		db.DB = db.Where("type=?", _type)
	} else {
		db.DB = db.Where("type in(?)", []int{0, 1})
	}
	if err = db.Order("type asc").Select("type, name, img, descimg").Find(&vs, "country=?", country).Error; err != nil {
		glog.Error("get_index_recommend_list fail! err=", err)
	}

	return
}

// 获取商户服务列表信息
func GetIndexRecommendList(country int, _type int, page, page_size int) (vs []common.RecommendIndex, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}

	vs, err, _ = IndexRecommendListCacheQuery(ch, PREFIX_RECOMMEND_INDEX, fmt.Sprintf("%v-%v", country, _type), EXPIRE_TIME, get_index_recommend_list, country, _type)

	for i, v := range vs {
		key := fmt.Sprintf("goods_%v_%v", country, v.Type)
		r := Recommend{}
		lf, e := r.List(key, page, page_size)
		if e != nil {
			err = e
			glog.Error("GetIndexRecommendList fail! err=", err)
		}
		vs[i].Rowset = lf.List
		vs[i].ListEnded = lf.ListEnded
		// if vs[i].Rowset, err = r.List(key, page, page_size); err != nil {
		// 	glog.Error("GetIndexRecommendList fail! err=", err)
		// }
	}

	return
}
