package client

import (
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 查询yunex用户充值地址
func Wallet_YunexAddressQuery(c *gin.Context) {
	_, ok := util.GetUid(c)
	if !ok {
		return
	}
	coin_type, _ := base.CheckQueryIntDefaultField(c, "coin", 1)
	token, _ := util.GetHeaderString(c, "X-Yf-Token")
	v, err := util.UserInfoGet(token)
	if err != nil {
		glog.Error("Wallet_HotCoinAddress fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}
	coin := common.GetCurrencyName(coin_type)
	var address string
	if address, err = util.QueryYunexAddress(v.Cc, v.Tel, coin); err != nil {
		if err.Error() == "USER_NOT_FOUND" {
			yf.JSON_Fail(c, common.ERR_YUNEX_USER_NOT_FOUND)
			return
		}
		glog.Error("Wallet_HotCoinAddress fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"address": address})
}

type yunexchargeSt struct {
	Coin    string  `json:"coin" binding:"required"`
	Address string  `json:"address"`
	OrderId string  `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount,string" binding:"required"`
}

// yunex充值回调
func Yunex_Charge_Callback(c *gin.Context) {
	if !util.Yunex_SignCheck(c) {
		return
	}
	v := yunexchargeSt{}
	if ok := util.UnmarshalReq(c, &v); !ok {
		return
	}

	db := db.GetTxDB(c)
	id, reason, err := share.Recharge_fromthird(db, "yunex", v.Address, v.Coin, v.Amount, v.OrderId)
	if err != nil {
		if reason != "" {
			yf.JSON_Fail(c, reason)
		} else {
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		}
		return
	}
	yf.JSON_Ok(c, gin.H{"order_id": fmt.Sprintf("%v", id)})
}

// 通过用户充值地址查询用户id
func Yunex_AddressQuery(c *gin.Context) {
	if !util.Yunex_SignCheck(c) {
		return
	}
	coin, _ := base.CheckQueryStringField(c, "coin")
	address, _ := base.CheckQueryStringField(c, "address")
	coin = strings.ToLower(coin)

	if (coin != "ybt" && coin != "kt") || address == "" {
		glog.Error("Yunex_AddressQuery faiL! coin=", coin, " address=", address)
		yf.JSON_Fail(c, "INVALID_ARGS")
		return
	}
	v, err := share.GetUserInfoByRechargeAddress(address)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, "NOT_FOUND")
			return
		}
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"user_id": v.UserId})
}

type depositSt struct {
	TxHash string `json:"order_id"`
	Reason string `json:"reason"`
	Status string `json:"status"`
}

// 提币回调接口
func Yunex_DepositNotify(c *gin.Context) {
	if !util.Yunex_SignCheck(c) {
		return
	}
	var req depositSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	// var status string
	// switch req.Status {
	// case 1:
	// 	status = "success"
	// case 3:
	// 	status = "confirming" // 审核中
	// default:
	// 	status = "failed"		// 其它视为失败处理
	// }

	w := share.WithDrawSt{TxHash: req.TxHash, Status: req.Status, Reason: req.Reason, Channel: common.CHANNEL_YUNEX}
	db := db.GetTxDB(c)
	if reason, err := share.WithDrawCallbackHandle(db, w); err != nil {
		if reason != "" {
			yf.JSON_Fail(c, reason)
		} else {
			yf.JSON_Fail(c, "INVALID_ARGS")
		}
		glog.Error("Yunex_DepositNotify fail! err=", err)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

type amountSt struct {
	Amount float64 `json:"amount"`
}

// 查询yunex帐户在yunbay平台里的余额信息
func Yunex_BalanceQuery(c *gin.Context) {
	if !util.Yunex_SignCheck(c) {
		return
	}
	third_id := conf.Config.ThirdAccount["yunex"].UserId
	if 0 == third_id {
		glog.Error("Yunex_BalanceQuery fail! third yunex user_id not define!")
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	var v common.UserAsset
	db := db.GetDB()
	if err := db.Find(&v, "user_id=?", third_id).Error; err != nil {
		glog.Error("Yunex_BalanceQuery fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	var amt amountSt
	if err := db.Model(&common.UserAssetDetail{}).Where("type=? and transaction_type=? and user_id=?", common.CURRENCY_KT, common.KT_TRANSACTION_PROFIT, conf.Config.ThirdAccount["yunex"].UserId).Select("sum(amount) as amount").Scan(&amt).Error; err != nil {
		glog.Error("Yunex_BalanceQuery fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, util.CoinBalance{YBT: v.TotalYbt, KT: v.TotalKt, Bonus: amt.Amount})
}
