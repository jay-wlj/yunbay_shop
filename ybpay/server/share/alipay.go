package share

import (
	"fmt"
	"yunbay/ybpay/common"
	"yunbay/ybpay/conf"

	"github.com/jie123108/glog"
	"github.com/smartwalle/alipay"
)

type Alipay struct {
	*alipay.AliPay
}

var g_aliClient *Alipay

func GetAliPay() *Alipay {
	if g_aliClient == nil {
		cfg := conf.Config.Alipay
		g_aliClient = &Alipay{alipay.New(cfg.Appid, cfg.Public, cfg.Private, conf.Config.Server.Debug)}
	}
	return g_aliClient
}

func GetTradeStatus(status string) int {
	switch status {
	case "TRADE_FINISHED", "TRADE_SUCCESS":
		return common.STATUS_OK
	case "TRADE_CLOSED":
		return common.STATUS_FAIL
	}
	return common.STATUS_INIT
}

// 统一订单
func (t *Alipay) TradeAppPay(req *OrderSt) (prepay_id, reason string, err error) {
	cfg := conf.Config.Alipay
	p := alipay.AliPayTradeAppPay{}
	p.Subject = req.Subject
	p.TotalAmount = req.Amount.String()
	//p.TotalAmount = fmt.Sprintf("%.2f", req.Amount)
	p.OutTradeNo = fmt.Sprintf("%v", req.Id) // 唯一订单号
	p.NotifyURL = cfg.NotifyUrl
	p.ProductCode = cfg.ProductCode

	str_over_time := conf.Config.Server.Ext["order_over_time"].(string)

	p.TimeoutExpress = str_over_time // 订单超时设置 取值范围：1m～15d 该参数数值不接受小数点， 如 1.5h，可转换为 90m

	prepay_id, err = t.AliPay.TradeAppPay(p)
	if err != nil {
		glog.Error("TradeAppPay faiL! err=", err)
		return
	}
	return
}

// 订单查询
func (t *Alipay) QueryOrder(id int64) (status int, reason string, err error) {
	v := alipay.AliPayTradeQuery{OutTradeNo: fmt.Sprintf("%v", id)}
	ret, err := t.TradeQuery(v)
	if err != nil {
		glog.Error("alipay_query faiL! err=", err, " id=", id)
		return
	}
	if ret.IsSuccess() {
		status = GetTradeStatus(ret.AliPayTradeQuery.TradeStatus)
		reason = ret.AliPayTradeQuery.SubMsg
		v := TxParms{Status: status, TradeId: ret.AliPayTradeQuery.OutTradeNo, TxHash: ret.AliPayTradeQuery.TradeNo, Account: ret.AliPayTradeQuery.BuyerLogonId, Amount: ret.AliPayTradeQuery.TotalAmount, Reason: ret.AliPayTradeQuery.SubMsg}
		if err = v.UpdateRmbRecharge(); err != nil {
			glog.Error("alipay_query faiL! err=", err, " id=", id)
			return
		}
	}
	return
}

type CloseReq struct {
	TxHash string
	Reason string
}

// 取消订单
func (t *Alipay) CloseOrder(id int64) (v CloseReq, success bool, err error) {
	p := alipay.AliPayTradeClose{OutTradeNo: fmt.Sprintf("%v", id)}
	ret, err1 := t.TradeClose(p)
	if err1 != nil {
		err = err1
		glog.Error("CloseAlipay faiL! err=", err)
		return
	}
	success = IsSuccess(ret)
	v.Reason = ret.AliPayTradeClose.SubMsg
	v.TxHash = ret.AliPayTradeClose.TradeNo
	if !success {
		if ret.AliPayTradeClose.SubCode == "ACQ.TRADE_NOT_EXIST" || ret.AliPayTradeClose.SubCode == "ACQ.TRADE_STATUS_ERROR" {
			success = true // 交易不存在也认为成功
		}
	}
	return
}

func IsSuccess(v *alipay.AliPayTradeCloseResponse) bool {
	if v.AliPayTradeClose.Code == alipay.K_SUCCESS_CODE {
		return true
	}
	return false
}
