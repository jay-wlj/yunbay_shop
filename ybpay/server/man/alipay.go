package man

import (
	"yunbay/ybpay/common"
	"yunbay/ybpay/server/share"
	"yunbay/ybpay/util"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/smartwalle/alipay"
)

func Alipay_pay(c *gin.Context) {
	var req share.OrderSt
	if ok := util.UnmarshalBodyAndCheck(c, &req); !ok {
		glog.Error("Alipay_pay fail! args invalid!", req)
		return
	}
	req.Channel = common.CHANNEL_ALIPAY
	db := db.GetTxDB(c)
	sign, reason, err := req.CreatePay(db)
	if err != nil {
		if reason == "" {
			reason = yf.ERR_SERVER_ERROR
		}
		yf.JSON_Fail(c, reason)
		return
	}
	yf.JSON_Ok(c, gin.H{"channel": common.CHANNEL_ALIPAY, "pay_sign": sign, "pay_id": req.Id})
}

type WithDrawSt struct {
	OrderId    int64   `json:"order_id,string"`
	TxHash     string  `json:"tx_hash"`
	Status     string  `json:"status"`
	Reason     string  `json:"reason"`
	FeeInEther float64 `json:"fee_in_ether,string"`
	Channel    int     `json:"channel"`
	TxType     string  `json:"tx_type"`
	BlockTime  string  `json:"block_time"`
}

// 支付宝转帐接口
func Alipay_Transfer(c *gin.Context) {
	var req alipay.AliPayFundTransToAccountTransfer
	if ok := util.UnmarshalBodyAndCheck(c, &req); !ok {
		glog.Error("Alipay_pay fail! args invalid!", req)
		return
	}
	ret, err := share.GetAliPay().FundTransToAccountTransfer(req)
	if err != nil {
		glog.Error("Alipay_Transfer fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	status := common.TX_STATUS_FAILED
	if ret.IsSuccess() {
		status = common.TX_STATUS_SUCCESS
	} else {
		// 系统繁忙 需重试
		if ret.Body.SubCode == "SYSTEM_ERROR" {
			status = common.TX_STATUS_CHECKPASS
		}
		glog.Error("Alipay_TransferQuery fail! sub_msg:", ret.Body.SubMsg)

	}

	order_id, _ := base.StringToInt64(req.OutBizNo)
	v := WithDrawSt{OrderId: order_id, TxHash: ret.Body.OrderId, Status: getstatustxtbyint(status), Reason: ret.Body.SubMsg, Channel: common.CHANNEL_ALIPAY, BlockTime: ret.Body.PayDate}
	yf.JSON_Ok(c, v)

	return
}

// 支付宝转帐订单查询
func Alipay_TransferQuery(c *gin.Context) {
	var req alipay.AliPayFundTransOrderQuery
	if ok := util.UnmarshalBodyAndCheck(c, &req); !ok {
		glog.Error("Alipay_pay fail! args invalid!", req)
		return
	}
	ret, err := share.GetAliPay().FundTransOrderQuery(req)
	if err != nil {
		glog.Error("Alipay_TransferQuery fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	status := common.TX_STATUS_FAILED
	if ret.IsSuccess() {
		status = common.TX_STATUS_SUCCESS
	} else {
		// 系统繁忙 需重试
		if ret.Body.SubCode == "SYSTEM_ERROR" {
			status = common.TX_STATUS_CHECKPASS
		}
		glog.Error("Alipay_TransferQuery fail! sub_msg:", ret.Body.SubMsg)

	}

	order_id, _ := base.StringToInt64(req.OutBizNo)
	v := WithDrawSt{OrderId: order_id, TxHash: ret.Body.OrderId, Status: getstatustxtbyint(status), Reason: ret.Body.SubMsg, Channel: common.CHANNEL_ALIPAY}
	yf.JSON_Ok(c, v)
	return
}

// 订单退款
func Alipay_Refund(c *gin.Context) {
	// var req alipay.AliPayTradeRefund
	// if ok := util.UnmarshalBodyAndCheck(c, &req); !ok {
	// 	glog.Error("Alipay_pay fail! args invalid!", req)
	// 	return
	// }
	// ret, err := share.GetAlipay().TradeRefund(req)
	// if err != nil {
	// 	glog.Error("Alipay_TransferQuery fail! err=", err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	// status := common.TX_STATUS_SUCCESS
	// if ret.IsSuccess() {
	// 	status = common.TX_STATUS_FAILED
	// 	glog.Error("Alipay_TransferQuery fail! sub_msg:", ret.Body.SubMsg)
	// }

	// order_id, _ := base.StringToInt64(req.OutBizNo)
	// v := WithDrawSt{OrderId:order_id, TxHash:ret.Body.OrderId, Status:getstatustxtbyint(status), Reason:ret.Body.SubMsg, Channel:common.CHANNEL_ALIPAY}
	// yf.JSON_Ok(c, v)
	return
}
