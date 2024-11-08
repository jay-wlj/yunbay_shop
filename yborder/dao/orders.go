package dao

import (
	"encoding/json"
	"fmt"
	"time"
	"yunbay/yborder/common"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"

	//"github.com/jinzhu/gorm"
	base "github.com/jay-wlj/gobaselib"
)

const (
	PREFIX_ORDERS_LIST         string = "order_ul:"
	PREFIX_ORDERS_INFO         string = "order_id"
	PREFIX_ORDERS_PRODUCT_LIST string = "order_pl:"
	//PREFIX_ORDERS_CNT string = "order_ucnt:"
	//EXPIRE_TIME time.Duration = time.Duration(240 * time.Hour)
)

func init() {
	cache.MakeHCacheQuery(&OrdersListCacheQuery)
	cache.MakeHCacheQuery(&OrdersStatusCountCacheQuery)
	cache.MakeHCacheQuery(&OrdersSaleCountCacheQuery)
	cache.MakeHCacheQuery(&OrdersProductListCacheQuery)
}

type Orders struct {
	*db.PsqlDB
}

type OrdersListData struct {
	List      []common.Orders `json:"list"`
	ListEnded bool            `json:"list_ended"`
	Total     int             `json:"total,omitempty"`
}

// 订单列表缓存
var OrdersListCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int64, int, int, int, int) ([]common.Orders, error), int64, int, int, int, int) ([]common.Orders, error, string)

// // 订单计数缓存
// var OrdersCountCacheQuery func(
// 	*cache.RedisCache, string, string, time.Duration, func(int64,int,int) (int, error), int64,int,int) (int, error, string)

// 订单各状态计数缓存
var OrdersStatusCountCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int64, int) ([]OrderStatusCntSt, error), int64, int) ([]OrderStatusCntSt, error, string)

// 订单售后计数缓存
var OrdersSaleCountCacheQuery func(
	*cache.RedisCache, string, string, time.Duration, func(int64, int) (int, error), int64, int) (int, error, string)

// 商品被卖出订单缓存
var OrdersProductListCacheQuery func(*cache.RedisCache, string, string, time.Duration, func(int64, int, int) (OrdersListData, error), int64, int, int) (OrdersListData, error, string)

func getUserOrdersList(user_id int64, iseller, status, page, page_size int) (ret []common.Orders, err error) {
	ret = []common.Orders{}
	db := db.GetDB().ListPage(page, page_size)
	if 1 == iseller { // 卖家
		db = db.Where("shield=0 and seller_userid=?", user_id)
	} else {
		db = db.Where("shield=0 and user_id=?", user_id)
	}

	if status > -1 {
		db = db.Where("status=?", status)
	}

	if 1 == iseller {
		db = db.Order("status asc")
	}
	if err = db.Order("update_time desc, id desc").Find(&ret).Error; err != nil {
		glog.Error("getUserCartList fail! err=", err)
		return
	}

	// if err := share.GetCartsProductInfos(&vs); err != nil {
	// 	glog.Error("getUserCartList fail! err=", err)
	// 	return
	// }

	return
}

// func getUserOrdersCount(user_id int64, iseller, status int)(total int, err error) {
// 	db := db.GetDB().Model(&common.Orders{})
// 	if 1 == iseller {
// 		db = db.Where("shield=0 and seller_userid=?", user_id)
// 	} else {
// 		db = db.Where("user_id=?", user_id)
// 	}

// 	if status > -1 {
// 		db = db.Where("status=?", status)
// 	}

// 	if err = db.Count(&total).Error; err != nil {
// 		glog.Error("getUserCartList fail! err=", err)
// 		return
// 	}

// 	return
// }

func (t *Orders) List(user_id int64, iseller, status, page, page_size int) (v OrdersListData, err error) {
	ch, err := cache.GetWriter(common.RedisApi)
	if err != nil {
		glog.Error("Cart_List fail! err=", err)
		return
	}

	cacheKey := PREFIX_ORDERS_LIST + fmt.Sprintf("%v-%v", user_id, iseller)
	field := fmt.Sprintf("%v:%v-%v", status, page, page_size)
	var ishit string
	expire_time := time.Duration(24 * time.Hour)
	v.List, err, ishit = OrdersListCacheQuery(ch, cacheKey, field, expire_time, getUserOrdersList, user_id, iseller, status, page, page_size)
	if err != nil {
		glog.Error("Orders_List fail! err=", err)
		return
	}
	//cache.MakeHCacheQuery(&GetOrderStatusCount)

	//cacheKey = PREFIX_ORDERS_CNT + fmt.Sprintf("%v-%v", user_id, iseller)
	//field = fmt.Sprintf("status_cnt", status)
	//expire_time =  time.Duration(240 * time.Hour)
	//v.Total, err, ishit = OrdersCountCacheQuery(ch, cacheKey, "status_cnt", expire_time, getUserOrdersCount, user_id, iseller, status)
	cnts, err := t.GetOrderStatusCount(user_id, iseller)
	if err != nil {
		glog.Error("Orders_List fail! GetOrderStatusCount err=", err)
		return
	}
	for _, c := range cnts {
		if c.Status == status {
			v.Total = c.Count
			break
		}
	}

	v.ListEnded = base.IsListEnded(page, page_size, len(v.List), v.Total)

	glog.Infof("Orders_List key:%v ishit:%v ", cacheKey, ishit)
	return
}

