package client

import (
	"yunbay/ybsearch/server/share"
	"yunbay/ybsearch/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

func Index(c *gin.Context) {
	keyword, _ := base.CheckQueryStringField(c, "keyword")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	sort, _ := base.CheckQueryStringField(c, "sort")
	sequence, _ := base.CheckQueryStringField(c, "sequence")

	total, ids, err := share.GetSphinx().Search(keyword, sort, sequence, page, page_size)
	if err != nil {
		glog.Error("Index fail! err=", err)
		//yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		yf.JSON_Ok(c, gin.H{"rowset": []interface{}{}, "list_ended": true, "total": 0})
		return
	}
	res, e := util.ListProductInfoByIds(ids)
	if e != nil {
		glog.Error("Index fail! err=", err)
		yf.JSON_Fail(c, e.Error())
		return
	}

	res.Total = total
	res.ListEnded = base.IsListEnded(page, page_size, len(ids), total)
	yf.JSON_Ok(c, gin.H{"rowset": res.List, "list_ended": res.ListEnded, "total": res.Total})
}
