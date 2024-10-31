package util

import (
	"github.com/jie123108/glog"
	"github.com/shopspring/decimal"
)

type Payorders struct {
	ProductId    int64                  `json:"product_id"`
	ProductSkuId int64                  `json:"product_sku_id"`
	UserId       int64                  `json:"user_id"`
	AddressId    int64                  `json:"address_id"`
	PayType      int                    `json:"pay_type" binding:"min=0,max=3"`
	Amount       decimal.Decimal        `json:"amount"`
	Quantity     int                    `json:"quantity"`
	Extinfos     map[string]interface{} `json:"extinfos"`
	RewardYbt    decimal.Decimal        `json:"reward_ybt"`
	Price        decimal.Decimal        `json:"price"`
}

func (t *Payorders) Do() (order_id int64, err error) {
	uri := "/man/order/pay/lotterys"
	err = post_info(uri, "yborder", nil, t, "id", &order_id, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("Payorders fail! err=", err)
		return
	}
	return
}
