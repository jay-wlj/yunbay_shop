package man

import (
	"fmt"
	"yunbay/ybgoods/common"
	"yunbay/ybgoods/server/share"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

type goodselfReq struct {
	common.PageReq
	FirstCategoryId  int    `form:"first_category_id" binding:"gte=0"`
	SecondCategoryId int    `form:"second_category_id" binding:"gte=0"`
	Status           int    `form:"status,default=-1"`
	Title            string `form:"title"`
	PublishArea      int    `form:"publish_area,default=-1"`
	ProductId        int64  `form:"product_id" binding:"gte=0"`
	UserId           int64  `form:"user_id" binding:"gte=0"`
	IsHid            int    `form:"is_hid"`
}

// 获取商品列表
func GoodsList(c *gin.Context) {
	title, _ := base.CheckQueryStringField(c, "title")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	first_category_id, _ := base.CheckQueryInt64Field(c, "first_category_id")
	second_category_id, _ := base.CheckQueryInt64Field(c, "second_category_id")
	product_id, _ := base.CheckQueryInt64Field(c, "product_id")
	user_id, _ := base.CheckQueryInt64Field(c, "user_id")
	status, _ := base.CheckQueryInt64DefaultField(c, "status", -1)
	publish_area, _ := base.CheckQueryIntDefaultField(c, "publish_area", -1)
	is_hid, _ := base.CheckQueryIntDefaultField(c, "is_hid", -1)
	start_time, _ := base.CheckQueryInt64Field(c, "start_time")
	end_time, _ := base.CheckQueryInt64Field(c, "end_time")
	sort, _ := base.CheckQueryStringField(c, "sort")
	sequence, _ := base.CheckQueryStringField(c, "sequence")

	db := db.GetDB()

	if first_category_id > 0 {
		vs := []common.ProductCategory{}
		if err := db.Select("id").Find(&vs, "parent_id=?", first_category_id).Error; err != nil {
			glog.Error("get_category_list fail! err=", err)
			return
		}
		ids := []int64{}
		for _, v := range vs {
			ids = append(ids, v.Id)
		}
		db.DB = db.Where("category_id in(?)", ids)
	} else if second_category_id > 0 {
		db.DB = db.Where("category_id =?", second_category_id)
	}

	if user_id > 0 {
		db.DB = db.Where("user_id=?", user_id)
	}
	if status > -1 {
		db.DB = db.Where("status=?", status)
	}
	if product_id > 0 {
		db.DB = db.Where("id=?", product_id)
	}
	if publish_area > -1 {
		db.DB = db.Where("publish_area=?", publish_area)
	}
	if is_hid > -1 {
		db.DB = db.Where("is_hid=?", is_hid)
	}

	if title != "" {
		db.DB = db.Where("title like ?", fmt.Sprintf("%%%v%%", title))
	}
	if start_time > 0 {
		db.DB = db.Where("update_time>=?", start_time)
	}
	if end_time > 0 {
		db.DB = db.Where("update_time<=?", end_time)
	}

	var total int
	if err := db.Model(&common.Product{}).Count(&total).Error; err != nil {
		glog.Error("GoodsList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	var order_by string
	if sequence != "desc" && sequence != "asc" {
		sequence = "desc"
	}

	if sort == "update_time" || sort == "sold" {
		order_by = sort + " " + sequence
	}
	if order_by != "" {
		order_by += ","
	}
	order_by += " id desc"

	vs := []common.ManProduct{}
	if err := db.ListPage(page, page_size).Order(order_by).Find(&vs).Error; err != nil {
		glog.Error("GoodsList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err := share.GetCategorys(vs); err != nil {
		glog.Error("GoodsList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"list": base.FilterStruct(vs, false, "images", "descimgs"), "list_ended": base.IsListEnded(page, page_size, len(vs), total), "total": total})
}

// 获取商品详细信息
func GoodsBackendInfo(c *gin.Context) {
	id, _ := base.CheckQueryInt64Field(c, "id")
	if id == 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var v common.ManProduct
	db := db.GetDB()
	if err := db.Find(&v, "id=?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("GoodsInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err := share.FormatProduct(&v.Product, true); err != nil {
		glog.Error("GoodsInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	vs := []common.ManProduct{}
	vs = append(vs, v)
	if err := share.GetCategorys(vs); err != nil {
		glog.Error("GoodsList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 获取商家信息
	var b common.Business
	if err := db.Select("user_id, type, company,name").Find(&b, "user_id=?", v.UserId).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GoodsList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	vs[0].Business = &b

	yf.JSON_Ok(c, vs[0])
}
