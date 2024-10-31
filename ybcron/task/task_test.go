package task

import (
	"fmt"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"testing"
	"yunbay/ybcron/conf"
)

func init() {

	_, err := conf.LoadConfig("../conf/config.yml")
	if err != nil {
		return
	}

	db.InitPsqlDb(conf.Config.PsqlUrl["asset"], conf.Config.Debug) // 默认db
	db.InitPsqlDb(conf.Config.PsqlUrl["api"], conf.Config.Debug)
	cache.InitRedis(conf.Config.Redis)
}

func TestCurrencyRate(t *testing.T) {
	ms := make(map[string]interface{})

	s := ms["contact_phone"]
	if tel, ok := s.(string); ok {
		fmt.Println(tel)
	}
	Currency_Update()
}
