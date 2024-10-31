package dao

import (
	"fmt"
	"testing"

	"yunbay/ybgoods/conf"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
)

func init() {
	conf, err := conf.LoadConfig("../conf/config.yml")
	if err != nil {
		return
	}
	if _, err := db.InitPsqlDb(conf.Server.PSQLUrl, conf.Server.Debug); err != nil {
		panic(err.Error())
	}
	cache.InitRedis(conf.Redis)
}

func TestIndex(t *testing.T) {
	var cs []int64

	fmt.Println(len(cs))
}
