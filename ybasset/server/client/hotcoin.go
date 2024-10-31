package client

import (
	"net/url"
	"yunbay/ybasset/common"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 获取token加密串
func Wallet_HotCoinTokenGet(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	token, _ := util.GetHeaderString(c, "X-Yf-Token")
	if token == "" {
		glog.Error("Wallet_HotCoinTokenGet fail! token is empty! user_id=", user_id)
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	token_str := util.AESEncrypt(token)
	token_str = url.QueryEscape(token_str)
	yf.JSON_Ok(c, gin.H{"token": token_str})
}

// 通过token获取用户id及电话
func HotCoinToken(c *gin.Context) {
	if !util.HotCoin_SignCheck(c) {
		return
	}
	token_str, _ := base.CheckQueryStringField(c, "token")
	token, err := util.AESDecrypt(token_str)
	if err != nil || token == "" {
		glog.Error("User_InfoGet fail! err=", err, "token=", token)
		yf.JSON_Fail(c, "INVALID_TOKEN")
		return
	}
	v, err := util.UserInfoGet(token)
	if err != nil {
		glog.Error("User_InfoGet fail! err=", err)
		yf.JSON_Fail(c, "INVALID_TOKEN")
		return
	}
	yf.JSON_Ok(c, gin.H{"user_id": v.UserId, "phone": v.Cc + v.Tel})
}

// 查询热币用户充值地址
func Wallet_HotCoinAddress(c *gin.Context) {
	_, ok := util.GetUid(c)
	if !ok {
		return
	}
	token, _ := util.GetHeaderString(c, "X-Yf-Token")
	v, err := util.UserInfoGet(token)
	if err != nil {
		glog.Error("Wallet_HotCoinAddress fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}
	var address string
	if address, err = util.QueryHotCoinAddress(v.Cc + v.Tel); err != nil {
		if err.Error() == "USER_NOT_FOUND" {
			yf.JSON_Fail(c, common.ERR_HOTCOIN_USER_NOT_FOUND)
			return
		}
		glog.Error("Wallet_HotCoinAddress fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	share.AddAddressChannel(address, common.CHANNEL_HOTCOIN) // 添加热币地址标识
	yf.JSON_Ok(c, gin.H{"address": address})
}

// 通过用户充值地址查询用户id
func Wallet_AddressQuery(c *gin.Context) {
	if !util.HotCoin_SignCheck(c) {
		return
	}
	coin, _ := base.CheckQueryStringField(c, "coin")
	address, _ := base.CheckQueryStringField(c, "address")

	if coin != "KT" || address == "" {
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

type hotchargeSt struct {
	Coin     string  `json:"coin" binding:"required"`
	UserId   int64   `json:"user_id"`
	Address  string  `json:"address"`
	OrderId  string  `json:"order_id" binding:"gt=0"`
	Amount   float64 `json:"amount,string" binding:"gt=0"`
	Platform string  `json:"platform"`
}

// 充值回调
func HotCoin_Charge_Callback(c *gin.Context) {
	if !util.HotCoin_SignCheck(c) {
		return
	}
	v := hotchargeSt{}
	if ok := util.UnmarshalReq(c, &v); !ok {
		return
	}

	db := db.GetTxDB(c)
	var err error
	address := v.Address
	// 效验该铁丝充值记录及地址
	if v.UserId > 0 {
		if address, err = share.GetAndSaveUserAddress(v.UserId); err != nil {
			glog.Error("HotCoin_Charge_Callback fail! err=", err, " user_id:", v.UserId)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	} else if v.Address == "" {
		yf.JSON_Fail(c, "INVALID_ARGS")
		return
	}

	id, reason, err := share.Recharge_fromthird(db, "hotcoin", address, v.Coin, v.Amount, v.OrderId)
	if err != nil {
		if reason != "" {
			yf.JSON_Fail(c, reason)
		} else {
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		}
		return
	}

	yf.JSON_Ok(c, gin.H{"order_id": id, "user_id": v.UserId})
}
