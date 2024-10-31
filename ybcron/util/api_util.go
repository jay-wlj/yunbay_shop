package util

import (
	"fmt"

	base "github.com/jay-wlj/gobaselib"
	//"yunbay/ybasset/common"
)

func GetRecommenders(user_id int64) (vs []int64, err error) {
	uri := fmt.Sprintf("/man/invite/beinvite?user_id=%v", user_id)
	if err = get_info(uri, "ybapi", "recommend_userids", &vs, false, EXPIRE_RES_INFO); err != nil {
		return
	}
	return
}

func GetInvitersByIds(user_ids []int64) (mids map[int64][]int64, err error) {
	mids = make(map[int64][]int64)
	uri := fmt.Sprintf("/man/invite/beinvites?user_ids=%v", base.Int64SliceToString(user_ids, ","))
	if err = get_info(uri, "ybapi", "inviters", &mids, false, EXPIRE_RES_INFO); err != nil {
		return
	}
	return

}

// type orderSt struct {
// 	OrderIds []int64 `json:"order_ids"`
// 	Status int `json:"status"`
// }

// // 设置订单状态为已完成，资金池中的冻结金额将由平台打款给卖家
// func SetPayStatus(order_ids []int64)(err error){
// 	uri := "/man/asset/payset"
// 	v := orderSt{OrderIds:order_ids, Status:common.ASSET_POOL_FINISH}
// 	if err = post_info(uri, "ybasset", nil, &v, "", nil, false, EXPIRE_RES_INFO); err != nil {
// 		return
// 	}
// 	return
// }

type WithDrawSt struct {
	OrderId    int64   `json:"order_id,string"`
	TxHash     string  `json:"tx_hash"`
	Status     string  `json:"status"`
	Reason     string  `json:"reason"`
	FeeInEther float64 `json:"fee_in_ether,string"`
	Channel    int     `json:"channel"`
	TxType     string  `json:"tx_type"`
}

// 调用提币回调接口通知其提币订单状态的改变
func WithdrawCallback(v WithDrawSt) (err error) {
	uri := "/man/wallet/withdraw/callback"
	if err = post_info(uri, "ybasset", nil, &v, "", nil, false, EXPIRE_RES_INFO); err != nil {
		return
	}
	return
}

type checkParams struct {
	Id   int64 `json:"id"`
	Pass bool  `json:"pass"`
}

// 提币审核接口
func WithdrawCheck(id int64, pass bool) (err error) {
	v := checkParams{Id: id, Pass: pass}
	uri := "/man/wallet/draw/check"
	headers := make(map[string]string)
	headers["x-yf-maner"] = "system"
	if err = post_info(uri, "ybasset", headers, &v, "", nil, false, EXPIRE_RES_INFO); err != nil {
		return
	}
	return
}

type dateSt struct {
	Date string `json:"date"`
}

// 调用ybt及kt发放接口
func ReleaseYbt(date string) (err error) {
	uri := "/man/ybt/reward/check"
	v := dateSt{Date: date}
	headers := map[string]string{"X-Yf-Maner": "system"}
	if err = post_info(uri, "ybasset", headers, &v, "", nil, false, 0); err != nil {
		return
	}
	return
}

// 调用ybt及kt发放接口
func ReleaseKt(date string) (err error) {
	uri := "/man/kt/reward/check"
	v := dateSt{Date: date}
	headers := map[string]string{"X-Yf-Maner": "system"}
	if err = post_info(uri, "ybasset", headers, &v, "", nil, false, 0); err != nil {
		return
	}
	return
}
