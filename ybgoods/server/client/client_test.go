package client

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"yunbay/ybgoods/conf"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func performGetRequest(r http.Handler, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", path, nil)
	//req.Header["X-Yf-Country"] = []string{"1"}
	req.Header["Content-Type"] = []string{"application/json"}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performPostRequest(r http.Handler, path, body string) *httptest.ResponseRecorder {
	reader := bytes.NewBufferString(body)
	req, _ := http.NewRequest("POST", path, reader)
	req.Header["Content-Type"] = []string{"application/json"}
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

// // 获取商品列表接口
// func TestGoodsList(t *testing.T) {
// 	router := gin.New()
// 	router.Use(func(c *gin.Context) {
// 		c.Set("user_id", int64(618371))
// 	})
// 	router.GET("/test", GoodsList)
// 	// RUN
// 	w := performGetRequest(router, "/test")
// 	// TEST
// 	assert.Equal(t, w.Code, 200)

// 	fmt.Println("body", w.Body)
// }

func TestMap(t *testing.T) {
	m := make(map[string]string)
	m["abc"] = "sf"
	m["acb"] = "sff"
	m["gdzf"] = "ssdff"
	m["gdff"] = "ssdff"
	for k, v := range m {
		fmt.Println("k=", k, " v=", v)
	}
	fmt.Println(m)

	vs := []int{1, 43, 12, 53, 12, 43, 23, 65}
	for i, v := range vs {
		if v == 12 {
			vs = append(vs[:i], vs[i+1:]...)
		}
	}
	fmt.Println(vs)
}

// 获取商品详情
func TestGoodsInfo(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(61871))
	})
	router.GET("/test", GoodsInfo)
	// RUN
	w := performGetRequest(router, "/test?id=7678")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 获取商品售后方式
func TestGoodsContact(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(61871))
	})
	router.GET("/test", GoodsContact)
	// RUN
	w := performGetRequest(router, "/test?id=65")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 添加规格属性
