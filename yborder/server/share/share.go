package share

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"yunbay/yborder/common"
	"yunbay/yborder/dao"
	"yunbay/yborder/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/shopspring/decimal"

	"github.com/jayden211/retag"
	"github.com/jie123108/glog"
	//"encoding/json"
)

func GetCurrencyTypeByCoin(coin string) int {
	coin = strings.ToLower(coin)
	switch coin {
	case "ybt":
		return common.CURRENCY_YBT
	case "kt":
		return common.CURRENCY_KT
	case "cny":
		return common.CURRENCY_RMB
	case "snet":
		return common.CURRENCY_SNET
	default:
		return -1
	}
}

func GetCoinByCurrencyType(txType int) string {
	switch txType {
	case common.CURRENCY_YBT:
		return "ybt"
	case common.CURRENCY_KT:
		return "kt"
	case common.CURRENCY_RMB:
		return "cny"
	case common.CURRENCY_SNET:
		return "snet"
	}
	return ""
}

// 获取币种转换兑换比例
func GetRatioCache(to_type, from_type int) (ratio float64, err error) {
	cache, err := cache.GetWriter(common.RedisPub)

	to_str := common.GetCurrencyName(to_type)
	from_str := common.GetCurrencyName(from_type)
	// 优先从缓存里获取
	if err == nil {
		key := fmt.Sprintf("asset_ratio_%v", to_str)
		ratio, err = cache.HGetF64(key, from_str)
	}
	// 否则从资产接口中获取
	if err != nil {
		var ratios map[string]float64
		ratios, err = util.YBAsset_GetRatio(to_type)
		if v, ok := ratios[from_str]; ok {
			ratio = v
		} else {
			glog.Error("GetRatioCache fail! to_type:", to_str, " no from_type:", from_str)
		}
	}
	return
}

// 获取购物车包含的商品信息
func GetCartsProductInfos(vs []common.Cart) (err error) {
	if len(vs) == 0 {
		return
	}
	// 接口调用
	pids := []util.ProdIds{}
	for _, v := range vs {
		if v.Product == nil {
			pids = append(pids, util.ProdIds{ProductId: v.ProductId, ProductModelId: v.ProductSkuId})
		}
	}

	var ms map[int64]*common.Product
	if ms, err = util.ListProductInfo(pids); err != nil {
		glog.Error("GetOrderProductInfo fail! err=", err)
		return
	}
	// 剔除非勾选规格
	for i, v := range vs {
		if k, ok := ms[v.ProductId]; ok {
			vs[i].Product = base.StructToMap(k)
		}
	}

	return
}

// 添加订单
func UpsertOrders(db *db.PsqlDB, v *common.Orders) (err error) {
	// if v.SellerUserId == 0 {
	// 	// 获取商品卖家用户id
	// 	var product common.Product
	// 	if err = db.Select("user_id").Where("id=?", v.ProductId).Find(&product).Error; err != nil {
	// 		glog.Errorf("UpsertOrders err=%v", err)
	// 		return
	// 	}
	// 	v.SellerUserId = product.UserId
	// }
	// 买家不能购物自己的商品
	// if v.SellerUserId == v.UserId {
	// 	glog.Error("buyer can not bu own's product! user_id:", v.UserId, " product_id:%v", v.SellerUserId)
	// 	err = fmt.Errorf(common.ERR_FORBIDDEN_BUY_OWNGOODS)
	// 	return
	// }

	o := dao.Orders{db}
	if err = o.Upsert(v); err != nil {
		glog.Error("UpsertOrders fail! err=", err)
		return
	}
	return
}

// // 获取订单当前商品一些快照信息
// func GetSnapOrdersProduct(db *dao.PsqlDB, v *common.Orders)(err error) {
// 	p := dao.ProModSt{ProductId:v.ProductId, ProductSkuId:v.ProductSkuId}
// 	var product common.Product
// 	product, err = dao.Product_ByModelId(p)
// 	if err != nil {
// 		glog.Error("Product_ByModelId fail! err=", err)
// 		return
// 	}

