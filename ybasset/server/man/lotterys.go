package man

import (
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"

	"github.com/jie123108/glog"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// 用户支付给抽奖帐号
func LotterysPay(c *gin.Context) {
	var req share.TransferPaySt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	req.To = conf.Config.SystemAccounts["lotterys_account"] // 将用户的资产转给抽奖帐号
	db := db.GetTxDB(c)
	if err := req.Transfer(db); err != nil {
		glog.Error("LotterysPay1 fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}
	yf.JSON_Ok(c, gin.H{"id": req.Id, "key": req.Key})
}

func TransferRefund(c *gin.Context) {
	var req share.TransferRefundSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	db := db.GetTxDB(c)
	if err := req.Refund(db); err != nil {
		glog.Error("LotterysRefund fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

type lottSt struct {
	OrderId int64 `json:"order_id"`
	//LotteryId    int64           `json:"lottery_id"`
	//PublishArea  int             `json:"publish_area"`
	CurrencyType int             `json:"currency_type"`
	Amount       decimal.Decimal `json:"amount"`
	SellerUserId int64           `json:"seller_userid"`
	SellerKt     decimal.Decimal `json:"seller_kt"`
}

type amountSt struct {
	Amount decimal.Decimal `json:"amount"`
}

// 将抽奖池里的该计划 根据订单id下单 生成订单流水记录
func LotterysOrderPay(c *gin.Context) {
	var req lottSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	v := common.YBAssetPool{OrderId: req.OrderId, CurrencyType: req.CurrencyType, PayerUserId: conf.Config.SystemAccounts["lotterys_account"]}
	v.SellerUserId = req.SellerUserId
	v.SellerAmount, _ = req.Amount.Float64()
	// 根据币种兑换相应的kt
	cache := share.RatioSt{From: common.GetCurrencyName(req.CurrencyType), To: "kt"}
	ratio, err := cache.Get()
	if err != nil {
		glog.Error("LotterysPay fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	v.SellerKt, _ = req.Amount.Mul(ratio).Float64() // 换算成kt 并作为结算给商户
	v.PayAmount, _ = req.Amount.Float64()
	v.PublishArea = common.PUBLISH_AREA_YBT
	v.Date = base.GetCurDay()

	db := db.GetTxDB(c)
	if reason, err := pay(db, conf.Config.SystemAccounts["lotterys_account"], []common.YBAssetPool{v}); err != nil {
		glog.Error("Asset_Pay fail! err=", err, " reason:", reason)
		yf.JSON_Fail(c, reason)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}
