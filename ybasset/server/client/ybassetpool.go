package client

import (
	//"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	//base "github.com/jay-wlj/gobaselib"

	//"github.com/jay-wlj/gobaselib/yf"
	"yunbay/ybasset/common"

	"github.com/jinzhu/gorm"
)

func YBAssetPool_Add(db *gorm.DB, v common.YBAssetPool) (err error) {
	if err = db.Create(&db).Error; err != nil {
		glog.Error("YBAssetPool_Add fail! err=", err)
		return
	}
	return
}
