package client

import (
	"yunbay/ybapi/common"
	"yunbay/ybapi/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"

	//"github.com/lib/pq"
	"github.com/jinzhu/gorm"
)

func Logistics_Upsert(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args common.Logistics
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}

	args.UserId = user_id

	db := db.GetTxDB(c)
	if err := db.Save(&args).Error; err != nil {
		glog.Error("Address_Add  fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"id": args.Id})
}

// func Logistics_List(c *gin.Context) {
// 	id, _ := base.CheckQueryInt64Field(c, "id")
// 	//page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
// 	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

// 	vs := []common.Logistics{}
// 	if err := db.GetDB().Find(&vs).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			yf.JSON_Ok(c, gin.H{})
// 			return
// 		}
// 		glog.Error("Logistics_List fail! err=", err)
// 		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
// 		return
// 	}
// 	list_ended := true
// 	if len(vs.Infos) == page_size {
// 		list_ended = false
// 	}
// 	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
// }

func Logistics_Info(c *gin.Context) {
	order_id, _ := base.CheckQueryInt64Field(c, "order_id")
	id, _ := base.CheckQueryInt64Field(c, "id")
	db := db.GetDB()
	if order_id == 0 && id == 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	if order_id > 0 {
		db.DB = db.Where("order_id=?", order_id)
	}
	if id > 0 {
		db.DB = db.Where("id=?", id)
	}

	var v common.Logistics
	if err := db.First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Logistics_Info fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 获取商品类型标题

	yf.JSON_Ok(c, v)
}
