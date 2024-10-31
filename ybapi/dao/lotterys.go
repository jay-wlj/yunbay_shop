package dao

import (
	"fmt"
	"strconv"
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybapi/conf"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

const (
	PREFIX_LOTTERYS_INFO            string = "lotterys_info"
	PREFIX_LOTTERYS_LIST            string = "lotterys_list"
	PREFIX_LOTTERYS_RECORD_LIST     string = "lotterys_record_list"
	PREFIX_LOTTERYS_RECORD_SELFLIST string = "lotterys_record_self_list"
	//EXPIRE_TIME          time.Duration = time.Duration(240 * time.Hour)
)

func init() {
	cache.MakeHCacheQuery(&LotterysInfoCacheQuery)
	cache.MakeHCacheQuery(&LotterysListCacheQuery)
	cache.MakeHCacheQuery(&LotterysRecordCacheQuery)
}

type Lotterys struct {
}

type LotterysData struct {
	List      []common.Lotterys `json:"list"`
	ListEnded bool              `json:"list_ended"`
	Total     int               `json:"total,omitempty"`
}

type LotterysRecordData struct {
	List      []common.LotterysRecord `json:"list"`
	ListEnded bool                    `json:"list_ended"`
	Total     int                     `json:"total,omitempty"`
}

var LotterysInfoCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int64) (common.Lotterys, error), int64) (common.Lotterys, error, string)

var LotterysListCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int, int) (LotterysData, error), int, int) (LotterysData, error, string)

var LotterysRecordCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int64, int64, int, int, int) (LotterysRecordData, error), int64, int64, int, int, int) (LotterysRecordData, error, string)

func getLotterysList(page, page_size int) (ret LotterysData, err error) {
	ret.List = []common.Lotterys{}
	db := db.GetDB().ListPage(page, page_size)

	if err = db.Order("status asc, update_time desc, id desc").Find(&ret.List, "hid=0").Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("getLotterysList fail! err=", err)
		return
	}

	for i, v := range ret.List {
		if v.Status == common.ACTIVITY_STATUS_END {
			ret.List[i].Invalid = v.Sold < v.Stock // 是否失效
		}
	}

	list_ended := true
	if page_size == len(ret.List) {
		list_ended = false
	}
	ret.ListEnded = list_ended

	return
}

func (t *Lotterys) List(page, page_size int) (v LotterysData, err error) {
	ch, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("Cart_List fail! err=", err)
		return
	}
	//cache.MakeHCacheQuery(&CartListCacheQuery)

	cacheKey := PREFIX_LOTTERYS_LIST
	field := fmt.Sprintf("%v-%v", page, page_size)
	var ishit string
	v, err, ishit = LotterysListCacheQuery(ch, cacheKey, field, EXPIRE_TIME, getLotterysList, page, page_size)

	glog.Infof("Cart_List key:%v ishit:%v ", cacheKey, ishit)
	return
}

func getLotterysInfo(lotterys_id int64) (ret common.Lotterys, err error) {
	db := db.GetDB()
	if err = db.First(&ret, "id=?", lotterys_id).Error; err != nil {
		glog.Error("getUserCartList fail! err=", err)
		return
	}
	return
}

func (t *Lotterys) Get(lotterys_id int64) (v common.Lotterys, err error) {
	ch, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("Cart_List fail! err=", err)
		return
	}
	//cache.MakeHCacheQuery(&CartListCacheQuery)
	v, err, _ = LotterysInfoCacheQuery(ch, PREFIX_LOTTERYS_INFO, strconv.FormatInt(lotterys_id, 10), EXPIRE_TIME, getLotterysInfo, lotterys_id)
	return
}

func (t *Lotterys) Refresh(lotterys_id int64) (err error) {
	// 删除缓存
	if ch, err := cache.GetWriter(common.RedisPub); err == nil {
		ch.HDel(PREFIX_LOTTERYS_INFO, strconv.FormatInt(lotterys_id, 10)) // 刷新详情
		ch.Del(PREFIX_LOTTERYS_LIST)                                      // 刷新列表缓存
	} else {
		glog.Error("Lotterys fail! err=", err)
	}

	return
}

