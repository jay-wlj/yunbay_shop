package client

import (
	"yunbay/ybasset/common"
	"yunbay/ybasset/dao"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func UserAsset_Get(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}

	v, err := dao.GetUserAsset(user_id, nil)
	if err != nil {
		glog.Error("UserAsset_Get fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if ratio, err := share.GetRmbRatio(); err == nil {
		v.RmbRatio = ratio
	}

	yf.JSON_Ok(c, v)
}

// 获取用户所有资产列表
func UserAsset_All(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	vs, _ := share.GetUserAllAsset(user_id)
	masset := make(map[int]common.UserAssetType)
	for i, v := range vs {
		masset[v.Type] = vs[i]
	}

	vs = []common.UserAssetType{}
	for i := common.CURRENCY_YBT; i < common.CURRENCY_UNKNOW; i += 1 {
		if i != common.CURRENCY_RMB {
			if v, ok := masset[i]; ok {
				vs = append(vs, v)
			} else {
				vs = append(vs, common.UserAssetType{Type: i})
			}
		}
	}
	yf.JSON_Ok(c, gin.H{"list": vs})
}

func UserAssetDetail_Add(c *gin.Context) {
	var args common.UserAssetDetail

	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}

	now := time.Now().Unix()
	args.CreateTime = now
	args.UpdateTime = now

	if err := db.GetTxDB(c).Create(&args).Error; err != nil {
		glog.Error("UserAssetDetail_Add fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

// 获取用户资产明细信息
func UserAssetDetail_List(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	_type, _ := base.CheckQueryIntDefaultField(c, "type", -1)
	transaction_type, _ := base.CheckQueryIntDefaultField(c, "transaction_type", -1)
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")

	vs := []common.UserAssetDetail{}
	db := db.GetDB()
	db.DB = db.Where("user_id=?", user_id)
	if _type >= 0 {
		db.DB = db.Where(" type=?", _type)
	}
	if transaction_type >= 0 {
		db.DB = db.Where("transaction_type=?", transaction_type)
	}
	if begin_date != "" {
		db.DB = db.Where("date>=?", begin_date)
	}
	if end_date != "" {
		db.DB = db.Where("date<=?", end_date)
	}

	var total int = 0
	if err := db.Model(&common.UserAssetDetail{}).Count(&total).Error; err != nil {
		glog.Error("UserAssetDetail_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err := db.ListPage(page, page_size).Order("create_time desc").Find(&vs).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("UserAssetDetail fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": base.IsListEnded(page, page_size, len(vs), total), "total": total})
}

// 获取昨日的kt收益金及ybt奖励
func UserAssetDetail_BonusInfo(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	v, err := dao.GetYesterDayBonus(day, user_id)
	if err != nil {
		glog.Error("UserAssetDetail_BeforeBonus failed! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, v)
}

// 获取往日ybt分红记录
func UserAssetDetail_YbtBonus(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")
	if page <= 0 {
		page = 1
	}
	if page_size <= 0 {
		page_size = 10
	}
	db := db.GetDB().ListPage(page, page_size).Where("user_id=?", user_id).Order("date desc")
	if begin_date != "" {
		db = db.Where("date>=?", begin_date)
	}
	if end_date != "" {
		db = db.Where("date<=?", end_date)
	}

	vs := []common.BonusYbtDetail{}
	if err := db.Find(&vs).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("UserAssetDetail_YbtBonus failed! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
}

// 获取往日kt分红记录
func UserAssetDetail_KtBonus(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")
	if page <= 0 {
		page = 1
	}
	if page_size <= 0 {
		page_size = 10
	}
	db := db.GetDB().ListPage(page, page_size).Where("user_id=?", user_id).Order("date desc")
	if begin_date != "" {
		db = db.Where("date>=?", begin_date)
	}
	if end_date != "" {
		db = db.Where("date<=?", end_date)
	}

	vs := []common.BonusKtDetail{}
	if err := db.Find(&vs).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("UserAssetDetail_YbtBonus failed! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
}

type amounttype struct {
	Channel int     `json:"channel"`
	Amount  float64 `json:"amount"`
}

// 获取ybt的累积的邀请奖励,活动奖励等
func UserAssetDetail_YbtInfo(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	type_, _ := base.CheckQueryIntDefaultField(c, "channel", -1)
	vs := []amounttype{}
	db := db.GetDB().Model(&common.UserAssetDetail{}).Where("user_id=? and type=?", user_id, common.CURRENCY_YBT).Group("transaction_type")
	if type_ > -1 {
		db = db.Where("transaction_type=?", type_)
	}
	rows, err := db.Select("transaction_type as channel, sum(amount) as amount").Rows()
	if err != nil {
		glog.Error("UserAssetDetail_YbtInfo failed! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var v amounttype
		db.ScanRows(rows, &v)
		vs = append(vs, v)
	}

	yf.JSON_Ok(c, gin.H{"list": vs})
}
