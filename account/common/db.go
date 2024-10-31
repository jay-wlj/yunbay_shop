package common

import (
	"github.com/jay-wlj/gobaselib/db"
)

type Account struct {
	Id         int64  `json:"user_id" gorm:"column:user_id;primary_key:true"`
	Cc         string `json:"cc"`
	Tel        string `json:"tel"`
	Password   string `json:"-"`
	Status     int    `json:"status"`
	UserType   int16  `json:"user_type"`
	Platform   string `json:"platform" view:"man"`
	Version    string `json:"version" view:"man"`
	DeviceId   string `json:"did,omitempty" view:"man"`
	Username   string `json:"username" view:"other"`
	Avatar     string `json:"avatar" view:"other"`
	ZJPassword string `json:"-" gorm:"column:zjpassword"`
	Ip         string `json:"ip" view:"man"`
	Birthday   string `json:"birthday"`
	Date       string `json:"date" view:"man"`
	CertStatus int    `json:"cert_status"`
	Country    int    `json:"country" view:"man"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
	Cert       *Cert  `json:"certinfos,omitempty" gorm:"column:-;ForeignKey:UserId"`
}

func (t *Account) TableName() string {
	return "account"
}

type Cert struct {
	Id          int64    `json:"id" gorm:"primary_key" view:"man"`
	UserId      int64    `json:"user_id" view:"man"`
	CardCountry string   `json:"card_country" `
	CardName    string   `json:"card_name" `
	CardId      string   `json:"card_id" `
	Status      int      `json:"status"`
	Reason      string   `json:"reason,omitempty"`
	Maner       string   `json:"maner" view:"man"`
	Country     int      `json:"country" view:"man"`
	CreateTime  int64    `json:"create_time" view:"man"`
	UpdateTime  int64    `json:"update_time"`
	CardImgs    db.JSONB `json:"card_imgs" gorm:"card_imgs" view:"man"`
}

func (t *Cert) TableName() string {
	return "cert"
}

type LoginRecord struct {
	Id         int64  `json:"-" gorm:"primary_key"`
	UserId     int64  `json:"user_id"`
	Ip         string `json:"ip"`
	Country    int    `json:"country"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

func (t *LoginRecord) TableName() string {
	return "login_record"
}

type ImToken struct {
	Id         int64  `json:"id" gorm:"primary_key"`
	UserId     int64  `json:"user_id"`
	Imid       string `json:"imid"`
	Token      string `json:"token"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

func (t *ImToken) TableName() string {
	return "imtoken"
}

type ThirdAccount struct {
	Id           int64    `json:"-" gorm:"primary_key"`
	UserId       int64    `json:"user_id"`
	ThirdName    string   `json:"-"`
	ThirdId      int64    `json:"third_id"`
	ThirdAccount db.JSONB `json:"third_account"`
	CreateTime   int64    `json:"create_time"`
	UpdateTime   int64    `json:"-"`
}

func (t *ThirdAccount) TableName() string {
	return "third_account"
}
