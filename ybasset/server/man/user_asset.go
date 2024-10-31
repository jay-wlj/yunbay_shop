package man

import (
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

type uIds struct {
	UserId   int64 `json:"user_id" binding:"required"`
	InviteId int64 `json:"invite_id"`
}

func UserAsset_Add(c *gin.Context) {
	var args uIds
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	now := time.Now().Unix()

	status := 0
	if v, ok := conf.Config.Server.Ext["init_account_lock"]; ok {
		if s, ok := v.(int); ok {
			status = s
		}
	}
	v := common.UserAsset{UserId: args.UserId, CreateTime: now, UpdateTime: now, Status: int16(status)} // 先冻结所有用户的资产
	db := db.GetTxDB(c)
	db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (user_id) DO update set update_time=%v", now))
	if err := db.Save(&v).Error; err != nil {
		glog.Error("UserAsset_Add fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	today := time.Now().Format("2006-01-02")
	// 添加注册空投记录 脚本定时赠送
	rs := []common.RewardRecord{}
	if conf.Config.RewardYbt["reg"] > 0 {
		rs = append(rs, common.RewardRecord{UserId: v.UserId, Status: common.STATUS_INIT, Type: common.CURRENCY_YBT, Amount: conf.Config.RewardYbt["reg"], Date: today, Reason: "注册奖励", Maner: "system", CreateTime: now, UpdateTime: now})
	}

	if args.InviteId > 0 && conf.Config.RewardYbt["inviter"] > 0 {
		rs = append(rs, common.RewardRecord{UserId: args.InviteId, InviteId: v.UserId, Status: common.STATUS_INIT, Type: common.CURRENCY_YBT, Amount: conf.Config.RewardYbt["inviter"], Date: today, Reason: "邀请注册奖励", Maner: "system", Lock: true, CreateTime: now, UpdateTime: now})
	}
	db.DB = db.Set("gorm:insert_option", "")
	for _, v := range rs {
		if err := db.Save(&v).Error; err != nil {
			glog.Error("UserAsset_Add fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}

	yf.JSON_Ok(c, gin.H{"id": v.Id})
}

type uIdLocks struct {
	UserId int64 `json:"user_id"`
	Status int16 `json:"status"`
}

func Asset_Lock(c *gin.Context) {
	var args uIdLocks
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	now := time.Now().Unix()

	v := common.UserAsset{UserId: args.UserId, Status: args.Status, CreateTime: now, UpdateTime: now}
	db := db.GetTxDB(c)
	db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (user_id) DO update set status=%v, update_time=%v", args.Status, now))
	if err := db.Save(&v).Error; err != nil {
		glog.Error("UserAsset_Add fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

type assetSt struct {
	common.UserAsset
	Ext   string `json:"ext"`
	KtExt string `json:"ext_kt"`
}

func Man_UserAssetList(c *gin.Context) {
	id, _ := base.CheckQueryInt64DefaultField(c, "id", 0)
	user_id, _ := base.CheckQueryInt64DefaultField(c, "user_id", -1)
	status, _ := base.CheckQueryIntDefaultField(c, "status", -1)
	str_sorts, _ := base.CheckQueryStringField(c, "sorts")
	str_orders, _ := base.CheckQueryStringField(c, "orders")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	var orders []string
	var sorts []string
	if str_orders != "" {
		orders = strings.Split(str_orders, ",")
	}
	if str_sorts != "" {
		sorts = strings.Split(str_sorts, ",")
	}

	db := db.GetDB()
	// 获取累积信息
	var info assetSt
	if err := db.Model(&common.UserAsset{}).Select("sum(total_ybt) as total_ybt, sum(normal_ybt) as normal_ybt, sum(lock_ybt) as lock_ybt, sum(freeze_ybt) as freeze_ybt, sum(total_kt) as total_kt, sum(normal_kt) as normal_kt, sum(lock_kt) as lock_kt, sum(total_snet) as total_snet, sum(normal_snet) as normal_snet, sum(lock_snet) as lock_snet").Scan(&info).Error; err != nil {
		glog.Error("Man_UserAssetList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	var u common.UserAsset
	if err := db.Find(&u, "user_id=?", conf.Config.ThirdAccount["yunex"].WithDrawId).Error; err != nil {
		glog.Error("Man_UserAssetList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 总可用ybt需要排除掉yunex提币入帐的帐号资产
	info.TotalYbt -= u.TotalYbt
	info.NormalYbt -= u.NormalYbt
	info.LockYbt -= u.LockYbt
	info.FreezeYbt -= u.FreezeYbt

	// 获取累积释放的ybt
	var ybasset common.YBAsset
	if err := db.Last(&ybasset).Error; err != nil {
		glog.Error("Man_UserAssetList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	outYbt := ybasset.TotalIssuedYbt - ybasset.TotalDestroyedYbt - info.NormalYbt - info.LockYbt
	info.Ext = fmt.Sprintf("在外YBT:%.4f(排除掉系统帐号%v的资产,为转出到yunex的帐号)", outYbt, conf.Config.ThirdAccount["yunex"].WithDrawId)

	seller_amount, rebat_amount, err := share.GetPoolAsset(db)
	if err != nil {
		glog.Error("Man_UserAssetList fail! err=", err)
		return
	}

	// 获取今日贡献值
	info.TotalKt += (seller_amount + rebat_amount)
	info.NormalKt += rebat_amount
	info.KtExt = fmt.Sprintf("售出冻结kt:%.4f，未分红的kt:%.4f ", seller_amount, rebat_amount)

	if id > 0 {
		db.DB = db.Where("id=?", id)
	}
	if user_id > -1 {
		db.DB = db.Where("user_id=?", user_id)
	}
	if status > -1 {
		db.DB = db.Where("status=?", status)
	}

	var total int64 = 0
	if err := db.Model(&common.UserAsset{}).Count(&total).Error; err != nil {
		glog.Error("Man_UserAssetList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 排序
	for i, v := range sorts {
		order := "desc"
		if len(order) > i && (orders[i] == "asc" || orders[i] == "desc") {
			order = orders[i]
		}
		db.DB = db.Order(fmt.Sprintf("%v %v", v, order))
	}

	db.DB = db.ListPage(page, page_size)
	vs := []common.UserAsset{}
	if err := db.Model(&common.UserAsset{}).Order("create_time desc").Find(&vs).Error; err != nil {
		glog.Error("Man_UserAssetList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "total": total, "list_ended": list_ended, "info": info})
}

func Man_UserAssetInfo(c *gin.Context) {
	user_id, _ := base.CheckQueryInt64DefaultField(c, "user_id", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	txType, _ := base.CheckQueryIntDefaultField(c, "type", 0)
	transaction_type, _ := base.CheckQueryIntDefaultField(c, "transaction_type", -1)
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")
	if user_id < 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	db := db.GetDB()

	if user_id > -1 {
		db.DB = db.Where("user_id=?", user_id)
	}
	if txType > -1 {
		db.DB = db.Where("type=?", txType)
	}
	if transaction_type > -1 {
		db.DB = db.Where("transaction_type=?", transaction_type)
	}
	if begin_date != "" {
		db.DB = db.Where("date>=?", begin_date)
	}
	if end_date != "" {
		db.DB = db.Where("date<=?", end_date)
	}

	var total int64 = 0
	if err := db.Model(&common.UserAssetDetail{}).Count(&total).Error; err != nil {
		glog.Error("Man_UserAssetList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	db.DB = db.ListPage(page, page_size)
	vs := []common.UserAssetDetail{}
	if err := db.Model(&common.UserAssetDetail{}).Order("create_time desc").Find(&vs).Error; err != nil {
		glog.Error("Man_UserAssetList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "total": total, "list_ended": list_ended})
}
