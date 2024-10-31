package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"yunbay/upload/conf"

	"github.com/gin-gonic/gin"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/stretchr/testify/assert"
	"github.com/yunge/sphinx"
)

func performGetRequest(r http.Handler, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", path, nil)
	//req.Header["X-Yf-Country"] = []string{"1"}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performPostRequest(r http.Handler, path string, body []byte) *httptest.ResponseRecorder {
	reader := bytes.NewBuffer(body)
	req, _ := http.NewRequest("POST", path, reader)
	req.Header["Content-Type"] = []string{"image/jpg"}
	req.Header["X-Yf-Third"] = []string{"1"}
	req.Header["X-Yf-Appid"] = []string{"upload"}
	req.Header["X-Yf-Rid"] = []string{"234"}
	req.Header["X-Yf-Hash"] = []string{"6d1cad06fabe7db0f2b955e98bcc78bfa84c1191"}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

var (
	sc *sphinx.Client
	//host = "/var/run/searchd.sock"
	host  = "172.17.6.140"
	index = "d_product"
	words = "手机"
)

func init() {
	_, err := conf.LoadConfig("../../conf/config.yml")
	if err != nil {
		fmt.Println("err=", err)
		return
	}
	if _, err := db.InitPsqlDb(conf.Config.Server.PSQLUrl, conf.Config.Server.Debug); err != nil {
		panic(err.Error())
	}
	// cache.InitRedis(conf.Redis)

}

// 搜索商品
func TestSearchGoods(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(61871))
	})
	router.POST("/test", UploadSimple)
	body, _ := ioutil.ReadFile("../../tools/test.jpg")
	// RUN
	w := performPostRequest(router, "/test", body)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}
