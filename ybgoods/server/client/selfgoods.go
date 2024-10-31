package client

import (
	"fmt"
	"yunbay/ybgoods/common"
	"yunbay/ybgoods/dao"
	"yunbay/ybgoods/server/share"
	"yunbay/ybgoods/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func GoodsSelfInfo(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req common.IdReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	var v common.ManProduct
	db := db.GetDB()
	if err := db.Find(&v, "user_id=? and id=?", user_id, req.Id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("GoodsInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 处理商品sku问题
	if err := share.FormatProduct(&v.Product, false); err != nil {
		glog.Error("GoodsInfo fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	vs := []common.ManProduct{}
	vs = append(vs, v)
	// 获取商品分类
	if err := share.GetCategorys(vs); err != nil {
		glog.Error("SelfGoodsList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 获取原价
	if err := share.GetCostPrice(vs); err != nil {
		glog.Error("SelfGoodsList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, base.FilterStruct(vs[0], false, "create_time", "update_time"))
}

type descImgs struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Path   string `json:"path"`
}

func GoodsUpsert(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}

	// 验证是否为商户
	var v common.Business
	if err := db.GetDB().First(&v, "user_id=?", user_id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_NOT_BUSINESS)
			return
		}
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	var req common.Product
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	// 详情图结构修改
	rdb := db.GetDB()
	ds := []interface{}{}
	for _, v := range req.Descimgs {
		if s, ok := v.(string); ok {
			var u common.Uploadfile
			// 获取图片的宽高
			if err := rdb.Find(&u, "path=?", s).Error; err != nil && err != gorm.ErrRecordNotFound {
				glog.Error("GoodsUpsert fail! err=", err)
				yf.JSON_Fail(c, err.Error())
				return
			}
			ds = append(ds, descImgs{Path: s, Width: u.Width, Height: u.Height})
		}
	}
	req.Descimgs = ds
	req.UserId = user_id

	db := db.GetTxDB(c)
	vs := []*common.Product{&req}
	if err := share.AddProduct(db, vs); err != nil {
		glog.Error("GoodsUpsert fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}

	db.Commit()
	yf.JSON_Ok(c, gin.H{"id": vs[0].Id})
	return
}

func GoodsOffine(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req common.IdOkSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	// 验证是否为商户
	var v common.Business
	db := db.GetDB()
	if err := db.First(&v, "user_id=?", user_id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_NOT_BUSINESS)
			return
		}
		glog.Error("GoodsOffine fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err := db.Model(&common.Product{}).Where("id=?", req.Id).Updates(base.Maps{"status": req.Status}).Error; err != nil {
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		glog.Error("GoodsOffine fail! err=", err)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

type attr struct {
	CategoryId int64    `json:"category_id"`
	Name       string   `json:"name"`
	Values     []string `json:"values"`
}

// 添加规格属性名及值
func GoodsAddSku(c *gin.Context) {
	_, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req attr
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	v := common.ProductAttrKey{CategoryId: req.CategoryId, Name: req.Name}
	for _, r := range req.Values {
		v.Values = append(v.Values, common.ProductAttrValue{Value: r})
	}
	db := db.GetTxDB(c)

	if err := db.Save(&v).Error; err != nil {
		glog.Error("GoodsAddSku fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	db.Commit()
	yf.JSON_Ok(c, gin.H{})
}

// 复制一个商品
func GoodsDuplicate(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req common.IdOkSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	db := db.GetDB()
	db.DB = db.Where("user_id=?", user_id)
	v, err := share.DuplicateGoods(db, req.Id, nil)
	if err != nil {
		glog.Error("GoodsDuplicate fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}

	yf.JSON_Ok(c, gin.H{"id": v.Id})
}

// 上下线商品
func GoodsStatusOne(c *gin.Context) {
	var req common.IdOkSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	db := db.GetTxDB(c)
	res := db.Model(&common.Product{}).Where("id=?", req.Id).Updates(base.Maps{"status": req.Status})
	if res.Error != nil {
		glog.Error("GoodsOffine fail! err=", res.Error)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if res.RowsAffected > 0 {
		db.AfterCommit(func() {
			dao.RefreshGoods(req.Id)
		})
	}
	yf.JSON_Ok(c, gin.H{})
}

func GoodsStatus(c *gin.Context) {
	var req common.IdsOkSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	ids := []int64{}
	for _, v := range req.Ids {
		ids = append(ids, v)
	}
	db := db.GetTxDB(c)
	res := db.Model(&common.Product{}).Where("id in(?)", ids).Updates(base.Maps{"status": req.Status})
	if res.Error != nil {
		glog.Error("GoodsOffine fail! err=", res.Error)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if res.RowsAffected > 0 {
		db.AfterCommit(func() {
			for _, id := range ids {
				dao.RefreshGoods(id)
			}
		})
	}
	yf.JSON_Ok(c, gin.H{})
}

type goodselfReq struct {
	common.PageReq
	FirstCategoryId  int    `form:"first_category_id" binding:"gte=0"`
	SecondCategoryId int    `form:"second_category_id" binding:"gte=0"`
	Status           int    `form:"status,default=-1"`
	Title            string `form:"title"`
	PublishArea      int    `form:"publish_area,default=1"`
}

// 获取自己发布的商品列表信息
func GoodsSelfList(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req goodselfReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	tdb := db.GetDB()
	db := db.GetDB()
	db.DB = db.Where("user_id=? and publish_area=?", user_id, req.PublishArea)

	if req.Status > -1 {
		db.DB = db.Where("status=?", req.Status)
	}
	if req.Title != "" {
		db.DB = db.Where("title like ?", fmt.Sprintf("%%%v%%", req.Title))
	}

	if req.SecondCategoryId > 0 {
		db.DB = db.Where("category_id=?", req.SecondCategoryId)
	} else if req.FirstCategoryId > 0 {
		// 获取一级分类下的所有二级分类id
		vs := []common.ProductCategory{}
		if err := tdb.Select("id").Find(&vs, "parent_id=?", req.FirstCategoryId).Error; err != nil {
			glog.Error("SelfGoodsList fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
		if len(vs) > 0 {
			ids := []int64{}
			for _, v := range vs {
				ids = append(ids, v.Id)
			}
			db.DB = db.Where("category_id in(?)", ids)
		}
	}

	var total int
	if err := db.Model(&common.Product{}).Count(&total).Error; err != nil {
		glog.Error("SelfGoodsList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	vs := []common.ManProduct{}
	if err := db.ListPage(req.Page, req.PageSize).Order("sold desc, update_time desc, id desc").Find(&vs).Error; err != nil {
		glog.Error("SelfGoodsList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 获取商品分类
	if err := share.GetCategorys(vs); err != nil {
		glog.Error("SelfGoodsList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 获取原价

	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": base.IsListEnded(req.Page, req.PageSize, len(vs), total), "total": total})
}
