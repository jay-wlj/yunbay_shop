package client

import (
	"yunbay/ybapi/common"
	"yunbay/ybapi/util"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
	//"time"
)

// 资讯推荐
func Notice_Recommend(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	platform, _ := util.GetPlatformVersionByContext(c)
	if platform != "web" { // 移动端默认给前条推荐
		page_size = 3
		page = 1
	}
	vs := []common.Notice{}
	db := db.GetDB().ListPage(page, page_size)
	country := util.GetCountry(c)
	if err := db.Where("status=? and country=?", 1, country).Order("status desc, update_time desc").Find(&vs).Error; err != nil {
		glog.Error("notice List fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	list_ended := true
	if page_size == len(vs) {
		list_ended = false
	}

	count := 0
	db.Model(&common.Notice{}).Where("status=?", 1).Count(&count)

	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended, "total": count})
}

func Notice_List(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	type_, _ := base.CheckQueryIntDefaultField(c, "type", -1)
	key, _ := base.CheckQueryStringField(c, "key")

	db := db.GetDB()
	country := util.GetCountry(c)
	db.DB = db.Where("status>=0 and country=?", country)
	if type_ > -1 {
		db.DB = db.Where("type=?", type_)
	}
	if key != "" {
		db.DB = db.Where("title like ?", fmt.Sprintf("%%%v%%", key))
	}
	var total int = 0
	if err := db.Model(&common.Notice{}).Count(&total).Error; err != nil {
		glog.Error("notice List fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	vs := []common.Notice{}
	if err := db.ListPage(page, page_size).Order("status desc, update_time desc").Find(&vs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("notice List fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": base.IsListEnded(page, page_size, len(vs), total), "total": total})
}

func Notice_Info(c *gin.Context) {
	id, _ := base.CheckQueryInt64Field(c, "id")

	v := common.Notice{Id: id}
	if err := db.GetDB().Find(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("notice List fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, v)
}
