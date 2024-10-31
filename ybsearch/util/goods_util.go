package util

import (
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"yunbay/ybapi/common"

	"github.com/jie123108/glog"
)

type ListRes struct {
	List      []common.Product `json:"list"`
	ListEnded bool             `json:"list_ended"`
	Total     int              `json:"total"`
}

func ListProductInfoByIds(product_ids []uint64) (res ListRes, err error) {
	res = ListRes{}
	uri := fmt.Sprintf("/man/list_by_ids?ids=%v", base.Uint64SliceToString(product_ids, ","))
	err = get_info(uri, "ybgoods", "", &res, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("ListProductInfoByIds fail! err=", err)
		return
	}

	return
}
