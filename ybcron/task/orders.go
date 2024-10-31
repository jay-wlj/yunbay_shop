package task

import (
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybcron/conf"
	"yunbay/ybcron/util"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 订单交易超时定时检查
func Orders_AutoCancelCheck() {
	//fmt.Println("Orders_AutoCancelCheck begin")
	db, err := db.InitPsqlDb(conf.Config.PsqlUrl["api"], conf.Config.Debug)
	if err != nil {
		fmt.Println("Orders_AutoCancelCheck end fail! err=", err)
	}
	db = db.Begin()
	if err := cancel_orders(db); err != nil {
		glog.Error("Orders_AutoCancelCheck end fail! err=", err)
		db.Rollback()
		return
	}
	db.Commit()

	// 定时检测充值支付订单状态
	// if err := util.CloseOverPayOrder(); err != nil {
	// 	glog.Error("RmbRecharge_OrderCheck fail! err=", err)
	// }
	//fmt.Println("Orders_AutoCancelCheck end success!")
}

// 订单确认超时定时检查
func Orders_AutoFinishCheck() {
	//fmt.Println("Orders_AutoFinishCheck begin")
	db, err := db.InitPsqlDb(conf.Config.PsqlUrl["api"], conf.Config.Debug)
	if err != nil {
		fmt.Println("Orders_AutoFinishCheck end fail! err=", err)
	}
	db1 := db.Begin()
	db = db.Begin()
	if err := finish_orders(db); err != nil {
		glog.Error("Orders_AutoFinishCheck end fail! err=", err)
		db.Rollback()
		return
	}
	db.Commit()

	//db1 := db.Begin()
	if err := auto_tip_orders(db1); err != nil {
		glog.Error("Orders_AutoTipCheck end fail! err=", err)
		db1.Rollback()
		return
	}
	db1.Commit()
	//fmt.Println("Orders_AutoFinishCheck end success!")
}

// 平台当日分发及销毁的ybt
func cancel_orders(db *gorm.DB) (err error) {
	//args := conf.Config.Rebat

	now := time.Now().Unix()

	// 查找所有待支付且已过锁定时间的订单
	vs := []common.Orders{}
	if err = db.Where("status=? and auto_cancel_time>0 and auto_cancel_time<?", common.ORDER_STATUS_UNPAY, now).Find(&vs).Error; err != nil {
		glog.Error("cancel_orders fail! err=", err)
		return
	}
	if len(vs) > 0 {
		ids := []int64{}
		user_ids := []int64{}
		seller_user_ids := []int64{}
		for _, v := range vs {
			ids = append(ids, v.Id)
			user_ids = append(user_ids, v.UserId)
			seller_user_ids = append(seller_user_ids, v.SellerUserId)
		}
		user_ids = base.UniqueInt64Slice(user_ids)
		seller_user_ids = base.UniqueInt64Slice(seller_user_ids)
		if err = util.CancelOrders(ids, user_ids, seller_user_ids); err != nil {
			glog.Error("cancel_orders fail! err=", err)
			return
		}
		// du := db.Model(common.Orders{}).Where("id in(?)", ids).Updates(map[string]interface{}{"status": common.ORDER_STATUS_CANCEL, "update_time": now})
		// if du.Error != nil || du.RowsAffected != int64(len(ids)) {
		// 	err = du.Error
		// 	glog.Error("cancel_orders RowsAffected fail! err=", err)
		// 	return
		// }
	}

	return
}

// 自动确认收货
func finish_orders(db *gorm.DB) (err error) {
	//args := conf.Config.Rebat

	now := time.Now().Unix()

	// 查找所有已发货且已过确认时间的订单
	vs := []common.Orders{}
	if err = db.Where("status=? and auto_finish_time>0 and auto_finish_time<?", common.ORDER_STATUS_SHIPPED, now).Find(&vs).Error; err != nil {
		glog.Error("cancel_orders fail! err=", err)
		return
	}
	if len(vs) > 0 {
		ids := []int64{}
		user_ids := []int64{}
		seller_user_ids := []int64{}
		for _, v := range vs {
			ids = append(ids, v.Id)
			user_ids = append(user_ids, v.UserId)
			seller_user_ids = append(seller_user_ids, v.SellerUserId)
		}
		user_ids = base.UniqueInt64Slice(user_ids)
		seller_user_ids = base.UniqueInt64Slice(seller_user_ids)
		if err = util.FinishOrders(ids, user_ids, seller_user_ids); err != nil {
			glog.Error("cancel_orders fail! err=", err)
			return
		}

		// du := db.Model(common.Orders{}).Where("id in(?)", ids).Updates(map[string]interface{}{"status": common.ORDER_STATUS_FINISH, "update_time": now})
		// err = du.Error
		// if err != nil || du.RowsAffected != int64(len(ids)) {
		// 	err = du.Error
		// 	glog.Error("cancel_orders RowsAffected fail! err=", err)
		// 	return
		// }

		// err = util.SetPayStatus(ids)
	}

	return
}

type orderidSt struct {
	OrderId int64 `json:"order_id"`
}

// 查询某开有订单状态的订单id列表
func getOrderIdsByDateStatus(date string, status int) (ids []int64, mids map[int64]bool, err error) {
	ids = []int64{}
	mids = make(map[int64]bool)
	db, err := db.InitPsqlDb(conf.Config.PsqlUrl["api"], conf.Config.Debug)
	if err != nil {
		fmt.Println("Orders_AutoCancelCheck end fail! err=", err)
	}

	var vs []orderidSt
	if err = db.Model(&common.OrderStatus{}).Where("date=? and status=?", date, status).Select("order_id").Scan(&vs).Error; err != nil {
		glog.Error("getorderIdsByDatestatus fail! err=", err)
		return
	}

	for _, v := range vs {
		ids = append(ids, v.OrderId)
		mids[v.OrderId] = true
	}
	return
}

// 付款的订单超过10分钟 则短信提醒商家
func auto_tip_orders(db *gorm.DB) (err error) {
	var orders []common.Orders
	now := time.Now().Unix()

	duration, _ := time.ParseDuration(conf.Config.Orders["newtime"])
	auto_tip := int64(duration.Seconds())
	if auto_tip > 0 {
		auto_tip = now - auto_tip
	}

	// 只发短信给实物商品的商家
	if err = db.Find(&orders, "status=? and product_type=? and update_time<? and (maninfos->>'sms_status' is null or (maninfos->>'sms_status')::integer <> ?)", common.ORDER_STATUS_PAYED, common.GOODS_TYPE_PHYSICAL, auto_tip, common.STATUS_OK).Error; err != nil {
		glog.Error("auto_tip_orders fail! err=", err)
		return
	}
	if len(orders) == 0 {
		return
	}
	uids := []int64{}
	ids := []int64{}
	for _, v := range orders {
		uids = append(uids, v.SellerUserId)
		ids = append(ids, v.Id)
	}
	// 更新已发送短信状态
	//sql := fmt.Sprintf("update %v set maninfos=maninfos || '{\"sms_status\":%v}' where id in([%v])", (common.Orders{}).TableName(), common.STATUS_OK);
	if err = db.Model(&common.Orders{}).Where("id in(?)", ids).Updates(map[string]interface{}{"maninfos": gorm.Expr("maninfos - 'sms_status' || '{\"sms_status\":1}'")}).Error; err != nil {
		glog.Error("auto_tip_orders fail! err=", err)
		return
	}
	// 去重uid
	uids = base.UniqueInt64Slice(uids)

	// 获取商家联系方式
	bs := []common.Business{}
	if err = db.Select("user_id, contact").Find(&bs, "user_id in(?)", uids).Error; err != nil {
		glog.Error("auto_tip_orders fail! err=", err)
		return
	}
	tels := []util.TelInfo{}
	for _, v := range bs {
		if tel, ok := v.Contact["contact_phone"].(string); ok {
			tels = append(tels, util.TelInfo{Tel: tel})
		}
	}
	if len(tels) > 0 {
		if _, err = util.SendTelsSms(tels, conf.Config.Orders["newtips"]); err != nil {
			if err1 := db.Model(&common.Orders{}).Where("id in(?)", ids).Updates(map[string]interface{}{"maninfos": gorm.Expr("maninfos - 'sms_status' || '{\"sms_status\":0}'")}).Error; err1 != nil {
				glog.Error("auto_tip_orders fail! err=", err, " fail_ids:", ids)
				return
			}
		}
	}

	// fail_ids, err := util.SendSms(uids, conf.Config.Orders["newtips"])
	// if err != nil {
	// 	glog.Error("auto_tip_orders fail! err=", err)
	// 	return
	// }
	// if len(fail_ids) > 0 {
	// 	if err1 := db.Model(&common.Orders{}).Where("id in(?)", fail_ids).Updates(map[string]interface{}{"maninfos": gorm.Expr("maninfos - 'sms_status' || '{\"sms_status\":0}'")}).Error; err1 != nil {
	// 		glog.Error("auto_tip_orders fail! err=", err, " fail_ids:", fail_ids)
	// 		return
	// 	}
	// }

	return
}

type ofId struct {
	OrderId int64 `json:"order_id"`
}

// 查询未确认状态的欧飞充值订单
func OfOrdersQuery() {
	ids := []ofId{}
	db, err := db.InitPsqlDb(conf.Config.PsqlUrl["api"], conf.Config.Debug)
	if err != nil {
		fmt.Println("Orders_AutoCancelCheck end fail! err=", err)
	}
	if err = db.Model(&common.OfOrder{}).Where("game_state=?", common.STATUS_INIT).Select("order_id").Limit(20).Scan(&ids).Error; err != nil {
		glog.Error("OfOrdersQuery fail! err=", err)
		return
	}
	if len(ids) > 0 {
		oids := []int64{}
		for _, v := range ids {
			oids = append(oids, v.OrderId)
		}
		if err := util.QueryOfOrders(oids); err != nil {
			glog.Error("OfOrdersQuery fail! err=", err)
			return
		}
	}
}
