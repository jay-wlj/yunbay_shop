package share

import (
	"errors"
	"fmt"
	"yunbay/yborder/common"
	"yunbay/yborder/dao"
	"yunbay/yborder/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/shopspring/decimal"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

type Order struct{}

type PayParmas struct {
	OrderIds     []int64         `json:"order_ids" binding:"gt=0"`
	CurrencyType int             `json:"pay_type"`
	Amount       decimal.Decimal `json:"amount,omitempty" binding:"required,gt=0"`
	UserId       int64           `json:"user_id"`
	Country      int
}

// 订单支付逻辑
func (t *Order) Pay(db *db.PsqlDB, req PayParmas, token string) (reason string, err error) {
	err = fmt.Errorf("ERR_SERVER_ERROR")
	// 获取所支付的订单商品信息
	var vs []common.Orders
	if err = db.Where("id in(?)", req.OrderIds).Find(&vs).Error; err != nil {
		glog.Error("Pay fail! err=", err)
		return
	}

	// 如果超时已取消状态的订单,则退回
	for _, v := range vs {
		if v.Status == common.ORDER_STATUS_CANCEL {
			if err = t.Refund(db, &v); err != nil {
				glog.Error("Order pay fail! err=", err)
				return
			}
		}
	}

	// 判断订单金额是否一致
	var amount decimal.Decimal
	for _, v := range vs {
		//amount.Add(v.TotalAmount.Mul(v.CurrencyPercent))
		amount = amount.Add(decimal.NewFromFloat(v.TotalAmount * v.CurrencyPercent))
	}

	if !req.Amount.Equal(amount) {
		glog.Errorf("pay amount:%v order amount:%v", req.Amount, amount)
		reason = common.ERR_ORDER_AMOUNT_INVALID
		err = fmt.Errorf(reason)
		return
	}

	mProdcutIds := make(map[int64]bool)
	vquantitys := []util.QuantitySt{}
	pools := []util.YBAssetPool{}
	for _, v := range vs {
		mProdcutIds[v.ProductId] = true

		// 卖家应得金额为减去贡献值后的金额
		selleramount := v.TotalAmount - v.RebatAmount
		if selleramount < 0 {
			glog.Errorf("selleramount <0? fail! TotalAmount:%v RebatAmount:%v", v.TotalAmount, v.RebatAmount)
			reason = common.ERR_ORDER_AMOUNT_INVALID
			err = fmt.Errorf(reason)
			return
		}

		// 添加平台交易资金池记录
		p := util.YBAssetPool{OrderId: v.Id, CurrencyType: v.CurrencyType, PayerUserId: v.UserId, PayAmount: v.TotalAmount, SellerUserId: v.SellerUserId, SellerAmount: selleramount, RebatAmount: v.RebatAmount, Status: 0, Country: req.Country, PublishArea: v.PublishArea}
		// 将rmb兑换成相应货币比例
		p.PayAmount = p.PayAmount * v.CurrencyPercent
		p.SellerAmount = p.SellerAmount * v.CurrencyPercent
		p.RebatAmount = p.RebatAmount * v.CurrencyPercent
		if v.Maninfos != nil {
			var man common.ManInfos
			man.ParseJsonb(v.Maninfos)
			//if m, ok := v.Maninfos["cost_price"].(float64); ok {
			p.SellerKt, _ = man.Amount.Round(6).Float64() // 商品实际成本价kt
			//}
		}

		pools = append(pools, p)

		vquantitys = append(vquantitys, util.QuantitySt{OrderId: v.Id, ProductId: v.ProductId, ProductSkuId: v.ProductSkuId, Quantity: -v.Quantity})
	}

	// 修改订单状态为已支付状态
	ids := []int64{}
	for _, v := range req.OrderIds {
		ids = append(ids, v)
	}
	res := db.Model(&common.Orders{}).Where("status = ? and id in(?)", common.ORDER_STATUS_UNPAY, req.OrderIds).Updates(map[string]interface{}{"status": common.ORDER_STATUS_PAYED})
	if err = res.Error; err != nil {
		glog.Error("Update order fail! err=", err)
		//yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if res.RowsAffected != int64(len(req.OrderIds)) {
		// 有订单状有变 应取消本次支付操作
		if 0 == res.RowsAffected {
			return
		}
		glog.Errorf("pay ids:%v order ids size:%v", len(req.OrderIds), res.RowsAffected)
		reason = common.ERR_ORDER_NOT_EXIST
		err = fmt.Errorf(reason)
		return
	}
	// 修改相应产品规格库存及付款回调
	if err = util.AddProductModelQuantity(vquantitys, true); err != nil {
		glog.Error("Order_Pay fail! Update product quantity fail! err=", err)
		//yf.JSON_Fail(c, err.Error())
		return
	}

	if token == "" {
		err = util.YBAsset_PayByUserId(pools, req.UserId)
	} else {
		err = util.YBAsset_Pay(pools, token)
	}

	if err != nil {
		// 支付失败 需相应减少已售数量
		for i, _ := range vquantitys {
			vquantitys[i].Quantity = -vquantitys[i].Quantity
		}
		util.AddProductModelQuantity(vquantitys)
		glog.Error("orders pay fail! err=", err)
		reason = err.Error()
		return
	}

	// 提交后刷新订单相关缓存
	db.AfterCommit(func() {
		o := &dao.Orders{}
		o.RefreshByOrderIds(req.OrderIds)

		for k := range mProdcutIds {
			o.RefreshByProductId(k)
		}
	})

	need_rebat_orders := []int64{}
	for _, v := range vs {
		if v.PublishArea == common.PUBLISH_AREA_REBAT {
			need_rebat_orders = append(need_rebat_orders, v.Id)
		}
	}
	if len(need_rebat_orders) > 0 {
		//go getOrderRebat(need_rebat_orders)	// 获取折扣专区的折扣
		for _, id := range need_rebat_orders {
			if err = util.AsynGenerateEosOrderRebat(id); err != nil {
				glog.Error("OrderPay fail! err=", err)
				return
			}
		}
	}

	// 是否有自动发货
	type ordersSt struct {
		OrderId int64 `json:"order_id"`
	}
	os := []ordersSt{}
	for _, v := range vs {
		if v.AutoDeliver {
			os = append(os, ordersSt{OrderId: v.Id})
			//vt := ordersSt{OrderId: v.Id}
		}
	}
	if len(os) > 0 {
		// 提交后才发送异步消息
		db.AfterCommit(func() {
			for _, v := range os {
				mq := common.MQUrl{Methond: "post", Uri: "/man/order/auto_deliver", AppKey: "yborder", Data: v, MaxTrys: -1}
				util.AsyncPublishMsg(mq) // 异步消息处理
			}
		})
	}

	err = nil // 成功
	return
}

// 取消订单
func (t *Order) Cancel(db *db.PsqlDB, id int64, f func(t *common.Orders) bool) (err error) {
	var v common.Orders
	if err = db.Find(&v, "id=?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(common.ERR_ORDER_NOT_EXIST)
		}
		glog.Error("OrderPay fail! err=", err)
		return
	}

	switch v.Status {
	case common.ORDER_STATUS_UNPAY, common.ORDER_STATUS_PAYED:
	case common.ORDER_STATUS_CANCEL, common.ORDER_STATUS_REFUND:
		return
	default:
		// 其它状态禁止取消订单
		err = errors.New(common.ERR_ORDER_FORBIDDEN_CANCEL)
		glog.Error("ERR_ORDERS_CANCEL_FAILED fail! orders.Status=", v.Status)
		return
	}

	if f != nil && !f(&v) {
		return
	}
	// if only_unpay && common.ORDER_STATUS_UNPAY != v.Status {
	// 	return
	// }

	// 自动发货商品禁止取消
	// if v.AutoDeliver && v.Status != common.ORDER_STATUS_UNPAY {
	// 	err = errors.New(common.ERR_VIRTUAL_ORDER_FORBIDDEN_CANCEL)
	// 	return
	// }

	status := common.ORDER_STATUS_CANCEL
	if v.Status == common.ORDER_STATUS_PAYED {
		status = common.ORDER_STATUS_REFUND
	}

	// 乐观锁确保避免多次调用
	res := db.Model(&v).Where("status in(?)", []int{common.ORDER_STATUS_UNPAY, common.ORDER_STATUS_PAYED}).Updates(base.Maps{"status": status})
	if err = res.Error; err != nil {
		glog.Error("OrderPay fail! err=", err)
		return
	}

	if 0 == res.RowsAffected {
		return
	}

	// 释放商品锁定库存
	// 修改相应产品规格库存及退款回调
	vquantity := []util.QuantitySt{{OrderId: v.Id, ProductId: v.ProductId, ProductSkuId: v.ProductSkuId, Quantity: v.Quantity}}
	// 支付失败 需相应减少已售数量 异步处理
	if err = util.AddProductModelQuantity(vquantity); err != nil {
		glog.Error("OrderPay fail! AsyncAddProductModelQuantity err=", err)
		return
	}

	// 处理后续流程
	switch status {
	case common.ORDER_STATUS_REFUND:
		if err = t.Refund(db, &v); err != nil {
			glog.Error("CancelOrders fail! Refund err=", err)
			return
		}
	}

	// 提交后刷新订单相关缓存
	db.AfterCommit(func() {
		o := &dao.Orders{}
		o.RefreshByOrderIds([]int64{id}) // 此处不要用v.Id 目的是不引用v,能尽快进行垃级回收
	})

	return
}

