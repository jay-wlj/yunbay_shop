package dao

import (
	"github.com/jay-wlj/gobaselib/cache"
	"strconv"
	"yunbay/ybapi/common"

	"github.com/jie123108/glog"
)

const (
	PREFIX_LOTTERYS_STATUS string = "lotterys_status"

	//EXPIRE_TIME          time.Duration = time.Duration(240 * time.Hour)
)

type LoterysId int64

// 获取当前状态
func (t *LoterysId) GetStatus() (status int, err error) {
	ch, e := cache.GetWriter(common.RedisApi)
	if err = e; err != nil {
		glog.Error("GetStatus fail! err=", err)
		return
	}
	var val string
	if val, err = ch.HGet(PREFIX_LOTTERYS_STATUS, strconv.FormatInt(int64(*t), 10)); err == nil {
		status, err = strconv.Atoi(val)
	}
	return
}

// 设置当前销售状态
func (t *LoterysId) SetStatus(status int) (err error) {
	ch, e := cache.GetWriter(common.RedisApi)
	if err = e; err != nil {
		glog.Error("SetStatus fail! err=", err)
		return
	}

	err = ch.HSet(PREFIX_LOTTERYS_STATUS, strconv.FormatInt(int64(*t), 10), status, 0)
	return
}
