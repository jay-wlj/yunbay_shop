package man

import (
	"yunbay/ybpay/common"
	"yunbay/ybpay/server/share"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

var m_drawstatus map[string]int

func init() {
	m_drawstatus = make(map[string]int)
	m_drawstatus["waiting"] = common.TX_STATUS_WAITING
	m_drawstatus["submitted"] = common.TX_STATUS_SUBMIT
	m_drawstatus["confirming"] = common.TX_STATUS_CONFIRM
	m_drawstatus["failed"] = common.TX_STATUS_FAILED
	m_drawstatus["success"] = common.TX_STATUS_SUCCESS
}

func getstatustxtbyint(status int) string {
	for k, v := range m_drawstatus {
		if v == status {
			return k
		}
	}
	return ""
}

// 定时关闭交易订单
func Trade_Close(c *gin.Context) {
	var vs []common.RmbRecharge
	now := time.Now().Unix()

	// 获取过期没有关闭的支付订单
	db := db.GetTxDB(c)
	if err := db.Find(&vs, "status=? and over_time<=?", common.STATUS_INIT, now).Limit(50).Error; err != nil {
		glog.Error("Trade_Close fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if len(vs) == 0 {
		yf.JSON_Ok(c, gin.H{})
		return
	}
	ids := []int64{}
	for _, v := range vs {
		ids = append(ids, v.Id)
	}

	// 先全部置为失败状态
	if err := db.Model(&common.RmbRecharge{}).Where("id in(?)", ids).Updates(map[string]interface{}{"status": common.STATUS_FAIL, "update_time": now}).Error; err != nil {
		glog.Error("Trade_Close fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	var err error
	var ok bool

	fail_ids := []int64{}
	for _, v := range vs {
		var r share.CloseReq
		ok = false
		switch v.Channel {
		case common.CHANNEL_ALIPAY:
			r, ok, err = share.GetAliPay().CloseOrder(v.Id)
		case common.CHANNEL_WEIXIN:
			r, ok, err = share.GetWeixin().CloseOrder(v.Id)
		}

		if err != nil || !ok {
			fail_ids = append(fail_ids, v.Id)
		} else {
			// 调用成功 更新rason及支付宝单号等
			if er1 := db.Model(&common.RmbRecharge{}).Where("id=?", v.Id).Updates(map[string]interface{}{"txhash": r.TxHash, "reason": r.Reason}).Error; er1 != nil {
				glog.Error("Trade_Close fail! err=", err)
			}
		}
	}

	// 将更新失败的订单置回初始状态
	if len(fail_ids) > 0 {
		if err := db.Model(&common.RmbRecharge{}).Where("id in(?)", fail_ids).Updates(map[string]interface{}{"status": common.STATUS_INIT, "update_time": now}).Error; err != nil {
			glog.Error("Trade_Close fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}
	yf.JSON_Ok(c, gin.H{})
}
