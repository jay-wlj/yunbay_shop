package client

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"

	//base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"yunbay/ybapi/common"
	"yunbay/ybapi/util"

	//"github.com/lib/pq"
	"github.com/jinzhu/gorm"
)

// 商家认证信息添加
func Business_Upsert(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}

	var args common.Business
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}

	now := time.Now().Unix()
	args.UpdateTime = now
	args.UserId = user_id
	if args.Id == 0 {
		args.CreateTime = now
	}

	db := db.GetTxDB(c)
	if err := db.Save(&args).Error; err != nil {
		glog.Error("Address_Add  fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"id": args.Id})
}

// 商家认证信息
func Business_Info(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}

	var v common.Business
	err := db.GetDB().Find(&v, "user_id=?", user_id).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("Business_Info fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if err == gorm.ErrRecordNotFound {
		yf.JSON_Ok(c, gin.H{})
		return
	}
	// v.ProTypes, err = ProductType_GetTitleByIds(v.ProdutTypes)
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		yf.JSON_Ok(c, gin.H{})
	// 		return
	// 	}
	// 	glog.Error("ProductType_GetTitleByIds  fail! err", err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	yf.JSON_Ok(c, v)
}

// 商家币种汇率信息
func Business_RatioSet(c *gin.Context) {

}
