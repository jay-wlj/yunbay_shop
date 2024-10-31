package client

import (
	"io/ioutil"
	"yunbay/ybpay/common"
	"yunbay/ybpay/conf"
	"yunbay/ybpay/server/share"
	"yunbay/ybpay/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	wx "github.com/jay-wlj/wxpay"
	"github.com/jie123108/glog"
)

func WeixinPay(c *gin.Context) {
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
	req.Channel = common.CHANNEL_WEIXIN
	req.RemoteIp = c.ClientIP()

	db := db.GetTxDB(c)
	prepayid, reason, err := req.CreatePay(db)
	if err != nil {
		if reason == "" {
			reason = yf.ERR_SERVER_ERROR
		}

		yf.JSON_Fail(c, reason)
		return
	}

	params, _ := share.GetWeixin().SignApp(prepayid)
	params.SetInt64("pay_id", req.Id)
	//yf.JSON_Ok(c, gin.H{"channel": common.CHANNEL_WEIXIN, "prepayid": sign, "pay_id": req.Id, "appid": cfg.Appid, "partnerid": cfg.Mchid, "timestamp": time.Now().Unix(), "noncestr": strconv.FormatInt(time.Now().UTC().UnixNano(), 10)})
	yf.JSON_Ok(c, params)
}

func WeixinNotify(c *gin.Context) {
	var noti wx.Params
	if conf.Config.Server.Test {
		noti = make(wx.Params)
		buf, _ := ioutil.ReadAll(c.Request.Body)
		noti = wx.XmlToMap(string(buf))
	} else {
		var err error
		if noti, err = share.GetWeixin().NotifyUrl(c.Request); err != nil {
			glog.Error("Alipay_Notify fail! err=", err)
			c.String(200, "error")
			return
		}
	}

	ret := wx.Notifies{} //make(wx.Params)

	if noti.IsSuccess() {
		// 判断订单是否已完成

		status := common.STATUS_OK

		v := share.TxParms{Channel: common.CHANNEL_WEIXIN, Status: status, TradeId: noti["out_trade_no"], TxHash: noti["transaction_id"], Account: noti["openid"], Amount: noti.GetString("total_fee")}

		if err := v.UpdateRmbRecharge(); err != nil {
			c.Data(200, "application/xml", []byte(ret.NotOK("again")))
			glog.Error("WeixinNotify fail! err=", err, " ret=", ret)
			return
		}
		c.Data(200, "application/xml", []byte(ret.OK()))
	} else {
		c.Data(200, "application/xml", []byte(ret.NotOK("again")))
		return
	}
	return
}

func WeixinRefundNotify(c *gin.Context) {
	var noti wx.Params
	if conf.Config.Server.Test {
		noti = make(wx.Params)
		buf, _ := ioutil.ReadAll(c.Request.Body)
		noti = wx.XmlToMap(string(buf))
	} else {
		var err error
		if noti, err = share.GetWeixin().RefundNotifyUrl(c.Request); err != nil {
			glog.Error("Alipay_Notify fail! err=", err)
			c.String(200, "error")
			return
		}
	}

	ret := wx.Notifies{} //make(wx.Params)
	//ret.SetString("return_code", wx.Fail)
	//ret.SetString("return_msg", "OK")

	if noti.IsSuccess() {
		status := common.STATUS_OK

		if noti.GetString("refund_status") != wx.Success {
			status = common.STATUS_FAIL
		}
		total_fee := noti.GetInt64("total_fee") / 100
		res := db.GetDB().Model(&common.RmbRefund{}).Where("id=? and total_fee=? and status=?", noti.GetInt64("out_refund_no"), total_fee,
			common.STATUS_INIT).Updates(base.Maps{"refund_fee": noti.GetInt64("refund_fee") / 100, "status": common.STATUS_OK, "refund_account": noti.GetString("refund_account"), "refund_recv_account": noti.GetString("refund_recv_account")})

		if err := res.Error; err != nil {
			//c.XML(200, ret.NotOK("again"))
			c.Data(200, "application/xml", []byte(ret.NotOK("again")))
			glog.Error("WeixinNotify fail! err=", err, " ret=", ret)
			return
		}

		if res.RowsAffected > 0 && status == common.STATUS_OK {
			// TODO 退款成功，更新订单退款信息
		}
		c.Data(200, "application/xml", []byte(ret.OK()))
	} else {
		c.Data(200, "application/xml", []byte(ret.OK()))
		return
	}
	return
}
