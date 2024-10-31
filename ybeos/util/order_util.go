package util

import (
	"yunbay/ybeos/common"

	"github.com/jie123108/glog"
)

type orderRebatSt struct {
	OrderId int64   `json:"order_id" binding:"required"`
	Rebat   float64 `json:"rebat" binding:"required,min=0,max=1"`
	TxHash  string  `json:"tx_hash" binding:"required"`
}

// 更新订单折扣
func UpdateOrderRebat(order_id int64, txHash string) (err error) {
	uri := "/man/order/rebat/update"

	v := orderRebatSt{OrderId: order_id, TxHash: txHash}
	if err = post_info(uri, "yborder", nil, v, "", nil, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("UpdateOrderRebat fail! err=", err)
		return
	}
	return
}

// 更新订单折扣
func UpdateOrderNotify(order_id int64, txHash string, notify *common.MQUrl) (err error) {
	if notify == nil {
		return
	}
	v := orderRebatSt{OrderId: order_id, TxHash: txHash}
	if err = post_info(notify.Uri, notify.AppKey, nil, v, "", nil, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("UpdateOrderNotify fail! err=", err)
		return
	}
	return
}
