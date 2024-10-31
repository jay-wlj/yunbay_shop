package dao

import (
	"yunbay/ybapi/common"
	"fmt"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"time"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

const (
	PREFIX_CART_LIST string        = "api_cart:"
	EXPIRE_TIME      time.Duration = time.Duration(240 * time.Hour)
)

func init() {
	cache.MakeHCacheQuery(&CartListCacheQuery)
}

type Cart struct {
	*db.PsqlDB
}

type CartListData struct {
	List      []common.Cart `json:"list"`
	ListEnded bool          `json:"list_ended"`
	Total     int           `json:"total,omitempty"`
}

var CartListCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int64, int, int, int) (CartListData, error), int64, int, int, int) (CartListData, error, string)

func getUserCartList(user_id int64, publish_area, page, page_size int) (ret CartListData, err error) {
	ret.List = []common.Cart{}
	db := db.GetDB().ListPage(page, page_size).Where("user_id=? and publish_area=?", user_id, publish_area)

	if err = db.Order("create_time desc").Find(&ret.List).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("getUserCartList fail! err=", err)
		return
	}

	list_ended := true
	if page_size == len(ret.List) {
		list_ended = false
	}
	ret.ListEnded = list_ended
	return
}

func (t *Cart) List(user_id int64, publish_area, page, page_size int) (v CartListData, err error) {
	ch, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("Cart_List fail! err=", err)
		return
	}
	//cache.MakeHCacheQuery(&CartListCacheQuery)

	cacheKey := PREFIX_CART_LIST + fmt.Sprintf("%v", user_id)
	field := fmt.Sprintf("%v:%v-%v", publish_area, page, page_size)
	var ishit string
	v, err, ishit = CartListCacheQuery(ch, cacheKey, field, EXPIRE_TIME, getUserCartList, user_id, publish_area, page, page_size)

	glog.Infof("Cart_List key:%v ishit:%v ", cacheKey, ishit)
	return
}

func (t *Cart) RefreshCache(user_id int64) (err error) {
	cacheKey := PREFIX_CART_LIST + fmt.Sprintf("%v", user_id)
	// 删除缓存
	if ch, err := cache.GetWriter(common.RedisPub); err == nil {
		ch.Del(cacheKey)
	} else {
		glog.Error("FlreshCache fail! err=", err)
	}
	return
}

func (t *Cart) DelByIds(user_id int64, ids []int64) (err error) {
	if t.PsqlDB != nil {
		if err = t.Delete(common.Cart{}, "user_id=? and id in (?)", user_id, ids).Error; err != nil {
			glog.Error("DelCarts fail! err=", err)
			return
		}
		t.RefreshCache(user_id)
	} else {
		glog.Error("DelCarts fail! db is nil")
	}
	return
}
