package client

import (
	"bytes"
	"fmt"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"net/http"
	"net/http/httptest"
	"testing"
	"yunbay/ybasset/conf"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func performGetRequest(r http.Handler, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", path, nil)
	req.Header["X-Yf-Country"] = []string{"1"}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performPostRequest(r http.Handler, path, body string) *httptest.ResponseRecorder {
	reader := bytes.NewBufferString(body)
	req, _ := http.NewRequest("POST", path, reader)
	req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	req.Header["X-Yf-Third"] = []string{"1"}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func init() {
	conf, err := conf.LoadConfig("../../conf/config.yml")
	if err != nil {
		return
	}
	if _, err := db.InitPsqlDb(conf.Server.PSQLUrl, conf.Server.Debug); err != nil {
		panic(err.Error())
	}
	cache.InitRedis(conf.Redis)
}

// LLT充值
func TestLLT_Recharge(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51883))
	})
	router.GET("/test", LLT_Recharge)
	// RUN
	w := performGetRequest(router, "/test?id=835")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 购买商品
func TestLLT_Recharge2(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51882))
	})
	router.POST("/test", LLT_Recharge)
	// RUN
	w := performPostRequest(router, "/test?id=835", `{"order_id":"1111","user_id":51887,"amount":8980,"platform":""}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}
