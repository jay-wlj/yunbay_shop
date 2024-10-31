package common

import (
	"github.com/jay-wlj/gobaselib/db"
	"github.com/shopspring/decimal"

	"github.com/jie123108/glog"
	jsoniter "github.com/json-iterator/go"
)

type IdSt struct {
	Id int64 `json:"id" binding:"gt=0"`
}

type ManInfos struct {
	Type     int             `json:"type,omitempty"`
	Amount   decimal.Decimal `json:"amount,omitempty"`
	OfKey    string          `json:"of_key"`
	OfAmount decimal.Decimal `json:"of_amount,omitempty"`
	TxId     string          `json:"tx_id,omitempty"`
	Voucher  *VoucherSt      `json:"voucher,omitempty"`
}

type VoucherSt struct {
	Type    int     `json:"type,omitempty"`
	Amount  float64 `json:"amount,omitempty"`
	TxId    *int64  `json:"tx_id,omitempty"`
	ThirdId int64   `json:"third_id"`
}

func (t *ManInfos) Parse(m *ProductSku) {
	t.Amount = m.Price
	if _, ok := m.Extinfo["voucher"]; ok {
		bExt, err := jsoniter.Marshal(m.Extinfo)
		if err != nil {
			glog.Error("manInfos parse fail! extinfos=", m.Extinfo)
			return
		}
		jsoniter.Unmarshal(bExt, t)
	}
}

func (t *ManInfos) ParseJsonb(m db.JSONB) {
	bExt, err := jsoniter.Marshal(m)
	if err != nil {
		return
	}
	jsoniter.Unmarshal(bExt, t)
}

type TelRecharge struct {
	Tel    string `json:"tel"`
	Amount int    `json:"amount"`
}

func (t *TelRecharge) ParseJsonb(m db.JSONB) {
	bExt, err := jsoniter.Marshal(m)
	if err != nil {
		return
	}
	jsoniter.Unmarshal(bExt, t)
}
