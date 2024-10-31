package client

import (
	"fmt"
	"net/url"
	"time"
	"yunbay/yborder/common"
	"yunbay/yborder/conf"
	"yunbay/yborder/dao"
	"yunbay/yborder/server/share"
	"yunbay/yborder/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

type shipSt struct {
	OrderId int64  `json:"order_id" binding:"gt=0"`
	Company string `json:"company"`
	Number  string `json:"number"`
}

// 卖家设置已发货状态
func Orders_Shipped(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	user_type, _ := util.GetUserType(c)
	if user_type < common.USER_TYPE_SELLER {
		glog.Error("Orders_Shipped fail! user_type=", user_type)
		yf.JSON_Fail(c, common.ERR_ORDER_FORBIDDENT_MODIFY)
		return
	}
	var args shipSt
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	var v common.Orders
	db := db.GetTxDB(c)
	if err := db.Where("seller_userid=? and id=?", user_id, args.OrderId).Find(&v).Error; err != nil {
		yf.JSON_Fail(c, common.ERR_ORDER_NOT_EXIST)
		return
	}
	// 只有订单状态在已付款状态下才能更改为已发货状态
	if v.Status != common.ORDER_STATUS_PAYED {
		glog.Error("Orders_Shipped fail! status=", v.Status)
		yf.JSON_Fail(c, common.ERR_ORDER_FORBIDDENT_MODIFY)
		return
	}
	// 添加物流信息
	lg := common.Logistics{OrderId: args.OrderId, UserId: user_id, Company: args.Company, Number: args.Number}
	if err := db.Save(&lg).Error; err != nil {
		glog.Error("Orders_Shipped fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	v.Status = common.ORDER_STATUS_SHIPPED
	v.LogisticsId = lg.Id

	duration, _ := time.ParseDuration(conf.Config.Orders.AutoFinishTime)
	v.AutoFinishTime = time.Now().Unix() + int64(duration.Seconds())
	// 修改订单状态为已发货状态 设置自动确认收货时间
	if err := db.Model(&v).Updates(map[string]interface{}{"status": v.Status, "auto_finish_time": v.AutoFinishTime, "logistics_id": v.LogisticsId, "update_time": v.UpdateTime}).Error; err != nil {
		glog.Error("Orders_Shipped fail! status=", v.Status)
		yf.JSON_Fail(c, common.ERR_ORDER_FORBIDDENT_MODIFY)
		return
	}
	// ybt专区的商品在商家发货后立即完成此笔订单
	if v.PublishArea == common.CURRENCY_YBT {
		ys := util.YBAssetStatus{OrderIds: []int64{args.OrderId}, Status: common.ASSET_POOL_FINISH, PublishArea: v.PublishArea}
		mq := common.MQUrl{Methond: "post", Uri: "/man/asset/payset", AppKey: "ybasset", Data: ys, MaxTrys: -1}
		if err := util.PublishMsg(mq); err != nil {
			glog.Error("YBAsset_PayStatus fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}

	db.AfterCommit(func() {
		o := &dao.Orders{}
		o.RefreshByOrderIds([]int64{args.OrderId})
	})

	yf.JSON_Ok(c, v)
}

type Payproducts struct {
	ProductId    int64 `json:"product_id"`
	ProductSkuId int64 `json:"product_sku_id"`
	Quantity     int   `json:"quantity"`
}

// 添加订单信息
type payorders struct {
	//OrderIds []int64 `json:"order_ids"`
	CartIds   []int64                `json:"cart_ids"`
	Products  []Payproducts          `json:"products"`
	AddressId int64                  `json:"address_id"`
	PayType   int                    `json:"pay_type" binding:"min=0,max=3"`
	Amount    decimal.Decimal        `json:"amount"`
	Quantity  int                    `json:"quantity"`
	Extinfos  map[string]interface{} `json:"extinfos"`
}

// 提交订单待支付
func Orders_PrePay(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args payorders
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	country := util.GetCountry(c)
	db := db.GetTxDB(c)
	order_ids := []int64{}
	args.CartIds = base.UniqueInt64Slice(args.CartIds)
	pay_type := args.PayType
	// 先将商品信息添加到订单
	if len(args.Products) > 0 {
		for _, v := range args.Products { // 生成商品订单
			if 0 == v.Quantity { // 商品数量默认为1
				v.Quantity = 1
			}
			o, err := snapOrder(db, user_id, v.ProductId, v.ProductSkuId, v.Quantity, pay_type, args.Extinfos)
			if err != nil {
				if err.Error() == common.ERR_FORBIDDEN_BUY_OWNGOODS {
					yf.JSON_Fail(c, err.Error())
					return
				}
				glog.Error("Orders_PrePay fail! err", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			order_ids = append(order_ids, o.Id) // 将订单添加到订单列表
		}
	}
	// 将购物车的商品添加到订单
	if len(args.CartIds) > 0 {
		carts, err := ListCartByIds(user_id, args.CartIds)
		if err != nil {
			glog.Error("Orders_PrePay fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
		if len(carts) != len(args.CartIds) {
			// 刷新下购物车缓存
			(&dao.Cart{}).RefreshCache(user_id)
			glog.Error("Orders_PrePay fail! err=len(carts) == len(args.CartIds)")
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
		for _, v := range carts {
			o, err := snapOrder(db, user_id, v.ProductId, v.ProductSkuId, v.Quantity, pay_type, args.Extinfos)
			if err != nil {
				if err.Error() == common.ERR_FORBIDDEN_BUY_OWNGOODS {
					yf.JSON_Fail(c, err.Error())
					return
				}
				glog.Error("Orders_PrePay fail! err", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			order_ids = append(order_ids, o.Id) // 将订单添加到订单列表
		}
	}

	//order_ids = base.UniqueInt64Slice(order_ids) // 去重id
	if len(order_ids) == 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	addrInfo := make(map[string]interface{})
	if args.AddressId > 0 {
		// 获取地址信息
		addr := common.UserAddress{Id: args.AddressId}
		if err := db.Find(&addr).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				yf.JSON_Fail(c, common.ERR_ADDRESS_NOT_EXIST)
				return
			}
			glog.Error("UserAddress fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
		addrInfo = base.StructToMap(addr)
	}

	// 设置订单的收货信息及待支付状态
	infos, err := share.Orders_ListUserByIds(db.DB, user_id, order_ids)
	if err != nil {
		glog.Error("Orders_ListUserByIds fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if len(infos) != len(order_ids) {
		glog.Error("order infos fail! len(infos)=", len(infos), " len(args.OrderIds):", len(order_ids))
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 保存当前商品一些快照信息
	if reason, err := share.SnapOrdersProducts(db, infos); err != nil {
		glog.Error("SnapOrdersProducts fail! err=", err)
		if reason != "" {
			yf.JSON_Fail(c, reason)
			return
		}
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 判断当时总支付金额是否一致
	// 获取当时rmvb->kt的比例
	// 获取rmb->kt汇率
	var fKtRatio float64 = 1
	if pay_type == common.CURRENCY_KT {
		fKtRatio, err = share.GetRatioCache(pay_type, common.CURRENCY_RMB)
		if err != nil {
			glog.Error("GetRatioCache fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}
	// 核对订单金额与支付金额是否一致
	amount := float64(0)
	for _, v := range infos {
		amount += v.TotalAmount
	}
	amount = amount * fKtRatio
	// 如果当时汇率波动 导致价格有变化 则需要提示重新获取最新汇率
	//if !base.IsEqual(amount, args.Amount) {
	if args.Amount.Sub(decimal.NewFromFloat(amount)).Abs().GreaterThan(decimal.NewFromFloat32(0.000001)) {
		glog.Error("PrePay fail! pay amount:", args.Amount, " orders amount:", amount)
		yf.JSON_Fail(c, common.ERR_ORDER_AMOUNT_INVALID)
		return
	}

	// 设置订单锁定时间(待支付到已支付时间,到期需自动关闭)
	duration, _ := time.ParseDuration(conf.Config.Orders.AutoCancelTime)
	auto_cancel_time := int64(duration.Seconds())
	if auto_cancel_time > 0 {
		auto_cancel_time += time.Now().Unix()
	}

	today := time.Now().Format("2006-01-02")
	// 修改地址及待付款状态 修改当时rmb->kt比例
	if err := db.Model(&common.Orders{}).Where("id in(?)", order_ids).Updates(common.Orders{Date: today, AddressInfo: addrInfo, Status: common.ORDER_STATUS_UNPAY, CurrencyType: pay_type, CurrencyPercent: fKtRatio, AutoCancelTime: auto_cancel_time, Country: country}).Error; err != nil {
		glog.Error("Orders_PrePay fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 清除购物车
	if len(args.CartIds) > 0 {
		cart := &dao.Cart{db}
		if err := cart.DelByIds(user_id, args.CartIds); err != nil {
			glog.Error("Orders_PrePay fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}

	db.AfterCommit(func() {
		o := &dao.Orders{}
		o.RefreshByOrderIds(order_ids)

		if auto_cancel_time > 0 {
			// 定时取消订单
			type orderSt struct {
				OrderIds []int64 `json:"order_ids"`
			}
			noti := orderSt{OrderIds: order_ids}
			err = util.DeferedPublishMsg(common.MQUrl{Methond: "POST", AppKey: "yborder", Uri: "/man/order/cancel", Data: noti, MaxTrys: -1}, duration)
			if err != nil {
				glog.Error("Orders_PrePay fail! err=", err)
				return
			}
		}
	})

	yf.JSON_Ok(c, gin.H{"order_ids": order_ids, "amount": amount})
}

// 生成快照订单
func snapOrder(db *db.PsqlDB, user_id, product_id, product_modelid int64, quantity int, pay_type int, extinfos map[string]interface{}) (o common.Orders, err error) {
	o = common.Orders{UserId: user_id, ProductId: product_id, Status: common.ORDER_STATUS_UNPAY, ProductSkuId: product_modelid, CurrencyType: pay_type, ExtInfos: extinfos, Quantity: quantity}
	// 保存当前商品一些快照信息
	// if err = share.GetSnapOrdersProduct(db, &o); err != nil {
	// 	glog.Errorf("OrdersProductSnap err=%v", err)
	// 	return
	// }

	if err = share.UpsertOrders(db, &o); err != nil {
		glog.Error("Orders_PrePay fail! err=", err)
		return
	}
	return
}

func Orders_GetById(d *gorm.DB, id int64) (v common.Orders, err error) {
	if d == nil {
		d = db.GetDB().DB
	}

	if err = d.Where("id=?", id).Find(&v).Error; err != nil {
		glog.Error("Orders_ListByIds fail! err=", err)
		return
	}
	return
}

// 获取买家的售出订单
func Orders_List(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	status, _ := base.CheckQueryIntDefaultField(c, "status", 1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	// vs := []common.Orders{}

	o := &dao.Orders{}
	v, err := o.List(user_id, 0, status, page, page_size)
	if err != nil {
		glog.Error("Orders_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, v)
}

// 获取卖家的售出订单
func Order_SellerList(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	user_type, _ := util.GetUserType(c)
	if user_type < common.USER_TYPE_SELLER {
		glog.Error("Orders_Shipped fail! user_type=", user_type)
		yf.JSON_Fail(c, common.ERR_USER_TYPE_NOT_SELLER)
		return
	}
	status, _ := base.CheckQueryIntDefaultField(c, "status", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	o := &dao.Orders{}
	v, err := o.List(user_id, 1, status, page, page_size)
	if err != nil {
		glog.Error("Order_SellerList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, v)
}

// 订单查询
func Order_SellerSearchList(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	user_type, _ := util.GetUserType(c)
	if user_type < common.USER_TYPE_SELLER {
		glog.Error("Orders_Shipped fail! user_type=", user_type)
		yf.JSON_Fail(c, common.ERR_USER_TYPE_NOT_SELLER)
		return
	}
	//user_id := 51903
	status, _ := base.CheckQueryIntDefaultField(c, "status", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	buy_uid, _ := base.CheckQueryInt64Field(c, "user_id")
	id, _ := base.CheckQueryInt64Field(c, "id")
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")
	export, _ := base.CheckQueryIntField(c, "export")

	db := db.GetDB()
	db.DB = db.Where("seller_userid=?", user_id)
	if status > -1 {
		db.DB = db.Where("status=?", status)
	}
	if id > 0 {
		db.DB = db.Where("id=?", id)
	}
	if buy_uid > 0 {
		db.DB = db.Where("user_id=?", buy_uid)
	}
	if begin_date != "" {
		db.DB = db.Where("date>=?", begin_date)
	}
	if end_date != "" {
		db.DB = db.Where("date<=?", end_date)
	}

	vs := []common.Orders{}
	// 导出报表
	if export > 0 {
		if err := db.Order("id desc").Find(&vs).Error; err != nil {
			glog.Error("Order_SellerSearchList fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}

		data, err := share.OrdersReport(&vs)
		if err != nil {
			glog.Error("Order_SellerSearchList fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=%s(%s-%s).xlsx", url.QueryEscape("订单报表"), begin_date, end_date))
		c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data.Bytes())
		return
	}

	var total int
	var err error
	if err = db.Model(&common.Orders{}).Count(&total).Error; err != nil {
		glog.Error("Order_SellerSearchList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err = db.ListPage(page, page_size).Order("id desc").Find(&vs).Error; err != nil {
		glog.Error("Order_SellerSearchList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 总交易额及贡献值
	type tOrders struct {
		TotalAmount float64 `json:"total_amount"`
		RebatAmount float64 `json:"rebat_amount"`
	}
	var t tOrders
	if err = db.Model(&common.Orders{}).Select("sum(quantity*(product->>'price')::double precision) as total_amount, sum(rebat_amount) as rebat_amount").Scan(&t).Error; err != nil {
		glog.Error("Order_SellerSearchList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": base.IsListEnded(page, page_size, len(vs), total), "total": total, "info": t})

}

// 订单详情
func Orders_Info(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	order_id, _ := base.CheckQueryInt64Field(c, "order_id")

	var v common.Orders

	db := db.GetDB()
	if err := db.Where("shield=0 and id=? and (user_id=? or seller_userid=?)", order_id, user_id, user_id).Find(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_ORDER_NOT_EXIST)
			return
		}
		glog.Error("Orders_Get fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if v.Status < common.ORDER_STATUS_PAYED {
		v.Now = time.Now().Unix()
	}
	if common.ORDER_STATUS_FINISH == v.Status {
		if common.GOODS_TYPE_CARD == v.ProductType {
			// 获取卡密信息
			vs := []common.OfCard{}
			if err := db.Find(&vs, "order_id=?", v.Id).Error; err != nil {
				glog.Error("Orders_Card fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			v.ExtInfos["cards"] = vs
		}
	}

	// if err := share.GetOrderProductInfo(&v); err != nil {
	// 	glog.Error("Orders_List fail! err=", err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	yf.JSON_Ok(c, v)
}

// 获取该订单卡密信息
func Orders_Card(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req idParams
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	// 验证该订单号属于该用户
	db := db.GetDB()
	var v common.Orders
	if err := db.Select("id").Find(&v, "user_id=? and id=?", user_id, req.OrderId).Error; err != nil {
		reason := yf.ERR_SERVER_ERROR
		if err == gorm.ErrRecordNotFound {
			reason = yf.ERR_NOT_EXISTS
		}
		yf.JSON_Fail(c, reason)
		return
	}

	vs := []common.OfCard{}
	if err := db.Find(&vs, "order_id=?", v.Id).Error; err != nil {
		glog.Error("Orders_Card fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"list": vs})
}

func Order_RebatInfo(c *gin.Context) {
	order_id, err := base.CheckQueryInt64Field(c, "order_id")
	if err != nil {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	var v common.Orders

	if err := db.GetTxDB(c).Where("shield=0 and id=?", order_id).Find(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_ORDER_NOT_EXIST)
			return
		}
		glog.Error("Orders_Get fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, v.ExtInfos)
}

type idParams struct {
	OrderId int64 `json:"order_id" form:"order_id" binding:"gt=0"`
}

// 买家取消订单
func Order_Cancel(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args idParams
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	db := db.GetTxDB(c)
	db.DB = db.Where("user_id=?", user_id)
	var order share.Order
	fcancel := func(t *common.Orders) bool {
		if t.AutoDeliver && t.Status != common.ORDER_STATUS_UNPAY {
			// 虚拟订单支付后用户不能取消
			return false
		}
		return true
	}
	if err := order.Cancel(db, args.OrderId, fcancel); err != nil {
		glog.Error("Order_Cancel fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

type idsParam struct {
	OrderIds []int64 `json:"order_ids" binding:"gt=0"`
}

// 买家删除订单
func Order_Del(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args idsParam
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	args.OrderIds = base.UniqueInt64Slice(args.OrderIds) // 去重
	if len(args.OrderIds) == 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	// 屏蔽订单
	db := db.GetDB().Model(&common.Orders{}).Where("shield=0 and user_id=? and status in(?) and id in(?)", user_id, []int{common.ORDER_STATUS_INIT, common.ORDER_STATUS_UNPAY, common.ORDER_STATUS_FINISH, common.ORDER_STATUS_CANCEL, common.ORDER_STATUS_REFUND}, args.OrderIds).
		Updates(map[string]interface{}{"shield": 1})
	if db.Error != nil {
		glog.Error("Order_Del fail! err=", db.Error)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if db.RowsAffected > 0 {
		o := &dao.Orders{}
		o.RefreshByOrderIds(args.OrderIds)
	}

	yf.JSON_Ok(c, gin.H{})
}

// 卖家取消订单
func Order_SellerCancel(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args idParams
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}

	var v share.Order
	db := db.GetTxDB(c)
	db.DB = db.Where("shield=0 and seller_userid=?", user_id)
	if err := v.Cancel(db, args.OrderId, nil); err != nil {
		glog.Error("Order_SellerCancel fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

type aftersale struct {
	OrderId int64 `json:"order_id" binding:"gt=0"`
	Finish  bool  `json:"finish"`
}

type smsContent struct {
	UserIds []int64 `json:"user_ids"`
	Content string  `jsn:"content"`
}

func Order_AfterSale(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args aftersale
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	var v common.Orders
	db := db.GetTxDB(c)
	if err := db.Where("shield=0 and id=? and (user_id=?)", args.OrderId, user_id).Find(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_ORDER_NOT_EXIST)
			return
		}
		glog.Error("Order_AfterSale fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	status := common.SALE_STATUS_ING
	if args.Finish {
		status = common.SALE_STATUS_END
	}
	if err := db.Model(&v).Updates(map[string]interface{}{"sale_status": status, "update_time": time.Now().Unix()}).Error; err != nil {
		glog.Error("Order_AfterSale fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	}

	db.AfterCommit(func() {
		o := &dao.Orders{}
		o.RefreshByOrderIds([]int64{args.OrderId})
	})

	// 售后中发送短信通知商家
	if status == common.SALE_STATUS_ING {
		c := smsContent{UserIds: []int64{v.SellerUserId}, Content: conf.Config.SmsText["aftersale"]}
		mq := common.MQUrl{Methond: "post", Uri: "/man/account/sms/send", AppKey: "account", Data: c}
		if err := util.PublishMsg(mq); err != nil {
			glog.Error("Order_AfterSale fail! send msg fail!")
		}
	}
	yf.JSON_Ok(c, gin.H{})
}

// 获取买家售后列表
func Orders_AfterSaleList(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	vs, _, err := getAfterList(user_id, false, page, page_size)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Orders_AfterSaleList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if page_size == len(vs) {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
}

// 获取卖家售后列表
func Orders_SellerAfterSaleList(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	vs, total, err := getAfterList(user_id, true, page, page_size)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Orders_SellerAfterSaleList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": base.IsListEnded(page, page_size, len(vs), total), "total": total})
}

func getAfterList(user_id int64, seller bool, page, page_size int) (vs []common.Orders, total int, err error) {
	vs = []common.Orders{}
	db := db.GetDB()
	db.DB = db.Where("shield=0 and sale_status>0")
	if seller {
		db.DB = db.Where("seller_userid=?", user_id)
	} else {
		db.DB = db.Where("user_id=?", user_id)
	}
	if seller {
		if err = db.Model(&common.Orders{}).Count(&total).Error; err != nil {
			glog.Error("getAfterList fail! err=", err)
			return
		}
	}
	if err = db.ListPage(page, page_size).Order("create_time desc").Find(&vs).Error; err != nil {
		glog.Error("getAfterList fail! err=", err)
		return
	}

	// if err = share.GetOrdersProductInfos(&vs); err != nil {
	// 	glog.Error("getAfterList fail! err=", err)
	// 	return
	// }
	return
}

// 获取买家订单状态计数
func Orders_Count(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	status, _ := base.CheckQueryIntDefaultField(c, "status", 0)
	vs, salecount, err := get_orders_typecount(user_id, false, status)
	if err != nil {
		glog.Error("Orders_Count fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "sale_count": salecount})
}

// 获取卖家订单状态计数
func Orders_SellerCount(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	status, _ := base.CheckQueryIntDefaultField(c, "status", 0)
	vs, salecount, err := get_orders_typecount(user_id, true, status)
	if err != nil {
		glog.Error("Orders_SellerCount fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "sale_count": salecount})
}

func get_orders_typecount(user_id int64, seller bool, status int) (vs []dao.OrderStatusCntSt, sale_status_count int, err error) {
	iseller := 0
	if seller {
		iseller = 1
	}
	o := &dao.Orders{}
	vs, err = o.GetOrderStatusCount(user_id, iseller)
	if err != nil {
		glog.Error("get_orders_typecount fail! GetOrderStatusCount err=", err)
		return
	}

	sale_status_count, err = o.GetOrderAfterSaleCount(user_id, iseller)
	return
}

func Orders_ListByProduct(c *gin.Context) {
	product_id, err := base.CheckQueryInt64Field(c, "product_id")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	if err != nil {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	o := &dao.Orders{}
	v, err := o.GetProductOrders(product_id, page, page_size)
	if err != nil {
		glog.Error("Orders_ListByProduct fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	vs := v.List
	// 获取用户信息
	user_ids := []int64{}
	for _, v := range vs {
		user_ids = append(user_ids, v.UserId)
	}
	m, err := util.GetUserInfoByIds(user_ids)
	if err != nil {
		glog.Error("Orders_ListByProduct fail! err=", err)
	}
	for i, v := range vs {
		if u, ok := m[v.UserId]; ok {
			vs[i].UserInfo = u
		}
	}
	v.List = vs
	yf.JSON_Ok(c, v)
}