// 完成订单
func (t *Order) Finish(db *db.PsqlDB, id int64) (err error) {

	// 获取所支付的订单商品信息
	var v common.Orders
	if err = db.Find(&v, "status=? and id = ?", common.ORDER_STATUS_SHIPPED, id).Error; err != nil {
		glog.Error("Orders_Finish fail! err=", err)
		return
	}

	// // 验证用户资金密码
	// err = util.AuthUserZJPassword(user_id, args.ZJPassword)
	// if err != nil {
	// 	yf.JSON_Fail(c, common.ERR_ZJPASSWORD_INVALID)
	// 	return
	// }

	// 修改订单状态为已完成
	res := db.Model(&v).Where("status=?", common.ORDER_STATUS_SHIPPED).Updates(map[string]interface{}{"status": common.ORDER_STATUS_FINISH})
	if err = res.Error; err != nil {
		glog.Error("Finish fail! err=", err)
		return
	}

	if res.RowsAffected > 0 {
		//token, _ := util.GetHeaderString(c, "X-Yf-Token")
		ys := util.YBAssetStatus{OrderIds: []int64{v.Id}, Status: common.ASSET_POOL_FINISH, PublishArea: v.PublishArea}
		// 提交后刷新订单相关缓存
		db.AfterCommit(func() {
			// 异步通知 订单完成
			mq := common.MQUrl{Methond: "post", Uri: "/man/asset/payset", AppKey: "ybasset", Data: ys, MaxTrys: -1}
			if err = util.PublishMsg(mq); err != nil {
				glog.Error("Orders_Finish fail! YBAsset_PayStatus err=", err)
				return
			}

			o := &dao.Orders{}
			o.RefreshByOrderIds([]int64{id})
		})
	}

	return
}

