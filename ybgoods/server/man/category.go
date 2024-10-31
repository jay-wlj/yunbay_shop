package man

import (
	"yunbay/ybgoods/common"
	"yunbay/ybgoods/util"

	"yunbay/ybgoods/dao"

	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"

	"github.com/gin-gonic/gin"
)

type categoryidReq struct {
	CategoryId int64 `form:"category_id" json:"category_id" binding:"gte=0"`
}

func CategoryUpsert(c *gin.Context) {
	var req common.ProductCategory
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	db := db.GetDB()
	if err := db.Save(&req).Error; err != nil {
		glog.Error("GoodsCategoryUpsert fail! err =", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"id": req.Id})
	dao.AddCategoryId(req.ParentId, req.Id)

}

func CategoryDel(c *gin.Context) {
	var req common.IdReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	// 是否有子分类
	vs, err := dao.GetCategoryList(req.Id, true)
	if err != nil {
		glog.Error("CategoryDel fail! err =", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if len(vs) > 0 {
		yf.JSON_Fail(c, "ERROR_DONOT_DELETE")
		return
	}

	// 是否有商品已使用该分类
	var count int64
	db := db.GetDB()
	if err = db.Model(&common.Product{}).Where("category_id =?", req.Id).Count(&count).Error; err != nil {
		glog.Error("CategoryDel fail! err =", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if count > 0 {
		yf.JSON_Fail(c, "ERROR_DONOT_DELETE")
		return
	}

	var v common.ProductCategory
	if err = db.Find(&v, "id=?", req.Id).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("CategoryDel fail! err =", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if err == nil {
		if err = db.Delete(&v).Error; err != nil {
			glog.Error("CategoryDel fail! err =", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
		dao.RemoveCategoryId(v.Id, v.ParentId)
	}

	yf.JSON_Ok(c, gin.H{})

}

// 重置缓存
func GoodsRedisReset(c *gin.Context) {

	if err := dao.ReloadCategoryId(); err != nil {
		glog.Error("GoodsRedisReset fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}
