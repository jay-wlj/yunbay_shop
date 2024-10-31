package share

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"yunbay/ybgoods/common"
	"yunbay/ybgoods/dao"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"github.com/shopspring/decimal"

	"github.com/jinzhu/gorm"

	"github.com/jie123108/glog"
)

func AddProduct(db *db.PsqlDB, vs []*common.Product) (err error) {
	// 如果没有属性 则添加属性 更新商品的价格
	for i := range vs {
		if err = AddProductOne(db, vs[i]); err != nil {
			glog.Error("AddProduct fail! err=", err)
			return
		}
	}

	for i := range vs {
		// 添加库存到缓存
		pq := dao.GoodsQuantity(vs[i].Id)
		for _, v := range vs[i].Skus {
			if err = pq.SetQuantity(v.Id, v.Stock); err != nil {
				glog.Error("GoodsUpsert SetQuantity fail! err=", err)
				return
			}
		}

		// 添加价格到缓存
		pp := dao.GoodsPrice(vs[i].Id)
		for _, v := range vs[i].Skus {
			if err = pp.SetPrice(v.Id, v.CostPrice, v.Price); err != nil {
				glog.Error("GoodsUpsert SetPrice fail! err=", err)
				return
			}
		}

		if len(vs[i].Skus) == 0 {
			pq.SetQuantity(0, vs[i].Stock)               // 无规格库存设置
			pp.SetPrice(0, vs[i].CostPrice, vs[i].Price) // 无规格价格设置
		}

	}
	return
}

func AddProductOne(db *db.PsqlDB, req *common.Product) (err error) {
	// 如果没有属性 则添加属性 更新商品的价格
	for i, v := range req.Skus {

		if len(v.Combines) > 0 {
			ss := []Sku{}
			vcs := []map[string]string{}
			buf, _ := json.Marshal(v.Combines)
			if err = json.Unmarshal(buf, &vcs); err != nil {
				buf, _ := json.Marshal(*req)
				glog.Error("AddProductOne fail! err=", err, " req=", string(buf))
				break
			}
			// 创建属性
			for _, v := range vcs {
				// 这里每一项里应该只有唯一的一个kv
				for k, t := range v {
					var s Sku
					if s.AttrId, err = dao.SetGetAttrKey(req.CategoryId, k); err != nil {
						glog.Error("SetGetAttrKey fail! err", err)
						return
					}

					if s.ValueId, err = dao.SetGetAttrValue(s.AttrId, t); err != nil {
						glog.Error("SetGetAttrKey fail! err", err)
						return
					}
					ss = append(ss, s)
				}
			}

			buf, _ = json.Marshal(ss)
			req.Skus[i].Sku.RawMessage = buf
		}
	}

	// 判断虚拟商品 maninfo信息是否合法
	if req.Type > 0 {
		switch req.Type {
		case common.GOODS_TYPE_TEL_RECHARGE: // 话费充值
			for _, v := range req.Skus {
				amount, ok := v.Extinfo["amount"].(string)
				nu, er := strconv.Atoi(amount)
				if !ok || er != nil || !(nu == 50 || nu == 100) {
					// 不含此规格的实际话费充值
					glog.Error("GoodsUpsert fail! tel recharge amount is empty or invalid! not (50|100) amount=", amount)
					err = errors.New(yf.ERR_ARGS_INVALID)
					return
				}
			}
		}
	}

	// 多规格商品选取最小价格的规格及库存
	var def_sku_i int
	if len(req.Skus) > 0 {
		var stock int64
		req.Price = req.Skus[0].Price
		for i, v := range req.Skus {
			if req.Price.GreaterThan(v.Price) {
				req.Price = v.Price
				def_sku_i = i
			}
			stock += v.Stock
		}
		req.Stock = stock
	}

	// 先删掉现有商品的规格
	if req.Id > 0 {
		if err = db.Delete(&common.ProductSku{}, "product_id=?", req.Id).Error; err != nil {
			glog.Error("GoodsUpsert fail! err=", err)
			return
		}
	}

	// 折扣默认为1
	if req.Discount.IsZero() {
		req.Discount = decimal.New(1, 0)
	}

	// 不要修改销量
	//req.Sold =

	if err = db.Save(&req).Error; err != nil {
		glog.Error("GoodsUpsert fail! err=", err)
		return
	}

	// 更新商品的默认规格为最小价格的规格
	if len(req.Skus) > 0 {
		if err = db.Table("product").Where("id=?", req.Id).Updates(base.Maps{"def_sku_id": req.Skus[def_sku_i].Id}).Error; err != nil {
			glog.Error("GoodsUpsert fail! err=", err)
			return
		}
	}

	// 保存商品成本价及售价
	pps := []common.ProductPrice{}
	for _, v := range req.Skus {
		pps = append(pps, common.ProductPrice{PId: req.Id, PSkuId: v.Id, CostPrice: v.CostPrice, Price: v.Price})
	}
	if len(req.Skus) == 0 {
		pps = append(pps, common.ProductPrice{PId: req.Id, PSkuId: 0, CostPrice: req.CostPrice, Price: req.Price})
	}

	for _, v := range pps {
		db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (p_id,p_sku_id) DO update set cost_price=%v,price=%v", v.CostPrice, v.Price))
		if err = db.Save(&v).Error; err != nil {
			glog.Error("AddProductOne fail! err=", err)
			return
		}
	}
	db.DB = db.Set("gorm:insert_option", "")

	// 刷新缓存
	dao.RefreshGoods(req.Id)
	return
}

// 复制一个商品
func DuplicateGoods(tdb *db.PsqlDB, id int64, f func(v *common.Product) bool) (v common.Product, err error) {
	if tdb == nil {
		tdb = db.GetDB()
	}
	//var v common.Product
	if err = tdb.Find(&v, "id=?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(yf.ERR_NOT_FOUND)
			return
		}
		err = errors.New(yf.ERR_SERVER_ERROR)
		return
	}

	if f != nil {
		if ok := f(&v); ok {
			return
		}
	}

	// 获取关联的商品规格

	err = db.GetDB().Model(v).Related(&v.Skus).Error
	if err != nil {
		glog.Error("GoodsDuplicate fail! err=", err)
		err = errors.New(yf.ERR_SERVER_ERROR)
		return
	}

	v.Id = 0
	for i := range v.Skus {
		v.Skus[i].Id = 0
		v.Skus[i].ProductId = 0
	}

	if err = tdb.Save(&v).Error; err != nil {
		glog.Error("GoodsDuplicate fail! err=", err)
		err = errors.New(yf.ERR_SERVER_ERROR)
		return
	}

	// 更新默认规格
	if len(v.Skus) > 0 {
		if err = tdb.Model(&v).Updates(base.Maps{"def_sku_id": v.Skus[0].Id}).Error; err != nil {
			glog.Error("GoodsDuplicate fail! err=", err)
			err = errors.New(yf.ERR_SERVER_ERROR)
			return
		}
	}
	return
}
