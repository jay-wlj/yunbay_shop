package man

import (
	"yunbay/ybasset/common"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

func YBAsset_List(c *gin.Context) {
	date, _ := base.CheckQueryStringField(c, "date")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")

	db := db.GetDB()

	if date != "" {
		db.DB = db.Where("date=?", date)
	}
	if begin_date != "" {
		db.DB = db.Where("date>=?", begin_date)
	}
	if end_date != "" {
		db.DB = db.Where("date<=?", end_date)
	}

	var total int
	var err error
	if err = db.Model(&common.YBAssetAll{}).Count(&total).Error; err != nil {
		glog.Error("YBAsset_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	vs := []common.YBAssetAll{}
	if err = db.ListPage(page, page_size).Order("date asc").Find(&vs).Error; err != nil {
		glog.Error("YBAsset_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	var total_info common.YBAsset
	if err = db.Model(&common.YBAssetAll{}).Select("sum(issue_ybt) as total_issue_ybt, sum(mining) as total_mining, sum(air_unlock) as total_air_unlock, sum(activity) as total_activity, sum(project) as total_project, sum(air_drop) as total_air_drop, sum(air_recover) as total_air_recover, sum(destoryed_ybt) as total_destroyed_ybt").Scan(&total_info).Error; err != nil {
		glog.Error("YBAsset_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	for i, _ := range vs {
		if vs[i].TotalPerynbay, err = getDateBonus(vs[i].Date); err != nil {
			glog.Error("YBAsset_List fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": base.IsListEnded(page, page_size, len(vs), total), "total": total, "info": total_info})
}

func getDateBonus(date string) (perynbay float64, err error) {
	db := db.GetDB()
	var v common.YBAsset
	if err = db.Model(&common.YBAssetAll{}).Where("date>=?", date).Select("sum(perynbay) as total_perynbay").Scan(&v).Error; err != nil {
		glog.Error("getDateBonus fail! err=", err)
		return
	}
	perynbay = v.TotalPerynbay
	return
}
