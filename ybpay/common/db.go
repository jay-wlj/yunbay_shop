package common

import (
	"github.com/jay-wlj/gobaselib/db"

	"github.com/shopspring/decimal"

	"github.com/lib/pq"
)

type AddressSource struct {
	Id         int64  `json:"id" gorm:"primary_key:id"`
	Address    string `json:"address"`
	Channel    int    `json:"channel"`
	CreateTime int64  `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64  `json:"-" gorm:"column:update_time"`
}

func (AddressSource) TableName() string {
	return "address_source"
}

type RmbRecharge struct {
	Id         int64           `json:"id" gorm:"primary_key:id"`
	Channel    int             `json:"channel"`
	UserId     int64           `json:"user_id,string" gorm:"column:user_id"`
	OrderIds   pq.Int64Array   `json:"order_ids"`
	Subject    string          `json:"subject"`
	AssetId    int64           `json:"asset_id"`
	TxHash     string          `json:"tx_hash" gorm:"column:txhash"`
	TxType     int             `json:"tx_type" gorm:"column:tx_type"`
	Account    string          `json:"address"`
	Amount     decimal.Decimal `json:"amount"`
	Date       string          `json:"date"`
	Status     int             `json:"status"`
	Reason     string          `json:"reason"`
	OverTime   int64           `json:"over_time"`
	CreateTime int64           `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64           `json:"-" gorm:"column:update_time"`
}

func (RmbRecharge) TableName() string {
	return "rmb_recharge"
}

type RmbRefund struct {
	db.Model
	OrderId           int64           `json:"order_id"`
	TotalFee          decimal.Decimal `json:"total_fee"`
	RefundFee         decimal.Decimal `json:"refund_fee"`
	RefundAccount     string          `json:"refund_account"`
	RefundRecvAccount string          `json:"refund_recv_account"`
	Status            int             `json:"status"`
}

func (RmbRefund) TableName() string {
	return "rmb_refund"
}
