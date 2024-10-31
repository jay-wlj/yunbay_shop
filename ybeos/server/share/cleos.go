package share

import (
	"yunbay/eosio"
	"yunbay/ybeos/conf"

	"github.com/jie123108/glog"
)

var (
	pri_key        string = "5K1YRsDxXExR92ipZm1FH5RhRQqoKbJqkYXcHsvaNqzjgsj9Uxu"
	yunbay_account string = "yunbayroot21"
	to_account     string = "lilydick1234"
)

var chain eosio.Chain

func init() {
}

func GetChain() eosio.Chain {
	if chain == nil {
		chain = eosio.NewEOSChain(conf.Config.Servers["cleos"])
	}
	return chain
}

func SendTransaction(to, memo string, amount int64) (txHash string, err error) {
	eoscfg := conf.Config.EOSConf
	if to == "" {
		to = eoscfg["to"]
	}
	if 0 == amount {
		amount = 1
	}

	tx := eosio.EosSendTransaction{Key: eoscfg["pri_key"], From: eoscfg["from"], To: to, Memo: memo, Amount: amount}
	txHash, err = GetChain().SendTransaction(&tx)
	if err != nil {
		glog.Error("SendTransaction err=", err)
		return
	}
	return
}

// func RunEos() {
// 	info, e := chain.GetCpu(yunbay_account)
// 	if e != nil {
// 		fmt.Println("GetCpu err=", e)
// 		return
// 	}
// 	fmt.Println("info:", *info)
// 	amount, err := chain.GetAccountBalance(to_account)
// 	if err != nil {
// 		fmt.Println("GetAccountBalance err=", err)
// 		return
// 	}
// 	fmt.Println("account amount:", amount)

// 	//fmt.Println("tx hash=", hash)

// 	tn, err := chain.GetTransaction("783935caaac8afb21e22ffdf1866dd56472d9f139bf0abc0218542001df4aeb7")
// 	if err != nil {
// 		fmt.Println("GetTransaction err=", err)
// 		return
// 	}
// 	fmt.Println("tn = ", *tn)
// 	return
// }
