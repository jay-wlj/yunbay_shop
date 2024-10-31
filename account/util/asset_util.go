package util

import (
	"github.com/jie123108/glog"
	"yunbay/account/common"
)

type drawSt struct {
	TxType int `json:"tx_type" binding:"required"`
	ToUserId *int64 `json:"to_user_id" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Comment string `json:"comment"`
}

func RechargeNotify(id int64) (err error) {	
	uri := "/man/rmb/recharge/notify"	
	headers := make(map[string]string)
	headers["x-yf-maner"] = "system"
	headers["x-yf-country"] = "1"		// 1为"china"国内版
	v := idSt{id}

	err = post_info(uri, "ybasset", headers, v, "", nil, false, EXPIRE_RES_INFO)
	if  err != nil {	
		glog.Error("RechargeNotify fail! err=", err)
		return
	}
	return
}

type idSt struct {
	RechargeId int64 `json:"recharge_id"`
}

func AsyncRechargeNotify(id int64) (err error) {	
	uri := "/man/rmb/recharge/notify"	
	headers := make(map[string]string)
	headers["x-yf-maner"] = "system"
	headers["x-yf-country"] = "1"		// 1为"china"国内版
	v := idSt{id}
	
	m := common.MQUrl{Methond:"post", AppKey:"ybasset", Uri:uri, Headers:headers, Data:v, MaxTrys:-1, Delay:"100ms"}
	err = PublishMsg(m)
	if  err != nil {	
		glog.Error("AsyncTransferKt fail! err=", err)
		return
	}
	return
}