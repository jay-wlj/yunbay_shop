package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"
	"yunbay/yborder/common"
	"yunbay/yborder/conf"
	"yunbay/yborder/util"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
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
	req.Header["X-Yf-Token"] = []string{"T1g0MkaNJMCCILP5aLK2cAQybjyTU0BUzl.79a"}
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
func TestCart_List(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51883))
	})
	router.GET("/test", Cart_List)
	// RUN
	w := performGetRequest(router, "/test?publish_area=0")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 购买商品
func TestPrePay(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51882))
	})
	router.POST("/test", Orders_PrePay)
	// RUN
	w := performPostRequest(router, "/test?id=835", `{"pay_type":0,"address_id":81,"amount":898.165135116513511651356,"extinfos":{"note":""},"products":[{"virtual":false,"product_id":1247,"product_sku_id":3292,"quantity":1}]}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 获取订单列表
func TestOrdersList(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51881))
	})
	router.GET("/test", Orders_List)
	// RUN
	w := performGetRequest(router, "/test?status=5")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 商家查询订单及报表导出
func TestSellerOrdersList(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51903))
		c.Set("user_type", int64(1))
	})
	router.GET("/test", Order_SellerSearchList)
	// RUN
	w := performGetRequest(router, "/test?begin_date=2019-08-01&status=4")
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

func TestOrdersSellerCount(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51887))
	})
	router.GET("/test", Orders_SellerCount)
	// RUN
	w := performGetRequest(router, "/test?status=2")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

func TestOrdersSellerCancel(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51903))
	})
	router.POST("/test", Order_SellerCancel)
	// RUN
	w := performPostRequest(router, "/test?status=2", `{"order_id":1605797}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

func t1() (err error) {
	defer func(err error) {
		if err != nil {
			fmt.Println("t1")
		}
	}(err)
	err = errors.New("sdfd")
	if err != nil {
		glog.Error("err=", err)
		return
	}
	return
}

// 测试号码是否可以充值
func TestOfPayTelCheck(t *testing.T) {

	err := util.GetOfpay().Tel_Check("15818717950", 100)
	err = t1()
	if err != nil {
		glog.Error("err=", err)
	}
}

type Work struct {
	gorm.Model
	UserName string `json:"user_name"`
	Age      int    `json:"age"`
}

func TestSample(t *testing.T) {
	v := Work{UserName: "wlj", Age: 32}
	db := db.GetDB()
	if ok := db.CreateTable(v).Error; ok == nil {
		if err := db.Create(&v).Error; err != nil {
			glog.Error("TestSample fail! err=", err)
			return
		}
	}

}

type IWriter struct {
	io.Writer
	a int
}

var done = make(chan bool)
var msg string

func aGoroutine() {
	<-done
	time.Sleep(time.Millisecond)
	msg = "hello, world"

}

// 返回生成自然数序列的管道: 2, 3, 4, ...
func GenerateNatural() chan int {
	ch := make(chan int)
	go func() {
		for i := 2; ; i++ {
			ch <- i
		}
	}()
	return ch
}

// 管道过滤器: 删除能被素数整除的数
func PrimeFilter(in <-chan int, prime int) chan int {
	out := make(chan int)
	go func() {
		for {
			if i := <-in; i%prime != 0 {
				out <- i
			}
		}
	}()
	return out
}

func worker(wg *sync.WaitGroup, cannel chan bool) {
	defer wg.Done()

	for {
		select {
		default:
			//fmt.Println("hello")
		// 正常工作
		case <-cannel:
			fmt.Println("exit")
			goto gt
			// 退出
		}
	}

gt:
}

type cReq struct {
	Name string `form:"name" binding:"required"`
	Age  int    ` binding:"required,gt=20"`
	PId  int    `json:"p_id"`
}

func c1(c *gin.Context) {
	var v cReq
	if err := c.BindJSON(&v); err != nil {
		c.JSON(200, gin.H{"err": err.Error()})
	}
	c.JSON(200, v)
}
func TestT1(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(52041))
	})
	router.POST("/test", c1)
	// RUN
	w := performPostRequest(router, "/test?name=sdf&Age=2432&create_time=1998-01-02&unixTime=1562400033", `{"name":"sdfasdf","age":34,"p_id":234}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

func TestOfPayCard(t *testing.T) {

	v, err := util.GetOfpay().CardWidthdraw(61561561516, "1711282", 3)

	if err != nil {
		glog.Error("err=", err)
	}
	fmt.Println(v)
}

func TestOrders_Card(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51887))
	})
	router.GET("/test", Orders_Card)
	// RUN
	w := performGetRequest(router, "/test?order_id=1605764")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}
