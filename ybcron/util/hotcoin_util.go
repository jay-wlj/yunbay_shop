	
package util

import (
	"net/url"
	"fmt"
	"github.com/jie123108/glog"
)


// // 创建帐号唯一的地址
// func GetUserWalletAddress(user_id int64, currency_type string) (address string, err error){
// 	uri := "/api/platform/user/deposit/"

// 	v := createaddress{UserId:user_id}
// 	var addr Addr
// 	err = post_chain(uri, "chain", nil, v, "", &addr, false, EXPIRE_RES_INFO)
// 	if err != nil {
// 		glog.Error("GetUserWalletAddress fail! err=", err)
// 		return
// 	}
// 	address = addr.Address
// 	return
// }

type ReasonSt struct {
	Reason string `json:"reason"`
}
type HotCoinWithDraw struct {
	Coin string `json:"coin"`
	UserId int64 `json:"user_id,omitempty"`
	Address string `json:"address"`
	OrderId string `json:"order_id,string"`	
	Amount float64 `json:"amount"`	
	Platform string `json:"platform"`
}

type sthotOrder struct {
	ReasonSt
	OrderId string `json:"order_id"`
	UserId int64 `json:"user_id"`
}

type addrHotSt struct {
	ReasonSt
	UserId int64 `json:"user_id"`
	Address string `json:"address"`
	Phone string `json:"phone"`
}
// 判断是否为热币地址接口
func IsHotCoinAddress(address string) (user_id int64, err error){
	uri := fmt.Sprintf("/api/user/address/query?coin=KT&address=%v&platform=YunBay", address)
	var ret addrHotSt
	err = get_hotcoin(uri, "hotcoin", "", &ret, false, EXPIRE_RES_INFO)
	if err != nil {
		if ret.Reason != "" {
			err = fmt.Errorf(ret.Reason)			
		}
		glog.Error("IsHotCoinAddress fail! err=", err)
		return
	}

	user_id = ret.UserId
	return
}

// 查询热币充值地址
func QueryHotCoinAddress(tel string) (address string, err error){
	uri := fmt.Sprintf("/api/user/deposit/address?coin=KT&phone=%v&platform=YunBay", url.QueryEscape(tel))
	
	var ret addrHotSt
	err = get_hotcoin(uri, "hotcoin", "", &ret, false, EXPIRE_RES_INFO)
	if err != nil {
		if ret.Reason != "" {
			err = fmt.Errorf(ret.Reason)			
		}
		glog.Error("QueryHotCoinAddress fail! err=", err)
		return
	}

	address = ret.Address
	return
}

// 帐号提币接口
func HotCoinWithdrawWallet(v HotCoinWithDraw) (order_id string, err error){
	uri := "/api/platform/user/deposit/"
	var ret sthotOrder
	err = post_hotcoin(uri, "hotcoin", nil, v, "", &ret, false, EXPIRE_RES_INFO)

	if err != nil {
		if ret.Reason != "" {
			err = fmt.Errorf(ret.Reason)			
		}
		glog.Error("HotCoinWithdrawWallet fail! err=", err, " args:", v)
		return
	}

	order_id = ret.OrderId
	return
}

type balanceSt struct {
	ReasonSt
	BTC float64 `json:"BTC"`
	ETH float64 `json:"ETH"`
	KT float64 `json:"KT"`
	CHT float64 `json:"CHT"`	
}
// 查询平台热币帐号余额
func HotCoinBalance() (kt_amount float64, err error){
	uri := "/api/platform/user/balance/query?platform=YunBay"
	var ret balanceSt
	err = get_hotcoin(uri, "hotcoin", "", &ret, false, EXPIRE_RES_INFO)
	if err != nil {
		if ret.Reason != "" {
			err = fmt.Errorf(ret.Reason)			
		}
		glog.Error("HotCoinBalance fail! err=", err)
		return
	}

	kt_amount = ret.KT
	return
}