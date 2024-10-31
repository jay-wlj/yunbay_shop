package util

import (
	"fmt"
	"github.com/jie123108/glog"
)

type BankSt struct {
	BankName string `json:"bank_name"`
	BankType int `json:"bank_type"`
}
func QueryBankName(card_id string) (v BankSt, err error) {
	uri := fmt.Sprintf("/v1/bank/query?card_id=%v", card_id)
	if err = get_info(uri, "ybpay", nil, "", &v, false, EXPIRE_RES_INFO); err != nil {	
		glog.Error("QueryBankName fail! err=", err)
		return
	}
	return
}