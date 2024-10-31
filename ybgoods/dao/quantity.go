package dao

import (
	"errors"
	"strconv"
	"yunbay/ybgoods/common"

	"github.com/jinzhu/gorm"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jie123108/glog"
)

const (
	PREFIX_GOODS_QUANTITY      string = "goods_stock:"
	PREFIX_GOODS_QUANTITY_LOCK string = "goods_quantity_lock"

	PREFIX_GOODS_SOLD        string = "sold:"
	PREFIX_GOODS_SOLD_INCKEY string = "inc_sold"
)

func init() {
	//cache.MakeHCacheQuery(&GoodsQuantityCacheQuery)
}

type GoodsQuantity int64

func (t *GoodsQuantity) GetQuantity(sku_id int64) (int64, error) {
	var ch *cache.RedisCache
	ch, err := cache.GetReader(common.RedisGoods)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return 0, err
	}
	key := PREFIX_GOODS_QUANTITY + strconv.FormatInt(int64(*t), 10)

	if ok, _ := ch.Exists(key).Result(); !ok {
		// 不存在 则读取数据库中的库存
		var v common.Product
		db := db.GetDB()
		if err = db.Find(&v, "id=?", *t).Error; err != nil {
			return 0, err
		}
		db.Model(&v).Related(&v.Skus)
		for _, s := range v.Skus {
			ch.HSet(key, strconv.FormatInt(s.Id, 10), s.Stock, 0)
		}
		if len(v.Skus) == 0 {
			ch.HSet(key, "0", v.Stock, 0)
		}
	}
	if sku_id == 0 {
		// 获取该商品下所有规格的库存
		ss := []int64{}
		if err := ch.HVals(key).ScanSlice(&ss); err != nil {
			glog.Error("GetDefaultCache fail! err=", err)
			return 0, err
		}
		var total int64
		for _, v := range ss {
			total += v
		}

		return total, nil
	}
	return ch.HGetI(key, strconv.FormatInt(sku_id, 10))
}

// 加减库存
func (t *GoodsQuantity) AddQuantity(sku_id, incr int64) (int64, error) {
	var ch *cache.RedisCache
	ch, err := cache.GetReader(common.RedisGoods)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return 0, err
	}
	key := PREFIX_GOODS_QUANTITY + strconv.FormatInt(int64(*t), 10)
	field := strconv.FormatInt(sku_id, 10)

	quantity, err := t.GetQuantity(sku_id)
	if err != nil {
		return 0, err
	}
	// 不限库存
	if quantity == -1 {
		return quantity, nil
	} else if quantity >= 0 && quantity+incr >= 0 {
		// 确保n>=0,以免超卖
		quantity, err = ch.HIncrBy(key, field, incr)
		if quantity < 0 {
			// 扣减失败，需要还原库存变化
			ch.HIncrBy(key, field, -incr)
			err = errors.New(common.ERR_PRODUCT_NOT_MORE)
			return 0, err
		}
		return quantity, err
	} else {
		// 失败
		err = errors.New(common.ERR_PRODUCT_NOT_MORE)
		return quantity, err
	}

	return 0, err
}

// 设置库存
func (t *GoodsQuantity) SetQuantity(sku_id, quantity int64) error {
	var ch *cache.RedisCache
	ch, err := cache.GetReader(common.RedisGoods)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return err
	}
	key := PREFIX_GOODS_QUANTITY + strconv.FormatInt(int64(*t), 10)
	field := strconv.FormatInt(sku_id, 10)

	return ch.HSet(key, field, quantity, 0)
}

