package client

import (
	"fmt"
	"yunbay/ybasset/common"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

type lltchargeSt struct {
	OrderId  string  `json:"order_id"`
	UserId   int64   `json:"user_id"`
	Amount   float64 `json:"amount"`
	Platform string  `json:"platform"`
}

// yunex充值回调
func LLT_Recharge(c *gin.Context) {
	// if !util.Yunex_SignCheck(c) {
	// 	return
	// }
	req := lltchargeSt{}
	if ok := util.UnmarshalReqParms(c, &req, "INVALID_ARGS"); !ok {
		return
	}

	// 判断用户id是否存在
	vs, err := util.UserInfoGetByUserIds([]int64{req.UserId})
	if err != nil {
		glog.Error("LLT_Recharge fail! err=", err, " user_id:", req.UserId)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if 0 == len(vs) {
		glog.Error("LLT_Recharge fail! user_id not exist :", req.UserId)
		yf.JSON_Fail(c, "NOT_FOUND")
		return
	}
	address, err := share.GetAndSaveUserAddress(req.UserId)
	if err != nil {
		glog.Error("LLT_Recharge fail! err=", err, " user_id:", req.UserId)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	db := db.GetTxDB(c)
	id, reason, err := share.Recharge_fromthird(db, "miner", address, "snet", req.Amount, req.OrderId)
	if err != nil {
		if reason != "" {
			yf.JSON_Fail(c, reason)
		} else {
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		}
		return
	}
	yf.JSON_Ok(c, gin.H{"order_id": fmt.Sprintf("%v", id)})
}

// 通过用户充值地址查询用户id
func LLT_RechargeQuery(c *gin.Context) {
	if !util.Yunex_SignCheck(c) {
		return
	}
	id, _ := base.CheckQueryStringField(c, "order_id")

	var v common.RechargeFlow
	db := db.GetDB()
	if err := db.Find("txhash=?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, "ORDER_NOT_EXIST")
			return
		}
		glog.Error("LLT_RechargeQuery fail! err=?", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"order_id": fmt.Sprintf("%v", v.Id), "amount": v.Amount})
}
