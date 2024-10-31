package util

import (
	"fmt"

	"github.com/jie123108/glog"
	"github.com/shopspring/decimal"
)

func YBAsset_GetRatio(to_type int) (ratios map[string]float64, err error) {
	uri := fmt.Sprintf("/v1/currency/ratios?to_type=%v", to_type)
	err = get_info(uri, "ybasset", "ratios", &ratios, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_GetRatio fail! err=", err)
		return
	}
	return
}

type YBAssetPool struct {
	OrderId      int64   `json:"order_id" gorm:"column:order_id"`
	CurrencyType int     `json:"currency_type" gorm:"column:currency_type"`
	PayerUserId  int64   `json:"payer_userid" gorm:"column:payer_userid"`
	PayAmount    float64 `json:"pay_amount" gorm:"column:pay_amount"`
	SellerUserId int64   `json:"seller_userid" gorm:"column:seller_userid"`
	SellerAmount float64 `json:"seller_amount" gorm:"column:seller_amount"`
	RebatAmount  float64 `json:"rebat_amount" gorm:"column:rebat_amount"`
	SellerKt     float64 `json:"seller_kt"`
	Status       int     `json:"status"`
	Country      int     `json:"country"`
	PublishArea  int     `json:"publish_area"`
}

func YBAsset_Pay(vs []YBAssetPool, token string) (err error) {
	uri := "/man/asset/pay"
	header := make(map[string]string)
	header["X-YF-Token"] = token
	err = post_info(uri, "ybasset", header, vs, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_Pay fail! err=", err)
		return
	}
	return
}

type TransferPaySt struct {
	Key        string          `json:"key"`
	CoinType   int             `json:"coin_type"`
	From       int64           `json:"from"`
	To         int64           `json:"to"`
	Amount     decimal.Decimal `json:"amount" binding:"required,gt=0"`
	ZJPassword string          `json:"zjpassword"`
	Token      string          `json:"token"`
}

func YBAsset_LotterysPay(v TransferPaySt) (err error) {
	uri := "/man/lotterys/transfer"
	err = post_info(uri, "ybasset", nil, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_LotterysPay fail! err=", err)
		return
	}
	return
}

func YBAsset_LotterysRefund(key string) (err error) {
	uri := "/man/assert/transfer/refund"
	type transferRefundSt struct {
		//Id  int64  `json:"id"`
		Key string `json:"key"`
	}
	v := transferRefundSt{Key: key}
	err = post_info(uri, "ybasset", nil, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_LotterysPay fail! err=", err)
		return
	}
	return
}

type LotterysPay struct {
	OrderId      int64   `json:"order_id"`
	CurrencyType int     `json:"currency_type"`
	PayAmount    float64 `json:"amount"`
	SellerUserId int64   `json:"seller_userid"`
}

func YBAsset_LotterysConfirm(v LotterysPay) (err error) {
	uri := "/man/lotterys/order/pay"
	err = post_info(uri, "ybasset", nil, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_LotterysPay fail! err=", err)
		return
	}
	return
}

type paynotokenSt struct {
	Pools  []YBAssetPool `json:"pools"`
	UserId int64         `json:"user_id"`
}

func YBAsset_PayByUserId(vs []YBAssetPool, user_id int64) (err error) {
	uri := "/man/asset/pay_by_userid"
	v := paynotokenSt{Pools: vs, UserId: user_id}
	err = post_info(uri, "ybasset", nil, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_Pay fail! err=", err)
		return
	}
	return
}

type YBAssetStatus struct {
	PublishArea int     `json:"publish_area"`
	OrderIds    []int64 `json:"order_ids"`
	Status      int     `json:"status"`
}

func YBAsset_PayStatus(v YBAssetStatus) (err error) {
	uri := "/man/asset/payset"

	err = post_info(uri, "ybasset", nil, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_Pay fail! err=", err)
		return
	}
	return
}

type payRebatSt struct {
	OrderId int64   `json:"order_id"`
	Rebat   float64 `json:"rebat"`
}

func YBAsset_UpdatePayAmount(order_id int64, rebat float64) (err error) {
	uri := "/man/asset/pay/rebat"
	v := payRebatSt{OrderId: order_id, Rebat: rebat}
	err = post_info(uri, "ybasset", nil, v, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("YBAsset_UpdatePayAmount fail! err=", err)
		return
	}
	return
}

type VoucherSt struct {
	Id         int64   `json:"id"`
	UserId     int64   `json:"user_id" binding:"required"` // 优买会帐号id
	Type       int     `json:"type"`
	Amount     float64 `json:"amount" binding:"gt=0"`
	Title      string  `json:"title"`
	UnlockTime int64   `json:"unlock_time"`
}

func Voucher_Recharge(v VoucherSt) (order_id int64, err error) {
	// uri := "/man/voucher/recharge"
	// v := voucherSt{UserId: user_id, Type: typen, Amount: amount, Title: title}
	// err = post_info(uri, "ybasset", nil, v, "order_id", &order_id, false, EXPIRE_RES_INFO)
	// if err != nil {
	// 	glog.Error("Voucher_Recharge fail! err=", err)
	// 	return
	// }
	uri := "/man/wallet/recharge/yunbay"
	err = post_info(uri, "youbuy_asset", nil, v, "id", &order_id, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("Voucher_Recharge fail! err=", err)
		return
	}
	return
}

// // 设置订单状态为已完成，资金池中的冻结金额将由平台打款给卖家
// func SetPayStatus(order_ids []int64)(err error){
// 	uri := "/man/asset/payset"
// 	v := orderSt{OrderIds:order_ids, Status:common.ASSET_POOL_FINISH}
// 	if err = post_info(uri, "ybasset", nil, &v, "", nil, false, EXPIRE_RES_INFO); err != nil {
// 		return
// 	}
// 	return
// }

type RewarSt struct {
	UserId int64           `json:"user_id" binding:"required"`
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

type activitySt struct {
	Activitys   []RewarSt `json:"activitys" binding:"required"`
	ReleaseType int       `json:"release_type"`
	FixDays     int       `json:"fixdays"`
	Reason      string    `json:"reason"`
}

const (
	YBT_REWARD_AIRDROP  = 0 // ybt空投奖励
	YBT_REWARD_ACTIVITY = 1 // ybt活动奖励
	YBT_REWARD_MING     = 2 // ybt挖矿奖励
	YBT_REWARD_PROJECT  = 3 // ybt项目方奖励
)

// 发放活动奖励ybt
func RewardYbt(vs []RewarSt) (err error) {
	uri := "/man/ybt/reward/activity"
	headers := make(map[string]string)
	headers["X-Yf-Maner"] = "system"
	d := activitySt{Activitys: vs, ReleaseType: YBT_REWARD_ACTIVITY}
	err = post_info(uri, "ybasset", headers, d, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("RewardYbt fail! err=", err)
		return
	}
	return
}
