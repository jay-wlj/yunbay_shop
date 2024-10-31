package util

import (
	"fmt"

	base "github.com/jay-wlj/gobaselib"

	"github.com/jie123108/glog"
)

type orderidSt struct {
	OrderId int64 `json:"order_id"`
}

// 获取某天有该状态的订单id
func GetOrderIdsByDateStatus(date string, status int) (ret []int64, err error) {
	uri := fmt.Sprintf("/man/order/status/query?date=%v&status=%v", date, status)

	if err = get_info(uri, "yborder", nil, "order_ids", &ret, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("GetOrderIdsByDateStatus fail! err=", err)
		return
	}

	return
}

type Orders struct {
	Id              int64                  `json:"id"`
	UserId          int64                  `json:"user_id"`
	ProductId       int64                  `json:"product_id"`
	ProductModeId   int64                  `json:"product_model_id"`
	Quantity        int                    `json:"quantity"`
	CurrencyType    int                    `json:"currency_type"`
	CurrencyPercent float64                `json:"currency_percent"`
	OtherAmount     float64                `json:"other_amount"`
	RebatAmount     float64                `json:"rebat_amount"`
	TotalAmount     float64                `json:"total_amount"`
	Status          int                    `json:"status"`
	AutoCancelTime  int64                  `json:"auto_cancel_time" gorm:"column:auto_cancel_time"`
	AutoFinishTime  int64                  `json:"auto_finish_time" gorm:"column:auto_finish_time"`
	Product         map[string]interface{} `json:"product"`
	PublishArea     int                    `json:"publish_area"`
	Country         int                    `json:"country"`
}

// 获取订单信息
func GetOrderByIds(order_ids []int64) (vs []Orders, err error) {
	uri := fmt.Sprintf("/man/order/list?ids=%v&page_size=%v", base.Int64SliceToString(order_ids, ","), len(order_ids))
	//mTitles = make(map[int64]string)
	vs = []Orders{}
	if err = get_info(uri, "yborder", nil, "list", &vs, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("GetDrawNoCheckAmountLimit fail! err=", err)
		return
	}

	return
}
