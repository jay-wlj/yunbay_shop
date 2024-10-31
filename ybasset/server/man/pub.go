package man

import (
	"strings"
	"yunbay/ybasset/common"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	//"yunbay/ybasset/common"
	//base "github.com/jay-wlj/gobaselib"
	//"time"
)

func CurrencyRatioSet(c *gin.Context) {
	var req share.RatioSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		glog.Error("CurrencyRatioSet fail! args invalid!", req)
		return
	}
	if err := req.UpdateRatio(); err != nil {
		glog.Error("CurrencyRatioSet fail! args invalid!", req)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

type drawSt struct {
	UserId   *int64  `json:"user_id,string" binding:"required,gte=0"`
	TxType   *int    `json:"tx_type"`
	ToUserId *int64  `json:"to_user_id"`
	Address  string  `json:"address"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Comment  string  `json:"comment"`
}

// 提币申请
func Wallet_Draw(c *gin.Context) {
	maner, err := util.GetHeaderString(c, "X-Yf-Maner")
	if maner == "" || err != nil {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var req drawSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	db := db.GetTxDB(c)
	if req.ToUserId != nil {
		// 获取用户地址
		var err error
		if req.Address, err = share.GetAndSaveUserAddress(*req.ToUserId); err != nil {
			glog.Error("Wallet_Draw cant find address by user_id:", *req.ToUserId)
			yf.JSON_FailEx(c, yf.ERR_ARGS_INVALID, gin.H{"msg": err.Error()})
			return
		}
	}

	if err = yf.Var(req.Address, "required"); err != nil {
		yf.JSON_FailEx(c, yf.ERR_ARGS_INVALID, gin.H{"msg": err.Error()})
		return
	}

	var id int64 = -1
	mext := make(map[string]interface{})
	mext["drawner"] = maner       // 提币人
	mext["comment"] = req.Comment // 提币说明
	d := share.DrawSt{Man: true, Id: &id, UserId: *req.UserId, ToUserId: req.ToUserId, TxType: req.TxType, Amount: req.Amount, Address: req.Address, Extinfos: mext}

	if reason, err := share.WithDraw(db, d); err != nil {
		glog.Error("Wallet_Draw fail! err=", err)
		if reason == "" {
			reason = yf.ERR_SERVER_ERROR
		}
		yf.JSON_Fail(c, reason)
		return
	}

	yf.JSON_Ok(c, gin.H{"id": id})
}

func CurrencyRatio(c *gin.Context) {
	var str_from, str_to string
	symbol, _ := base.CheckQueryStringField(c, "symbol")
	virtual_type, _ := base.CheckQueryIntField(c, "type")
	//user_id, _ := base.CheckQueryInt64Field(c, "user_id")
	if symbol == "" {
		from, _ := base.CheckQueryIntField(c, "from")
		to, _ := base.CheckQueryIntField(c, "to")
		str_from = common.GetCurrencyName(from)
		str_to = common.GetCurrencyName(to)
	} else {
		ss := strings.Split(symbol, "_")
		if len(ss) > 1 {
			str_from = ss[0]
			str_to = ss[1]
		}
	}

	if str_from == "" || str_to == "" {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	cache := share.RatioSt{From: str_from, To: str_to, Type: virtual_type}
	ratio, err := cache.Get()
	if err != nil {
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"ratio": ratio})
}
