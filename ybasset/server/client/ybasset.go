package client

import (
	//"time"
	"yunbay/ybasset/common"
	"yunbay/ybasset/dao"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

type dateSt struct {
	Date string `json:"date"`
}

// 获取昨天(某天)平台资产信息
func YBAssetDetail_Get(c *gin.Context) {
	date, _ := base.CheckQueryStringField(c, "date")
	db := db.GetDB()

	if date != "" {
		db.DB = db.Where("date=?", date)
	}
	var v common.YBAssetDetail
	var err error
	if err = db.Order("date desc").First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("UserAsset_Get fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if v.Profit > 0 {
		v.ProfitRate = (v.Perynbay * 365 * (v.Mining + v.AirUnlock)) / v.Profit
	}

	yf.JSON_Ok(c, v)
}

// 按日获取平台资产明细信息列表
func YBAssetDetail_List(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	// vs := []common.YBAssetDetail{}
	// db := db.GetDB()

	// today := time.Now().Format("2006-01-02")

	// if err := db.ListPage(page, page_size).Where("date <> ?", today).Order("date desc").Find(&vs).Error; err != nil&&err!=gorm.ErrRecordNotFound {
	// 	glog.Error("YBAssetDetail_List fail! err=", err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	vs, err := dao.ListYBAssetDetail(page, page_size)
	if err != nil {
		glog.Error("YBAssetDetail_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	ended := true
	if page_size == len(vs) {
		ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": ended})
}

// 获取昨天(某天)平台资产信息
func YBAsset_Get(c *gin.Context) {
	date, _ := base.CheckQueryStringField(c, "date")
	db := db.GetDB()

	if date != "" {
		db.DB = db.Where("date=?", date)
	}
	var v common.YBAsset
	if err := db.Order("date desc").First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("UserAsset_Get fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, v)
}

// 按日获取平台帐户资产列表信息
func YBAsset_List(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	vs := []common.YBAsset{}
	db := db.GetDB()

	if err := db.Order("date desc").Limit(page_size).Offset((page - 1) * page_size).Find(&vs).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("YBAsset_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	ended := true
	if page_size == len(vs) {
		ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": ended})
}

type stYbtReleaseSt struct {
	AssetDetail common.YBAssetDetail `json:"asset_detail"`
	Asset       common.YBAsset       `json:"asset"`
	Date        string               `json:"date"`
}

func YBAsset_ReleaseInfo(c *gin.Context) {
	var ybinfo common.YBAssetDetail
	var ybtotal common.YBAsset
	db := db.GetDB()
	if err := db.Last(&ybinfo, "ybt_status=?", common.STATUS_OK).Error; err != nil {
		glog.Error("YBAsset_ReleaseInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if err := db.Find(&ybtotal, "date=?", ybinfo.Date).Error; err != nil {
		glog.Error("YBAsset_ReleaseInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	var ret stYbtReleaseSt
	ret.AssetDetail = ybinfo
	ret.Asset = ybtotal
	ret.Date = ybinfo.Date
	yf.JSON_Ok(c, ret)
}

// 获取某日挖矿难度
func YBAsset_Diffcult(c *gin.Context) {
	date, _ := base.CheckQueryStringField(c, "date")
	if date == "" {
		date = time.Now().Format("2006-01-02") // 默认取当天日期
	}

	val, err := get_difficult(date)
	if err != nil {
		glog.Error("YBAsset_Diffcult fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"date": date, "difficult": val})
}

func get_difficult(date string) (val float64, err error) {
	key := "diffcult"
	redis, er := cache.GetWriter(common.RedisPub)
	if er != nil {
		glog.Error("get_difficult GetDefaultCache fail! err=", er)
	}

	if redis != nil {
		val, err = redis.HGetF64(key, date)
		if err == nil {
			return
		}
	}
	var v common.YBAssetDetail

	db := db.GetDB()
	if err = db.Find(&v, "date=?", date).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("YBAsset_Diffcult fail! err=", err)
		return
	}
	val = v.Difficult

	if redis != nil && val > 0 {
		redis.HSet(key, date, val, 24*time.Hour)
	}
	if base.IsEqual(val, 0) {
		cur_date, e := time.Parse("2006-01-02", date)
		if e != nil {
			err = e
			glog.Error("get_difficult fail! err=", err)
			return
		}
		val, err = get_difficult(cur_date.AddDate(0, 0, -1).Format("2006-01-02"))
	}
	return
}
