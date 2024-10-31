package share

import (
	"fmt"
	"strconv"
	"time"
	"yunbay/ybpay/common"
	"yunbay/ybpay/conf"

	base "github.com/jay-wlj/gobaselib"

	"github.com/shopspring/decimal"

	wx "github.com/jay-wlj/wxpay"
	"github.com/jie123108/glog"
)

type WxClient struct {
	*wx.Client
}

var g_wxclient *WxClient

func GetWeixin() *WxClient {
	if g_wxclient == nil {
		cfg := conf.Config.Weixin
		account := wx.NewAccount(cfg.Appid, cfg.Mchid, cfg.Appkey, cfg.Sanbox)
		g_wxclient = &WxClient{wx.NewClient(account)}
	}
	return g_wxclient
}

func (t *WxClient) SignApp(prepay_id string) (p wx.Params, err error) {
	p = make(wx.Params)
	cfg := conf.Config.Weixin

	p.SetString("appid", cfg.Appid)
	p.SetString("partnerid", cfg.Mchid)
	p.SetString("prepayid", prepay_id)
	p.SetString("package", "Sign=WXPay")
	p.SetString("noncestr", strconv.FormatInt(time.Now().UTC().UnixNano(), 10))
	p.SetInt64("timestamp", time.Now().Unix())
	p.SetString("sign", t.Sign(p))
	return
}

// 统一下单
func (t *WxClient) TradeAppPay(req *OrderSt) (prepayid, reason string, err error) {
	p := make(wx.Params)
	p.SetString("body", req.Subject)
	p.SetString("out_trade_no", fmt.Sprintf("%v", req.Id))
	amount := req.Amount.Mul(decimal.New(int64(100), 0)).IntPart() // 将单位为元换成分
	//amount = math.Ceil(amount * 100) // 将单位为元换成分
	p.SetInt64("total_fee", amount)
	p.SetString("spbill_create_ip", req.RemoteIp)
	cfg := conf.Config.Weixin
	p.SetString("notify_url", cfg.NotifyUrl)
	p.SetString("trade_type", "APP")

	var params wx.Params
	if params, err = t.UnifiedOrder(p); err != nil {
		glog.Error("TradeAppPay fail! UnifiedOrder err=", err)
		return
	}

	if params.IsSuccess() {
		prepayid = params.GetString("prepay_id")
	} else {
		reason = params.GetString("return_msg")
		err = fmt.Errorf(reason)
	}
	return
}

// 查询交易
func (t *WxClient) QueryOrder(id int64) (status int, reason string, err error) {
	v := make(wx.Params)
	v["out_trade_no"] = base.Int64ToString(id)
	var ret wx.Params
	ret, err = t.OrderQuery(v)
	if err != nil {
		glog.Error("QueryOrder faiL! err=", err, " id=", id)
		return
	}
	status = common.STATUS_FAIL
	if v.IsSuccess() {
		status = common.STATUS_OK
	}
	reason = ret["err_code_des"]
	amount := float64(ret.GetInt64("total_fee")) / 100 // 将以分为单位转换成元为单位
	p := TxParms{Status: status, TradeId: ret["out_trade_no"], TxHash: ret["transaction_id"], Account: ret["openid"], Amount: base.Float64ToString(amount), Reason: reason}
	if err = p.UpdateRmbRecharge(); err != nil {
		glog.Error("alipay_query faiL! err=", err, " id=", id)
		return
	}
	return
}

// 取消交易
func (t *WxClient) CloseOrder(id int64) (r CloseReq, success bool, err error) {
	v := make(wx.Params)
	v["out_trade_no"] = base.Int64ToString(id)
	var ret wx.Params
	ret, err = t.Client.CloseOrder(v)
	if err != nil {
		glog.Error("alipay_cancel fail! err=", err, " id=", id)
		return
	}
	success = ret.IsSuccess()

	r.TxHash = ret["transaction_id"]
	r.Reason = ret["err_code_des"]

	// if ret.IsSuccess() {
	// 	// 交易关闭成功
	// 	if err = db.Model(&common.RmbRecharge{}).Where("id=? and status=?", id, common.STATUS_INIT).Updates(map[string]interface{}{"status": common.STATUS_FAIL}).Error; err != nil {
	// 		glog.Error("alipay_cancel fail! err=", err)
	// 		return
	// 	}
	// }
	return
}
