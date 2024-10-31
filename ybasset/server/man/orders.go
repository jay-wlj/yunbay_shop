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

func BonusOrders_List(c *gin.Context) {
	id, _ := base.CheckQueryInt64DefaultField(c, "id", 0)
	order_id, _ := base.CheckQueryInt64DefaultField(c, "order_id", -1)
	//status, _ := base.CheckQueryIntDefaultField(c, "status", -2)
	// begin_date,_ := base.CheckQueryStringField(c, "begin_date")
	// end_date,_ := base.CheckQueryStringField(c, "end_date")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	db := db.GetDB()
	if id > 0 {
		db.DB = db.Where("id=?", id)
	}
	if order_id > -1 {
		db.DB = db.Where("order_id=?", order_id)
	}

	var total int64 = 0
	if err := db.Model(&common.Ordereward{}).Count(&total).Error; err != nil {
		glog.Error("BonusOrders_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	db.DB = db.ListPage(page, page_size)
	vs := []common.Ordereward{}
	if err := db.Order("id desc").Find(&vs).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("TradeFlow_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_page": list_ended, "total": total})
}

func BonusOrders_Info(c *gin.Context) {
	order_id, _ := base.CheckQueryInt64DefaultField(c, "order_id", -1)
	if order_id <= 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	db := db.GetDB()
	var v common.Ordereward
	var err error
	if err = db.Find(&v, "order_id=?", order_id).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("TradeFlow_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if err == gorm.ErrRecordNotFound {
		yf.JSON_Ok(c, gin.H{})
		return
	}

	yf.JSON_Ok(c, v)
}
