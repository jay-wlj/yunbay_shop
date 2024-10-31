package client

import (
	"yunbay/ybgoods/common"
	"yunbay/ybgoods/dao"
	"yunbay/ybgoods/server/share"
	"yunbay/ybgoods/util"

	"github.com/jay-wlj/gobaselib/yf"

	"github.com/jie123108/glog"

	"github.com/gin-gonic/gin"
)

// app首页人气最高，获取人气最高的商品列表接口
func GoodsListHighest(c *gin.Context) {
	var req common.PageReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	v, err := dao.GetHighGoodsList(common.PUBLISH_AREA_KT, req.Page, req.PageSize, nil)
	if err != nil {
		glog.Error("GoodsListHighest fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if err := share.FormatProductList(v.List); err != nil {
		glog.Error("GoodsListHighest fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, v)
}

// 平台优选、最新上线商品列表
func GoodsIndexRecommend(c *gin.Context) {
	var req common.PageReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	country := util.GetCountry(c)

	vs, err := dao.GetIndexRecommendList(country, -1, req.Page, req.PageSize)
	if err != nil {
		glog.Error("GoodsIndexRecommend fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	for i := range vs {
		if ps, ok := vs[i].Rowset.([]common.Product); ok {
			if err := share.FormatProductList(ps); err != nil {
				glog.Error("GoodsIndexRecommend fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			vs[i].Rowset = ps
		}
	}

	yf.JSON_Ok(c, gin.H{"list": vs})
}

type indexMoreReq struct {
	common.PageReq
	Type int `form:"type" binding:"gte=0"`
}

func GoodsIndexRecommendMore(c *gin.Context) {
	var req indexMoreReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	country := util.GetCountry(c)
	vs, err := dao.GetIndexRecommendList(country, req.Type, req.Page, req.PageSize)
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
			vs[i].Rowset = ps
		}
	}
	yf.JSON_Ok(c, vs[0])

}

// YBT购买专区接口
func GoodsYbt(c *gin.Context) {
	var req common.PageReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	v, err := dao.GetHighGoodsList(common.PUBLISH_AREA_YBT, req.Page, req.PageSize, nil)
	if err != nil {
		glog.Error("GoodsListHighest fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err := share.FormatProductList(v.List); err != nil {
		glog.Error("GoodsYbt fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, v)
}

// 随机折扣专区列表接口
func GoodsDiscountList(c *gin.Context) {
	var req common.PageReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	_type := 2
	country := util.GetCountry(c)
	vs, err := dao.GetIndexRecommendList(country, _type, req.Page, req.PageSize)
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
			for i := range ps {
				ps[i].PublishArea = common.PUBLISH_AREA_REBAT // 设置为折扣专区商品
			}
			if err := share.FormatProductList(ps); err != nil {
				glog.Error("GoodsIndexRecommend fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			vs[i].Rowset = ps
		}
	}
	yf.JSON_Ok(c, vs[0])
}
