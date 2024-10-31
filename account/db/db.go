package db

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var g_db *gorm.DB

func InitDB(sqlDSN, sqlDeiver string) *gorm.DB {
	// return gorm.Open()
	return g_db
}

func GetDB() *gorm.DB {
	return g_db
}

func GetTxDB(c *gin.Context) *gorm.DB {
	return g_db
}
