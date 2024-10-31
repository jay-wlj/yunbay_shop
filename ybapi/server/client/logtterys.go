package client

import (
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybapi/dao"
	"yunbay/ybapi/server/share"
	"yunbay/ybapi/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func Lotterys(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	v := dao.Lotterys{}
	ret, err := v.List(page, page_size)
	if err != nil {
		glog.Error("Lotterys fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	now := time.Now().Unix()
	for i := range ret.List {
		ret.List[i].Now = now
	}

	yf.JSON_Ok(c, ret)
}

func Lotterys_Info(c *gin.Context) {
	id, _ := base.CheckQueryInt64Field(c, "id")
	if id <= 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var v common.Lotterys
	if err := db.GetDB().Find(&v, "id=?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, yf.ERR_NOT_EXISTS)
			return
		}
		glog.Error("Lotterys_Info fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 获取商品详情
	p, err := util.GetProductInfo(v.PId, 0)
	if err != nil {
		glog.Error("Lotterys_Info fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}
	v.Product = base.FilterStruct(p, true, "title", "images", "descimgs", "contact", "info", "publish_area").(map[string]interface{})
	v.Now = time.Now().Unix()
	yf.JSON_Ok(c, v)
}

func Lotterys_Record(c *gin.Context) {
	id, _ := base.CheckQueryInt64Field(c, "id")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	if id <= 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	v := dao.Lotterys{}
	ret, err := v.ListRecord(id, page, page_size)
	if err != nil {
		glog.Error("Lotterys fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, ret)
}

func Lotterys_SelfRecord(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	status, _ := base.CheckQueryIntDefaultField(c, "status", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	v := dao.Lotterys{}
	ret, err := v.ListUserRecord(user_id, status, page, page_size)
	if err != nil {
		glog.Error("Lotterys fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, ret)
}

type payLotterys struct {
	share.LotterysPayParmas
	Token      string `json:"-"`
	ZJPassword string `json:"zjpassword" binding:"gt=0"`
}

func Lotterys_Pay(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req payLotterys
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	req.From = user_id
	// 验证用户资金密码
	req.Token, _ = util.GetHeaderString(c, "X-Yf-Token")
	err := util.AuthUserZJPassword(req.Token, req.ZJPassword)
	if err != nil {
		yf.JSON_Fail(c, common.ERR_ZJPASSWORD_INVALID)
		return
	}

	db := db.GetTxDB(c)
	var o share.Lotterys
	id, err := o.Pay(db, req.LotterysPayParmas)
	if err != nil {
		glog.Error("Order_Pay fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}
	db.Commit()
	yf.JSON_Ok(c, gin.H{"id": id})
}

type confirmLotterys struct {
	LotterysRecordId int64                  `json:"lotterys_record_id"`
	AddressId        int64                  `json:"address_id"`
	Extinfos         map[string]interface{} `json:"extinfos"`
}

// 确认订单
func Lotterys_Confirm(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req confirmLotterys
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	db := db.GetTxDB(c)

	var v common.LotterysRecord
	var err error
	if err = db.Find(&v, "id=? and user_id=? and status=?", req.LotterysRecordId, user_id, common.LOTTERYS_STATUS_YES).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, yf.ERR_NOT_EXISTS)
			return
		}
		glog.Error("Lotterys_Confirm fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if v.OrderStatus == common.STATUS_OK {
		// 已确认订单
		yf.JSON_Ok(c, gin.H{})
		return
	}

	res := db.Model(&common.LotterysRecord{}).Where("id=? and status=? and order_status=?", req.LotterysRecordId, common.LOTTERYS_STATUS_YES, common.STATUS_INIT).Updates(base.Maps{"order_status": common.STATUS_OK})
	if err = res.Error; err != nil {
		glog.Error("Lotterys_Confirm fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if 0 == res.RowsAffected {
		yf.JSON_Ok(c, gin.H{})
		return
	}
	// 获取活动详情
	var ds dao.Lotterys
	ls, e := ds.Get(v.LotterysId)
	if err = e; err != nil {
		glog.Error("Lotterys_Confirm fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	o := util.Payorders{ProductId: ls.PId, UserId: user_id, Quantity: 1, PayType: ls.Coin, Extinfos: req.Extinfos, RewardYbt: ls.RewardYbt, Price: ls.Price, Amount: ls.Amount, AddressId: req.AddressId}
	if o.Extinfos == nil {
		o.Extinfos = make(map[string]interface{})
	}
	o.Extinfos["lotterys_record_id"] = req.LotterysRecordId
	o.Extinfos["lotterys_id"] = v.LotterysId
	var order_id int64
	if order_id, err = o.Do(); err != nil {
		glog.Error("Lotterys_Confirm fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}

	// // 先将商品信息添加到订单
	// o, err := snapOrder(db, user_id, ls.PId, 0, 1, ls.Coin, req.Extinfos)
	// if err != nil {
	// 	if err.Error() == common.ERR_FORBIDDEN_BUY_OWNGOODS {
	// 		yf.JSON_Fail(c, err.Error())
	// 		return
	// 	}
	// 	glog.Error("Orders_PrePay fail! err", err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }

	// addrInfo := make(map[string]interface{})
	// if req.AddressId > 0 {
	// 	// 获取地址信息
	// 	addr := common.UserAddress{Id: req.AddressId}
	// 	if err := db.Find(&addr).Error; err != nil {
	// 		if err == gorm.ErrRecordNotFound {
	// 			yf.JSON_Fail(c, common.ERR_ADDRESS_NOT_EXIST)
	// 			return
	// 		}
	// 		glog.Error("UserAddress fail! err=", err)
	// 		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 		return
	// 	}
	// 	addrInfo = base.StructToMap(addr)
	// }

	// o.TotalAmount, _ = ls.Amount.Float64() // 订单总价为抽奖金额

	// o.ExtInfos["pay_price"] = []common.PayPrice{common.PayPrice{Coin: common.GetCurrencyName(o.CurrencyType), PayType: o.CurrencyType, PredictYbt: ls.RewardYbt, SalePrice: ls.Price}} // 保存商品原价

	// // 设置订单的收货信息及待支付状态
	// infos := []common.Orders{o}
	// // 保存当前商品一些快照信息
	// if reason, err := share.SnapOrdersProducts(db, infos); err != nil {
	// 	glog.Error("SnapOrdersProducts fail! err=", err)
	// 	if reason != "" {
	// 		yf.JSON_Fail(c, reason)
	// 		return
	// 	}
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }

	// // 修改地址及待付款状态 修改当时rmb->kt比例
	// if err := db.Model(&common.Orders{}).Where("id =?", o.Id).Updates(common.Orders{Date: base.GetCurDay(), AddressInfo: addrInfo, Status: common.ORDER_STATUS_PAYED, Country: country}).Error; err != nil {
	// 	glog.Error("Orders_PrePay fail! err=", err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }

	// // 下单支付 将活动计划的总价/奖品数量 支付给商家
	// amount, _ := ls.Price.Float64()
	// if err = util.YBAsset_LotterysConfirm(util.LotterysPay{PayAmount: amount, CurrencyType: ls.Coin, OrderId: o.Id, SellerUserId: o.SellerUserId}); err != nil {
	// 	glog.Error("Orders_PrePay fail! err", err)
	// 	yf.JSON_Fail(c, err.Error())
	// 	return
	// }

	lotterys_id := v.LotterysId
	// 提交后刷新订单相关缓存
	db.AfterCommit(func() {
		ds.RefreshRecord(lotterys_id, user_id) // 更新抽奖记录缓存
	})
	yf.JSON_Ok(c, gin.H{"order_id": order_id})
}

func Lotterys_Key(c *gin.Context) {
	s, _ := dao.GetUniqueKey()
	yf.JSON_Ok(c, gin.H{"key": s})
}
