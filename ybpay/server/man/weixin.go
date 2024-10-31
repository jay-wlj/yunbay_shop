package man

import (
	"yunbay/ybpay/common"
	"yunbay/ybpay/conf"
	"yunbay/ybpay/server/share"
	"yunbay/ybpay/util"

	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
	wx "github.com/jay-wlj/wxpay"
	"github.com/jie123108/glog"
)

func Weixin_Refund(c *gin.Context) {
	var req common.RmbRefund
	if ok := util.UnmarshalBodyAndCheck(c, &req); !ok {
		return
	}

	db := db.GetTxDB(c)
	if err := db.Save(&req).Error; err != nil {
		glog.Error("Weixin_Refund fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	m := make(wx.Params)
	m.SetInt64("out_trade_no", req.OrderId)
	m.SetInt64("out_refund_no", req.Id)
	m.SetString("total_fee", req.TotalFee.Mul(decimal.New(int64(100), 0)).String())
	m.SetString("refund_fee", req.RefundFee.Mul(decimal.New(int64(100), 0)).String())
	m.SetString("notify_url", conf.Config.Weixin.RefundNotifyUrl)

	p, err := share.GetWeixin().Refund(m)
	if err != nil {
		glog.Error("Weixin_Refund fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if m.IsSuccess() {
		msg := p.GetString("err_code_des")
		glog.Error("Weixin_Refund fail! weixin refund err=", msg)
		yf.JSON_Fail(c, msg)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}
