package util

import (
	"github.com/jie123108/glog"
	"fmt"
	//base "github.com/jay-wlj/gobaselib"
	"github.com/smartwalle/alipay"
	//"yunbay/ybasset/common"
)

type WithDrawRetSt struct {
	OrderId int64 `json:"order_id,string"`
	TxHash string `json:"tx_hash"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	FeeInEther float64 `json:"fee_in_ether,string"`
	Channel int `json:"channel"`
	TxType string `json:"tx_type"`
	BlockTime string `json:"block_time"`
}


func Ali_Withdraw(id int64, amount float64, account string) (ret WithDrawRetSt, err error) {
	uri := "/man/alipay/trade/trasfer"
	strAmount := fmt.Sprintf("%.2f", amount)
	v := alipay.AliPayFundTransToAccountTransfer{OutBizNo:fmt.Sprintf("%v", id), Amount:strAmount, PayeeType:"ALIPAY_LOGONID", PayeeAccount:account}	
	if err = post_info(uri, "ybpay", nil, v, "", &ret, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("WithDrawAli fail! err=", err)
		return
	}
	return
}

// 关闭超时的订单
func CloseOverPayOrder() (err error) {
	uri := "/man/trade/close"		
	if err = post_info(uri, "ybpay", nil, nil, "", nil, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("CloseOverPayOrder fail! err=", err)
		return
	}
	return
}