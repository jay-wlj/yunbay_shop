package man

import (
	"strings"
	"time"

	"github.com/jay-wlj/gobaselib/db"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"

	//"github.com/jie123108/glog"
	"yunbay/ybasset/common"
	"yunbay/ybasset/dao"
	"yunbay/ybasset/server/share"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"

	//"github.com/jinzhu/gorm"
	"yunbay/ybasset/util"
)

var m_drawstatus map[string]int

func init() {
	m_drawstatus = make(map[string]int)
	m_drawstatus["waiting"] = common.TX_STATUS_WAITING
	m_drawstatus["submitted"] = common.TX_STATUS_SUBMIT
	m_drawstatus["confirming"] = common.TX_STATUS_CONFIRM
	m_drawstatus["failed"] = common.TX_STATUS_FAILED
	m_drawstatus["success"] = common.TX_STATUS_SUCCESS
}

type chargeSt struct {
	Symbol          string  `json:"symbol" binding:"oneof=ybt kt snet"` // oneof用空格隔开
	ContractAddress string  `json:"contract_address"`
	TxHash          string  `json:"tx_hash" binding:"required"`
	FromAddress     string  `json:"from_address"`
	Address         string  `json:"coin_address" binding:"required"`
	UserId          int64   `json:"user_id,string"`
	Amount          float64 `json:"amount" binding:"required"`
	BlockTime       string  `json:"block_time"`
}

// 通过地址查询内盘用户的资产信息等
func Man_BalanceByAddress(c *gin.Context) {
	address, _ := base.CheckQueryStringField(c, "address")
	addrs := strings.Split(address, ",")

	// 通过地址查询对应的帐号id
	var uw []common.UserWallet
	if err := db.GetDB().Find(&uw, "bind_address in(?)", addrs).Error; err != nil {
		glog.Error("Man_BalanceByAddress fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	ids := []int64{}
	for _, v := range uw {
		ids = append(ids, v.UserId)
	}
	var vs []common.UserAsset
	if err := db.GetDB().Find(&vs, "user_id in(?)", ids).Error; err != nil {
		glog.Error("Man_BalanceByAddress fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, vs)
}

type uIdt struct {
	UserId int64 `json:"user_id"`
	Type   int   `json:"type"`
}

// 获取指定用户的充值地址
func Wallet_Address(c *gin.Context) {
	var req uIdt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	str_type := common.GetCurrencyName(req.Type)
	if str_type == "" {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	// 先从库中查询
	var err error
	address, err := share.GetAndSaveUserAddress(req.UserId)
	//address, err := dao.GetUserWalletAddress(req.UserId)
	if err != nil {
		glog.Error("Wallet_Address fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// if address == "" {
	// 	// 创建帐号相关的地址
	// 	switch req.Type {
	// 	case common.CURRENCY_YBT, common.CURRENCY_KT:
	// 	default:
	// 		yf.JSON_Fail(c, common.ERR_TYPE_NOT_SUPPORT)
	// 		return
	// 	}

	// 	address, err = util.GetUserWalletAddress(req.UserId, "kt")
	// 	if err != nil {
	// 		glog.Error("GetUserWalletAddress fail! err=", err)
	// 		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 		return
	// 	}
	// 	if err = share.AddUserWalletAddress(req.UserId, address, req.Type); err != nil {
	// 		glog.Error("Wallet_Address save fail! err=", err)
	// 		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 		return
	// 	}

	// }
	yf.JSON_Ok(c, gin.H{"user_id": req.UserId, "bind_address": address})
}

// 充值回调
func Wallet_Charge_Callback(c *gin.Context) {
	if !util.Chain_SignCheck(c) {
		return
	}
	req := []chargeSt{}
	if ok := util.UnmarshalReqParms(c, &req, "INVALID_ARGS"); !ok {
		return
	}
	vs := []string{}
	today := time.Now().Format("2006-01-02")

	db := db.GetTxDB(c)
	for _, v := range req {
		// 效验该用户充值记录及地址
		if address, err := dao.GetUserWalletAddress(v.UserId); err == nil && address == v.Address {
			//if err = db.Find(&w, "user_id=? and bind_address=?", v.UserId, v.Address).Error; err == nil {
			txType := share.GetCurrencyTypeByCoin(v.Symbol)
			if txType < -1 {
				glog.Error("Wallet_Charge_Callback fail! txType:", txType, " symbol:", v.Symbol)
				yf.JSON_Fail(c, "INVALID_ARGS")
				return
			}
			// txType := common.CURRENCY_KT
			// transaction_type := common.TRANSACTION_RECHARGE
			// if strings.ToLower(v.Symbol) == "ybt" {
			// 	txType = common.CURRENCY_YBT
			// 	transaction_type = common.YBT_TRANSACTION_RECHARGE
			// }

			// 添加充值交易明细记录
			u := common.UserAssetDetail{UserId: v.UserId, Amount: v.Amount, Type: txType, TransactionType: common.TRANSACTION_RECHARGE, Date: today}
			if err = db.Save(&u).Error; err != nil {
				glog.Error("Chain_Charge fail! UserAssetDetail err=", err)
				yf.JSON_Fail(c, "SYSTEM_ERR")
				return
			}
			// 添加到充值流水记录中
			fw := common.RechargeFlow{UserId: v.UserId, TxHash: v.TxHash, FromAddress: v.FromAddress, AssetId: u.Id, Address: v.Address, TxType: txType, Amount: v.Amount, BlockTime: v.BlockTime, Date: today}
			if err = db.Save(&fw).Error; err != nil {
				glog.Error("Chain_Charge fail! UserAssetDetail err=", err)
				yf.JSON_Fail(c, "SYSTEM_ERR")
				return
			}

			vs = append(vs, v.TxHash)
		} else {
			glog.Error("Wallet_Charge_Callback fail! user_id=", v.UserId, " err=", err)
		}
	}

	yf.JSON_Ok(c, vs)
}

// 提币回调
func Wallet_Withdraw_Callback(c *gin.Context) {
	// if !util.Chain_SignCheck(c){
	// 	return
	// }
	var req share.WithDrawSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	if _, ok := m_drawstatus[req.Status]; !ok {
		glog.Error("Chain_Withdraw_Callback fail! org status=", req.Status)
		return
	}
	db := db.GetTxDB(c)
	if reason, err := share.WithDrawCallbackHandle(db, req); err != nil {
		if reason != "" {
			yf.JSON_Fail(c, reason)
		} else {
			yf.JSON_Fail(c, "INVALID_ARGS")
		}
		glog.Error("Wallet_Withdraw_Callback fail! err=", err)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}
