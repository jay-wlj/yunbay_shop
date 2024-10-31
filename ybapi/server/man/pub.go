package man

import (
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybapi/dao"
	"yunbay/ybapi/util"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
	//"github.com/lib/pq"
	//"github.com/jinzhu/gorm"
)

// 反馈列表查询
func ManFeedback_List(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	vs := []common.Feedback{}
	if err := db.GetDB().Order("create_time desc").Limit(page_size).Offset((page - 1) * page_size).Find(&vs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Invite_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	list_ended := true
	if page_size == len(vs) {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
}

// banner更新
func ManBanner_Upsert(c *gin.Context) {
	_, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args []common.Banner
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}

	v := dao.Banner{db.GetTxDB(c)}
	if err := v.Upsert(args); err != nil {
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

// 推荐更新
func RecommendUpsert(c *gin.Context) {
	// _, ok := util.GetUid(c)
	// if !ok {
	// 	return
	// }
	var args common.ProductRecommend
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	if args.Id > 0 {
		args.UpdateTime = time.Now().Unix()
	}
	v := dao.ProductRecommend{}
	if err := v.Upsert(args); err != nil {
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}
