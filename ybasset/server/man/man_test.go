package man

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"yunbay/ybasset/conf"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

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

// 获取购物车列表
func TestCurrencyRate(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51883))
	})
	router.GET("/test", CurrencyRatio)
	// RUN
	w := performGetRequest(router, "/test?symbol=snet_kt")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 设置货币兑换汇率
func TestCurrencyRatioSet(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51883))
	})
	router.POST("/test", CurrencyRatioSet)
	// RUN
	w := performPostRequest(router, "/test?symbol=ybt_kt", `{"from":"snet","to":"kt","ratio":"2.5", "type":1}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// pay
func TestLotterysTrasnfer(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51883))
	})
	router.POST("/test", LotterysPay)
	// RUN
	w := performPostRequest(router, "/test?symbol=ybt_kt", `{"key":"lotterys_1","coin_type":3,"from":51885,"to":0,"amount":"2500","zjpassword":"","token":""}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

/*
(?s) - 允许.匹配任何字符，包括换行
\( - 字面( char
(.*) - 子匹配1捕获任何零个或多个字符
\) - 字面)符号。
*/

func TestRegexp1(t *testing.T) {
	s := "尺寸(S,M,L,XL)"
	res, _ := regexp.Compile(`(?s)\((.*)\)`)
	a := res.FindString(s)
	//strings.Trim(a, "(")
	a = strings.Trim(a, "()")
	fmt.Println(a)
}

func TestRegexp2(t *testing.T) {
	s := "<img alt=\"\" src=\"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/09/2_20180925174905.jpg\" style=\"width: 380px; height: 568px;\" /><img alt=\"\" src=\"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/09/2_20180925174905.jpg\" style=\"width: 380px; height: 568px;\" />"
	res, _ := regexp.Compile(`src=\"([^"]*)\".*?width: (\d*)px; height: (\d*)px;`)

	a := res.FindAllStringSubmatch(s, -1)

	type imgs struct {
		Img    string
		Widht  int
		Height int
	}

	vs := []imgs{}
	for _, v := range a {
		fmt.Println(v)
		p := imgs{}
		for i, j := range v {
			switch i {
			case 1:
				p.Img = j
			case 2:
				p.Widht, _ = strconv.Atoi(j)
			case 3:
				p.Height, _ = strconv.Atoi(j)
			}
		}
		vs = append(vs, p)
	}

	fmt.Println(vs)
}

// refund
func TestLotterysPay(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51883))
	})
	router.POST("/test", LotterysOrderPay)
	// RUN
	w := performPostRequest(router, "/test?symbol=ybt_kt", `{"order_id":1605860,"currency_type":3,"amount":5000,"seller_userid":0}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// refund
func TestLotterysRefund(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51883))
	})
	router.POST("/test", TransferRefund)
	// RUN
	w := performPostRequest(router, "/test?symbol=ybt_kt", `{"key":"lotterys_12"}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 设置货币兑换汇率
func TestWallet_Charge_Callback(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51883))
	})
	router.POST("/test", Wallet_Charge_Callback)
	// RUN
	w := performPostRequest(router, "/test?symbol=ybt_kt", `[{"block_time":"2019-11-24 09:46:22", "contract_address":"", "symbol":"", "tx_hash":"0xd3921383090dd4f1c0ea4c8c3d6ddcdb5ce5b35d7c9264042c0347fcd5aa7eab", "coin_address":"0x75dce4581ee5efe03bd481d826f2b795a30dac98", "from_address":"0x0f3bd8b2f28b24b3e275e39549ee0bdc33992e19", "user_id":"119", "amount":498.234}]`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}
