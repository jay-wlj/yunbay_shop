package client

import (
	"yunbay/ybasset/common"
	"yunbay/ybasset/server/share"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
)

type currencyReq struct {
	FromType int `form:"from_type"`
}

func CurrencyRatio(c *gin.Context) {
	from_type, _ := base.CheckQueryIntField(c, "from_type")
	to_type, _ := base.CheckQueryIntField(c, "to_type")

	from_type_str := common.GetCurrencyName(from_type)
	to_type_str := common.GetCurrencyName(to_type)
	if from_type_str == "" || to_type_str == "" {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	ratio := share.GetRatio(from_type_str, to_type_str)
	yf.JSON_Ok(c, gin.H{"ratio": ratio, "from_type": from_type, "to_type": to_type})
}

func CurrencyAllRatio(c *gin.Context) {
	to_type, _ := base.CheckQueryIntField(c, "to_type")
	to_type_str := common.GetCurrencyName(to_type)
	if to_type_str == "" {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	ratios := share.GetRatios(to_type_str)
	yf.JSON_Ok(c, gin.H{"ratios": ratios, "to_type": to_type})
}
