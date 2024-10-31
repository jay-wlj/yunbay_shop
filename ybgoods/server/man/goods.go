package man

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"yunbay/ybgoods/common"
	"yunbay/ybgoods/dao"
	"yunbay/ybgoods/server/share"
	"yunbay/ybgoods/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 批量插入商品
func GoodsUpsert(c *gin.Context) {
	var req []*common.Product
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	db := db.GetTxDB(c)
	if err := share.AddProduct(db, req); err != nil {
		yf.JSON_Fail(c, err.Error())
		return
	}
	db.Commit()
	yf.JSON_Ok(c, gin.H{})
}

func GoodsInfo(c *gin.Context) {
	var req common.IdSkuReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	v, err := dao.GetGoodsInfo(req.Id)

	//if err := db.Find(&v, "id=?", id).Error; err != nil {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("GoodsInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err := share.FormatProduct(&v, true); err != nil {
		glog.Error("GoodsInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 只取当前sku_id
	if req.SkuId > 0 {
		for i, s := range v.Skus {
			if s.Id == req.SkuId {
				v.Skus = v.Skus[i : i+1]
				break
			}
		}
	}
	yf.JSON_Ok(c, v)
}

// 获取商品的售后联系方式接口
func GoodsContact(c *gin.Context) {
	var req common.IdReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	var v common.Product
	db := db.GetDB()
	if err := db.Select("id, contact").Find(&v, "id=?", req.Id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("GoodsContact fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, base.FilterStruct(v, true, "id", "contact"))
}

type QuantitySt struct {
	OrderId      int64 `json:"order_id"`
	ProductId    int64 `json:"product_id"`
	ProductSkuId int64 `json:"product_sku_id"`
	Quantity     int64 `json:"quantity"`
	Stock        int64 `json:"-"`
}

var alreadydostock = errors.New("repeted notify")

// 执行加减库存
func (t *QuantitySt) Do() (err error) {
	var ok bool
	ch, e := cache.GetWriter(common.RedisGoods)
	if e != nil {
		err = e
		glog.Error("QuantitySt fail! err=", err)
		return
	}
	// 分布式锁获取
	key := "p_stock_lock"
	field := fmt.Sprintf("%v-%v-%v-%v", t.OrderId, t.ProductId, t.ProductSkuId, t.Quantity)

	ch.Expire(key, 1*time.Hour)
	if ok, err = ch.HSetNx(key, field, t.Quantity); err != nil {
		glog.Error("QuantitySt Do fail! err=", err)
		return
	}

	if !ok {
		err = alreadydostock
		return
	}
	//ch.HSet(key, field, t.Quantity, 1*time.Hour) // 设置过期时间

	defer func() {
		if err != nil {
			ch.HDel(key, field) // 需释放分布式锁
		}
	}()

	// 执行商品库存的缓存加减
	gq := dao.GoodsQuantity(t.ProductId)
	if t.Stock, err = gq.AddQuantity(t.ProductSkuId, t.Quantity); err != nil {
		//gq.AddQuantity(v.ProductSkuId, -v.Quantity) // 恢复当前失败的库存
		glog.Error("QuantitySt fail! AddQuantity err=", err)
		return
	}
	return
}

func (t *QuantitySt) Restore() (err error) {
	gq := dao.GoodsQuantity(t.ProductId)
	_, err = gq.AddQuantity(t.ProductSkuId, -t.Quantity)
	if err != nil {
		glog.Error("QuantitySt Restore fail! err=", err)
	}
	key := "product_stock_lock"
	field := fmt.Sprintf("%v-%v-%v-%v", t.OrderId, t.ProductId, t.ProductSkuId, t.Quantity)

	ch, e := cache.GetWriter(common.RedisGoods)
	if e != nil {
		err = e
		glog.Error("QuantitySt Restore fail! err=", err)
		return
	}

	ch.HDel(key, field) // 删除分布式锁
	return
}

// // 商品库存增量修改
// func GoodsPlusQuantity(c *gin.Context) {
// 	var req []QuantitySt
// 	if ok := util.UnmarshalReq(c, &req); !ok {
// 		return
// 	}

// 	gq := dao.GoodsQuantity(req.ProductSkuId)
// 	if ok, _ := gq.AddQuantity(req.Quantity); ok {
// 		if err := db.GetDB().Model(&common.ProductSku{}).Where("id=? ", req.ProductSkuId).Updates(base.Maps{"stock": gorm.Expr("stock+?", req.Quantity), "sold": gorm.Expr("sold+?", -req.Quantity)}).Error; err != nil {
// 			// 还原库存缓存
// 			gq.AddQuantity(-req.Quantity)
// 			glog.Error("GoodsPlusQuantity fail! err=", err)
// 			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
// 			return
// 		}
// 		yf.JSON_Ok(c, gin.H{})
// 		return
// 	}
// 	yf.JSON_Fail(c, common.ERR_PRODUCT_NOT_MORE)
// }

// 商品库存增量修改
func GoodsPlusQuantity(c *gin.Context) {
	var req []QuantitySt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	ok_ids := []int{}
	var err error
	//var ok bool

	for i := range req {
		if err = req[i].Do(); err != nil {
			// 还原库存缓存
			glog.Error("GoodsPlusQuantity fail! err=", err)
			break
		}
		ok_ids = append(ok_ids, i)
	}

	if len(ok_ids) != len(req) {
		// 没有全部商品扣减库存失败 则恢复库存处理
		for _, i := range ok_ids {
			req[i].Restore()
		}

		// 如果是取消库存 遇到已经释放的库存或商品不存在时 也返回成功
		if len(req) == 1 && req[0].Quantity > 0 && (err == gorm.ErrRecordNotFound || err == alreadydostock) {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		yf.JSON_Fail(c, common.ERR_PRODUCT_NOT_MORE)
		return
	}

	// 更新数据库操作 TODO 改为定时从缓存增量中更新

	db := db.GetTxDB(c)
	for _, v := range req {
		// 有规格的商品
		if v.ProductSkuId > 0 {
			res := db.Model(&common.ProductSku{}).Where("id=? ", v.ProductSkuId).Updates(base.Maps{"stock": v.Stock, "sold": gorm.Expr("sold+?", -v.Quantity)})
			if err = res.Error; err != nil {
				// 还原库存缓存
				glog.Error("GoodsPlusQuantity fail! err=", err)
				break
			}
		} else {
			// 无规格情况
			res := db.Model(&common.Product{}).Where("id=? ", v.ProductId).Updates(base.Maps{"stock": v.Stock, "sold": gorm.Expr("sold+?", -v.Quantity)})
			if err = res.Error; err != nil {
				// 还原库存缓存
				glog.Error("GoodsPlusQuantity fail! err=", err)
				break
			}
		}
	}

	if err != nil {
		// 没有全部商品扣减库存失败 则恢复库存处理
		for _, i := range ok_ids {
			req[i].Restore()
		}
		yf.JSON_Fail(c, common.ERR_PRODUCT_NOT_MORE)
		return
	}
	db.Commit()

	// 刷新缓存
	for _, v := range req {
		dao.RefreshGoods(v.ProductId)
	}
	yf.JSON_Ok(c, gin.H{})
}

type quantityReq struct {
	ProductId    int64 `json:"product_id" binding:"gt=0"`
	ProductSkuId int64 `json:"product_sku_id" binding:"gte=0"`
	Stock        int64 `json:"stock" binding:"gte=0"`
	Sold         int64 `json:"sold" binding:"gte=0"`
}

// 设置商品库存
func GoodsQuantitySet(c *gin.Context) {
	var req quantityReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	ydb := db.GetTxDB(c)
	f := func() (err error) {
		if req.ProductSkuId > 0 {
			if err = ydb.Model(&common.ProductSku{}).Where("product_id=? and id=?", req.ProductSkuId, req.ProductId).Updates(base.Maps{"stock": req.Stock, "sold": req.Sold}).Error; err != nil {
				glog.Error("GoodsQuantitySet  fail! err=", err)
			}
		} else {
			if err = ydb.Model(&common.Product{}).Where("id=?", req.ProductId).Updates(base.Maps{"stock": req.Stock, "sold": req.Sold}).Error; err != nil {
				glog.Error("GoodsQuantitySet  fail! err=", err)
			}
		}

		pq := dao.GoodsQuantity(req.ProductId)
		if err = pq.SetQuantity(req.ProductSkuId, req.Stock); err != nil {
			glog.Error("GoodsQuantitySet  fail! err=", err)
			return
		}

		ydb.AfterCommit(func() {
			dao.RefreshGoods(req.ProductId)
		})
		return
	}
	if err := f(); err != nil {
		yf.JSON_Fail(c, err.Error())
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

// 搜索商品
func GoodsListByIds(c *gin.Context) {
	str_ids, _ := base.CheckQueryStringField(c, "ids")
	ids := base.StringToInt64Slice(str_ids, ",")

	db := db.GetDB()
	vs := []common.Product{}
	order_sql := fmt.Sprintf("array_position('{%v}'::bigint[], id)", base.Int64SliceToString(ids, ",")) // 按ids顺序排序
	if err := db.Order(order_sql).Select("id, images[1:1], info, title, publish_area, price, rebat").Find(&vs, "status=?  and is_hid=0 and id in(?)", common.STATUS_OK, ids).Error; err != nil {
		glog.Error("GoodsSearch fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if err := share.FormatProductList(vs); err != nil {
		glog.Error("GoodsSearch fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"list": vs})
}

type ProdIds struct {
	ProductId    int64
	ProductSkuId int64
}

// 批量获取指定规格的商品列表
func GoodsListDetail(c *gin.Context) {
	str, _ := base.CheckQueryStringField(c, "list_id_model")
	str_list := strings.Split(str, ",")

	pids := []int64{}
	//skuids := []int64{}
	mskus := make(map[int64][]int64)
	for _, s := range str_list {
		var v ProdIds
		if _, err := fmt.Sscanf(s, "%v-%v", &v.ProductId, &v.ProductSkuId); err != nil {
			glog.Error("GoodsListDetail fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
			return
		}
		mskus[v.ProductId] = append(mskus[v.ProductId], v.ProductSkuId)
		pids = append(pids, v.ProductId)
		//skuids = append(skuids, v.ProductSkuId)
	}
	pids = base.UniqueInt64Slice(pids)
	if 0 == len(pids) {
		yf.JSON_Ok(c, gin.H{"list": []interface{}{}})
		return
	}

	db := db.GetDB()
	vs := []common.Product{}
	if err := db.Find(&vs, "id in(?)", pids).Error; err != nil {
		glog.Error("GoodsListDetail fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 获取
	for i, v := range vs {
		if err := share.FormatProduct(&vs[i], true); err != nil {
			glog.Error("GoodsListDetail fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
		// 去掉多余的规格
		//db.Model(&vs[i]).Related(&vs[i].Skus)
		mvs := make(map[int64][]*common.ProductSku)
		if m, ok := mskus[v.Id]; ok {
			for j, sku := range vs[i].Skus {
				for _, k := range m {
					if k == sku.Id {
						mvs[v.Id] = append(mvs[v.Id], vs[i].Skus[j:j+1]...)
					}
				}
			}
			vs[i].Skus = mvs[v.Id]
		}
	}
	yf.JSON_Ok(c, gin.H{"list": base.FilterStruct(vs, false, "sku", "descimgs", "create_time", "update_time")})
}

// 上下线商品
func GoodsHidOne(c *gin.Context) {
	var req common.IdOkSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	db := db.GetTxDB(c)
	res := db.Model(&common.Product{}).Where("id=?", req.Id).Updates(base.Maps{"is_hid": req.Status, "hid_cause": req.Reason})
	if res.Error != nil {
		glog.Error("GoodsOffine fail! err=", res.Error)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if res.RowsAffected > 0 {
		dao.RefreshGoods(req.Id)
	}
	yf.JSON_Ok(c, gin.H{})
}

func GoodsHid(c *gin.Context) {
	var req common.IdsOkSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	ids := []int64{}
	for _, v := range req.Ids {
		ids = append(ids, v)
	}
	db := db.GetTxDB(c)
	res := db.Model(&common.Product{}).Where("id in(?)", ids).Updates(base.Maps{"is_hid": req.Status, "hid_cause": req.Reason})
	if res.Error != nil {
		glog.Error("GoodsOffine fail! err=", res.Error)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if res.RowsAffected > 0 {
		for _, id := range ids {
			dao.RefreshGoods(id)
		}
	}
	yf.JSON_Ok(c, gin.H{})
}

// 获取商品价格列表
func GoodsPriceListByIds(c *gin.Context) {
	str_ids, _ := base.CheckQueryStringField(c, "ids")
	ids := base.StringToInt64Slice(str_ids, ",")

	db := db.GetDB()
	vs := []common.ProductPrice{}

	if err := db.Select("p_id, p_sku_id, cost_price, price").Find(&vs, "p_id in(?)", ids).Error; err != nil {
		glog.Error("GoodsPriceListByIds fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": vs})
}
