package client

import (
	"yunbay/ybgoods/dao"
	"yunbay/ybgoods/util"

	"github.com/jay-wlj/gobaselib/yf"

	"github.com/jie123108/glog"

	"github.com/gin-gonic/gin"
)

type categoryidReq struct {
	CategoryId int64 `form:"category_id" binding:"gte=0"`
}

func CategoryList(c *gin.Context) {
	var req categoryidReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	vs, err := dao.GetCategoryList(req.CategoryId, true)
	if err != nil {
		glog.Error("CategoryList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": vs})
}

func CategoryFirst(c *gin.Context) {
	var req categoryidReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	vs, err := dao.GetCategoryList(0, false)
	if err != nil {
		glog.Error("CategoryList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": vs})
}
