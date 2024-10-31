package client

import (
	"strings"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	//"time"
)

func Wallet_ChargeSwitch(c *gin.Context) {
	yf.JSON_Ok(c, gin.H{"list": conf.Config.Switch})
}

// type drawSt struct {
// 	Channel *int `json:"channel"`
// 	TxType *int `json:"tx_type" binding:"required,min=0,max=1"`
// 	Amount float64 `json:"amount"`
// 	Fee float64 `json:"fee"`
// 	AddressId int64 `json:"address_id"`
// 	Address string `json:"address"`
// 	ZJPassword string `json:"zjpassword"`
// 	Code string `json:"code"`
// }

func Wallet_Fee(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	txType, _ := base.CheckQueryIntDefaultField(c, "tx_type", -1)
	address, _ := base.CheckQueryStringField(c, "address")
	to_id, _ := base.CheckQueryInt64Field(c, "to_uid")
	chl, _ := base.CheckQueryIntDefaultField(c, "channel", common.CHANNEL_UNKNOW)

	if chl >= common.CHANNEL_ALIPAY && chl <= common.CHANNEL_BANK {
		txType = common.CURRENCY_RMB
	}
	fees, channel, err := share.GetFeeByAddress(to_id, address, &chl)
	if err != nil {
		glog.Error("Wallet_Fee fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// str_type := common.GetCurrencyName(txType)
	for _, v := range fees {
		if v.Type == txType {
			// 限制ybt的提取数量
			if v.Type == common.CURRENCY_YBT {
				user_type, ok := util.GetUtype(c)
				if !ok {
					return
				}
				// 仅限制商家用户的提币
				if user_type == common.USER_TYPE_BUSINESS {
					var max float64 = 0
					if max, err = share.GetUserWithdarwAvalable(user_id, v.Type, v.Max); err != nil {
						glog.Error("Wallet_Fee fail! GetUserWithdarwAvalable err=", err)
						yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
						return
					}
					if max >= 0 {
						v.Max = &max
					}
				}
			} else { // rmb限额
				var max float64 = 0
				if max, err = share.GetUserWithdarwAvalable(user_id, v.Type, v.Max); err != nil {
					glog.Error("Wallet_Fee fail! GetUserWithdarwAvalable err=", err)
					yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
					return
				}
				if max >= 0 {
					v.Max = &max
				}
			}

			yf.JSON_Ok(c, gin.H{"list": []conf.Fee{v}, "channel": channel})
			return
		}
	}

	yf.JSON_Ok(c, gin.H{"list": fees, "channel": channel})
}

// 提币申请
func Wallet_Draw(c *gin.Context) {
	user_type, ok := util.GetUtype(c)
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	req := share.DrawSt{}
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	if (req.Channel != nil && *req.Channel > common.CHANNEL_BANK) || req.Amount < 0 || req.Amount < req.Fee {
		glog.Error("Wallet_Draw fail!")
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	token, _ := util.GetHeaderString(c, "X-Yf-Token")
	var id int64

	req.UserType = int(user_type)
	req.UserId = user_id
	req.Token = token
	req.Id = &id

	db := db.GetTxDB(c)
	if reason, err := share.WithDraw(db, req); err != nil {
		glog.Error("Wallet_DrawAlipay fail! err=", err)
		if reason == "" {
			reason = yf.ERR_SERVER_ERROR
		}
		yf.JSON_Fail(c, reason)
		return
	}
	yf.JSON_Ok(c, gin.H{"id": id})
}

func Wallet_Withdraw_List(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	txType, _ := base.CheckQueryIntDefaultField(c, "type", common.CURRENCY_KT)
	//country, _ := base.CheckQueryIntDefaultField(c, "country", -1)
	//country := util.GetCountry(c)

	fromType := txType
	if fromType == common.CURRENCY_RMB {
		fromType = common.CURRENCY_KT
	}

	vs := []common.WithdrawFlow{}
	db := db.GetDB()
	db.DB = db.Model(&common.WithdrawFlow{}).Where("user_id=? and tx_type=?", user_id, fromType)
	// if country > -1 {
	// 	db.DB = db.Where("country=?", country)
	// }
	var total int = 0
	if err := db.Count(&total).Error; err != nil {
		glog.Error("Wallet_Address_Upsert faiL! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if err := db.ListPage(page, page_size).Order("create_time desc").Find(&vs).Error; err != nil {
		glog.Error("Wallet_Address_Upsert faiL! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	for i, _ := range vs {
		vs[i].Amount += vs[i].Fee
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": base.IsListEnded(page, page_size, len(vs), total), "total": total})
}

func Wallet_Recharge_List(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	txType, _ := base.CheckQueryIntDefaultField(c, "type", common.CURRENCY_KT)

	if txType == common.CURRENCY_RMB {
		txType = common.CURRENCY_KT
	}

	db := db.GetDB()
	db.DB = db.Model(&common.RechargeFlow{}).Where("user_id=? and tx_type=?", user_id, txType)
	var total int = 0
	if err := db.Count(&total).Error; err != nil {
		glog.Error("Wallet_Recharge_List faiL! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	vs := []common.RechargeFlow{}
	if err := db.ListPage(page, page_size).Order("create_time desc").Find(&vs).Error; err != nil {
		glog.Error("Wallet_Address_Upsert faiL! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": base.IsListEnded(page, page_size, len(vs), total), "total": total})
}

type AliSt struct {
	Account     string `json:"account" binding:"required"`
	AccountName string `json:"account_name" binding:"required"`
}

type rmbdrawSt struct {
	AliSt
	Amount     float64 `json:"amount" binding:"required,gte=0.1"`
	Fee        float64 `json:"fee"`
	ZJPassword string  `json:"zjpassword" binding:"required"`
	Code       string  `json:"code"`
}

// 支付宝提现接口
func Wallet_DrawAlipay(c *gin.Context) {
	user_type, ok := util.GetUtype(c)
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req rmbdrawSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	channel := common.CHANNEL_ALIPAY
	txType := common.CURRENCY_KT
	token, _ := util.GetHeaderString(c, "X-Yf-Token")
	country := util.GetCountry(c)
	var id int64

	extinfos := make(map[string]interface{})
	extinfos["ali_info"] = req.AliSt

	d := share.DrawSt{Id: &id, UserId: user_id, UserType: int(user_type), Token: token, Channel: &channel, TxType: &txType, Amount: req.Amount, Fee: req.Fee, Address: req.Account, ZJPassword: req.ZJPassword, Code: req.Code, Extinfos: extinfos, Country: country}
	db := db.GetTxDB(c)
	if reason, err := share.WithDraw(db, d); err != nil {
		glog.Error("Wallet_DrawAlipay fail! err=", err)
		if reason == "" {
			reason = yf.ERR_SERVER_ERROR
		}
		yf.JSON_Fail(c, reason)
		return
	}
	yf.JSON_Ok(c, gin.H{"id": id})
}

type BankSt struct {
	Name     string `json:"account_name" binding:"required"`
	BankName string `json:"bank_name" binding:"required"`
	CardId   string `json:"bank_card" binding:"required"`
	Area     string `json:"bank_area"`
	Branch   string `json:"bank_branch"`
}

type bankdrawSt struct {
	Amount     float64 `json:"amount" binding:"required,gte=0.1"`
	Fee        float64 `json:"fee"`
	BankParmas BankSt  `json:"bank_info" binding:"required"`
	ZJPassword string  `json:"zjpassword" binding:"required"`
	Code       string  `json:"code"`
}

// 提现到银行卡
func Wallet_DrawBank(c *gin.Context) {
	user_type, ok := util.GetUtype(c)
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req bankdrawSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	// 检验银行卡号
	if !util.Luhn(req.BankParmas.CardId) {
		yf.JSON_Fail(c, common.ERR_BANK_CARDID_ERROR)
		glog.Error("Wallet_DrawBank fail! ERR_BANK_CARDID_ERROR card_id:", req.BankParmas.CardId)
		return
	}
	// 检验开户行
	if bank, err := util.QueryBankName(req.BankParmas.CardId); err == nil {
		if req.BankParmas.BankName != bank.BankName {
			glog.Error("Wallet_DrawBank fail! ERR_BANK_NAME_ERROR card_id:", req.BankParmas.CardId, " is:", bank.BankName, " not:", req.BankParmas.BankName)
			yf.JSON_Fail(c, common.ERR_BANK_NAME_ERROR)
			return
		}
	} else {
		glog.Error("Wallet_DrawBank fail! err=", err)
	}

	channel := common.CHANNEL_BANK
	txType := common.CURRENCY_KT
	token, _ := util.GetHeaderString(c, "X-Yf-Token")
	country := util.GetCountry(c)
	var id int64

	extinfos := make(map[string]interface{})
	extinfos["bank_info"] = req.BankParmas
	d := share.DrawSt{Id: &id, UserId: user_id, UserType: int(user_type), Token: token, Channel: &channel, TxType: &txType, Amount: req.Amount, Fee: req.Fee, Address: req.BankParmas.CardId, ZJPassword: req.ZJPassword, Code: req.Code, Extinfos: extinfos, Country: country}
	db := db.GetTxDB(c)
	if reason, err := share.WithDraw(db, d); err != nil {
		glog.Error("Wallet_DrawAlipay fail! err=", err)
		if reason == "" {
			reason = yf.ERR_SERVER_ERROR
		}
		yf.JSON_Fail(c, reason)
		return
	}
	yf.JSON_Ok(c, gin.H{"id": id})
}

type payInfo struct {
	UserId     int64   `json:"user_id" binding:"required" binding:"gt=0"`
	Amount     float64 `json:"amount" binding:"gt=0"`
	TxType     int     `json:"type" binding:"oneof=0 1 3 4 5"`
	Type       int     `json:"tx_type" binding:"oneof=0 1 3 4 5"`
	ZJPassword string  `json:"zjpassword" binding:"required"`
}

// 平台用户内部转帐操作
func Wallet_Transfer(c *gin.Context) {
	var req payInfo
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	if 0 == req.TxType {
		// 兼容客户端传的tx_type参数
		req.TxType = req.Type
	}
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	if user_id == req.UserId {
		yf.JSON_Fail(c, common.ERR_FORBIDDEN_TRANSFER_TO_OWN)
		return
	}
	country := util.GetCountry(c)
	// 验证支付密码
	token, _ := util.GetHeaderString(c, "X-Yf-Token")
	if err := util.AuthUserZJPassword(token, req.ZJPassword); err != nil {
		yf.JSON_Fail(c, err.Error())
		return
	}
	// 验证通过 开始转帐操作
	today := time.Now().Format("2006-01-02")
	now := time.Now().Unix()
	ua := []common.UserAssetDetail{}
	ua = append(ua, common.UserAssetDetail{UserId: user_id, Type: req.TxType, TransactionType: common.TRANSACTION_TRANSFER, Amount: -req.Amount, Date: today})   // 转出
	ua = append(ua, common.UserAssetDetail{UserId: req.UserId, Type: req.TxType, TransactionType: common.TRANSACTION_TRANSFER, Amount: req.Amount, Date: today}) // 转入

	var err error
	db := db.GetTxDB(c)
	for _, v := range ua {
		if err = db.Save(&v).Error; err != nil {
			reason := yf.ERR_SERVER_ERROR
			if strings.Index(err.Error(), "cannot be negtive") > 0 {
				reason = common.ERR_MONEY_NOT_MORE
			}
			glog.Error("Wallet_Transfer fail! err=", err)
			yf.JSON_Fail(c, reason)
			return
		}
	}
	// 添加资产出入记录
	r := common.RechargeFlow{UserId: req.UserId, AssetId: ua[1].Id, Channel: common.CHANNEL_CHAIN, FlowType: common.FLOW_TYPE_YUNBAY, TxType: req.TxType, Amount: req.Amount, Date: today, Country: country, TxHash: util.GetSnowflake().Generate().String()}
	if err = db.Save(&r).Error; err != nil {
		glog.Error("Wallet_Transfer fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	w := common.WithdrawFlow{UserId: user_id, Channel: common.CHANNEL_CHAIN, FlowType: common.FLOW_TYPE_YUNBAY, TxType: req.TxType, ToUserId: req.UserId, Amount: req.Amount, Status: common.TX_STATUS_SUCCESS, Country: country, Maner: "system", Date: today, CreateTime: now, UpdateTime: now}
	if err = db.Save(&w).Error; err != nil {
		glog.Error("Wallet_Transfer fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}