func TestGoodsAddSku(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.POST("/test", GoodsAddSku)
	// RUN
	//w := performPostRequest(router, "/test?id=126", `{"category_id":1,"name":"尺寸","values":["S", "M", "L", "XL", "XXL"]}`)
	w := performPostRequest(router, "/test?id=126", `{"category_id":1,"name":"颜色","values":["白色", "黄色", "绿色", "蓝色"]}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	t.Log("body=", w.Body)
}

// 发布商品
func TestGoodsUpsert(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.POST("/test", GoodsUpsert)
	// RUN
	//w := performPostRequest(router, "/test?id=126", `{"coin_type":"dt","user_id":51890,"title":"衣服","info":"好看的衣服","images":["https://file.yunbay.com/upload/img/03/fa/03fae1582fc61144457ba2723e97cfe68b6b83cc.jpg"],"descimgs":[{"height":1056,"path":"https://file.yunbay.com/upload/img/03/fa/03fae1582fc61144457ba2723e97cfe68b6b83cc.jpg","width":700}],"virtual":false,"canreturn":false,"total_quantity":0,"total_sold_quantity":1,"contact":{"contact_email":"1145882248@qq.com","contact_name":"李莹","contact_phone":"13760372004"},"status":0,"create_time":1534235677,"skus":[{"price":200,"stock":100,"sku":[{"attr_id":1,"value_id":2},{"attr_id":2,"value_id":6}]},{"price":205,"stock":20,"sku":[{"attr_id":1,"value_id":3},{"attr_id":2,"value_id":9}]}]}`)
	w := performPostRequest(router, "/test?id=126", `{"publish_area":1,"category_id":23,"user_id":51890,"title":"京东钢蹦","info":"京东钢蹦100元","images":["http://file.lkyundong.com/productpics/1531300299314.jpg","http://file.lkyundong.com/productpics/1531300299616.jpg"],"descimgs":[{"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/384506521415412479.jpg","width":380,"height":741},{"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/767556930838305502.jpg","width":380,"height":1171},{"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/603223481896683291.jpg","width":380,"height":449},{"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/356595627444232425.jpg","width":380,"height":652},{"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/255047625516641018.jpg","width":380,"height":391},{"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/389093234297012225.jpg","width":380,"height":379}],"price":100,"stock":100, "rebat":0.3,"extinfo":{"of_key":"gdgb100"}}`)
	//w := performPostRequest(router, "/test?id=126", `{"category_id":482,"cost_price":"0","descimgs":[{"height":741,"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/384506521415412479.jpg","width":380},{"height":1171,"path":"http://wechat.lkyundong.com/userfiles/9934c4c316dd459aa10f220202d9da02/images/product/2018/07/767556930838305502.jpg","width":380}],"discount":"1","images":["http://file.lkyundong.com/productpics/1531300299314.jpg","http://file.lkyundong.com/productpics/1531300299616.jpg"],"info":"京东钢蹦","price":"10","skus":[{"combines":[{"规格":"10元"}],"extinfo":{"of_key":"gdgb10"},"price":"10","sold":2,"stock":98},{"combines":[{"规格":"30元"}],"extinfo":{"of_key":"gdgb30"},"price":"30","sold":5,"stock":95},{"combines":[{"规格":"50元"}],"extinfo":{"of_key":"gdgb50"},"price":"50","sold":3,"stock":97},{"combines":[{"规格":"100元"}],"extinfo":{"of_key":"gdgb100"},"price":"100","sold":3,"stock":97},{"combines":[{"规格":"200元"}],"extinfo":{"of_key":"gdgb200"},"price":"200","sold":6,"stock":94},{"combines":[{"规格":"500元"}],"extinfo":{"of_key":"gdgb500"},"price":"500","sold":4,"stock":96}],"sold":10,"stock":600,"title":"京东钢蹦","type":2,"user_id":51890}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 下线商品
func TestGoodsOffine(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.POST("/test", GoodsOffine)
	// RUN
	w := performPostRequest(router, "/test?id=126", `{"id":126,"status":0}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	t.Log("body=", w.Body)
}

type product struct {
	Id    int64
	Price float64
}

func (product) TableName() string {
	return "product"
}
func TestLog(t *testing.T) {
	//s := "规格(多口充电版(61w 充电器)MacBook Pro  13,多口充电版(871w 充电器)MacBook Pro  15)"
	// s := "颜色(女童(粉蓝黄),男童(蓝黑绿))"
	// res, _ := regexp.Compile(`(.*?)\((.*)\)`)
	// as := res.FindStringSubmatch(s)
	// fmt.Println(as)
	// aa := strings.Split(as[2], ",")

	// fmt.Println("aa", aa[0], " aa1", aa[1])

	var v product
	db := db.GetDB()
	if err := db.Find(&v, "id=?", 2).Error; err != nil {
		glog.Error("TestLog fail! err=", err)
		return
	}
	v.Price = 234.6345345345
	if err := db.Save(&v).Error; err != nil {
		glog.Error("TestLog fail! err=", err)
		return
	}
}

// 获取人气最高的商品列表接口
func TestGoodsListHighest(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(61871))
	})
	router.GET("/test", GoodsListHighest)
	// RUN
	w := performGetRequest(router, "/test")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 获取分类商品列表
func TestGoodsByCategory(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(61871))
	})
	router.GET("/test", GoodsByCategory)
	// RUN
	w := performGetRequest(router, "/test?category_id=0")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 获取一級分类商品列表
func TestGoodsByFirstCategory(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(61871))
	})
	router.GET("/test", GoodsByFirstCategory)
	// RUN
	w := performGetRequest(router, "/test?category_id=370")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 获取分类商品列表
func TestCategoryList(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(618371))
	})
	router.GET("/test", CategoryList)
	// RUN
	w := performGetRequest(router, "/test?category_id=0")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 获取首页平台优选和最新上线推荐商品数据接口
func TestGoodsIndexRecommend(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(618371))
	})
	router.GET("/test", GoodsIndexRecommend)
	// RUN
	w := performGetRequest(router, "/test")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 获取首页平台优选和最新上线推荐商品数据接口
func TestGoodsIndexRecommendMore(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(618371))
	})
	router.GET("/test", GoodsIndexRecommendMore)
	// RUN
	w := performGetRequest(router, "/test?type=1")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// YBT购买专区接口
func TestGoodsYbt(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(618371))
	})
	router.GET("/test", GoodsYbt)
	// RUN
	w := performGetRequest(router, "/test?type=1")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 商家获取自己的商品信息
func TestGoodsSelfInfo(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51903))
	})
	router.GET("/test", GoodsSelfInfo)
	// RUN
	w := performGetRequest(router, "/test?id=7658")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 商家获取自己的商品列表
func TestGoodsSelfList(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(118994))
	})
	router.GET("/test", GoodsSelfList)
	// RUN
	w := performGetRequest(router, "/test")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 复制一个商品
func TestGoodsDuplicate(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51879))
	})
	router.POST("/test", GoodsDuplicate)
	// RUN
	w := performPostRequest(router, "/test?type=1", `{"id":3412}`)
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 获取折扣商品列表接口
func TestGoodsDiscountList(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(6171))
	})
	router.GET("/test", GoodsDiscountList)
	// RUN
	w := performGetRequest(router, "/test?id=2")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}
