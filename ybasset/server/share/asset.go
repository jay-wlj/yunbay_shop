package share

import (
	"yunbay/ybasset/common"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
)

type amountSt struct {
	Amount float64 `json:"amount"`
}

func GetPoolAsset(db *db.PsqlDB) (seller_amount, rebat_amount float64, err error) {
	// 获取商品售出金额冻结款
	var seller amountSt
	if err = db.Model(&common.YBAssetPool{}).Where("currency_type=? and status=?", common.CURRENCY_KT, common.STATUS_INIT).Select("sum(seller_amount) as amount ").Scan(&seller).Error; err != nil {
		glog.Error("GetPoolAsset fail! err=", err)
		return
	}
	// 获取今日贡献值
	seller_amount = seller.Amount
	// 获取未发放的kt
	var rebat amountSt
	if err = db.Model(&common.YBAssetDetail{}).Where("kt_status=?", common.STATUS_INIT).Select("sum(profit) as amount").Scan(&rebat).Error; err != nil {
		glog.Error("GetPoolAsset fail! err=", err)
		return
	}
	rebat_amount = rebat.Amount
	return
}

func GetUserAllAsset(user_id int64) (vs []common.UserAssetType, err error) {
	vs = []common.UserAssetType{}
	db := db.GetDB()
	if err = db.Order("type asc").Find(&vs, "user_id=?", user_id).Error; err != nil {
		glog.Error("GetUserAllAsset fail! err=", err)
		return
	}
	return
}
