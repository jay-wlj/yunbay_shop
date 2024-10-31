package client

import (
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybasset/util"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

type voucherPay struct {
	VoucherId  int64   `json:"voucher_id" binding:"gt=0"`
	PayAmount  float64 `json:"pay_amount" binding:"gt=0"`
	UserId     int64   `json:"user_id" validte:"gt=0"`
	ZJPassword string  `json:"zjpassword" binding:"required"`
}

// 消费券支付
func Voucher_Pay(c *gin.Context) {
	var req voucherPay
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	// 不能转帐给自己
	if user_id == req.UserId {
		yf.JSON_Fail(c, common.ERR_FORBIDDEN_TRANSFER_TO_OWN)
		return
	}
	// 验证支付密码
	token, _ := util.GetHeaderString(c, "X-Yf-Token")
	if err := util.AuthUserZJPassword(token, req.ZJPassword); err != nil {
		yf.JSON_Fail(c, err.Error())
		return
	}
	// 获取消费券
	var v common.Voucher
	var err error
	db := db.GetTxDB(c)
	if err = db.Find(&v, "id=? and user_id=?", req.VoucherId, user_id).Error; err != nil {
		glog.Error("Voucher_Pay fail! err=", err)
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_VOUCHER_NOT_EXIST)
			return
		}
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if v.Amount < req.PayAmount {
		yf.JSON_Fail(c, common.ERR_AMOUNT_EXCEED)
		return
	}

	now := time.Now().Unix()
	vr := common.VoucherRecord{VoucherId: req.VoucherId, ToUid: req.UserId, Amount: -req.PayAmount, CreateTime: now, UpdateTime: now}
	if err = db.Save(&vr).Error; err != nil { // 从消费券中消费指定金额
		glog.Error("Voucher_Pay fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 将指定金额转入商家帐号中
	today := time.Now().Format("2006-01-02")
	ua := common.UserAssetDetail{UserId: req.UserId, Type: v.Type, TransactionType: common.TRANSACTION_TRANSFER, Amount: req.PayAmount, Date: today}
	if err = db.Save(&ua).Error; err != nil {
		glog.Error("Voucher_Pay fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	snowflake.NewNode(1)
	// 添加资产出入帐记录
	country := util.GetCountry(c)
	asset_recharge := common.RechargeFlow{UserId: req.UserId, Channel: common.CHANNEL_CHAIN, FlowType: common.FLOW_TYPE_YUNBAY, AssetId: ua.Id, TxType: v.Type, Amount: req.PayAmount, Country: country, Date: today, TxHash: util.GetSnowflake().Generate().String()}
	if err = db.Save(&asset_recharge).Error; err != nil {
		glog.Error("Voucher_Pay fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	asset_withdraw := common.WithdrawFlow{UserId: user_id, ToUserId: req.UserId, TxType: v.Type, Amount: req.PayAmount, Channel: common.CHANNEL_CHAIN, FlowType: common.FLOW_TYPE_YUNBAY, Status: common.TX_STATUS_SUCCESS, Country: country, Date: today, CreateTime: now, UpdateTime: now, Maner: "system"}
	if err = db.Save(&asset_withdraw).Error; err != nil {
		glog.Error("Voucher_Pay fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err = util.PublishMsg(common.MQUrl{Methond: "POST", AppKey: "ybasset", Uri: "/man/voucher/record/update", Data: idSt{vr.Id}}); err != nil {
		glog.Error("Voucher_Pay fail! Voucher_Pay err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

// 获取代金券详情
func Voucher_Info(c *gin.Context) {
	_type, _ := base.CheckQueryInt64Field(c, "type")
	if 0 == _type {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	db := db.GetDB()
	info := common.VoucherInfo{}
	if err := db.Find(&info, "type=?", _type).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
			return
		}
		glog.Error("Voucher_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	//v := base.SelectStructView(info, "not man")
	v := base.FilterStruct(info, false, "create_time", "update_time", "id")
	yf.JSON_Ok(c, v)
	return
}

// 获取用户的代金券列表
func Voucher_List(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	db := db.GetDB()

	vis := []common.VoucherInfo{}
	if err := db.Order("type asc").Find(&vis).Error; err != nil {
		glog.Error("Voucher_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	vs := []common.Voucher{}
	if err := db.Order("type asc").Find(&vs, "user_id=?", user_id).Error; err != nil {
		glog.Error("Voucher_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	ms := make(map[int]common.Voucher)
	for i, v := range vs {
		ms[v.Type] = vs[i]
	}

	for i, _ := range vis {
		if m, ok := ms[vis[i].Type]; ok {
			vis[i].Voucher = m
		}
	}
	ls := base.FilterStruct(vis, false, "create_time", "update_time", "id")
	//ls := base.SelectStructView(vis, "not man")

	yf.JSON_Ok(c, gin.H{"list": ls})
}

// 获取用户代金券消费记录列表
func Voucher_RecordList(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	typen, _ := base.CheckQueryIntDefaultField(c, "type", -1)
	status, _ := base.CheckQueryIntDefaultField(c, "status", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	amount_min, _ := base.CheckQueryFloat64Field(c, "min_amount")
	amount_max, _ := base.CheckQueryFloat64Field(c, "max_amount")

	vs := []common.Voucher{}
	d := db.GetDB()
	if typen >= 0 {
		d.DB = d.Where("type=?", typen)
	}
	if err := d.Find(&vs, "user_id=?", user_id).Error; err != nil {
		glog.Error("Voucher_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	vrs := []common.VoucherRecord{}
	list_ended := true
	if len(vs) == 0 {
		yf.JSON_Ok(c, gin.H{"list": vrs, "list_ended": list_ended})
		return
	}

	db := db.GetDB()
	db.DB = db.ListPage(page, page_size)

	switch status {
	case 0:
		db.DB = db.Where("amount>0")
	case 1:
		db.DB = db.Where("amount<0")
	}
	if amount_min > 0.0001 {
		db.DB = db.Where("abs(amount)>=?", amount_min)
	}
	if amount_max > 0.0001 {
		db.DB = db.Where("abs(amount)<=?", amount_max)
	}

	voucher_ids := []int64{}
	ms := make(map[int64]interface{})
	for i, v := range vs {
		//ms[v.Id] = base.SelectStructView(vs[i], "record")
		ms[v.Id] = base.FilterStruct(vs[i], true, "type", "info")

		voucher_ids = append(voucher_ids, v.Id)
	}

	if err := db.Order("update_time desc").Find(&vrs, "voucher_id in (?)", voucher_ids).Error; err != nil {
		glog.Error("Voucher_RecordList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if len(vrs) == page_size {
		list_ended = false
	}
	if len(vrs) > 0 {
		for i, _ := range vrs {
			vrs[i].VoucherInfo = ms[vrs[i].VoucherId]
		}
	}

	yf.JSON_Ok(c, gin.H{"list": vrs, "list_ended": list_ended})
}
