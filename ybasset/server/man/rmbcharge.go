package man

import (
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"
	"fmt"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"

	//"github.com/jinzhu/gorm"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"
	base "github.com/jay-wlj/gobaselib"
	"time"
)

// 服务启动时处理支付宝支付成功还没入帐的订单
func HandlePreRecharge() {
	defer func() {
		if err := recover(); err != nil {
			glog.Error("panic HandlePreRecharge err=", err)
			//glog.Error("panic HandlePreRecharge err=", err)
		}
	}()

	var vs []common.RmbRecharge

	d := db.GetDB()
	// 查找支付宝支付成功还没入帐的订单
	if err := d.Find(&vs, "status=? and asset_id=0", common.STATUS_OK).Error; err != nil {
		glog.Error("HandlePreRecharge fail! err=", err)
		return
	}
	for _, v := range vs {
		// 处理完
		ctx := db.GetTxDB(nil)
		if err := payTransferKt(ctx, v.Id, 1); err != nil {
			ctx.Rollback()
		}
		ctx.Commit()
	}
}

type idSt struct {
	RechargeId int64 `json:"recharge_id"`
}

func payTransferKt(db *db.PsqlDB, id int64, country int) (err error) {
	var v common.RmbRecharge
	if err = db.Find(&v, "id=? and status=?", id, common.STATUS_OK).Error; err != nil {
		glog.Error("payTransfer fail! err=", err)
		return
	}
	// 这里需要将相应的rmb换算成对应的kt
	amount_kt := v.Amount
	if fRate := share.GetRatio("cny", "kt"); !base.IsEqual(fRate, 0) {
		amount_kt = amount_kt * fRate
	}
	if v.AssetId == 0 {
		// 划扣相应rmb的kt给该用户
		from := conf.Config.SystemAccounts["rmb_account"]
		var idAsset int64
		extinfos := make(map[string]interface{})
		extinfos["rmb_amount"] = v.Amount

		if idAsset, err = share.WalletTransferTo(db, from, v.UserId, common.CURRENCY_KT, amount_kt, country, extinfos); err != nil {
			glog.Error("payTransfer fail! err=", err)
			return
		}
		res := db.Model(&v).Where("id=? and status=? and asset_id=0", id, common.STATUS_OK).Updates(map[string]interface{}{"asset_id": idAsset, "update_time": time.Now().Unix()})
		if err = res.Error; err != nil {
			glog.Error("payTransfer fail! err=", err)
			return
		}
		v.AssetId = idAsset

		// 划扣相应kt后 通知api服务继续支付订单
		if res.RowsAffected > 0 && len(v.OrderIds) > 0 {
			d := util.PayParmas{UserId: v.UserId, CurrencyType: common.CURRENCY_KT, Amount: amount_kt, OrderIds: v.OrderIds}
			go payOrders(d)

			// headers := make(map[string]string)
			// headers["x-yf-country"] = "1"	// 国内版本支付
			// m := common.MQUrl{Methond:"post", AppKey:"ybapi", Uri:"/man/order/pay", Data:d, Headers:headers, MaxTrys:-1}
			// if err := util.PublishMsg(m); err != nil {
			// 	glog.Error("RmbRechargeNotify fail! err=", err)
			// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			// 	return
			// }
		}
	}
	return
}

func RmbRechargeNotify(c *gin.Context) {
	// maner, err := util.GetHeaderString(c, "X-Yf-Maner")
	// if maner == "" || err != nil {
	// 	yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
	// 	return
	// }
	var req idSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	country := util.GetCountry(c)

	db := db.GetTxDB(c)
	if err := payTransferKt(db, req.RechargeId, country); err != nil {
		glog.Error("RmbRechargeNotify fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

func payOrders(v util.PayParmas) {
	if err := util.PayOrders(v); err != nil {
		if err.Error() != "ERR_ORDER_NOT_EXIST" { // 订单已超时或已被支付取消等
			glog.Error("RmbRechargeNotify fail! err=", err)
			s := fmt.Sprintln("payOrders fail! err=", err)
			util.SendDingTextTalkToMe(s)
			return
		}
	}
}
