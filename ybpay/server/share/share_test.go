package share

import (
	"yunbay/ybpay/conf"
	"bytes"
	"encoding/xml"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"net/http"
	"net/http/httptest"

	//"strings"

	"testing"

	"github.com/stretchr/testify/assert"
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
	//cache.InitRedis(conf.Redis)
}

func TestMapToXml(t *testing.T) {
	params := make(yf.StringMap)
	params["Name"] = "jay d"
	params["age"] = "35"
	buf, _ := xml.Marshal(params)
	t.Log("xml=", string(buf))
}
func TestXmlToMap(t *testing.T) {

	s := `
	<xml>
	<return_code>SUCCESS</return_code>
	<return_msg><![CDATA[ok]]></return_msg>
	<sandbox_signkey><![CDATA[e173ae784f9d645d585e62569c8d9d58]]></sandbox_signkey>
  	</xml>`
	params := make(yf.StringMap)
	//decoder := xml.NewDecoder(strings.NewReader(s))
	err := xml.Unmarshal([]byte(s), &params)
	assert.Equal(t, err, nil)

	t.Log("params=", params)
}

// 获取sandboxkey
func TestGetSandBoxKey(t *testing.T) {

	params, err := GetWeixin().GetSandboxKey()
	assert.Equal(t, err, nil)

	t.Log("params=", params)
}

// 创建预支付接口
func TestWeixinPrePay(t *testing.T) {
	// router := gin.New()
	// router.Use(func(c *gin.Context) {
	// 	c.Set("user_id", int64(61871))
	// })
	// router.GET("/test", Recommend_Business)
	// // RUN
	// w := performGetRequest(router, "/test?page_size=5")
	// TEST
	req := &OrderSt{Id: 123453, Amount: 2000, RemoteIp: "202.104.136.37", Subject: "云贝商城支付"}
	pre_id, reason, err := GetWeixin().TradeAppPay(req)
	assert.Equal(t, err, nil)

	t.Log("pre_id=", pre_id, " reason=", reason)
}

// 关闭订单接口
func TestWeixinClose(t *testing.T) {
	// router := gin.New()
	// router.Use(func(c *gin.Context) {
	// 	c.Set("user_id", int64(61871))
	// })
	// router.GET("/test", Recommend_Business)
	// // RUN
	// w := performGetRequest(router, "/test?page_size=5")
	// TEST
	//req := &OrderSt{Id: 1234, Amount: 200, RemoteIp: "202.104.136.37", Subject: "云贝商城支付"}
	pre_id, ok, err := GetWeixin().CloseOrder(1234)
	assert.Equal(t, err, nil)

	t.Log("pre_id=", pre_id, " reason=", ok)
}
