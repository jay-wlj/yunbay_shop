package common

import (
	"github.com/jay-wlj/gobaselib/db"
	"github.com/shopspring/decimal"

	"github.com/jie123108/glog"
	jsoniter "github.com/json-iterator/go"
)

type IdReq struct {
	Id int64 `json:"id" form:"id" binding:"gt=0"`
}
type IdSkuReq struct {
	IdReq
	SkuId int64 `form:"sku_id" binding:"gte=0"` // sku_id>=0
}

type PageReq struct {
	Page     int `form:"page,default=1" binding:"gt=0"`
	PageSize int `form:"page_size,default=10" binding:"gt=0,lte=100"`
}

type IdOkSt struct {
	Id     int64  `json:"id" binding:"gt=0"`
	Status int    `json:"status" binding:"oneof=-1 0 1`
	Reason string `json:"reason,omitempty"`
}

type IdsOkSt struct {
	Ids    []int64 `json:"ids"`
	Status int     `json:"status" binding:"oneof=-1 0 1`
	Reason string  `json:"reason,omitempty"`
}

type ManInfos struct {
	//CostPrice float64    `json:"cost_price"`
	Voucher *VoucherSt `json:"voucher,omitempty"`
}

type VoucherSt struct {
	Type    int             `json:"type,omitempty"`
	Amount  decimal.Decimal `json:"amount,omitempty"`
	TxId    *int64          `json:"tx_id,omitempty"`
	ThirdId int64           `json:"third_id"`
}

func (t *ManInfos) Parse(m *ProductSku) bool {
	//t.CostPrice = m.CostPrice
	if _, ok := m.Extinfo["voucher"]; ok {
		bExt, err := jsoniter.Marshal(m.Extinfo)
		if err != nil {
			glog.Error("manInfos parse fail! extinfos=", m.Extinfo)
			return false
		}
		jsoniter.Unmarshal(bExt, t)
		return true
	}
	return false
}

func (t *ManInfos) ParseJsonb(m db.JSONB) {
	bExt, err := jsoniter.Marshal(m)
	if err != nil {
		return
	}
	jsoniter.Unmarshal(bExt, t)
}
