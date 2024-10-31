package man

import (
	"yunbay/ybasset/common"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func TradeFlow_List(c *gin.Context) {
	country, _ := base.CheckQueryIntDefaultField(c, "country", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	vs := []common.TradeFlow{}
	var total int64 = 0
	db := db.GetDB()

	if country > -1 {
		db.DB = db.Model(&common.TradeFlow{})
		db.DB = db.Where("country=?", country)
	} else {
		db.DB = db.Table("tradeflow_all")
	}
	if err := db.Count(&total).Error; err != nil {
		glog.Error("TradeFlow_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err := db.ListPage(page, page_size).Order("date desc").Find(&vs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("TradeFlow_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended, "total": total})
}
