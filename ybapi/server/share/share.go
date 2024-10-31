package share

import (
	"fmt"
	"strings"
	"yunbay/ybapi/common"
	"yunbay/ybapi/util"

	"github.com/jay-wlj/gobaselib/cache"

	"github.com/jayden211/retag"
	"github.com/jie123108/glog"
	//"encoding/json"
)

func GetCurrencyTypeByCoin(coin string) int {
	coin = strings.ToLower(coin)
	switch coin {
	case "ybt":
		return common.CURRENCY_YBT
	case "kt":
		return common.CURRENCY_KT
	case "cny":
		return common.CURRENCY_RMB
	case "snet":
		return common.CURRENCY_SNET
	default:
		return -1
	}
}

func GetCoinByCurrencyType(txType int) string {
	switch txType {
	case common.CURRENCY_YBT:
		return "ybt"
	case common.CURRENCY_KT:
		return "kt"
	case common.CURRENCY_RMB:
		return "cny"
	case common.CURRENCY_SNET:
		return "snet"
	}
	return ""
}

// 获取币种转换兑换比例
func GetRatioCache(to_type, from_type int) (ratio float64, err error) {
	cache, err := cache.GetWriter(common.RedisPub)

	to_str := common.GetCurrencyName(to_type)
	from_str := common.GetCurrencyName(from_type)
	// 优先从缓存里获取
	if err == nil {
		key := fmt.Sprintf("asset_ratio_%v", to_str)
		ratio, err = cache.HGetF64(key, from_str)
	}
	// 否则从资产接口中获取
	if err != nil {
		var ratios map[string]float64
		ratios, err = util.YBAsset_GetRatio(to_type)
		if v, ok := ratios[from_str]; ok {
			ratio = v
		} else {
			glog.Error("GetRatioCache fail! to_type:", to_str, " no from_type:", from_str)
		}
	}
	return
}

func Filter_Obj(v interface{}, tag string) (obj interface{}) {
	obj = retag.Convert(v, retag.NewView("json", tag))
	return
}
