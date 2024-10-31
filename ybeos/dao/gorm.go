package dao

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/postgres"
)

type PsqlDB struct {
	*gorm.DB
}

func (v *PsqlDB)ListPage(page, page_size int)(*gorm.DB){
	if page <= 0 {
		page = 1
		glog.Error("ListPage page:", page, " <= 0")
	}
	if page_size <= 0 {
		page_size = 10
		glog.Error("ListPage page_size:", page_size, " <= 0")
	}
	return v.Limit(page_size).Offset((page-1)*page_size)	
}

func (v *PsqlDB)GetDB()(*gorm.DB){
	return v.DB
}


var m_psqlDb map[string]*gorm.DB 

func init() {
	m_psqlDb = make(map[string]*gorm.DB)
}


// 默认db连接
var m_db *gorm.DB

func InitPsqlDb(psqlUrl string, debug bool) (*gorm.DB, error) {
	if db, ok := m_psqlDb[psqlUrl]; ok {
		return db, nil
	}
	
	db, err := gorm.Open("postgres", psqlUrl)
	if err != nil {
		glog.Fatalf("open postgresql(%v) failed! err: %v", psqlUrl, err)
		panic("open postgresql fail!")		
	}
	fmt.Println("open sql ok")
	glog.Infof("open psql(%s) ok!", psqlUrl)
	
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	//db.SingularTable(true) // 如果设置为true,`User`的默认表名为`user`,使用`TableName`设置的表名不受影响
	db.LogMode(debug)

	m_psqlDb[psqlUrl] = db

	if m_db == nil {
		m_db = db
	}
	fmt.Println("open database success!")
	return db, nil
}

func NewPsqlDb()(*PsqlDB){
	return &PsqlDB{m_db}
}

func GetDefaultDb()(*PsqlDB){
	var err error
	if m_db == nil {
		panic(fmt.Sprintf("GetDefaultDb fail! err=%v", err))
	}

	return &PsqlDB{m_db}
}


func GetPsqlDb(c *gin.Context)(*PsqlDB){
	if c != nil {
		conn, exist := c.Get("sqldao")
		if exist {
			return conn.(*PsqlDB)
		}
	}
	return &PsqlDB{m_db}	// 返回默认的db	
}