func (t *Orders) RefreshCache(user_id int64, iseller int) (err error) {
	cacheKey := PREFIX_ORDERS_LIST + fmt.Sprintf("%v-%v", user_id, iseller)
	// 删除缓存
	if ch, err := cache.GetWriter(common.RedisApi); err == nil {
		ch.Del(cacheKey)
	} else {
		glog.Error("FlreshCache fail! err=", err)
	}
	return
}

func (t *Orders) DelByIds(user_id int64, ids []int64) (err error) {
	if t.PsqlDB != nil {
		if err = t.Delete(common.Orders{}, "user_id=? and id in (?)", user_id, ids).Error; err != nil {
			glog.Error("DelOrders fail! err=", err)
			return
		}
		t.RefreshCache(user_id, 0)
	} else {
		glog.Error("DelOrders fail! db is nil")
	}
	return
}

type orderUserIdSt struct {
	UserId       int64 `json:"user_id"`
	SellerUserId int64 `json:"seller_user_id"`
}

func (t *Orders) Upsert(v *common.Orders) (err error) {
	if err = t.Save(v).Error; err != nil {
		glog.Error("UpsertOrders fail! err=", err)
		return
	}
	t.UpdateUserId(v.Id, v.UserId, v.SellerUserId)
	return
}

func (t *Orders) UpdateUserId(id int64, user_id, seller_user_id int64) {
	ch, err := cache.GetWriter(common.RedisApi)
	if err != nil {
		glog.Error("GetApiCache fail! err=", err)
		return
	}
	m := orderUserIdSt{user_id, seller_user_id}
	key := fmt.Sprintf("%v:%v", PREFIX_ORDERS_INFO, id)
	val, _ := json.Marshal(m)
	if err = ch.HSet(key, "user_id", val, 0); err != nil {
		glog.Error("GetApiCache fail! err=", err)
		return
	}
	return
}

func (t *Orders) GetUserById(id int64) (user_id, seller_userid int64, err error) {
	ch, err := cache.GetWriter(common.RedisApi)
	if err != nil {
		glog.Error("GetApiCache fail! err=", err)
		return
	}
	key := fmt.Sprintf("%v:%v", PREFIX_ORDERS_INFO, id)
	var val []byte
	if val, err = ch.HGetB(key, "user_id"); err == nil {
		var m orderUserIdSt
		if err = json.Unmarshal(val, &m); err == nil {
			user_id = m.UserId
			seller_userid = m.SellerUserId
			return
		}
	}

	// 从库中查找
	var v common.Orders
	if err = db.GetDB().Select("user_id, seller_userid").Find(&v, "id=?", id).Error; err != nil {
		glog.Error("GetUserById fail! err=", err)
		return
	}
	user_id = v.UserId
	seller_userid = v.SellerUserId

	return
}

func (t *Orders) RefreshByOrderIds(ids []int64) (err error) {
	ch, err := cache.GetWriter(common.RedisApi)
	if err != nil {
		glog.Error("GetApiCache fail! err=", err)
		return
	}

	user_ids := []int64{}
	seller_user_ids := []int64{}
	fail_ids := []int64{}
	for _, v := range ids {
		key := fmt.Sprintf("%v:%v", PREFIX_ORDERS_INFO, v)
		var val []byte
		if val, err = ch.HGetB(key, "user_id"); err == nil {
			var m orderUserIdSt
			if err = json.Unmarshal(val, &m); err == nil {
				user_ids = append(user_ids, m.UserId)
				seller_user_ids = append(seller_user_ids, m.SellerUserId)
			}
		} else {
			fail_ids = append(fail_ids, v)
		}
	}
	// 不在缓存中则从数据库中获取订单的买家和商家
	if len(fail_ids) > 0 {
		db := db.GetDB()
		us := []common.Orders{}
		if err = db.Find(&us, "id in(?)", fail_ids).Error; err != nil {
			glog.Error("RefreshByOrderIds fail! err=", err)
		}
		for _, v := range us {
			user_ids = append(user_ids, v.UserId)
			seller_user_ids = append(seller_user_ids, v.SellerUserId)
			t.UpdateUserId(v.Id, v.UserId, v.SellerUserId)
		}
	}

	// 刷新买家及卖家的订单列表缓存
	user_ids = base.UniqueInt64Slice(user_ids)
	seller_user_ids = base.UniqueInt64Slice(seller_user_ids)

	for _, v := range user_ids {
		t.RefreshCache(v, 0)
	}
	for _, v := range seller_user_ids {
		t.RefreshCache(v, 1)
	}
	return
}

// 订单状态计数缓存
type OrderStatusCntSt struct {
	// OrderStatusCnt map[string]int64 `json:"oscnt"`
	Status int `json:"status"`
	Count  int `json:"count"`
}

