package common

import (
	"github.com/jay-wlj/gobaselib/db"

	"github.com/lib/pq"
)

type Account struct {
	Id         int64  `gorm:"primary_key:user_id" json:"user_id"`
	Cc         string `gorm:"column:cc" json:"cc"`
	Tel        string `gorm:"column:tel" json:"tel"`
	Password   string `gorm:"column:password" json:"-"`
	Status     int16  `gorm:"column:status;null" json:"-"`
	UserType   int16  `gorm:"column:user_type" json:"user_type"`
	Platform   string `gorm:"column:platform" json:"platform"`
	Version    string `gorm:"column:version"  json:"version"`
	DeviceId   string `gorm:"column:device_id" json:"-"`
	Username   string `gorm:"column:username" json:"username"`
	Avatar     string `gorm:"column:avatar" json:"avatar"`
	Birthday   string `gorm:"column:birthday" json:"-"`
	ZJPassword string `gorm:"column:zjpassword" json:"-"`
	CertStatus int    `gorm:"column:cert_status" json:"cert_status"`
	Ip         string `gorm:"column:ip" json:"-"`
	Date       string `gorm:"column:date" json:"date"`
	CreateTime int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime int64  `orm:"column:update_time" json:"-"`
}

func (t *Account) TableName() string {
	return "account"
}

type IMToken struct {
	Id         int64  `json:"id" gorm:"primary_key:id"`
	UserId     int64  `json:"user_id" gorm:"column:user_id"`
	ImId       string `json:"accid" gorm:"column:imid"`
	Token      string `json:"im_token" gorm:"column:token"`
	CreateTime int64  `json:"create_time" gorm:"column:create_time"`
	UpdateTime int64  `json:"-" gorm:"column:update_time"`
}

func (IMToken) TableName() string {
	return "imtoken"
}

type IMgs struct {
	db.Model
	Type   int           `json:"type"`
	Msg    db.JSONB      `json:"msg"`
	Uids   pq.Int64Array `json:"uids"`
	OkUids pq.Int64Array `json:"ok_uids"`

	Status int `json:"status"`
}

func (IMgs) TableName() string {
	return "im_msgs"
}
