package util

import (
	"fmt"
	"strings"
	"yunbay/ybapi/common"

	base "github.com/jay-wlj/gobaselib"

	"github.com/shopspring/decimal"

	"github.com/jie123108/glog"
)

func GetProductInfo(product_id, product_sku_id int64) (v common.Product, err error) {
	uri := fmt.Sprintf("/man/info?id=%v&sku_id=%v", product_id, product_sku_id)
	err = get_info(uri, "ybgoods", "", &v, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("GetProductInfo fail! err=", err)
		return
	}
	return
}

type ProdIds struct {
	ProductId      int64
	ProductModelId int64
}

func (t ProdIds) String() string {
	return fmt.Sprintf("%v-%v", t.ProductId, t.ProductModelId)
}
func ListProductInfo(vs []ProdIds) (m map[int64]*common.Product, err error) {
	m = make(map[int64]*common.Product)
	//ret := make(map[string]common.Product)

	var str string
	for _, v := range vs {
		str += v.String()
		str += ","
	}
	str = strings.TrimRight(str, ",")

	uri := fmt.Sprintf("/man/list-detail?list_id_model=%v", str)
	glog.Info(uri)
	ps := []common.Product{}
	err = get_info(uri, "ybgoods", "list", &ps, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("ListProductInfoByIds fail! err=", err)
		return
	}
	// by, _ := json.Marshal(ret)
	// glog.Info("product:", string(by))

	for i, v := range ps {
		m[v.Id] = &ps[i]
	}
	return
}

func ListProductInfoByIds(product_ids []int64) (ms map[int64]common.Product, err error) {
	vs := []common.Product{}
	uri := fmt.Sprintf("/man/list_by_ids?ids=%v", base.Int64SliceToString(product_ids, ","))
	err = get_info(uri, "ybgoods", "", &vs, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("ListProductInfoByIds fail! err=", err)
		return
	}

	ms = make(map[int64]common.Product)
	for _, v := range vs {
		ms[v.Id] = v
	}
	return
}

type ProductPrice struct {
	PId       int64           `json:"p_id"`
	PSkuId    int64           `json:"p_sku_id"`
	CostPrice decimal.Decimal `json:"cost_price"`
	Price     decimal.Decimal `json:"price"`
}

// 获取商品原价
func ListProductPriceByIds(product_ids []int64) (ms map[int64]map[int64]*ProductPrice, err error) {
	vs := []ProductPrice{}
	uri := fmt.Sprintf("/man/price/list_by_ids?ids=%v", base.Int64SliceToString(product_ids, ","))
	err = get_info(uri, "ybgoods", "list", &vs, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("ListProductPriceByIds fail! err=", err)
		return
	}

	ms = make(map[int64]map[int64]*ProductPrice)
	for i, v := range vs {
		m, ok := ms[v.PId]
		if !ok {
			m = make(map[int64]*ProductPrice)
		}
		m[v.PSkuId] = &vs[i]
		ms[v.PId] = m
	}
	return
}

type QuantitySt struct {
	OrderId      int64 `json:"order_id"`
	ProductId    int64 `json:"product_id"`
	ProductSkuId int64 `json:"product_sku_id"`
	Quantity     int   `json:"quantity"`
}
type CallBackSt struct {
	AppKey string      `json:"app_key"`
	Uri    string      `json:"uri"`
	Method string      `json:"method"`
	Body   interface{} `json:"body"`
}

// 修改相应产品规格库存
func AddProductModelQuantity(vs []QuantitySt, sync ...bool) (err error) {
	if len(vs) == 0 {
		return
	}

	if sync != nil && len(sync) > 0 && sync[0] {
		// 同步调用
		err = post_info("/man/quantity/add", "ybgoods", nil, vs, "", nil, false, EXPIRE_RES_INFO)
		if err != nil {
			glog.Error("AddProductModelQuantity fail! err=", err)
			return
		}
		return
	}

	// 异步调用
	err = PublishMsg(common.MQUrl{Methond: "POST", AppKey: "ybgoods", Uri: "/man/quantity/add", Data: vs, MaxTrys: -1})
	if err != nil {
		glog.Error("AsyncAddProductModelQuantity fail! err=", err)
		return
	}

	return
}
