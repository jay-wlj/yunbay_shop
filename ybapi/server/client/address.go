package client

import (
	"yunbay/ybapi/common"
	"yunbay/ybapi/util"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"

	//"github.com/lib/pq"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jinzhu/gorm"
)

func Address_Upsert(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}

	var args common.UserAddress
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
	var selcount int64

	if err := db.Model(&common.UserAddress{}).Where("user_id=? and \"default\"=?", user_id, true).Count(&selcount).Error; err != nil {
		glog.Error("Address_Add fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 没有勾选则设为默认地址
	if 0 == selcount {
		args.Default = true
	}

	if err := db.Save(&args).Error; err != nil {
		glog.Error("Address_Add  fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 将其它选项置非勾选状态
	if selcount > 1 || (args.Default && (selcount != 0)) {
		if err := db.Model(&common.UserAddress{}).Where("user_id=? and id <> ?", user_id, args.Id).Update(map[string]interface{}{"default": false}).Error; err != nil {
			glog.Error("Address_Add fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}
	yf.JSON_Ok(c, gin.H{"id": args.Id})
}

func Address_List(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}

	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	vs := []common.UserAddress{}
	db := db.GetDB()
	db.DB = db.ListPage(page, page_size).Where("user_id=?", user_id)

	// 获取商品列表信息
	if err := db.Order("\"default\" desc, create_time desc").Find(&vs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Product_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}

	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
}

func Address_Del(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args common.IdSt
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}

	db := db.GetTxDB(c)

	if err := db.Delete(&common.UserAddress{}, "id=? and user_id=?", args.Id, user_id).Error; err != nil {
		glog.Error("Address_Del fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	var selcount int = 0
	if err := db.Model(&common.UserAddress{}).Where("user_id=? and \"default\"=?", user_id, true).Count(&selcount).Error; err != nil {
		glog.Error("Address_Del fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 设置最新的为默认值
	if selcount == 0 {
		var v common.UserAddress
		if err := db.Model(&common.UserAddress{}).Last(&v, "user_id=?", user_id).Error; err != nil && err != gorm.ErrRecordNotFound {
			glog.Error("Address_Del fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
		if v.Id > 0 {
			if err := db.Model(&common.UserAddress{}).Where("id=?", v.Id).Updates(map[string]interface{}{"default": true}).Error; err != nil {
				glog.Error("Address_Del fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
		}
	}
	yf.JSON_Ok(c, gin.H{})
}