// 获取销量
func (t *GoodsQuantity) GetSold(sku_id int64) (n int64, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisGoods)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}
	key := PREFIX_GOODS_SOLD + strconv.FormatInt(int64(*t), 10)

	if ok, _ := ch.Exists(key).Result(); !ok {
		// 不存在 则读取数据库中的库存
		db := db.GetDB()
		var sold int64
		if sku_id > 0 {
			var v common.ProductSku

			if err = db.Find(&v, "product_id=? and id=?", *t, sku_id).Error; err != nil && err != gorm.ErrRecordNotFound {
				return
			}
			sold = v.Sold
		} else {
			var v common.Product
			if err = db.Find(&v, " id=?", *t).Error; err != nil && err != gorm.ErrRecordNotFound {
				return
			}
			sold = v.Sold
		}

		ch.HSet(key, strconv.FormatInt(sku_id, 10), sold, 0)
	}

	n, err = ch.HGetI(key, strconv.FormatInt(sku_id, 10))
	return
}

// +-销量
func (t *GoodsQuantity) AddSold(sku_id int64, incr int64) (n int64, err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisGoods)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}
	key := PREFIX_GOODS_SOLD + strconv.FormatInt(int64(*t), 10)

	n, err = t.GetSold(sku_id)
	if err != nil {
		glog.Error("AddSold fail! err=", err)
		return
	}
	n, err = ch.HIncrBy(key, strconv.FormatInt(sku_id, 10), incr)
	if err == nil {
		// TODO 有待优化
		if sku_id > 0 {
			// 更新总销量
			ch.HIncrBy(key, "0", incr)
		}
	}
	// 添加到销量增量key中
	ch.SAdd(PREFIX_GOODS_SOLD_INCKEY, key)
	return
}

// // 每隔段时间销量增量持久到数据库中
// func (t *GoodsQuantity) PersistSave() (n int64, err error) {
// 	var ch *cache.RedisCache

// 	ch, err = cache.GetReader(common.RedisPub)
// 	if err != nil {
// 		glog.Error("GetHighGoodsList fail! err=", err)
// 		return
// 	}

// 	// 一次只更新50个增量key
// 	keys := ch.SPopN(PREFIX_GOODS_HIGHLIST, 50).Val()

// 	ydb := db.GetTxDB(nil)
// 	fnDo = func (err error){
// 		for _, key := range keys {
// 			// 解析出商品id
// 			str_pid := strings.TrimPrefix(key, PREFIX_GOODS_SOLD)
// 			if pid, e := strconv.Atoi(str_pid); e != nil {
// 				glog.Error("PersistSave pid not number! err", e, " pid:", str_pid)
// 				continue
// 			}

// 			// 获取该商品下所有规格库存
// 			vals := ch.HGetAll(key)
// 			msold := make(map[int]int)
// 			var sid int
// 			var sold int
// 			for k, v := range vals {
// 				if sid, err = strconv.Atoi(k); e != nil {
// 					glog.Error("PersistSave sku_id not number! err", e, " sku_id:", k)
// 					continue
// 				}
// 				if sold, err = strconv.Atoi(v); e != nil{
// 					glog.Error("PersistSave sku_id val not number! err", e, " val:", v)
// 					continue
// 				}
// 				msold[sid] = sold
// 			}

// 			// 更新到数据库中
// 			for sid, v := range msold {
// 				switch sid {
// 				case 0:
// 					if err = ydb.Model(&common.Product{}).Where("id=?", pid).Updates(base.Maps{"sold":v}).Error(); err != nil {
// 						glog.Error("PersistSave fail! err=", err)
// 						return
// 					}
// 				default:
// 					if err = ydb.Model(&common.ProductSkud{}).Where("product_id=? and id=?", pid, sid).Updates(base.Maps{"sold":v}).Error(); err != nil {
// 						glog.Error("PersistSave fail! err=", err)
// 						return
// 					}
// 				}
// 			}
// 		}
// 		return
// 	}

// 	if err = fnDo(); err != nil {
// 		// 需还原移除的增量key
// 		ch.SAdd(PREFIX_GOODS_HIGHLIST, keys...)
// 		ydb.Rollback()
// 		return
// 	}
// 	ydb.Commit()
// 	return
// }