func getLotterysRecordList(lotterys_id, user_id int64, status, page, page_size int) (ret LotterysRecordData, err error) {
	ret.List = []common.LotterysRecord{}
	db := db.GetDB().ListPage(page, page_size)

	if lotterys_id > 0 {
		db = db.Where("lotterys_id=?", lotterys_id)
	}
	if user_id > 0 {
		db = db.Where("user_id=?", user_id)
	}
	if status >= 0 {
		db = db.Where("status=?", status)
	}
	if err = db.Order("create_time desc").Find(&ret.List).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("getUserCartList fail! err=", err)
		return
	}

	list_ended := true
	if page_size == len(ret.List) {
		list_ended = false
	}
	ret.ListEnded = list_ended

	// 处理url

	for i := range ret.List {
		ret.List[i].Url = conf.Config.Ext["eos_url"].(string) + ret.List[i].Hash
	}
	// 获取计划标题等
	ch := &Lotterys{}
	if lotterys_id > 0 {
		if v, e := ch.Get(lotterys_id); e == nil {
			p := base.FilterStruct(v, true, "product", "coin", "pirce")
			for i := range ret.List {
				ret.List[i].Lotterys = p
			}
		}
	} else {
		for i := range ret.List {
			if v, e := ch.Get(ret.List[i].LotterysId); e == nil {
				p := base.FilterStruct(v, true, "product", "coin", "price")
				ret.List[i].Lotterys = p
			}
		}
	}

	return
}

func (t *Lotterys) ListRecord(lotterys_id int64, page, page_size int) (v LotterysRecordData, err error) {
	ch, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("Cart_List fail! err=", err)
		return
	}
	//cache.MakeHCacheQuery(&CartListCacheQuery)

	cacheKey := fmt.Sprintf("%v:%v", PREFIX_LOTTERYS_RECORD_LIST, lotterys_id)
	field := fmt.Sprintf("%v-%v", page, page_size)
	var ishit string
	v, err, ishit = LotterysRecordCacheQuery(ch, cacheKey, field, EXPIRE_TIME, getLotterysRecordList, lotterys_id, 0, -1, page, page_size)

	glog.Infof("Cart_List key:%v ishit:%v ", cacheKey, ishit)
	return
}

func (t *Lotterys) ListUserRecord(user_id int64, status, page, page_size int) (v LotterysRecordData, err error) {
	ch, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("Cart_List fail! err=", err)
		return
	}
	//cache.MakeHCacheQuery(&CartListCacheQuery)
	cacheKey := fmt.Sprintf("%v:%v", PREFIX_LOTTERYS_RECORD_SELFLIST, user_id)
	field := fmt.Sprintf("%v-%v-%v", status, page, page_size)
	var ishit string
	v, err, ishit = LotterysRecordCacheQuery(ch, cacheKey, field, EXPIRE_TIME, getLotterysRecordList, 0, user_id, status, page, page_size)

	glog.Infof("Cart_List key:%v ishit:%v ", cacheKey, ishit)
	return
}

func (t *Lotterys) RefreshRecord(lotterys_id, user_id int64) (err error) {
	// 删除缓存
	if ch, err := cache.GetWriter(common.RedisPub); err == nil {
		ch.Del(fmt.Sprintf("%v:%v", PREFIX_LOTTERYS_RECORD_LIST, lotterys_id))
		if user_id > 0 {
			ch.Del(fmt.Sprintf("%v:%v", PREFIX_LOTTERYS_RECORD_SELFLIST, user_id))
		}

		// 刷新列表缓存
	} else {
		glog.Error("Lotterys fail! err=", err)
	}

	return
}

func (t *Lotterys) RefreshUserRecord(user_id int64) (err error) {
	// 删除缓存
	if ch, err := cache.GetWriter(common.RedisPub); err == nil {

		ch.Del(fmt.Sprintf("%v:%v", PREFIX_LOTTERYS_RECORD_SELFLIST, user_id))

		// 刷新列表缓存
	} else {
		glog.Error("Lotterys fail! err=", err)
	}

	return
}
