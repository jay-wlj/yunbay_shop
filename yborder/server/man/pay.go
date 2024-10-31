package man

import (
	//"fmt"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"yunbay/yborder/server/share"
	"yunbay/yborder/util"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

// 订单支付接口
func Orders_Pay(c *gin.Context) {
	var args share.PayParmas
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	args.Country = util.GetCountry(c)
	db := db.GetTxDB(c)
	var o share.Order
	if reason, err := o.Pay(db, args, ""); err != nil {
		glog.Error("Order_Pay fail! err=", err)
		yf.JSON_Fail(c, reason)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}
