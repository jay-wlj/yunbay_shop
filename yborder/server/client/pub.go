package client

import (
	"encoding/json"
	"strconv"
	"yunbay/yborder/common"
	"yunbay/yborder/conf"
	"yunbay/yborder/server/share"
	"yunbay/yborder/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

func TelCheck(c *gin.Context) {
	tel, _ := base.CheckQueryStringField(c, "tel")
	val, _ := base.CheckQueryIntDefaultField(c, "amount", 100)
	if !yf.ValidTel(tel) {
		yf.JSON_Fail(c, common.ERR_TEL_NOT_SUPPORT_RECHARGE)
		return
	}
	err := util.GetOfpay().Tel_Check(tel, val)
	if err != nil {
		glog.Error("TelCheck fail! err=", err)
		yf.JSON_Fail(c, common.ERR_TEL_NOT_SUPPORT_RECHARGE, err.Error())
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

// 欧飞订单号结果通知
func OfRetNotify(c *gin.Context) {
	var ret_code, sporder_id, err_msg string
	var userid, sign string

	var ok bool

	if ret_code, ok = c.GetPostForm("ret_code"); !ok {
		body, _ := json.Marshal(c.Request.PostForm)
		glog.Error("lost ret_code! params=", string(body))
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	if sporder_id, ok = c.GetPostForm("sporder_id"); !ok {
		body, _ := json.Marshal(c.Request.PostForm)
		glog.Error("lost sporder_id! params=", string(body))
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	if userid, ok = c.GetPostForm("userid"); !ok {
		body, _ := json.Marshal(c.Request.PostForm)
		glog.Error("lost userid! params=", string(body))
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	if sign, ok = c.GetPostForm("sign"); !ok {
		body, _ := json.Marshal(c.Request.PostForm)
		glog.Error("lost sign! params=", string(body))
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	s := userid + ret_code + sporder_id + conf.Config.OfPay.AppSecret
	calc_sign := yf.Md5Hex([]byte(s))
	// 云贝商城回调接口没有进行签名
	if sign != "" && calc_sign != sign {
		body, _ := json.Marshal(c.Request.PostForm)
		glog.Error("sign error sign:", sign, " calc_sign:", calc_sign, "ret_code:", ret_code, "sporder_id:", sporder_id, "userid:", userid, " params=", string(body))

		yf.JSON_Fail(c, yf.ERR_SIGN_ERROR)
		return
	}
	//ordersuccesstime, _ = c.GetPostForm("ordersuccesstime")
	err_msg, _ = c.GetPostForm("err_msg")
	retCode, _ := strconv.Atoi(ret_code)

	order_id, err := base.StringToInt64(sporder_id)
	if err != nil {
		glog.Error("lost sporder_id! params=", sporder_id)
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	v := common.OfOrder{OrderId: order_id, GameState: retCode, ErrMsg: err_msg}
	if err = share.OfRetNotify(nil, &v); err != nil {
		glog.Error("OfRetNotify fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}
	c.String(200, "Y")
}
