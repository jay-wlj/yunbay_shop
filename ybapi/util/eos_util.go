package util

import (
	"fmt"
	"yunbay/ybapi/common"

	"github.com/jie123108/glog"
)

type transactionSt struct {
	OrderId int64  `json:"order_id"`
	Memo    string `json:"memo"`
}

type tranSt struct {
	common.MQUrl `json:"mqurl"`
	Memo         string `json:"memo"`
	Id           int64  `json:"id"`
}

// // 提交eos订单交易
// func GenerateEosOrderRebat(order_id int64) (err error) {
// 	uri := "/man/transaction/push"
// 	v := transactionSt{OrderId: order_id, Memo: fmt.Sprintf("%v", order_id)}
// 	err = post_info(uri, "ybeos", nil, v, "", nil, false, EXPIRE_RES_INFO)
// 	if err != nil {
// 		glog.Error("GenerateEosOrderRebat fail! err=", err)
// 		return
// 	}
// 	return
// }

// 异步提交eos订单交易
func AsynGenerateEosOrderRebat(order_id int64) (err error) {
	//v := transactionSt{OrderId: order_id, Memo: fmt.Sprintf("%v", order_id)}
	v := tranSt{Id: order_id, Memo: fmt.Sprintf("%v", order_id), MQUrl: common.MQUrl{AppKey: "yborder", Methond: "post", Uri: "/man/order/rebat/update"}}
	mq := common.MQUrl{Methond: "post", Uri: "/man/transaction/push", AppKey: "ybeos", Data: v, MaxTrys: -1}
	if err = PublishMsg(mq); err != nil {
		glog.Error("AsynGenerateEosOrderRebat fail! err=", err)
		return
	}
	
	return
}

// 异步提交eos订单交易
func AsynGenerateEosLotterysHash(order_id int64, memo string) (err error) {
	v := tranSt{Id: order_id, Memo: memo, MQUrl: common.MQUrl{AppKey: "ybapi", Methond: "get", Uri: "/man/lotterys/record/hash"}}
	mq := common.MQUrl{Methond: "post", Uri: "/man/transaction/push", AppKey: "ybeos", Data: v, MaxTrys: -1}
	if err = PublishMsg(mq); err != nil {
		glog.Error("AsynGenerateEosOrderRebat fail! err=", err)
		return
	}
	return
}