// 退款处理
func (t *Order) Refund(db *db.PsqlDB, v *common.Orders) (err error) {
	res := db.Model(&common.Orders{}).Where("id=?", v.Id).Updates(map[string]interface{}{"status": common.ORDER_STATUS_REFUND})
	if err = res.Error; err != nil {
		glog.Error("Update product quantity fail! err=", err)
		return
	}
	if 0 == res.RowsAffected {
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

	id := v.Id
	// 提交后刷新订单相关缓存
	db.AfterCommit(func() {
		o := &dao.Orders{}
		o.RefreshByOrderIds([]int64{id})
	})
	return
}

func (t *Order) RefundById(db *db.PsqlDB, id int64) (err error) {
	var v common.Orders
	if err = db.Select("id, product_id, product_sku_id, quantity, publish_area").First(&v, "id=?", id).Error; err != nil {
		glog.Error("RefundById fail! err=", err)
		return
	}
	return t.Refund(db, &v)
}

// 获取订单id列表信息
func Orders_ListUserByIds(d *gorm.DB, user_id int64, ids []int64) (vs []common.Orders, err error) {
	vs = []common.Orders{}
	if len(ids) == 0 {
		return
	}
	if d == nil {
		d = db.GetDB().DB
	}

	if err = d.Where("id in(?) and status<?", ids, common.ORDER_STATUS_PAYED).Find(&vs).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("Orders_ListByIds fail! err=", err)
		return
	}
	return
}
