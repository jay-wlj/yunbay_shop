package share

import (
	"yunbay/ybapi/common"

	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
)

func get_logistic_ids(ids []int64) (vs []common.Logistics, err error) {
	if err = db.GetDB().Select("id,company,number").Find(&vs, "id in(?)", ids).Error; err != nil {
		glog.Error("Orders_Report fail! err=", err)
		return
	}
	return
}
