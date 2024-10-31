package man

import (
	"fmt"
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybapi/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

func Notice_Upsert(c *gin.Context) {
	var args common.Notice
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	db := db.GetTxDB(c)
	now := time.Now().Unix()
	var err error
	if args.Id == 0 {
		args.CreateTime = now
		args.UpdateTime = now
		err = db.Save(&args).Error
	} else {
		err = db.Model(&args).Updates(map[string]interface{}{"type": args.Type, "title": args.Title, "linkurl": args.Linkurl, "context": args.Context, "update_time": now, "country": args.Country}).Error
	}

	if err != nil {
		glog.Error("Notice_Upsert fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"id": args.Id})
}

type idSt struct {
	Id int64 `json:"id" binding:"gt=0"`
}

func Notice_Del(c *gin.Context) {
	var args idSt
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	v := common.Notice{Id: args.Id}
	if err := db.GetTxDB(c).Delete(&v).Error; err != nil {
		glog.Error("Notice_Delete fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"id": args.Id})
}

type recommendSt struct {
	Ids    []int64 `json:"ids" binding:"gt=0"`
	Status int     `json:"status"`
}

// 资讯推荐
func Notice_RecommendUpsert(c *gin.Context) {
	var args recommendSt
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	now := time.Now().Unix()
	db := db.GetTxDB(c)
	if err := db.Model(&common.Notice{}).Where("id in(?)", args.Ids).Updates(map[string]interface{}{"status": args.Status, "update_time": now}).Error; err != nil {
		glog.Error("Notice_RecommendUpsert fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

func Notice_List(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	type_, _ := base.CheckQueryIntDefaultField(c, "type", 0)
	id, _ := base.CheckQueryInt64Field(c, "id")
	title, _ := base.CheckQueryStringField(c, "title")
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")
	country, _ := base.CheckQueryIntDefaultField(c, "country", 0)

	vs := []common.Notice{}
	db := db.GetDB()
	if type_ >= 0 {
		db.DB = db.Where("type=?", type_)
	}
	db.DB = db.Where("country=?", country)
	if id > 0 {
		db.DB = db.Where("id=?", id)
	}
	if title != "" {
		db.DB = db.Where("title like ?", fmt.Sprintf("%%%v%%", title))
	}
	if begin_date != "" {
		if nTime, err := time.Parse("2006-01-02", begin_date); err == nil {
			db.DB = db.Where("update_time >= ?", nTime.Unix())
		}
	}
	if end_date != "" {
		if nTime, err := time.Parse("2006-01-02", end_date); err == nil {
			db.DB = db.Where("update_time < ?", nTime.AddDate(0, 0, 1).Unix())
		}
	}
	if err := db.ListPage(page, page_size).Order("status desc, update_time desc").Find(&vs).Error; err != nil {
		glog.Error("notice List fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	list_ended := true
	if page_size == len(vs) {
		list_ended = false
	}

	count := 0
	db.Model(&common.Notice{}).Count(&count)

	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended, "total": count})
}
