package util

import (
	"fmt"

	"github.com/jie123108/glog"
)

func GetRecommenders(user_id int64) (vs []int64, err error) {
	uri := fmt.Sprintf("/v1/invite/beinvite?user_id=%v", user_id)
	if err = get_info(uri, "ybapi", nil, "recommend_userids", &vs, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("GetRecommenders fail! err=", err)
		return
	}
	return
}

type BusinessSt struct {
	UserId int64   `json:"user_id"`
	Type   int     `json:"type"`
	Amount float64 `json:"amount"`
	Rebat  float64 `json:"rebat"`
}

func UpdateBusinessAmount(vs []BusinessSt) (err error) {
	uri := "/man/business/amount/update"
	if err = post_info(uri, "ybapi", nil, vs, "", nil, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("UpdateBusinessAmount fail! err=", err)
		return
	}
	return

}

type drawLimitSt struct {
	KT  float64 `json:"KT"`
	YBT float64 `json:"YBT"`
}

// 获取提币免审额限制接口
func GetDrawNoCheckAmountLimit(channel int) (ybt, kt float64, err error) {
	uri := fmt.Sprintf("/man/setting/get-draw-limit-one?type=%v", channel)
	var v drawLimitSt
	if err = get_info(uri, "ybproduct", nil, "", &v, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("GetDrawNoCheckAmountLimit fail! err=", err)
		return
	}
	ybt = v.YBT
	kt = v.KT
	return
}
