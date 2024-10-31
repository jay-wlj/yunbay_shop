package util

import (
	"fmt"
	"github.com/jie123108/glog"
)

func YBAsset_GetRatio(to_type int)(ratios map[string]float64, err error) {
	uri := fmt.Sprintf("/v1/currency/ratios?to_type=%v", to_type)
	err = get_info(uri, "ybasset", "ratios", &ratios, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_GetRatio fail! err=", err)
		return
	}
	return
}

type YBAssetPool struct {
	OrderId int64 `json:"order_id" gorm:"column:order_id"`
	CurrencyType int `json:"currency_type" gorm:"column:currency_type"`
	PayerUserId int64 `json:"payer_userid" gorm:"column:payer_userid"`
	PayAmount float64 `json:"pay_amount" gorm:"column:pay_amount"`
	SellerUserId int64 `json:"seller_userid" gorm:"column:seller_userid"`
	SellerAmount float64 `json:"seller_amount" gorm:"column:seller_amount"`
	RebatAmount float64 `json:"rebat_amount" gorm:"column:rebat_amount"`
	Status int `json:"status"`
}

func YBAsset_Pay(vs []YBAssetPool, token string)(err error) {
	uri := "/man/asset/pay"
	header := make(map[string]string)
	header["X-YF-Token"] = token
	err = post_info(uri, "ybasset", header, vs, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_Pay fail! err=", err)
		return
	}
	return
}

type YBAssetStatus struct {
	OrderIds []int64 `json:"order_ids"`
	Status int `json:"status"`
}
func YBAsset_PayStatus(v YBAssetStatus, token string)(err error) {
	uri := "/man/asset/payset"
	header := make(map[string]string)
	if token != "" {
		header["X-YF-Token"] = token
	}
	
	err = post_info(uri, "ybasset", header, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_Pay fail! err=", err)
		return
	}
	return
}

// // 设置订单状态为已完成，资金池中的冻结金额将由平台打款给卖家
// func SetPayStatus(order_ids []int64)(err error){
// 	uri := "/man/asset/payset"
// 	v := orderSt{OrderIds:order_ids, Status:common.ASSET_POOL_FINISH}
// 	if err = post_info(uri, "ybasset", nil, &v, "", nil, false, EXPIRE_RES_INFO); err != nil {
// 		return
// 	}
// 	return
// }

