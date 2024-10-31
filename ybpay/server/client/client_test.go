package client

import (
	"yunbay/ybpay/conf"
	"bytes"
	"fmt"
	"github.com/jay-wlj/gobaselib/db"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func performGetRequest(r http.Handler, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performPostRequest(r http.Handler, path, body string) *httptest.ResponseRecorder {
	reader := bytes.NewBufferString(body)
	req, _ := http.NewRequest("POST", path, reader)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func init() {
	conf, err := conf.LoadConfig("../../conf/config.yml")
	if err != nil {
		return
	}
	_, err = db.InitPsqlDb(conf.Server.PSQLUrl, conf.Server.Debug)
	if err != nil {
		panic(err.Error())
	}

	//cache.InitRedis(conf.Redis)
}

func TestWeixinNotify(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(6137341))
	})
	router.POST("/test", WeixinNotify)
	// RUN
	w := performPostRequest(router, "/test", "<xml>354f</xml>")
	// TEST
	assert.Equal(t, w.Code, 200)

	t.Log("body=", w.Body)
}

func TestWeixinPrePay(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51885))
	})
	router.POST("/test", WeixinPay)
	// RUN
	w := performPostRequest(router, "/test", `{"amount":10}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println(w.Body.String())
}

func TestAliPrePay(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51885))
	})
	router.POST("/test", Alipay_pay)
	// RUN
	w := performPostRequest(router, "/test", `{"order_ids":[1605412],  "amount":1}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println(w.Body.String())
}
