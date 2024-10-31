package share

import (
	"encoding/json"
	"fmt"
	"strings"
	"yunbay/ybgoods/common"
	"yunbay/ybgoods/conf"
	"yunbay/ybgoods/dao"
	"yunbay/ybgoods/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

func GetThirdId(c *gin.Context) (third_id int, ok bool) {
	ok = true
	third_id = util.GetThirdId(c)
	if third_id == 0 {
		ok = false
		glog.Error("third_id not exist! ")
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	return
}

func IsExstCoinType(third_id int, coin_type string) bool {
	m, ok := conf.Config.Coins[third_id]
	if !ok {
		return false
	}
	for _, t := range m {
		if t == coin_type {
			return true
		}
	}
	return false
}

func GetAttrName(id int64) (name string, err error) {
	var v common.ProductAttrKey
	if err = db.GetDB().First(&v, "id=?", id).Error; err != nil {
		glog.Error("GetAttrName fail! err=", err)
		return
	}
	name = v.Name
	return
}

func GetAttrValues(id ...int64) (vs []common.ProductAttrValue, err error) {
	vs = []common.ProductAttrValue{}
	if err = db.GetDB().Find(&vs, "id in(?)", id).Error; err != nil {
		glog.Error("GetAttrName fail! err=", err)
		return
	}
	return
}

// func getUnitPrice(_type int) (ratio decimal.Decimal, err error) {
// 	ch, e := cache.GetReader("pub")
// 	if e != nil {
// 		err = e
// 		glog.Error("getUnitPrice fail! err=", err)
// 		return
// 	}
// 	var val string

// 	field := common.GetCurrencyName(_type) + "kt"
// 	if val, err = ch.HGet("currency_ratio", field); err == nil {
// 		ratio, err = decimal.NewFromString(val)
// 	}

// 	if err != nil || ratio.IsZero() {
// 		// 从接口里获取

// 	}

// 	val, err = ch.Get("ybp" + key)
// 	if err != nil || val == "" {
// 		var v common.Setting
// 		if err = db.GetDB().Find(&v, "setting_key=?", key).Error; err != nil {
// 			glog.Error("getUnitPrice fail! err=", err)
// 			return
// 		}
// 		val = v.SettingValue
// 		ch.Set("ybp"+key, val, 0)
// 	}
// 	rt, err = decimal.NewFromString(val)
// 	return
// }

func getdifficult() (dif float64, err error) {
	ch, e := cache.GetReader("pub")
	if e != nil {
		err = e
		glog.Error("getdifficult fail! err=", err)
		return
	}
	dif, err = ch.HGetF64("diffcult", base.GetCurDay())
	if err != nil {
		dif, err = util.GetDifficult()
	}
	return
}

func getVoucherPrice(_type int, price, rebat decimal.Decimal) (ratio decimal.Decimal, err error) {
	if ratio, err = getCurrencyRatio(_type, common.CURRENCY_RMB); err != nil {
		glog.Error("getVoucherPrice fail! err=", err)
		return
	}
	ratio = price.Mul(ratio).Mul(rebat.Add(decimal.New(1, 0))).Round(2)
	return
}

func getCurrencyRatio(from, to int, virtual ...int) (val decimal.Decimal, err error) {
	ch, e := cache.GetReader("pub")
	if e != nil {
		err = e
		glog.Error("getdifficult fail! err=", err)
		return
	}
	var dif string
	var virtual_type int
	filed := common.GetCurrencyName(from) + common.GetCurrencyName(to)
	if virtual != nil && len(virtual) > 0 && virtual[0] > 0 {
		virtual_type = virtual[0]
		filed += fmt.Sprintf("_%v", virtual_type)
	}
	if dif, err = ch.HGet("currency_ratio", filed); err == nil {
		val, err = decimal.NewFromString(dif)
	}
	if err != nil {
		val, err = util.GetCurrencyRatio(from, to, virtual_type)
	}
	return
}

type Sku struct {
	AttrId  int64 `json:"attr_id"`
	ValueId int64 `json:"value_id"`
}

// 动态调整商品价格
func price_adjust(v *common.Product) (err error) {
	if common.PUBLISH_AREA_KT != v.PublishArea && common.PUBLISH_AREA_REBAT != v.PublishArea {
		return
	}

	if len(v.Skus) == 0 {
		switch v.PublishArea {
		case common.PUBLISH_AREA_KT:
			var man common.ManInfos
			man.ParseJsonb(v.Extinfo)
			if man.Voucher != nil {
				// 处理代金券的问题
				if v.Price, err = getVoucherPrice(man.Voucher.Type, man.Voucher.Amount, v.Rebat); err != nil {
					glog.Error("FormatProduct fail! getVoucherPrice err=", err)
					return
				}
			}
		case common.PUBLISH_AREA_REBAT:
			v.Price = v.Price.Mul(decimal.NewFromFloat32(1 + 0.2)).Truncate(4) // 折扣商品价格提高20% 以降低低折扣风险
		}
		return
	}
	for i := range v.Skus {
		var man common.ManInfos
		man.Parse(v.Skus[i])
		if man.Voucher != nil {
			// 处理代金券的问题
			if v.Skus[i].Price, err = getVoucherPrice(man.Voucher.Type, man.Voucher.Amount, v.Rebat); err != nil {
				glog.Error("FormatProduct fail! getVoucherPrice err=", err)
				return
			}
			// // 更新common.PayPrice里的价格
			// if len(v.Skus[i].PayPrice) > 0 {
			// 	v.Skus[i].PayPrice[0].Price = v.Skus[i].Price
			// 	v.Skus[i].PayPrice[0].PredictYbt, _ = v.Skus[i].Price.Mul(v.Rebat).Mul(decimal.NewFromFloat(0.7)).Div(decimal.NewFromFloat(diffcult).Truncate(4)).Float64()
			// }
		}

		if v.PublishArea == common.PUBLISH_AREA_REBAT {
			v.Skus[i].Price = v.Skus[i].Price.Mul(decimal.NewFromFloat32(1 + 0.2)).Truncate(4) // 折扣商品价格提高20% 以降低低折扣风险
		}
	}

	return
}
func FormatProduct(v *common.Product, man bool) (err error) {
	db := db.GetDB()
	// 获取关联的商品规格
	err = db.Model(v).Select("id,sku,stock,sold,price,img,combines,extinfo").Related(&v.Skus).Error
	if err != nil {
		glog.Error("GoodsInfo fail! err=", err)
		return
	}

	if v.PublishArea == common.PUBLISH_AREA_KT {
		v.Canreturn = true
	}

	// 兑换专区
	var ybt_unit decimal.Decimal
	var snet_unit decimal.Decimal

	var diffcult float64
	if ybt_unit, err = getCurrencyRatio(common.CURRENCY_YBT, common.CURRENCY_KT); err != nil {
		return
	}

	if snet_unit, err = getCurrencyRatio(common.CURRENCY_SNET, common.CURRENCY_KT, int(v.Type)); err != nil {
		return
	}

	if diffcult, err = getdifficult(); err != nil {
		glog.Error("getdifficult fail! err=", err)
	}

	if err = price_adjust(v); err != nil {
		glog.Error("FormatProduct fail! voucher_handle err=", err)
		return
	}

	if len(v.Skus) == 0 {
		if v.PredictYbt, v.PayPrice, err = get_payprice(v.PublishArea, v.Price, v.Rebat, ybt_unit, snet_unit, v.Discount, diffcult); err != nil {
			glog.Error("FormatProduct fail! get_payprice err=", err)
			return
		}
		v.Attrs = []common.Attr{}
		return
	}

	// 处理sku里的common.PayPrice字段
	for i, f := range v.Skus {
		if _, v.Skus[i].PayPrice, err = get_payprice(v.PublishArea, f.Price, v.Rebat, ybt_unit, snet_unit, v.Discount, diffcult); err != nil {
			glog.Error("FormatProduct fail! get_payprice err=", err)
			return
		}
	}

	if man {
		v.Attrs = []common.Attr{}
		return
	}

	// 处理商品sku问题
	mskus := make(map[int64][]int64)
	attr_key_ids := []int64{} // 属性key排序
	var once bool = true
	for i, f := range v.Skus {
		as := []Sku{}
		if err = json.Unmarshal(f.Sku.RawMessage, &as); err != nil {
			glog.Error("GoodsInfo sku fail! err=", err)
			return
		}
		for _, a := range as {
			if once {
				attr_key_ids = append(attr_key_ids, a.AttrId)
			}

			mskus[a.AttrId] = append(mskus[a.AttrId], a.ValueId)
		}
		once = false

		if f.Img == "" {
			v.Skus[i].Img = v.Images[0]
		}
	}
	// 去重属性id
	for k, v := range mskus {
		mskus[k] = base.UniqueInt64Slice(v)
	}
	attrs := []common.Attr{}
	for _, k := range attr_key_ids {
		//for k, v := range mskus {
		v, ok := mskus[k]
		if !ok {
			glog.Error("FormatProduct no attr_key_ids vals k=", k)
			continue
		}
		var at common.ProductAttrKey
		if at, err = dao.GetGoodsAttrInfo(k); err != nil {
			glog.Error("GetAttrName sku fail! err=", err)
			return
		}

		ar := common.Attr{Id: k, Name: at.Name}
		if len(v) == len(at.Values) {
			ar.Values = at.Values
		} else {
			// 选择拥有的属性值
			mvalues := make(map[int64]*common.ProductAttrValue)
			for i, val := range at.Values {
				mvalues[val.Id] = &at.Values[i]
			}

			for _, f := range v {
				if _, ok := mvalues[f]; !ok {
					glog.Error("FormatProduct no values value_id=", f, " values_ids=", mvalues)
					continue
				}
				ar.Values = append(ar.Values, *mvalues[f])
			}
		}
		attrs = append(attrs, ar)
	}
	v.Attrs = attrs // 获取商品规格属性
	return
}

// 格式化商品列表信息
func FormatProductList(vs []common.Product) (err error) {
	var ybt_unit decimal.Decimal
	var snet_unit decimal.Decimal
	var snet_unit_virtual decimal.Decimal

	var diffcult float64
	if ybt_unit, err = getCurrencyRatio(common.CURRENCY_YBT, common.CURRENCY_KT); err != nil {
		glog.Error("FormatProductList fail! getUnitPrice err=", err)
		return
	}
	if snet_unit, err = getCurrencyRatio(common.CURRENCY_SNET, common.CURRENCY_KT); err != nil {
		glog.Error("FormatProductList fail! getUnitPrice err=", err)
		return
	}
	if snet_unit_virtual, err = getCurrencyRatio(common.CURRENCY_SNET, common.CURRENCY_KT, 1); err != nil {
		return
	}
	if diffcult, err = getdifficult(); err != nil {
		glog.Error("getdifficult fail! err=", err)
	}

	for i := range vs {
		// 是否为代金券
		price_adjust(&vs[i])
		v := &vs[i]
		// 兑换专区商品 需要将其它币种价格计算出
		snet_ratio := snet_unit
		if v.Type > 0 {
			snet_ratio = snet_unit_virtual
		}
		if v.PredictYbt, v.PayPrice, err = get_payprice(v.PublishArea, v.Price, v.Rebat, ybt_unit, snet_ratio, v.Discount, diffcult); err != nil {
			glog.Error("FormatProduct fail! get_payprice err=", err)
			return
		}

		if v.PublishArea == common.PUBLISH_AREA_KT {
			v.PayPrice = nil // 商品列表不用给出pay_price字段
		}
	}
	return
}

func get_payprice(publish_area int, price, rebat, ybt_unit, snet_unit, discount decimal.Decimal, diffcult float64) (preict_ybt float64, ps []common.PayPrice, err error) {
	ps = []common.PayPrice{}
	switch publish_area {
	case common.PUBLISH_AREA_KT:
		// 只有kt支付才会释放ybt
		ybt := price.Mul(rebat).Mul(decimal.NewFromFloat(0.7)).Div(decimal.NewFromFloat(diffcult)).Truncate(4)
		preict_ybt, _ = ybt.Float64()
		ps = append(ps, common.PayPrice{Coin: "KT", PayType: common.CURRENCY_KT, OriginPrice: price, Price: price.Mul(discount), UnitPrice: decimal.New(1, 0), PredictYbt: preict_ybt})
	case common.PUBLISH_AREA_YBT, common.PUBLISH_AREA_LOTTERYS:
		ybt := common.PayPrice{Coin: "YBT", PayType: common.CURRENCY_YBT, OriginPrice: price.Div(ybt_unit).Truncate(4), UnitPrice: ybt_unit}
		ybt.Price = ybt.OriginPrice.Mul(discount)
		snet := common.PayPrice{Coin: "SNET", PayType: common.CURRENCY_SNET, OriginPrice: price.Div(snet_unit).Truncate(4), UnitPrice: snet_unit}

		snet.Price = snet.OriginPrice.Mul(discount)

		ps = append(ps, ybt)
		ps = append(ps, snet)
	case common.PUBLISH_AREA_REBAT:
		predict_ybt, _ := price.Mul(rebat).Mul(decimal.NewFromFloat(0.7)).Div(decimal.NewFromFloat(diffcult)).Truncate(4).Float64()
		ps = append(ps, common.PayPrice{Coin: "KT", PayType: common.CURRENCY_KT, OriginPrice: price, Price: price.Mul(discount), UnitPrice: decimal.New(1, 0), PredictYbt: predict_ybt})
		ybt := common.PayPrice{Coin: "YBT", PayType: common.CURRENCY_YBT, OriginPrice: price.Div(ybt_unit).Truncate(4), UnitPrice: ybt_unit}
		ybt.Price = ybt.OriginPrice.Mul(discount)
		snet := common.PayPrice{Coin: "SNET", PayType: common.CURRENCY_SNET, OriginPrice: price.Div(snet_unit).Truncate(4), UnitPrice: snet_unit}
		snet.Price = snet.OriginPrice.Mul(discount)

		ps = append(ps, ybt)
		ps = append(ps, snet)

		// 微调价格 计算最低折扣
		for i := range ps {
			//ps[i].Price = v.Price.Mul(decimal.NewFromFloat32(1 + 0.2)).Truncate(4)                              // 控制精度在小数点后4拉
			ps[i].LowestDiscountPrice, _ = ps[i].Price.Mul(decimal.NewFromFloat(0.1111)).Truncate(4).Float64()  // 最低折扣为11.11%
			ps[i].HighestDiscountPrice, _ = ps[i].Price.Mul(decimal.NewFromFloat(0.9999)).Truncate(4).Float64() // 最高折扣为99.99%
		}
	}

	return
}

func GetCategorys(vs []common.ManProduct) (err error) {

	// 获取所有二级分类id
	ids := []int64{}
	for _, v := range vs {
		ids = append(ids, v.CategoryId)
	}
	db := db.GetDB()
	if len(ids) > 0 {
		ids = base.UniqueInt64Slice(ids)
		// 获取二级分类的id及一级分类id标题
		cs2 := []common.ProductCategory{}
		if err = db.Select("id,title,parent_id").Find(&cs2, "id in(?)", ids).Error; err != nil {
			glog.Error("getcategorys faiL! err=", err)
			return
		}
		// 获取一级分类的标题
		ids = ids[0:0]
		for _, v := range cs2 {
			ids = append(ids, v.ParentId)
		}
		ids = base.UniqueInt64Slice(ids)
		cs1 := []common.ProductCategory{}
		if err = db.Select("id,title,parent_id").Find(&cs1, "id in(?)", ids).Error; err != nil {
			glog.Error("getcategorys faiL! err=", err)
			return
		}
		ms2 := make(map[int64]*common.ProductCategory)
		for i, v := range cs2 {
			ms2[v.Id] = &cs2[i]
		}

		ms1 := make(map[int64]*common.ProductCategory)
		for i, v := range cs1 {
			ms1[v.Id] = &cs1[i]
		}

		// 更新到商品列表中
		ids = ids[0:0]
		for i, v := range vs {
			if m, ok := ms2[v.CategoryId]; ok {
				if m1, ok := ms1[m.ParentId]; ok {
					vs[i].Categories = append(vs[i].Categories, m1)
				}
				vs[i].Categories = append(vs[i].Categories, m)
			}
		}
	}

	// 更新到商品列表中
	ids = ids[0:0]
	for _, v := range vs {

		if v.DefSkuId > 0 {
			ids = append(ids, v.DefSkuId)
		}
	}

	// 批量获取默认规格id
	if len(ids) > 0 {
		ss := []common.ProductSku{}
		if err = db.Select("id,combines").Find(&ss, "id in(?)", ids).Error; err != nil {
			glog.Error("getcategorys faiL! err=", err)
			return
		}
		ms := make(map[int64]*common.ProductSku)
		for i, v := range ss {
			ms[v.Id] = &ss[i]
		}
		for i, v := range vs {
			if m, ok := ms[v.DefSkuId]; ok {
				var title string
				vcs := []map[string]string{}
				buf, _ := json.Marshal(m.Combines)
				if err = json.Unmarshal(buf, &vcs); err == nil {
					for _, v := range vcs {
						for k, t := range v {
							title += fmt.Sprintf("%v:%v ", k, t)
						}
					}
				}

				vs[i].DefSku = strings.TrimRight(title, " ")
				//vs[i].Product.Skus = append(vs[i].Product.Skus, m)
			}
		}
	}

	return
}

// 获取原价 TODO 需要重构
func GetCostPrice(vs []common.ManProduct) (err error) {

	ids := []int64{}
	for _, v := range vs {
		ids = append(ids, v.Product.Id)
	}

	ps := []common.ProductPrice{}
	if err = db.GetDB().Select("p_id, p_sku_id, cost_price, price").Find(&ps, "p_id in(?)", ids).Error; err != nil {
		glog.Error("GetCostPrice fail! err=", err)
		return
	}

	ms := make(map[int64]map[int64]*common.ProductPrice)
	for i, v := range ps {
		m, ok := ms[v.PId]
		if !ok {
			m = make(map[int64]*common.ProductPrice)
		}
		m[v.PSkuId] = &ps[i]
		ms[v.PId] = m
	}

	for i, v := range vs {
		if m, ok := ms[v.Id]; ok {
			if mm, ok := m[0]; ok {
				vs[i].CostPrice = mm.CostPrice
			}
			for k, s := range v.Skus {
				if mm, ok := m[s.Id]; ok {
					vs[i].Skus[k].CostPrice = mm.CostPrice
				}
			}
		}
	}

	return
}
