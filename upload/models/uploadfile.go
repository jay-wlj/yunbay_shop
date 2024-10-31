package models

import (
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jie123108/glog"
)

/**************	table struct 			*******/
type Uploadfile struct {
	db.Model
	Rid      string   `orm:"column(rid)" json:"rid"`
	AppId    string   `orm:"column(appid)" json:"appid" gorm:"column:appid"`
	Hash     string   `orm:"column(hash);null" json:"hash"`
	Size     int      `orm:"column(size);null" json:"size"`
	Path     string   `orm:"column(path)" json:"path"`
	Width    int      `orm:"column(width);null"  json:"width"`
	Height   int      `orm:"column(height);null"  json:"height"`
	Duration int      `orm:"column(duration);null"  json:"duration"`
	Extinfo  db.Jsonb `orm:"column(extinfo);null" json:"extinfo"`
}

func (t *Uploadfile) TableName() string {
	return "uploadfile"
}

// GetAccountById retrieves Account by Id. Returns error if
// Id doesn't exist

func GetUploadfileByRid(rid string) (v *Uploadfile, err error) {
	db := db.GetDB()
	var m Uploadfile
	if err = db.First(&m, "rid=?", rid).Error; err != nil {
		glog.Error("GetUploadfileByRid fail! err=", err)
		return
	}

	v = &m
	return
}
