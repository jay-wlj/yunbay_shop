package client

import (
	"yunbay/ybapi/dao"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
	//"yunbay/ybapi/common"
)

func Recommend_Index(c *gin.Context) {
	v := dao.ProductRecommend{}
	vs, err := v.IndexRecommend()
	if err != nil {
		glog.Error("RecommendIndex fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, vs)
}

func Recommend_List(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	type_, _ := base.CheckQueryIntDefaultField(c, "type", 0)

	v := dao.ProductRecommend{}
	vs, total, err := v.List(type_, page, page_size)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("RecommendList fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	list_ended := true
	if page_size == len(vs) {
		list_ended = false
	}

	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended, "total": total})
}