func get_orders_typecount(user_id int64, seller int) (vs []OrderStatusCntSt, err error) {
	db := db.GetDB().Model(&common.Orders{}).Where("shield=0 and status>0")
	if 1 == seller {
		db = db.Where("seller_userid=?", user_id)
	} else {
		db = db.Where("user_id=?", user_id)
	}

	rows, err := db.Group("status").Select("status, count(*) as count").Rows()
	if err != nil {
		glog.Error("get_orders_typecount fail! err=", err)
		return
	}
	vs = []OrderStatusCntSt{}
	for rows.Next() {
		var v OrderStatusCntSt
		db.ScanRows(rows, &v)
		vs = append(vs, v)
	}

	var total = 0
	for _, v := range vs {
		total += v.Count
	}
	vs = append(vs, OrderStatusCntSt{Status: -1, Count: total})

	return
}

func get_orders_salecount(user_id int64, seller int) (count int, err error) {
	db := db.GetDB().Model(&common.Orders{}).Where("shield=0 and status>0")
	if 1 == seller {
		db = db.Where("seller_userid=?", user_id)
	} else {
		db = db.Where("user_id=?", user_id)
	}

	// 获取售后中个数
	if err = db.Where("sale_status=?", common.SALE_STATUS_ING).Count(&count).Error; err != nil {
		glog.Error("get_orders_typecount fail! err=", err)
		return
	}

	return
}

func (t *Orders) GetOrderStatusCount(user_id int64, iseller int) (vs []OrderStatusCntSt, err error) {
	ch, err := cache.GetWriter(common.RedisApi)
	if err != nil {
		glog.Error("GetOrderStatusCount fail! err=", err)
		return
	}

	cacheKey := PREFIX_ORDERS_LIST + fmt.Sprintf("%v-%v", user_id, iseller)
	//field := fmt.Sprintf("%v", status)
	expire_time := time.Duration(30 * 24 * time.Hour)
	vs, err, _ = OrdersStatusCountCacheQuery(ch, cacheKey, "status_cnt", expire_time, get_orders_typecount, user_id, iseller)
	if err != nil {
		glog.Error("GetOrderStatusCount fail! err=", err)
		return
	}
	//glog.Infof("GetOrderStatusCount key:%v ishit:%v ", cacheKey, ishit)
	return
}

func (t *Orders) GetOrderAfterSaleCount(user_id int64, iseller int) (count int, err error) {
	ch, err := cache.GetWriter(common.RedisApi)
	if err != nil {
		glog.Error("GetOrderStatusCount fail! err=", err)
		return
	}

	cacheKey := PREFIX_ORDERS_LIST + fmt.Sprintf("%v-%v", user_id, iseller)
	//field := fmt.Sprintf("%v", status)
	expire_time := time.Duration(30 * 24 * time.Hour)
	count, err, _ = OrdersSaleCountCacheQuery(ch, cacheKey, "sale_cnt", expire_time, get_orders_salecount, user_id, iseller)
	if err != nil {
		glog.Error("GetOrderAfterSaleCount fail! err=", err)
		return
	}
	return
}

func (t *Orders) RefreshByProductId(product_id int64) (err error) {
	ch, err := cache.GetWriter(common.RedisApi)
	cacheKey := PREFIX_ORDERS_PRODUCT_LIST + fmt.Sprintf("%v", product_id)
	ch.Del(cacheKey)
	return
}

func (t *Orders) GetProductOrders(product_id int64, page, page_size int) (v OrdersListData, err error) {
	ch, err := cache.GetWriter(common.RedisApi)
	if err != nil {
		glog.Error("GetApiCache fail! err=", err)
		return
	}

	cacheKey := PREFIX_ORDERS_PRODUCT_LIST + fmt.Sprintf("%v", product_id)
	field := fmt.Sprintf("%v-%v", page, page_size)
	var ishit string
	expire_time := time.Duration(240 * time.Hour)

	v, err, ishit = OrdersProductListCacheQuery(ch, cacheKey, field, expire_time, getOrdersProductList, product_id, page, page_size)
	if err != nil {
		glog.Error("Orders_List fail! err=", err)
		return
	}
	glog.Infof("Orders_productList key:%v ishit:%v ", cacheKey, ishit)
	return
}

func getOrdersProductList(product_id int64, page, page_size int) (v OrdersListData, err error) {
	db := db.GetDB()
	db.DB = db.Where("product_id=? and status>=? and status<>?", product_id, common.ORDER_STATUS_PAYED, common.ORDER_STATUS_CANCEL)
	if err = db.Model(&common.Orders{}).Count(&v.Total).Error; err != nil {
		glog.Error("getOrdersProductList fail! err=", err)
		return
	}
	v.List = []common.Orders{}
	if err = db.ListPage(page, page_size).Order("create_time desc").Find(&v.List).Error; err != nil {
		glog.Error("Orders_ListByProduct fail! err=", err)
		return
	}

	v.ListEnded = base.IsListEnded(page, page_size, len(v.List), v.Total)

	return
}
