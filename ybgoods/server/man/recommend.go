package man

import (
	"fmt"

	base "github.com/jay-wlj/gobaselib"

	"yunbay/ybgoods/common"
	"yunbay/ybgoods/dao"
	"yunbay/ybgoods/server/share"
	"yunbay/ybgoods/util"

	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/jie123108/glog"

	"github.com/gin-gonic/gin"
)

// 推荐app首页人气最高，获取人气最高的商品列表接口
func RecommendUpsert(c *gin.Context) {
	var req common.RecommendIndex
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	db := db.GetTxDB(c)

	// 随机折扣 需要复制相同的商品
	if common.PUBLISH_AREA_REBAT == req.Type {
		ids := base.UniqueInt64Slice(req.ProductIds)
		new_ids := []int64{}
		for _, v := range ids {
			if p, err := share.DuplicateGoods(db, v, func(f *common.Product) bool {
				if f.PublishArea == common.PUBLISH_AREA_REBAT {
					return true
				}
				f.PublishArea = common.PUBLISH_AREA_REBAT // 修改发布专区为随机折扣专区
				return false                              // 需要复制商品
			}); err == nil {
				new_ids = append(new_ids, p.Id)
			} else {
				glog.Error("RecommendUpsert fail! err=", err)
				yf.JSON_Fail(c, err.Error())
				return
			}
		}
		req.ProductIds = new_ids
	}

	db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (type, country) DO update set name='%v', img='%v', descimg='%v', product_ids='{%v}'::bigint[]",
		req.Name, req.Img, req.Descimg, base.Int64SliceToString(req.ProductIds, ",")))

	if err := db.Save(&req).Error; err != nil {
		glog.Error("RecommendIndex fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 刷新缓存
	if err := dao.RefreshIndexRecommend(req); err != nil {
		glog.Error("RecommendIndex RefreshIndexRecommend fail! err=", err)
	}

	db.Commit()
	yf.JSON_Ok(c, gin.H{})
}

type recReq struct {
	common.PageReq
	Type    int `form:"type,default=-1"`
	Country int `form:"country"`
}

// 获取推荐列表信息
func RecommendList(c *gin.Context) {
	var req recReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	vs, err := dao.GetIndexRecommendList(req.Country, req.Type, req.Page, req.PageSize)
	if err != nil {
		glog.Error("GoodsIndexRecommend fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if len(vs) == 0 {
		yf.JSON_Ok(c, gin.H{})
		return
	}
	for i := range vs {
		if ps, ok := vs[i].Rowset.([]common.Product); ok {
			if err := share.FormatProductList(ps); err != nil {
				glog.Error("GoodsIndexRecommend fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			mps := []common.ManProduct{}
			for i := range ps {
				mps = append(mps, common.ManProduct{Product: ps[i]})
			}
			if err := share.GetCategorys(mps); err != nil {
				glog.Error("GoodsList fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}

			vs[i].Rowset = mps
		}
	}
	yf.JSON_Ok(c, vs[0])
}
