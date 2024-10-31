package man

import (
	"fmt"
	"time"
	"yunbay/yborder/common"
	"yunbay/yborder/dao"

	"github.com/gin-gonic/gin"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"

	"encoding/json"
	"regexp"
	"strings"
	"yunbay/yborder/server/share"
	"yunbay/yborder/util"
)

// 订单列表查询接口
func Orders_List(c *gin.Context) {
	status, _ := base.CheckQueryIntDefaultField(c, "status", -1)
	sale_status, _ := base.CheckQueryIntDefaultField(c, "sale_status", -1)
	str_ids, _ := base.CheckQueryStringField(c, "ids")
	buyer_userid, _ := base.CheckQueryInt64DefaultField(c, "buyer_userid", -1)
	seller_userid, _ := base.CheckQueryInt64DefaultField(c, "seller_userid", -1)
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	str_sorts, _ := base.CheckQueryStringField(c, "sorts")
	str_orders, _ := base.CheckQueryStringField(c, "orders")
	// publish_area, _ := base.CheckQueryIntDefaultField(c, "publish_area", 1)
	currency_type, _ := base.CheckQueryIntDefaultField(c, "currency_type", 1)

	var orders []string
	var sorts []string
	if str_orders != "" {
		orders = strings.Split(str_orders, ",")
	}
	if str_sorts != "" {
		sorts = strings.Split(str_sorts, ",")
	}
	db := db.GetDB()
	db.DB = db.Model(&common.Orders{})
	if status > 0 {
		db.DB = db.Where("status=?", status)
	} else {
		db.DB = db.Where("status>?", common.ORDER_STATUS_INIT)
	}
	if sale_status >= 0 {
		db.DB = db.Where("sale_status=?", sale_status)
	}
	if str_ids != "" {
		ids := base.StringToInt64Slice(str_ids, ",")
		if len(ids) > 0 {
			db.DB = db.Where("id in (?)", ids)
		}
	}
	if buyer_userid > -1 {
		db.DB = db.Where("user_id=?", buyer_userid)
	}
	if seller_userid > -1 {
		db.DB = db.Where("seller_userid=?", seller_userid)
	}
	if begin_date != "" {
		db.DB = db.Where("date>=?", begin_date)
	}
	if end_date != "" {
		db.DB = db.Where("date<=?", end_date)
	}
	// db.DB = db.Where("publish_area=?", publish_area)
	db.DB = db.Where("currency_type=?", currency_type)

	var count int64 = 0
	if err := db.Count(&count).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("Orders_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	var db2 *gorm.DB
	var totalResult TotalAmountResult
	db2 = db.DB
	db2.Select("sum(total_amount) as total_trading, sum(rebat_amount) as total_rebat").Scan(&totalResult)

	db.DB = db.ListPage(page, page_size)
	// 排序
	for i, v := range sorts {
		order := "desc"
		if len(order) > i && (orders[i] == "asc" || orders[i] == "desc") {
			order = orders[i]
		}
		db.DB = db.Order(fmt.Sprintf("%v %v", v, order))
	}

	vs := []common.Orders{}
	if err := db.Order("create_time desc").Find(&vs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("UserAssetDetail fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 处理订单商品的成本、售价及规格
	// for i := range vs {
	// 	p := vs[i].Product
	// 	switch currency_type {
	// 	case 0:
	// 		if skus, ok := p["skus"].([]interface{}); ok {
	// 			for j := range skus {
	// 				if s, ok := skus[j].(map[string]interface{}); ok {
	// 					id := s.()
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	// if err := share.GetOrdersProductInfos(&vs); err != nil {
	// 	glog.Error("GetOrdersProductInfos fail! err=", err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	list_ended := true
	if page_size == len(vs) {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended, "total": count, "total_amount": gin.H{"total_trading": totalResult.TotalTrading, "total_rebat": totalResult.TotalRebat}})
}

type TotalAmountResult struct {
	TotalTrading float64 // 累计总交易额
	TotalRebat   float64 // 累计总贡献值
}

// 订单详情
func Orders_Info(c *gin.Context) {
	order_id, _ := base.CheckQueryInt64Field(c, "order_id")
	if order_id == 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var v common.Orders
	if err := db.GetTxDB(c).Where("id=?", order_id).Find(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_ORDER_NOT_EXIST)
			return
		}
		glog.Error("Orders_Get fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// if err := share.GetOrderProductInfo(&v); err != err {
	// 	glog.Error("GetOrderProductInfo fail! err=", err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	yf.JSON_Ok(c, v)
}

type orderSt struct {
	OrderIds      []int64 `json:"order_ids" binding:"gt=0"`
	UserIds       []int64 `json:"user_ids"`
	SellerUserIds []int64 `json:"seller_user_ids"`
}

// 超时自动取消订单
func Orders_Cancel(c *gin.Context) {
	var req orderSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	fcancel := func(t *common.Orders) bool {
		if common.ORDER_STATUS_UNPAY == t.Status {
			// 只取消未支付状态下的订单
			return true
		}
		return false
	}

	t := share.Order{}
	for _, v := range req.OrderIds {
		db := db.GetTxDB(nil)
		if err := t.Cancel(db, v, fcancel); err != nil && err.Error() != common.ERR_ORDER_FORBIDDEN_CANCEL {
			glog.Error("cancel_orders fail! err=", err)
			yf.JSON_Fail(c, err.Error())
			db.Rollback()
			return
		}
		db.Commit()
	}

	o := &dao.Orders{}
	// 删除买家及卖家缓存
	for _, v := range req.UserIds {
		o.RefreshCache(v, 0)
	}
	for _, v := range req.SellerUserIds {
		o.RefreshCache(v, 1)
	}
	yf.JSON_Ok(c, gin.H{})
}

// 超时自动确认收货订单
func Orders_Finish(c *gin.Context) {
	var req orderSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	t := share.Order{}
	fail_ids := []int64{}
	for _, v := range req.OrderIds {
		d := db.GetTxDB(nil)
		if err := t.Finish(d, v); err != nil {
			glog.Error("Orders_Finish fail! err=", err)
			d.Rollback()
		}

		d.Commit()
	}

	// // 付款给卖家
	// v := util.YBAssetStatus{OrderIds: req.OrderIds, Status: common.ASSET_POOL_FINISH}
	// if err := util.YBAsset_PayStatus(v, ""); err != nil {
	// 	glog.Error("Orders_Finish fail! YBAsset_PayStatus err=", err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }

	o := &dao.Orders{}
	// 删除买家及卖家缓存
	for _, v := range req.UserIds {
		o.RefreshCache(v, 0)
	}
	for _, v := range req.SellerUserIds {
		o.RefreshCache(v, 1)
	}
	yf.JSON_Ok(c, gin.H{"fail_ids": fail_ids})
}

type orderidSt struct {
	OrderId int64 `json:"order_id"`
}

// 查询某开有订单状态的订单id
func Orders_StatusQuery(c *gin.Context) {
	date, _ := base.CheckQueryStringField(c, "date")
	status, _ := base.CheckQueryIntField(c, "status")

	var vs []orderidSt
	db := db.GetDB()
	if err := db.Model(&common.OrderStatus{}).Where("date=? and status=?", date, status).Select("order_id").Scan(&vs).Error; err != nil {
		glog.Error("Orders_StatusQuery fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	ids := []int64{}
	for _, v := range vs {
		ids = append(ids, v.OrderId)
	}
	yf.JSON_Ok(c, gin.H{"order_ids": ids})
}

type OrderRebatSt struct {
	OrderId int64 `json:"order_id" binding:"gt=0"`
	//Rebat float64  `json:"rebat" binding:"required,min=0,max=1"`
	TxHash string `json:"tx_hash" binding:"required"`
}

type PaidAffiche struct {
	Rebate       float64 `json:"rebate"`
	Username     string  `json:"username"`
	ProductTitle string  `json:"product_title"`
	UpdatedTime  int64   `json:"updated_time"`
}

// 更新订单折扣
func Orders_RebatUpdate(c *gin.Context) {
	var req OrderRebatSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	// 将获得hash值非(1～9)的字符去掉后，取末尾4位数字作为小数点后的数字以生成小数，再通过小数转换为百分数
	r, _ := regexp.Compile("[^1-9]")
	bt := r.ReplaceAll([]byte(req.TxHash), []byte{}) // 将非0-9的字符剔除
	str_rebat := string(bt[len(bt)-4:])
	rebat, _ := base.StringToFloat64(str_rebat)
	rebat /= 10000 // 转换成百分比
	if base.IsEqual(rebat, base.FLOAT_MIN) {
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	db := db.GetTxDB(c)
	var v common.Orders
	err := db.Find(&v, "id=? and publish_area=? and status>=? and extinfos->'rebat_hash' is null", req.OrderId, common.PUBLISH_AREA_REBAT, common.ORDER_STATUS_PAYED).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Orders_RebatUpdate fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	offer_amount := v.TotalAmount * (1 - rebat)
	// 更新订单折扣价格
	res := db.Model(&common.Orders{}).Where("id=? and publish_area=? and status>=? and extinfos->'rebat_hash' is null", req.OrderId, common.PUBLISH_AREA_REBAT, common.ORDER_STATUS_PAYED).
		Updates(map[string]interface{}{"rebat_amount": v.RebatAmount * rebat, "extinfos": gorm.Expr("extinfos || ?", fmt.Sprintf("{\"rebat\":%v, \"rebat_hash\":\"%v\", \"pay_amount\":%v,\"offer_amount\":%v}", rebat, req.TxHash, v.TotalAmount*rebat, offer_amount))})

	if res.Error != nil {
		glog.Error("Orders_RebatUpdate fail! err=", res.Error)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if 0 == res.RowsAffected {
		glog.Error("Orders_RebatUpdate no Rowsaffected! order_id:", req.OrderId)
		yf.JSON_Ok(c, gin.H{})
		return
	}

	// 退款用户折扣金额
	if err := util.YBAsset_UpdatePayAmount(req.OrderId, rebat); err != nil {
		glog.Error("Orders_RebatUpdate fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	o := &dao.Orders{}
	o.RefreshByOrderIds([]int64{req.OrderId}) // 刷新订单列表缓存 又更新折扣状态

	// 成交记录写入缓存
	ids := []int64{v.UserId}
	result, _ := util.GetUserInfoByIds(ids)
	var username string
	for _, value := range result {
		if account, ok := value.(util.Account); ok {
			if account.Username == "" { // 取电话
				var byteSlice []byte
				byteSlice = []byte(account.Tel)
				for i := 3; i < 7; i++ {
					byteSlice[i] = '*'
				}
				username = string(byteSlice)
			} else {
				username = account.Username
			}
		}
	}
	// 恭喜alice 以 11.25% 折扣获得了 SNET矿机
	// rebate username product_title => list
	var product common.Product
	binary, _ := json.Marshal(v.Product)
	json.Unmarshal(binary, &product)
	var paidAffiche PaidAffiche
	paidAffiche.Rebate = rebat
	paidAffiche.Username = username
	paidAffiche.ProductTitle = product.Title
	nowTime := time.Now()
	paidAffiche.UpdatedTime = nowTime.Unix()
	b, _ := json.Marshal(paidAffiche) // 结构体转json字符串
	redisCache, err := cache.GetWriter(common.RedisApi)
	if err != nil {
		glog.Error("Get redis cache fail! err=", err)
	} else {
		redisCache.LPush("paid_affiche_list", string(b))
	}

	yf.JSON_Ok(c, gin.H{})
}

func get_logistic_ids(ids []int64) (vs []common.Logistics, err error) {
	if err = db.GetDB().Select("id,company,number").Find(&vs, "id in(?)", ids).Error; err != nil {
		glog.Error("Orders_Report fail! err=", err)
		return
	}
	return
}

// 订单表格导出
func Orders_Report(c *gin.Context) {
	seller_userid, _ := base.CheckQueryInt64DefaultField(c, "seller_userid", -1)
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")

	ydb := db.GetDB()
	if begin_date != "" {
		ydb.DB = ydb.Where("date>=?", begin_date)
	}
	if end_date != "" {
		ydb.DB = ydb.Where("date<=?", end_date)
	}
	// 获取时间段内已发货的订单id列表
	vss := []common.OrderStatus{}
	if err := ydb.Select("id,order_id").Find(&vss, "status=?", common.ORDER_STATUS_SHIPPED).Error; err != nil {
		glog.Error("Orders_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	ids := []int64{}
	for _, v := range vss {
		ids = append(ids, v.OrderId)
	}
	ids = base.UniqueInt64Slice(ids)

	db := db.GetDB()
	if seller_userid > -1 {
		db.DB = db.Where("seller_userid=?", seller_userid)
	}

	// 排序
	vs := []common.Orders{}
	if err := db.Order("update_time asc,id asc").Find(&vs, "id in(?)", ids).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("UserAssetDetail fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	buf, err := share.OrdersReport(&vs)
	if err != nil {
		glog.Error("Orders_Report fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=订单报表(%s-%s).xlsx", begin_date, end_date))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// 添加订单信息
type payorders struct {
	ProductId    int64                  `json:"product_id"`
	ProductSkuId int64                  `json:"product_sku_id"`
	UserId       int64                  `json:"user_id"`
	AddressId    int64                  `json:"address_id"`
	PayType      int                    `json:"pay_type" binding:"min=0,max=3"`
	Amount       decimal.Decimal        `json:"amount"`
	Quantity     int                    `json:"quantity"`
	Extinfos     map[string]interface{} `json:"extinfos"`
	RewardYbt    decimal.Decimal        `json:"reward_ybt"`
	Price        decimal.Decimal        `json:"price"`
}

func Orders_CreateByLotterys(c *gin.Context) {
	var req payorders
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	var err error
	db := db.GetTxDB(c)

	o := common.Orders{UserId: req.UserId, ProductId: req.ProductId, Status: common.ORDER_STATUS_UNPAY, ProductSkuId: req.ProductSkuId, CurrencyType: req.PayType, ExtInfos: req.Extinfos, Quantity: req.Quantity}
	if err = share.UpsertOrders(db, &o); err != nil {
		glog.Error("Orders_CreateByLotterys fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	addrInfo := make(map[string]interface{})
	if req.AddressId > 0 {
		// 获取地址信息
		addr := common.UserAddress{Id: req.AddressId}
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

	o.TotalAmount, _ = req.Amount.Float64() // 订单总价为抽奖金额

	o.ExtInfos["pay_price"] = []common.PayPrice{common.PayPrice{Coin: common.GetCurrencyName(o.CurrencyType), PayType: o.CurrencyType, PredictYbt: req.RewardYbt, SalePrice: req.Price}} // 保存商品原价

	// 设置订单的收货信息及待支付状态
	infos := []common.Orders{o}
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

	country := util.GetCountry(c)
	// 修改地址及待付款状态 修改当时rmb->kt比例
	if err := db.Model(&common.Orders{}).Where("id =?", o.Id).Updates(common.Orders{Date: base.GetCurDay(), AddressInfo: addrInfo, Status: common.ORDER_STATUS_PAYED, Country: country}).Error; err != nil {
		glog.Error("Orders_PrePay fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 下单支付 将活动计划的总价/奖品数量 支付给商家
	amount, _ := req.Price.Float64()
	if err = util.YBAsset_LotterysConfirm(util.LotterysPay{PayAmount: amount, CurrencyType: o.CurrencyType, OrderId: o.Id, SellerUserId: o.SellerUserId}); err != nil {
		glog.Error("Orders_PrePay fail! err", err)
		yf.JSON_Fail(c, err.Error())
		return
	}

	db.AfterCommit(func() {
		(&dao.Orders{}).RefreshCache(req.UserId, 0) // 更新用户订单列表缓存
	})

	yf.JSON_Ok(c, gin.H{"id": o.Id})
}
