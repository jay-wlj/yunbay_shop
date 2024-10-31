package dao

import (
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"yunbay/ybgoods/common"

	"github.com/jie123108/glog"
)

type goods struct {
}

func (t *goods) GetByIds(ids []int64) (vs []common.Product, err error) {
	vs = []common.Product{}
	if len(ids) == 0 {
		return
	}
	order_sql := fmt.Sprintf("array_position('{%v}'::bigint[], id)", base.Int64SliceToString(ids, ",")) // 按ids顺序排序
	if err = db.GetDB().Order(order_sql).Select("id,category_id,type,publish_area,title,price,rebat,stock,sold,images[1:1],extinfo").Find(&vs, "status=? and is_hid=0 and id in(?)", common.STATUS_OK, ids).Error; err != nil {
		glog.Error("GetByIds fail! err=", err)
		return
	}

	return
}

func (t *goods) ListMore(pre_id int64, page_size int, exclude_ids []int64) (rs []interface{}, err error) {

	return
}
