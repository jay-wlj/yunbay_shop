package util

import (
	"yunbay/ybgoods/common"

	"github.com/jie123108/glog"
)

func AddSomeProduct(vs []common.Product) (err error) {
	uri := "/man/upsert"
	err = post_info(uri, "ybgoods", nil, vs, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("SaveToDb fail! err=", err)
		return
	}
	return
}