// 	info := base.SelectStructView(product, "order")

// 	v.Product = info
// 	v.SellerUserId = product.UserId		// 获取商品卖家用户id
// 		//err = db.Model(&v).Updates(map[string]interface{}{"product":v.Product, "update_time":time.Now().Unix()}).Error

// 	return
// }

// 提交订单后保存当前商品一些快照信息
func SnapOrdersProducts(db *db.PsqlDB, vs []common.Orders) (reason string, err error) {
	// pids := []dao.ProModSt{}
	// for _, v := range vs {
	// 	if v.Product == nil {
	// 		pids = append(pids, dao.ProModSt{ProductId:v.ProductId, ProductSkuId:v.ProductSkuId})
	// 	}
	// }
	o := dao.Orders{db}
	if len(vs) > 0 {

		for i, v := range vs {
			if v.Product == nil {
				var p common.Product

				if p, err = util.GetProductInfo(v.ProductId, v.ProductSkuId); err != nil {
					glog.Error("SnapOrdersProducts fail! err=", err)
					return
				}

				//info := Filter_Obj(p, "order")
				//info := base.SelectStructView(p, "order")
				info := base.FilterStruct(p, true, "id", "images", "rebat", "type", "skus", "title", "price", "pay_price", "extinfo")
				vs[i].SellerUserId = p.UserId
				vs[i].Product = info.(map[string]interface{})
				vs[i].PublishArea = p.PublishArea
				vs[i].ProductType = p.Type

				// 刷新订单对应的商家id
				o.UpdateUserId(vs[i].Id, vs[i].UserId, vs[i].SellerUserId)

				// 买家不能购物自己的商品
				if vs[i].SellerUserId == vs[i].UserId {
					glog.Error("buyer can not bu own's product! user_id:", vs[i].UserId, " product_id:%v", vs[i].SellerUserId)
					reason = common.ERR_FORBIDDEN_BUY_OWNGOODS
					err = fmt.Errorf(common.ERR_FORBIDDEN_BUY_OWNGOODS)
					return
				}

				// 无规格获取商品价格
				m := &common.ProductSku{Price: p.Price, PayPrice: p.PayPrice, Extinfo: p.Extinfo}

				// 获取当前规格的商品售价信息
				//maninfos := "'{}'"
				man := &common.ManInfos{}
				for i, s := range p.Skus {
					if s.Id == v.ProductSkuId {
						m = &p.Skus[i]
						break
					}
				}

				//maninfos = fmt.Sprintf("'{\"cost_price\": %v}'", m.CostPrice)
				man.Parse(m)
				man.Amount = man.Amount.Mul(decimal.New(int64(vs[i].Quantity), 0)) // bugfix 真实售价kt=商品kt售价 * 数量
				switch p.Type {
				case common.GOODS_TYPE_TEL_RECHARGE:
					// 如果是话费充值，则需要检测 extinfos里是否有充值的手机号
					tel, ok := v.ExtInfos["tel"].(string)
					if !ok || !yf.ValidTel(tel) {
						// 效验tel是否合法
						err = errors.New(common.ERR_TEL_INVALID)
						glog.Error("ERR_TEL_INVALID tel:", tel)
						reason = common.ERR_TEL_INVALID
						return
					}

					if amount, ok := m.Extinfo["amount"].(string); ok {
						if man.OfAmount, err = decimal.NewFromString(amount); err != nil {
							glog.Error("Extinfo amount fail! err=", err)
							return
						}
					} else {
						glog.Error("Extinfo  fail! no amount!")
						err = errors.New(yf.ERR_SERVER_ERROR)
						return
					}

					// 调用of手机话费套餐接口是否支持该手机号充值
					if err = util.GetOfpay().Tel_Check(tel, int(man.OfAmount.IntPart())); err != nil {
						glog.Error("TelCheck fail! err=", err)
						reason = common.ERR_TEL_NOT_SUPPORT_RECHARGE
						return
					}
					vs[i].AutoDeliver = true // 虚拟商品自动发货
				case common.GOODS_TYPE_CARD:
					if key, ok := m.Extinfo["of_key"].(string); ok {
						man.OfKey = key // 将商品扩展字段的of_key赋值给订单maninfos
					}
					vs[i].AutoDeliver = true // 虚拟商品自动发货
				case common.GOODS_TYPE_YOUBUY:
					if man.Voucher != nil {
						man.Voucher.Amount *= float64(vs[i].Quantity)
						vs[i].AutoDeliver = true

						// 检测优买会帐号是否存在
						var y *util.ThirdAccount
						y, err = util.GetYoubuyAccount(vs[i].UserId)
						if err != nil || y == nil {
							reason = err.Error()
							return
						}
						man.Voucher.ThirdId = y.ThirdId
					}
				}

				vs[i].Product["price"] = m.Price // 更新商品规格里的kt食品到商品字段中
				// 兼容积分抽奖
				if p.PublishArea == common.PUBLISH_AREA_LOTTERYS {
					if v.ProductSkuId > 0 {
						for i, s := range p.Skus {
							if s.Id == v.ProductSkuId {
								s.PayPrice = vs[i].ExtInfos["pay_price"].([]common.PayPrice)
								vs[i].Product["skus"] = []common.ProductSku{s}
								break
							}
						}
					} else {
						vs[i].Product["pay_price"] = vs[i].ExtInfos["pay_price"].([]common.PayPrice)
						vs[i].Product["skus"] = []common.ProductSku{}
					}
				} else {
					// 多币种支付时 处理用户支付的币种方式
					if m.PayPrice != nil && len(m.PayPrice) > 1 {
						var paySt *common.PayPrice
						for i := range m.PayPrice {
							if m.PayPrice[i].PayType == v.CurrencyType {
								paySt = &m.PayPrice[i]
							}
						}
						//coin := GetCoinByCurrencyType(v.CurrencyType)
						//paySt, ok := m.PayPrice[strings.ToUpper(coin)]
						if paySt == nil {
							glog.Error("GetCoinByCurrencyType coin empty! txType:", v.CurrencyType)
							err = fmt.Errorf(yf.ERR_SERVER_ERROR)
							return
						}
						//vs[i].Product["price"] = paySt.SalePrice // 更新商品规格从到商品字段中
						total_amount := paySt.SalePrice.Mul(decimal.New(int64(vs[i].Quantity), 0))
						vs[i].TotalAmount, _ = total_amount.Float64()
						if p.PublishArea == common.PUBLISH_AREA_KT { // 只有kt专区的订单才参与挖矿
							vs[i].RebatAmount, _ = p.Rebat.Mul(total_amount).Float64()
						}
					} else { // 默认处理
						total_amount := m.Price.Mul(decimal.New(int64(vs[i].Quantity), 0))
						vs[i].TotalAmount, _ = total_amount.Float64()
						if p.PublishArea == common.PUBLISH_AREA_KT { // 只有kt专区的订单才参与挖矿
							vs[i].RebatAmount, _ = p.Rebat.Mul(total_amount).Float64()
						}
					}
				}

				if err = db.Model(&vs[i]).Updates(map[string]interface{}{"seller_userid": vs[i].SellerUserId, "auto_deliver": vs[i].AutoDeliver, "maninfos": base.StructToMap(man), "total_amount": vs[i].TotalAmount, "rebat_amount": vs[i].RebatAmount, "product": vs[i].Product, "product_type": vs[i].ProductType, "publish_area": vs[i].PublishArea, "update_time": time.Now().Unix()}).Error; err != nil {
					glog.Error("SnapOrdersProducts fail! err=", err)
					return
				}

			}
		}
	}
	return
}

func Filter_Obj(v interface{}, tag string) (obj interface{}) {
	obj = retag.Convert(v, retag.NewView("json", tag))
	return
}
