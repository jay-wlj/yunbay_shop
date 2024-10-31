package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"yunbay/yborder/common"
	"yunbay/yborder/server/man"
	"yunbay/yborder/server/share"
	"yunbay/yborder/util"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

type payConfirm struct {
	share.PayParmas
	ZJPassword string `json:"zjpassword" binding:"gte=6"`
}

// 订单支付接口
func Order_Pay(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args payConfirm
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	args.UserId = user_id
	args.Amount = args.Amount.Round(6) // 取小数后六位

	// 验证用户资金密码
	token, _ := util.GetHeaderString(c, "X-Yf-Token")
	err := util.AuthUserZJPassword(token, args.ZJPassword)
	if err != nil {
		yf.JSON_Fail(c, common.ERR_ZJPASSWORD_INVALID)
		return
	}
	args.Country = util.GetCountry(c)

	db := db.GetTxDB(c)
	var o share.Order
	if reason, err := o.Pay(db, args.PayParmas, ""); err != nil {
		glog.Error("Order_Pay fail! err=", err)
		if reason == "" {
			reason = err.Error()
		}
		yf.JSON_Fail(c, reason)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

type orderid struct {
	OrderId int64 `json:"order_id" binding:"gt=0"`
}

// 退款操作
func Refund(db *db.PsqlDB, v *common.Orders) (err error) {
	if v.Status != common.ORDER_STATUS_PAYED {
		glog.Error("ERR_ORDERS_CANCEL_FAILED fail! orders.Status=", v.Status, " not ", common.ORDER_STATUS_PAYED)
		err = fmt.Errorf(common.ERR_ORDER_FORBIDDEN_CANCEL)
		return
	}

	// 折扣专区订单在生成折扣前不可退款
	if v.PublishArea == common.PUBLISH_AREA_REBAT {
		// if _, ok := v.ExtInfos["rebat_hash"]; !ok {
		// 	glog.Error("ERR_ORDERS_CANCEL_FAILED fail! rebat order has generating rebat")
		err = fmt.Errorf(common.ERR_ORDER_FORBIDDEN_CANCEL)
		return
		//}
	}

	// // 验证用户资金密码
	// err = util.AuthUserZJPassword(user_id, args.ZJPassword)
	// if err != nil {
	// 	yf.JSON_Fail(c, common.ERR_ZJPASSWORD_INVALID)
	// 	return
	// }
	// 验证用户帐户钱包
	// var userasset common.UserAsset
	// if err = db.Where("user_id=?", user_id).Find(&userasset).Error; err != nil {
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }

	nTime := time.Now()
	now := nTime.Unix()

	// // 修改相应产品规格库存
	// if err = share.ModifyProductModelQuantity(db, v.ProductSkuId, -v.Quantity); err != nil {
	// 	glog.Errorf("Update product quantity fail! err=",err)
	// 	return
	// }

	// 修改订单状态为已取消
	// order.Status = common.ORDER_STATUS_CANCEL
	// order.UpdateTime = 	now
	if err = db.Model(v).Updates(map[string]interface{}{"status": common.ORDER_STATUS_REFUND, "update_time": now}).Error; err != nil {
		glog.Error("Update product quantity fail! err=", err)
		return
	}

	// // 执行退款
	ys := util.YBAssetStatus{OrderIds: []int64{v.Id}, Status: common.ASSET_POOL_CANCEL, PublishArea: v.PublishArea}
	if err = util.YBAsset_PayStatus(ys); err != nil {
		glog.Error("YBAsset_PayStatus fail! err=", err)
		return
	}

	// 修改相应产品规格库存及退款回调
	vquantitys := []util.QuantitySt{util.QuantitySt{OrderId: v.Id, ProductId: v.ProductId, ProductSkuId: v.ProductSkuId, Quantity: v.Quantity}}
	// 支付失败 需相应减少已售数量 异步处理
	util.AddProductModelQuantity(vquantitys)
	// if err = util.AddProductModelQuantity(vquantitys, token, &util.CallBackSt{AppKey:"ybasset", Uri:"/man/asset/payset", Method:"post", Body:ys}); err != nil {
	// 	glog.Errorf("Order_Pay fail! Update product quantity fail! err=",err)
	// 	return
	// }

	return
	// 修改平台交易资金池记录 将卖家所得额给买家，余下手续费用归平台所有用于后面分红
	// var p common.YBAssetPool
	// if err = db.Where("payer_userid=? and order_id=?", user_id, args.OrderId).Find(&p).Error; err != nil {
	// 	glog.Errorf("Get YBAssetPool fail! err=",err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	// if p.Status != common.ASSET_POOL_LOCK {
	// 	glog.Errorf("OrderFinish fail! pool status:%v is not ASSET_POOL_LOCK",p.Status)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	// p.Status = common.ASSET_POOL_CANCEL
	// if err = db.Model(&p).Update("status").Error; err != nil {
	// 	glog.Error("YBAssetPool_Add fail! err=", err)
	// 	return
	// }

	// // 改成触发器完成
	// // 添加用户资金变化记录(触发器会自动计算用户帐户金额)
	// detail := common.UserAssetDetail{UserId:user_id, Type:common.CURRENCY_KT, TransactionType:common.KT_TRANSACTION_RETURND, Amount:p.SellerAmount, CreateTime:now, UpdateTime:now}
	// if err = db.Create(&detail).Error; err != nil {
	// 	glog.Errorf("Create UserAssetDetail fail! err=",err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
}

// // 订单取消退款接口
// func Order_Refund(c *gin.Context) {
// 	_, ok := util.GetUid(c)
// 	if !ok {
// 		return
// 	}
// 	var args orderid
// 	if ok := util.UnmarshalReq(c, &args); !ok {
// 		return
// 	}
// 	db := db.GetTxDB(c)
// 	// 获取所支付的订单商品信息
// 	order, err := Orders_GetById(db.DB, args.OrderId)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			yf.JSON_Fail(c, common.ERR_ORDER_NOT_EXIST)
// 			return
// 		}
// 		glog.Error("Orders_ListByIds fail! err=", err)
// 		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
// 		return
// 	}

// 	token, _ := util.GetHeaderString(c, "X-Yf-Token")
// 	if err := Refund(db, &order, token); err != nil {
// 		glog.Error("Refund fail! err=", err)
// 		yf.JSON_Fail(c, err.Error())
// 		return
// 	}
// 	o := &dao.Orders{}
// 	o.RefreshByOrderIds([]int64{args.OrderId})
// 	yf.JSON_Ok(c, gin.H{})
// }

// 订单完成接口
func Order_Finish(c *gin.Context) {
	_, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args orderid
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	db := db.GetTxDB(c)
	var v share.Order
	if err := v.Finish(db, args.OrderId); err != nil {
		glog.Error("Order_Finish fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}
	// 修改平台交易资金池记录 将卖家所得额给买家，余下手续费用归平台所有用于后面分红
	// var p common.YBAssetPool
	// if err = db.Where("payer_userid=? and order_id=?", user_id, args.OrderId).Find(&p).Error; err != nil {
	// 	glog.Errorf("Get YBAssetPool fail! err=",err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	// if p.Status != common.ASSET_POOL_LOCK {
	// 	glog.Errorf("OrderFinish fail! pool status:%v is not ASSET_POOL_LOCK",p.Status)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }

	// if err = db.Model(&p).Updates(map[string]interface{}{"status":common.ASSET_POOL_FINISH, "update_time":now}).Error; err != nil {
	// 	glog.Error("YBAssetPool_Add fail! err=", err)
	// 	return
	// }

	// 改成触发器完成
	// 添加用户资金变化记录(触发器会自动计算用户帐户金额)
	// detail := common.UserAssetDetail{UserId:p.SellerUserId, Type:common.CURRENCY_KT, TransactionType:common.KT_TRANSACTION_SELLER, Amount:p.SellerAmount, CreateTime:now, UpdateTime:now}
	// if err = db.Create(&detail).Error; err != nil {
	// 	glog.Errorf("Create UserAssetDetail fail! err=",err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }

	yf.JSON_Ok(c, gin.H{})
}

// 折扣专区购买公告
func RebatePaidAfficheList(c *gin.Context) {
	redisCache, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("redis cache connect is fail! error = ", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	var afficheList []string
	afficheList, _ = redisCache.LRange("paid_affiche_list", 0, 5).Result()
	stringList := strings.Replace(fmt.Sprint(afficheList), " ", ",", -1)
	var paidAfficheList []man.PaidAffiche
	json.Unmarshal([]byte(stringList), &paidAfficheList)
	var longTime int64 = 80
	var num int = 0
	returnData := []man.PaidAffiche{}
	nowTime := time.Now().Unix()
	for _, v := range paidAfficheList {
		if (nowTime - v.UpdatedTime) > longTime {
			break
		} else {
			// glog.Error(num)
			returnData = append(returnData, v)
			num++
		}
	}
	// 判断长度是否超过5，超过后，后面的记录清空
	redisCache.LTrim("paid_affiche_list", 0, 5)
	yf.JSON_Ok(c, returnData)
}
