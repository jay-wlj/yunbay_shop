package util

import (
	"fmt"

	"github.com/jie123108/glog"
	"github.com/shopspring/decimal"
)

func GetDifficult() (dif float64, err error) {
	uri := "/v1/ybasset/difficult"

	err = get_info(uri, "ybasset", "difficult", &dif, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("GetDifficult fail! err=", err)
		return
	}
	return
}

func GetCurrencyRatio(from, to, virtual int) (val decimal.Decimal, err error) {
	if virtual > 0 {
		virtual = 1
	}
	uri := fmt.Sprintf("/man/currency/ratio?from=%v&to=%v&type=%v", from, to, virtual)

	err = get_info(uri, "ybasset", "ratio", &val, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("GetCurrencyRatio fail! err=", err)
		return
	}
	return
}
