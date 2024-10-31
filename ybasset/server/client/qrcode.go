package client

import (
	"yunbay/ybasset/common"
	"yunbay/ybasset/server/share"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"
	"net/url"
	"regexp"
	"strings"

	"github.com/jie123108/glog"

	"github.com/gin-gonic/gin"
)

type qrData struct {
	Content string `json:"code_str" binding:"required"`
}

// 收款
type CodeCollect struct {
	UserId int64   `json:"user_id"`
	TxType int     `json:"type,omitempty"`
	Amount float64 `json:"amount,omitempty"`
}

// 此处要修改结构内的变量 必须用指针结构的函数才能改变
func (t *CodeCollect) Parse(query_str string) (err error) {
	querys, err := url.ParseQuery(query_str)
	if err != nil {
		err = fmt.Errorf("ERR_ARGS_MISSING")
		return
	}
	t.UserId, err = base.CheckStringToInt64(querys.Get("user_id"))
	if err != nil {
		return
	}
	t.TxType, err = base.CheckStringToInt(querys.Get("type"))
	if err != nil {
		return
	}

	t.Amount, err = base.CheckStringToFloat64(querys.Get("amount"))
	if err != nil {
		return
	}
	return
}

// 通过扫码获取用户，币种及金额信息
func Qrcode_Query(c *gin.Context) {
	// var req qrData
	// if ok := util.UnmarshalReqParms(c, &req, "ERR_QRCODE_NOT_SUPPORT"); !ok {
	// 	return
	// }
	content, err := base.CheckQueryStringField(c, "code_str")
	content = strings.TrimSpace(content)
	if err != nil || content == "" {
		glog.Error("Qrcode_Query err=", err)
		yf.JSON_Fail(c, "ERR_ARGS_MISSING")
		return
	}
	var v CodeCollect

	// 如果是转帐地址
	if strings.HasPrefix(content, "0x") {
		if ok, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", content); !ok {
			goto not_supoort
		}
		// 查找地址所属用户
		u, err := share.GetUserInfoByRechargeAddress(content)
		if err != nil {
			goto not_supoort
		}
		v.UserId = u.UserId
		yf.JSON_Ok(c, v)
		return
	} else if strings.HasPrefix(content, "ybp://") {
		// 如果是支付连接
		content = strings.TrimPrefix(content, "ybp://")

		if v.Parse(content) != nil {
			glog.Error("Qrcode_Query err=", err, " content=", content)
			yf.JSON_Fail(c, "ERR_ARGS_MISSING")
			return
		}
		yf.JSON_Ok(c, v)
		return
	}
not_supoort:
	yf.JSON_Fail(c, common.ERR_QRCODE_NOT_SUPPORT)
}
