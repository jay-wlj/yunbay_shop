package dao

import (
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"strconv"
	"yunbay/ybgoods/common"

	"github.com/shopspring/decimal"

	"github.com/jie123108/glog"
)

const (
	PREFIX_GOODS_PRICE string = "goods_price:"
)

func init() {
	//cache.MakeHCacheQuery(&GoodsQuantityCacheQuery)
}

type GoodsPrice int64

func (t *GoodsPrice) GetPrice(sku_id int64) (decimal.Decimal, error) {
	var ch *cache.RedisCache
	ch, err := cache.GetReader(common.RedisGoods)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return decimal.New(0, 0), err
	}
	key := PREFIX_GOODS_PRICE + strconv.FormatInt(int64(*t), 10)

	if ok, _ := ch.Exists(key).Result(); !ok {
		// 不存在 则读取数据库中的库存
		var v common.ProductPrice
		db := db.GetDB()
		if err = db.Find(&v, "p_id=? and p_sku_id=?", *t, sku_id).Error; err != nil {
			return decimal.New(0, 0), err
		}
		ch.HSet(key, strconv.FormatInt(sku_id, 10)+"_cost_price", v.CostPrice.String(), 0)
		ch.HSet(key, strconv.FormatInt(sku_id, 10)+"_price", v.Price.String(), 0)
	}

	val, err := ch.HGet(key, strconv.FormatInt(sku_id, 10)+"_price")
	if err != nil {
		return decimal.New(0, 0), err
	}
	return decimal.NewFromString(val)
}

func (t *GoodsPrice) GetCostPrice(sku_id int64) (decimal.Decimal, error) {
	var ch *cache.RedisCache
	ch, err := cache.GetReader(common.RedisGoods)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return decimal.New(0, 0), err
	}
	key := PREFIX_GOODS_PRICE + strconv.FormatInt(int64(*t), 10)

	if ok, _ := ch.Exists(key).Result(); !ok {
		// 不存在 则读取数据库中的库存
		var v common.ProductPrice
		db := db.GetDB()
		if err = db.Find(&v, "p_id=? and p_sku_id=?", *t, sku_id).Error; err != nil {
			return decimal.New(0, 0), err
		}
		ch.HSet(key, strconv.FormatInt(sku_id, 10)+"_cost_price", v.CostPrice.String(), 0)
		ch.HSet(key, strconv.FormatInt(sku_id, 10)+"_price", v.Price.String(), 0)
	}

	val, err := ch.HGet(key, strconv.FormatInt(sku_id, 10)+"_cost_price")
	if err != nil {
		return decimal.New(0, 0), err
	}
	return decimal.NewFromString(val)
}

// 设置成本及售价
func (t *GoodsPrice) SetPrice(sku_id int64, cost_price, price decimal.Decimal) error {
	var ch *cache.RedisCache
	ch, err := cache.GetReader(common.RedisGoods)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return err
	}
	key := PREFIX_GOODS_PRICE + strconv.FormatInt(int64(*t), 10)

	if err = ch.HSet(key, strconv.FormatInt(sku_id, 10)+"_cost_price", cost_price.String(), 0); err != nil {
		glog.Error("SetPrice fail! err=", err)
		return err
	}

	if err = ch.HSet(key, strconv.FormatInt(sku_id, 10)+"_price", price.String(), 0); err != nil {
		glog.Error("SetPrice fail! err=", err)
		return err
	}

	return nil
}
