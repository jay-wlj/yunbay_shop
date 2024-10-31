package man

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"yunbay/yborder/common"
	"yunbay/yborder/conf"
	"yunbay/yborder/server/share"
	"yunbay/yborder/util"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

	jsoniter "github.com/json-iterator/go"
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

func TestTest1(t *testing.T) {
	s := `[{"category_id":462,"create_time":1564544064,"def_sku_id":3293,"descimgs":[{"height":741,"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/384506521415412479.jpg","width":380},{"height":1171,"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/767556930838305502.jpg","width":380},{"height":449,"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/603223481896683291.jpg","width":380},{"height":652,"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/356595627444232425.jpg","width":380},{"height":391,"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/255047625516641018.jpg","width":380},{"height":379,"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/389093234297012225.jpg","width":380}],"id":1248,"images":["http://file.lkyundong.com/productpics/1531300299314.jpg","http://file.lkyundong.com/productpics/1531300299616.jpg"],"info":"好看的衣服","price":"200","publish_area":1,"rebat":"0.05","skus":[{"combines":{"内存":"4G","尺寸":"L","颜色":"淡蓝色"},"create_time":1564544064,"id":3293,"img":"","price":"200","sku":[{"attr_id":3,"value_id":221},{"attr_id":4,"value_id":15},{"attr_id":37,"value_id":222}],"sold":0,"stock":100,"update_time":1564544064}],"sold":0,"status":1,"stock":240,"title":"衣服","type":0,"update_time":1564544064,"user_id":51869}]`
	vs := []common.Product{}
	err := jsoniter.Unmarshal([]byte(s), &vs)
	fmt.Println("err=", err)
	fmt.Println("vs=", vs)
}

/*
(?s) - 允许.匹配任何字符，包括换行
\( - 字面( char
(.*) - 子匹配1捕获任何零个或多个字符
\) - 字面)符号。
*/

type n int

func (a n) Print() {
	fmt.Println("n=", a)
}
func (a *n) PrintB() {
	fmt.Println("n=", *a)
}

func TestRegexp1(t *testing.T) {
	var a n
	defer a.Print()
	defer a.PrintB()
	defer func() {
		a.Print()
	}()
	defer func() {
		a.PrintB()
	}()
	a = 3
}

func TestOrders_Report(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51887))
	})
	router.GET("/test", Orders_Report)
	// RUN
	w := performGetRequest(router, "/test?begin_date=2019-08-01&end_date=2019-08-31")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
	f, err := os.Create("./1.xlsx")
	if err != nil {
		fmt.Println("err=", err)
		return
	}
	f.Write(w.Body.Bytes())
	f.Close()
}

func TestChannel(t *testing.T) {
	util.AsynGenerateEosLotterysHash(11, "4MR636KLtiuM") // 上链处理
}

func TestChannel2(t *testing.T) {
	fmt.Println(share.OfRetNotify(nil, &common.OfOrder{OrderId: 1605764, GameState: 1, Cards: []common.OfCard{common.OfCard{Cardno: "234", Cardpws: "432"}}}))

}
