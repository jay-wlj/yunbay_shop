package share

import (
	"fmt"
	"strings"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"
	"yunbay/ybasset/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	yf "github.com/jay-wlj/gobaselib/yf"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
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

// 根据第三方平台获取渠道充值类型
func getchannelByThirdPlat(plat string) int {
	switch plat {
	case "hotcoin":
		return common.CHANNEL_HOTCOIN
	case "yunex":
		return common.CHANNEL_YUNEX
	}
	return -1
}

func GetThirdUserIdByChannel(channel int) (third_id int64) {
	third_id = -1
	switch channel {
	case common.CHANNEL_HOTCOIN:
		third_id = conf.Config.ThirdAccount["hotcoin"].UserId
	case common.CHANNEL_YUNEX:
		third_id = conf.Config.ThirdAccount["yunex"].UserId
	default:
		s := fmt.Sprintf("channel's user_id not define! channel:%v", channel)
		glog.Error(s)
	}
	return
}

// 获取第三方提币进帐帐号
func GetThirdWithdarwIdByChannel(channel int) (third_id int64) {
	third_id = -1
	switch channel {
	case common.CHANNEL_HOTCOIN:
		third_id = conf.Config.ThirdAccount["hotcoin"].WithDrawId
	case common.CHANNEL_YUNEX:
		third_id = conf.Config.ThirdAccount["yunex"].WithDrawId
	case common.CHANNEL_ALIPAY, common.CHANNEL_WEIXIN, common.CHANNEL_BANK:
		third_id = conf.Config.SystemAccounts["rmb_withdraw_account"]

	default:
		s := fmt.Sprintf("channel's user_id not define! channel:%v", channel)
		glog.Error(s)
	}
	return
}

func GetThirdPlatFromBonusId(bonus_id int64) string {
	for k, v := range conf.Config.ThirdAccount {
		if v.BonusId == bonus_id {
			return k
		}
	}
	return ""
}

// 从第三方平台帐户充值到平台用户
func Recharge_fromthird(db *db.PsqlDB, third_plat string, address string, coin string, amount float64, txHash string) (id int64, reason string, err error) {
	defer func() {
		if err != nil {
			// 发生错误 需要报警处理
			s := fmt.Sprintf("第三方平台[%v]充值失败, 错误码:%v", third_plat, err.Error())
			util.SendDingTextTalk(s, []string{"15818717950"})
		}
	}()

	txType := GetCurrencyTypeByCoin(coin)
	if txType < 0 {
		glog.Error("Recharge_fromthird fail! coin not define! coin:", coin)
		reason = "INVALID_ARGS"
		err = fmt.Errorf(reason)
		return
	}
	third, ok := conf.Config.ThirdAccount[third_plat]
	if !ok || third.UserId == 0 {
		s := fmt.Sprintln("Recharge_fromthird fail! third plat:", third_plat, " is not define!")
		glog.Error(s)
		reason = "INVALID_ARGS"
		err = fmt.Errorf(s)
		return
	}
	third_id := third.UserId // 获取第三方平台帐户id

	// 查询第三方order_id是否已存在
	if txHash != "" {
		var o common.RechargeFlow
		if err = db.Find(&o, "txhash=?", txHash).Error; err != nil && err != gorm.ErrRecordNotFound {
			glog.Error("Recharge_fromthird fail! err=", err)
			return
		}
		// 该笔充值订单已存在
		if o.Id > 0 {
			id = o.Id
			return
		}
	}

	// 效验该铁丝充值记录及地址
	if address == "" {
		reason = "INVALID_ARGS"
		err = fmt.Errorf("INVALID_ARGS")
		return
	}
	var w common.UserWallet
	if err = db.Find(&w, "bind_address=?", address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			reason = "NOT_FOUND"
			return
		}
		glog.Error("recharge_fromthird fail! err=", err)
		reason = yf.ERR_SERVER_ERROR
		return
	}

	// 获取第三方帐号资金
	var hot common.UserAsset
	if err = db.Find(&hot, "user_id=?", third_id).Error; err != nil {
		glog.Error("yunex user not found! user_id:", third_id)
		reason = "INVALID_ARGS"
		return
	}
	var normal_amount float64

	switch txType {
	case common.CURRENCY_YBT:
		normal_amount = hot.NormalYbt
	case common.CURRENCY_KT:
		normal_amount = hot.NormalKt
	case common.CURRENCY_SNET:
		normal_amount = hot.NormalSnet
	}

	// 报警检测
	var alarm_email string
	var bTips bool
	if str_alarm_blance, ok := third.Ext["alarm_balance"]; ok {
		alarm_balance, _ := base.StringToFloat64(str_alarm_blance)
		if alarm_balance >= normal_amount {
			// 第三方帐户余额不足 及时通知
			if alarm_email, ok = third.Ext["alarm_email"]; ok {
				s := fmt.Sprintf("您帐号[%v]余额已不足%v，请及时充值！剩余%v:%v 本次用户充值需划转:%v", third_id, alarm_balance, common.GetCurrencyName(txType), normal_amount, amount)
				util.PublishMsg(common.MQMail{Receiver: []string{alarm_email}, Subject: "Yunbay商城", Content: s})
				util.SendDingTextTalk(s, []string{"15818717950"})
				bTips = true
			}
		}
	}

	if normal_amount < amount {
		// 忽略已经提示的
		if !bTips {
			s := fmt.Sprintf("您帐号[%v]余额已不足，无法划转，请及时充值！剩余%v:%v  本次用户充值需划转:%v", third_id, common.GetCurrencyName(txType), normal_amount, amount)
			if alarm_email != "" {
				util.PublishMsg(common.MQMail{Receiver: []string{alarm_email}, Subject: "Yunbay商城", Content: s})
			}
			util.SendDingTextTalk(s, []string{"15818717950"})
		}

		glog.Error("hotcoin normal  not more! normal=", normal_amount, " user recharge =", amount)
		reason = "INSUFFICIENT_FUNDS"
		err = fmt.Errorf("INSUFFICIENT_FUNDS")
		return
	}

	today := time.Now().Format("2006-01-02")

	// 将第三方帐号的资金划扣给用户
	u := common.UserAssetDetail{UserId: hot.UserId, Amount: -amount, Type: txType, TransactionType: common.KT_TRANSACTION_PICKUP, Date: today}
	if err = db.Save(&u).Error; err != nil {
		glog.Error("Yunex_Charge_Callback fail! err=", err)
		reason = yf.ERR_SERVER_ERROR
		return
	}

	// 添加用户充值记录
	u = common.UserAssetDetail{UserId: w.UserId, Amount: amount, Type: txType, TransactionType: common.KT_TRANSACTION_RECHARGE, Date: today}
	if err = db.Save(&u).Error; err != nil {
		glog.Error("Yunex_Charge_Callback fail! err=", err)
		reason = yf.ERR_SERVER_ERROR
		return
	}

	// 添加到充值流水记录中
	f := common.RechargeFlow{Channel: getchannelByThirdPlat(third_plat), FlowType: common.FLOW_TYPE_YUNBAY, AssetId: u.Id, UserId: w.UserId, TxHash: txHash, FromAddress: "", Address: w.BindAddress, TxType: txType, Amount: amount, Date: today}

	// 保存充值记录
	if err = db.Save(&f).Error; err != nil {
		glog.Error("Chain_Charge fail! err=", err)
		reason = "INSUFFICIENT_FUNDS"
		return
	}
	id = f.Id
	return
}
