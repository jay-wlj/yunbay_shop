package client

import (
	"yunbay/ybpay/common"
	//"yunbay/ybpay/dao"
	"yunbay/ybpay/server/share"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

func Bank_Query(c *gin.Context) {
	card_id, _ := base.CheckQueryStringField(c, "card_id")
	if card_id == "" || len(card_id) < 6 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	// 取前6位
	card_id = card_id[:6]
	v, ok := share.QueryBank(card_id)
	if !ok {
		yf.JSON_Fail(c, common.ERR_NOT_SUPPORT_BANK)
		return
	}
	v.CardId = card_id
	yf.JSON_Ok(c, v)
}

func Trade_Query(c *gin.Context) {
	id, err := base.CheckQueryInt64Field(c, "id")
	if err != nil {
		glog.Error("Trade_Query fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	v, err := share.TradeQuery(id)
	if err != nil {
		glog.Error("Trade_Query fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	if v.Status != common.STATUS_OK {
		db := db.GetTxDB(c).Begin()

		// 查询交易
		var err error
		switch v.Channel {
		case common.CHANNEL_ALIPAY:
			v.Status, v.Reason, err = share.GetAliPay().QueryOrder(id)
		case common.CHANNEL_WEIXIN:
			v.Status, v.Reason, err = share.GetWeixin().QueryOrder(id)
		}

		if err != nil {
			db.Rollback()
		}
		db.Commit()

	}

	yf.JSON_Ok(c, gin.H{"id": id, "status": v.Status, "reason": v.Reason})
}
