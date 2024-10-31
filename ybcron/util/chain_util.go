
package util

import (
	"github.com/jie123108/glog"
)


type Addr struct {
	Address string `json:"coin_address" binding:"required,eth_addr"`
}

type createaddress struct {
	UserId int64 `json:"userid,string"`
}
// 创建帐号唯一的地址
func GetUserWalletAddress(user_id int64, currency_type string) (address string, err error){
	uri := "/api/charge/address"

	v := createaddress{UserId:user_id}
	var addr Addr
	err = post_chain(uri, "chain", nil, v, "", &addr, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("GetUserWalletAddress fail! err=", err)
		return
	}
	address = addr.Address
	return
}

type WithDraw struct {
	Symbol string `json:"symbol"`
	UserId int64 `json:"user_id,string"`
	OrderId int64 `json:"order_id,string"`
	Address string `json:"address"`
	Amount float64 `json:"amount,string"`	
}
type stOrder struct {
	OrderId int64 `json:"order_id,string" binding:"required"`
}

// 帐号提币接口
func ChainWithdrawWallet(v WithDraw) (err error){
	uri := "/api/withdraw"
	var ret stOrder
	err = post_chain(uri, "chain", nil, v, "", &ret, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("GetUserWalletAddress fail! err=", err)
		return
	}
	if v.OrderId != ret.OrderId {
		glog.Error("WithdrawWallet fail! v.OrderId != ret.OrderId")
	}
	return
}

type TxStatus struct {
	Status string `json:"status" binding:"required"`
	TxHash string `json:"tx_hash" binding:"required"`
	FeeInEther float64 `json:"fee_in_ether"`
	Reason string `json:"reason"`
}

// 查询提币交易
func ChainTxQuery(order_id int64)(ret TxStatus, err error){
	uri := "/api/withdraw/query"
	v := stOrder{OrderId:order_id}
	err = post_chain(uri, "chain", nil, v, "", &ret, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("TxQuery fail! err=", err)
		return
	}

	return
}
