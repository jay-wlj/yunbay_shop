package share

import (
	"github.com/jay-wlj/gobaselib/db"
	"yunbay/ybgoods/common"

	"github.com/jie123108/glog"
)

func GetCategoryByParentId(parent_id int64) (vs []common.ProductCategory, err error) {
	// 如果没有属性 则添加属性 更新商品的价格
	vs = []common.ProductCategory{}
	if err = db.GetDB().Find(&vs, "parent_id=?", parent_id).Error; err != nil {
		glog.Error("GetCategoryByParentId fail! err=", err)
		return
	}

	return
}

