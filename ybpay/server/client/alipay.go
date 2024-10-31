package client

import (
	"yunbay/ybpay/common"
	"yunbay/ybpay/conf"
	"yunbay/ybpay/server/share"
	"yunbay/ybpay/util"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/smartwalle/alipay"
)

func init() {
}

func Alipay_pay(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req share.OrderSt
	if ok := util.UnmarshalBodyAndCheck(c, &req); !ok {
		glog.Error("Alipay_pay fail! args invalid!", req)
		return
	}
	req.UserId = user_id
	req.Channel = common.CHANNEL_ALIPAY
	req.RemoteIp = c.ClientIP()

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

// 异步通知付款成功或失败
func Alipay_Notify(c *gin.Context) {
	var noti *alipay.TradeNotification
	if conf.Config.Server.Test {
		var req alipay.TradeNotification
		if ok := util.UnmarshalBodyAndCheck(c, &req); !ok {
			return
		}
		noti = &req
	} else {
		var err error
		if noti, err = share.GetAliPay().GetTradeNotification(c.Request); err != nil {
			glog.Error("Alipay_Notify fail! err=", err)
			c.String(200, "error")
			return
		}
	}

	if noti != nil { // 传进的签名等于计算出的签名，说明请求合法
		// 判断订单是否已完成
		//if noti.TradeStatus == "TRADE_FINISHED" || noti.TradeStatus == "TRADE_SUCCESS" { //交易成功
		status := share.GetTradeStatus(noti.TradeStatus)
		v := share.TxParms{Status: status, TradeId: noti.OutTradeNo, TxHash: noti.TradeNo, Account: noti.BuyerLogonId, Amount: noti.TotalAmount}

		err := v.UpdateRmbRecharge()
		if err != nil {
			glog.Error("Alipay_Notify fail! err=", err)
			c.String(200, "error")
			return
		}
		if status == common.STATUS_OK {
			c.Set("resp_tx", true)
			c.String(200, "success")
			return
		} else {
			//contro.Ctx.WriteString("error")
			c.String(200, "error")
			return
		}
	} else {
		c.String(200, "error")
		return
	}
	return
}
