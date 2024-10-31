package share

import (
	"yunbay/ybpay/common"
	"yunbay/ybpay/conf"
	"github.com/jay-wlj/gobaselib/db"
	"io/ioutil"

	//base "github.com/jay-wlj/gobaselib"
	"encoding/csv"
	"strings"

	"github.com/jie123108/glog"
)

type BankSt struct {
	BankName string `json:"bank_name"`
	CardType int    `json:"card_type"`
	CardId   string `json:"card_id"`
	CardIcon string `json:"card_icon"`
}

var mBid map[string]BankSt

func init() {
	mBid = make(map[string]BankSt)
}

func InitBankId(cfg string) (err error) {
	var bt []byte
	if bt, err = ioutil.ReadFile(cfg); err != nil {
		glog.Error("InitBankId fail! err=", err)
		return
	}
	cv := csv.NewReader(strings.NewReader(string(bt)))
	var ss [][]string
	if ss, err = cv.ReadAll(); err != nil {
		glog.Error("InitBankId fail! err=", err)
		return
	}
	bank_cfg := conf.Config.BankCfg

	for i := 1; i < len(ss); i++ {
		s := ss[i]
		if len(s) >= 4 {
			card_type := 0 // 储蓄卡
			if strings.Index(s[1], "信用卡") >= 0 {
				card_type = 1 // 信用卡
			}

			if v, ok := bank_cfg.Banks[s[0]]; ok {
				b := BankSt{BankName: s[0], CardType: card_type, CardIcon: v}
				if b.CardIcon == "" { // 使用默认图标
					b.CardIcon = bank_cfg.DefaultIcon
				}
				mBid[s[4]] = b
			}
		}
	}
	return
}

// 查询银行卡
func QueryBank(card_id string) (v BankSt, ok bool) {
	v, ok = mBid[card_id]
	return
}

// 交易id查询
func TradeQuery(id int64) (r common.RmbRecharge, err error) {
	db := db.GetDB()
	if err = db.Find(&r, "id=?", id).Error; err != nil {
		glog.Error("TradeQuery fail! err=", err)
		return
	}
	return
}
