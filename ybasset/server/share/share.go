package share

import (
	"fmt"
	"github.com/jay-wlj/gobaselib/db"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"

	"github.com/jie123108/glog"
	//base "github.com/jay-wlj/gobaselib"
)

// 发放ybt邀请奖励(从平台帐户中划扣) (type_:[2,3,4,5])
func DeliverYbtReward(db *db.PsqlDB, type_ int, user_id int64, amount float64) (err error) {
	today := time.Now().Format("2006-01-02")

	if type_ < common.YBT_TRANSACTION_CONSUME && type_ > common.YBT_TRANSACTION_ACTIVITY {
		err = fmt.Errorf("DeliverYbtReward type is error! type:%v", type_)
		return
	}
	v := common.UserAssetDetail{UserId: 0, Type: common.CURRENCY_YBT, TransactionType: type_, Amount: -amount, Date: today}
	vs := []common.UserAssetDetail{v}
	u := common.UserAssetDetail{UserId: user_id, Type: common.CURRENCY_YBT, TransactionType: type_, Amount: amount, Date: today}

	vs = append(vs, u)
	for _, v := range vs {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("DeliverInviteReward fail! err=", err)
			return
		}
	}
	return
}

// 获取提币手续费用
func CalcDrawFee(pChannel *int, txType int, address string, amount float64) (fee, min float64, channel int, err error) {
	// str_type := common.GetCurrencyName(txType)
	var fees []conf.Fee
	fees, channel, err = GetFeeByAddress(0, address, pChannel)
	if err != nil {
		glog.Error("CalcDrawFee fail! err=", err)
		return
	}
	if pChannel != nil {
		if *pChannel >= common.CHANNEL_ALIPAY && *pChannel <= common.CHANNEL_BANK {
			txType = common.CURRENCY_RMB // 提现rmb类型
		}
	}
	//glog.Info("CalcDrawFee address:", address, " fees:", fees)
	for _, v := range fees {
		if v.Type == txType {
			switch v.Feetype {
			case 0:
				fee = v.Val
				min = v.Min
			case 1:
				fee = amount * v.Val
				min = v.Min
			default:
				break
			}
			break
		}
	}

	return
}

// 货币提取是否可用
func CanWidthDraw(txType int) (enable bool) {
	// str_type := common.GetCurrencyName(txType)
	enable = false
	for _, v := range conf.Config.Switch {
		if v.Type == txType {
			enable = v.Withdraw
		}
	}

	return
}
