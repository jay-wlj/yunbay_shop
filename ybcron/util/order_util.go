package util

import (
	"fmt"

	base "github.com/jay-wlj/gobaselib"
	//"yunbay/ybasset/common"
)

type orderSt struct {
	OrderIds      []int64 `json:"order_ids"`
	UserIds       []int64 `json:"user_ids"`
	SellerUserIds []int64 `json:"seller_user_ids"`
}

// 超时取消订单
func CancelOrders(order_ids, user_ids, seller_user_ids []int64) (err error) {
	uri := "/man/order/cancel"
	v := orderSt{OrderIds: order_ids, UserIds: user_ids, SellerUserIds: seller_user_ids}
	headers := map[string]string{"X-Yf-Maner": "system"}
	if err = post_info(uri, "yborder", headers, &v, "", nil, false, EXPIRE_RES_INFO); err != nil {
		return
	}
	return
}

// 自动确认收货订单
func FinishOrders(order_ids, user_ids, seller_user_ids []int64) (err error) {
	uri := "/man/order/finish"
	v := orderSt{OrderIds: order_ids, UserIds: user_ids, SellerUserIds: seller_user_ids}
	headers := map[string]string{"X-Yf-Maner": "system"}
	if err = post_info(uri, "yborder", headers, &v, "", nil, false, EXPIRE_RES_INFO); err != nil {
		return
	}
	return
}

// 查询欧飞订单
func QueryOfOrders(order_ids []int64) (err error) {
	uri := fmt.Sprintf("/man/order/of/query?ids=%v", base.Int64SliceToString(order_ids, ","))
	if err = get_info(uri, "yborder", "", nil, false, EXPIRE_RES_INFO); err != nil {
		return
	}
	return
}
