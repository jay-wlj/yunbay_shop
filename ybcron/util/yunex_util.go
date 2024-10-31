
package util

import (
	"strings"
	"fmt"
	"github.com/jie123108/glog"
	base "github.com/jay-wlj/gobaselib"
)


type YunexAccount struct {
	UserId int64 `json:"user_id,string"`
	Useable float64 `json:"useable,string"`
	Total float64 `json:"total,string"`	
	Freeze float64 `json:"freeze,string"`
}

type snapAccount struct {
	Snaps []YunexAccount `json:"snaps"`
	Total int64 `json:"total,string"`
	Symbol string `json:"symbol"`
}
type addrYunex struct {
	ReasonSt
	UserId int64 `json:"user_id" `
}

type YunexWithDraw struct {
	Plat string `json:"plat"`
	Coin string `json:"coin"`
	Address string `json:"address"`
	OrderId string `json:"order_id,string"`	
	Amount float64 `json:"amount,string"`	
}

type yunexAccount struct {
	UserId int64 `json:"user_id,string"`
	Address string `json:"string"`
}
type yunexUserInfo struct {
	Plat string `json:"plat"`
	Zonenum string `json:"zone_num"`
	Mobile string `json:"mobile"`
	Symbol string `json:"symbol"`
}

type CoinBalance struct {
	YBT float64 `json:"YBT"`
	KT float64 `json:"KT"`
}


// // 帐号提币接口
// func YunexWithdrawWallet(v YunexWithDraw) (order_id string, err error){
// 	uri := "/api/platform/user/deposit/"
// 	var ret sthotOrder
// 	err = post_yunex(uri, "yunex", nil, v, "", &ret, false, EXPIRE_RES_INFO)
// 	if err != nil {
// 		if ret.Reason != "" {
// 			err = fmt.Errorf(ret.Reason)			
// 		}
// 		glog.Error("YunexWithdrawWallet fail! err=", err, " args:", v)		
// 		return
// 	}
// 	order_id = ret.OrderId
// 	return
// }


// 获取云网持有ybt的用户
func SnapYunexYbtAccount(start, count int, day string) (total int64, vs []YunexAccount, err error) {
	uri := "/api/coin/bonus/snap?" + fmt.Sprintf("start=%v&count=%v&day=%v", start, count, day)
	var ret snapAccount
	if err = get_yunex(uri, "yunex", "", &ret, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("SnapYunexYbtAccount fail! err=", err)
		return
	}
	total = ret.Total
	vs = ret.Snaps
	return
}


// // 获取yunex平台帐户余额信息
// func GetYunbayAccountInYunex() (m map[string]float64, err error) {
// 	uri := "http://yunex.io/api/yunex/credit/query/"	
// 	if err = get_yunex(uri, "yunex", "", &m, false, EXPIRE_RES_INFO); err != nil {
// 		glog.Error("SnapYunexYbtAccount fail! err=", err)
// 		return
// 	}
// 	return
// }

type YunexKtBonus struct {
	ToUid int64 `json:"to_uid"`
	Symbol string `json:"symbol"`
	Amount float64 `json:"amount,string"`
	Date string `json:"date"`
	OrderId int64 `json:"order_id,string"`	
}

type YunexKtBonusRet struct {
	OrderId int64 `json:"order_id,string"`
	Reason int `json:"reason"`
}

type bonusYunex struct {
	Bonus []YunexKtBonus `json:"bonus"`
}
// kt分红转帐给云网用户
func BonusYunexKt(vs []YunexKtBonus) (fail []YunexKtBonusRet, err error) {
	uri := "/api/coin/bonus/transfer"
	v := bonusYunex{Bonus:vs}
	if err = post_yunex(uri, "yunex", nil, v, "fail", &fail, false, EXPIRE_RES_INFO); err != nil {
		glog.Error("BonusYunexKt fail! err=", err)
		return
	}
	return
}


// 查询yunex充值地址
func QueryYunexAddress(zone, tel, coin string) (addr string, err error){
	uri := "/api/pub/user/deposit/bind/"
	var ret yunexAccount

	v := yunexUserInfo{Plat:"yunbay", Zonenum:zone, Mobile:tel, Symbol:coin}
	err = RequestYunExApi(uri, "POST", "yunex", nil, v, "", &ret, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("IsYunexAddress fail! err=", err)
		return
	}

	addr = ret.Address
	return
}

// 判断是否为yunex地址接口 否:user_id=0
func IsYunexAddress(address, coin string) (user_id int64, err error){
	//uri := fmt.Sprintf("http://a.yunex.io/api/pub/address/check/?plat=yunbay&coin=%v&address=%v", coin, address)
	uri := fmt.Sprintf("/api/pub/address/check/?plat=yunbay&coin=%v&address=%v", coin, address)
	var ret addrYunex
	err = RequestYunExApi(uri, "GET", "yunex",nil, nil, "", &ret, false, EXPIRE_RES_INFO)
	if err != nil {
		if ret.Reason != "" {
			err = fmt.Errorf(ret.Reason)			
		}
		glog.Error("IsYunexAddress fail! err=", err)
		return
	}

	user_id = ret.UserId
	return
}

type YunexDepositRet struct {
	TxHash string `json:"order_id"`
	Reason string `json:"reason"`
	Status int `json:"status"`
}

// 帐号提币接口
func YunexWithdrawWallet(data YunexWithDraw) (dataStatus YunexDepositRet, err error) {
	//uri := "http://a.yunex.io/api/pub/user/deposit/"
	uri := "/api/pub/user/deposit/"
	err = RequestYunExApi(uri, "POST", "yunex", nil, data, "", &dataStatus, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("/api/pub/user/deposit/ is fail, err:", err)
	}
    return
}

// 获取yunex平台帐户余额信息
func GetYunExBalance() (m map[string]float64, err error) {
	//uri := "http://a.yunex.io/api/pub/account/balance/?plat=yunbay"
	uri := "/api/pub/account/balance/?plat=yunbay"
	dataStatus := make(map[string]string)
	err = RequestYunExApi(uri, "GET", "yunex", nil, nil, "", &dataStatus, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("/api/pub/account/balance/ is fail, err:", err)
	}
	m = make(map[string]float64)
	for k, v := range dataStatus {
		if m[strings.ToLower(k)], err = base.StringToFloat64(v); err != nil {
			return
		}
	}
    return
}