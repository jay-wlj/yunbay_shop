package client

import (
	"yunbay/ybgoods/common"
	"yunbay/ybgoods/server/share"
	"yunbay/ybgoods/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"yunbay/ybgoods/dao"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// func GoodsList(c *gin.Context) {
// 	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
// 	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
// 	coin_type, _ := base.CheckQueryStringField(c, "coin_type")
// 	coin_type = strings.ToLower(coin_type)

// 	if coin_type == "" {
// 		coin_type = "usdt"
// 	}
// 	vs, err := dao.GetGoodsList(third_id, coin_type, page, page_size)
// 	if err != nil {
// 		glog.Error("GoodsList fail! err=", err)
// 		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
// 		return
// 	}
// 	list_ended := true
// 	if len(vs) == page_size {
// 		list_ended = false
// 	}
// 	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
// }

func GoodsInfo(c *gin.Context) {
	var req common.IdSkuReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	v, err := dao.GetGoodsInfo(req.Id)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, yf.ERR_NOT_EXISTS)
			return
		}
		glog.Error("GoodsInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 处理商品sku问题
	if err := share.FormatProduct(&v, false); err != nil {
		glog.Error("GoodsInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 只取当前sku_id
	if req.SkuId > 0 {
		for i, s := range v.Skus {
			if s.Id == req.SkuId {
				v.Skus = v.Skus[i : i+1]
				break
			}
		}
	}
	yf.JSON_Ok(c, base.FilterStruct(v, false, "create_time", "update_time", "status"))
}

// 获取商品的售后联系方式接口
func GoodsContact(c *gin.Context) {
	var req common.IdReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	var v common.Product
	db := db.GetDB()
	if err := db.Select("id, contact").Find(&v, "id=?", req.Id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, yf.ERR_NOT_EXISTS)
			return
		}
		glog.Error("GoodsContact fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, v.Contact)
}

type cateryReq struct {
	common.PageReq
	CategoryId int64  `form:"category_id"`
	Sort       string `form:"sort"`
	Sequence   string `form:"sequence"`
}

// 获取分类商品列表
func GoodsByCategory(c *gin.Context) {
	var req cateryReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	db := db.GetDB()
	db.DB = db.Where("publish_area=? and status=? and is_hid=0", common.PUBLISH_AREA_KT, common.STATUS_OK)
	if req.CategoryId > 0 {
		db.DB = db.Where("category_id=?", req.CategoryId)
	}

	var total int
	if err := db.Model(&common.Product{}).Count(&total).Error; err != nil {
		glog.Error("GoodsByCategory fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	switch req.Sort {
	case "sold", "rebat", "price":
		if req.Sequence != "desc" {
			req.Sequence = "asc"
		}
		req.Sequence += ",id desc"
		db.DB = db.Order(req.Sort + " " + req.Sequence)
	}

	vs := []common.Product{}
	// 这里images只取第一个图像
	if err := db.ListPage(req.Page, req.PageSize).Select("id,images[1:1],price,title,rebat,publish_area").Find(&vs).Error; err != nil {
		glog.Error("GoodsByCategory fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if err := share.FormatProductList(vs); err != nil {
		glog.Error("GoodsIndexRecommend fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": base.IsListEnded(req.Page, req.PageSize, len(vs), total), "total": total})
}


func GoodsByFirstCategory(c *gin.Context) {
	var req cateryReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	v, err := dao.GetCategoryGoods(req.CategoryId, req.Page, req.PageSize)
	if err != nil {
		glog.Error("GoodsByFirstCategory fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}	
	if err := share.FormatProductList(v.List); err != nil {
		glog.Error("GoodsByFirstCategory fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, v)
}