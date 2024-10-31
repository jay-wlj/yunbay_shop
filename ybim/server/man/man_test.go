package man

import (
	"bytes"
	"fmt"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"net/http"
	"net/http/httptest"
	"testing"
	"yunbay/ybgoods/conf"

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
	if _, err := db.InitPsqlDb(conf.Server.PSQLUrl, conf.Server.Debug); err != nil {
		panic(err.Error())
	}
	cache.InitRedis(conf.Redis)
}

// 推荐商品
func TestRecommendIndex(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.POST("/test", MsgSend)
	// RUN
	w := performPostRequest(router, "/test?ids=62,64,65", `{"type":0, "to":[51903], "content":"你是谁"}`)
	fmt.Println("body=", w.Body)
}

// 推荐商品
func TestRegisterIMUser(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.POST("/test", RegisterIMUser)
	// RUN
	w := performPostRequest(router, "/test?ids=62,64,65", `{"from_inviteid":0,"tel":"15171636640","user_id":52041}`)
	fmt.Println("body=", w.Body)
}
