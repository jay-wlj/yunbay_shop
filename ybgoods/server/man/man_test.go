package man

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"yunbay/ybgoods/conf"

	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

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
func TestRecommendUpsert(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.POST("/test", RecommendUpsert)
	// RUN
	w := performPostRequest(router, "/test?ids=62,64,65", `{"name":"\u6700\u65b0\u4e0a\u7ebf","img":"https:\/\/file.yunbay.com\/upload\/img\/30\/42\/30420d1a9afb2bcb60335812569af4435a59ce17.jpg","descimg":"https:\/\/file.yunbay.com\/upload\/img\/1b\/46\/1b4605b0e20ceccf91aa278d10e81fad64e24e27.jpg","product_ids":[7643],"type":0,"country":1}`)
	fmt.Println("body=", w.Body)
}

// 搜索商品
func TestGoodsList(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.GET("/test", GoodsList)
	// RUN
	w := performGetRequest(router, "/test?ids=385")

	fmt.Println("body=", w.Body)
}

// 管理后台商品详情
func TestGoodsBackendInfo(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.GET("/test", GoodsBackendInfo)
	// RUN
	w := performGetRequest(router, "/test?id=840")

	fmt.Println("body=", w.Body)
}

// 搜索商品
func TestGoodsListByIds(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.GET("/test", GoodsListByIds)
	// RUN
	w := performGetRequest(router, "/test?ids=62,64,65")

	fmt.Println("body=", w.Body)
}

// 获取商品价格
func TestGoodsPriceListByIds(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.GET("/test", GoodsPriceListByIds)
	// RUN
	w := performGetRequest(router, "/test?ids=2515,2563")

	fmt.Println("body=", w.Body)
}

// 搜索商品
func TestGoodsListDetail(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.GET("/test", GoodsListDetail)
	// RUN
	w := performGetRequest(router, "/test?list_id_model=7673-0")

	fmt.Println("body=", w.Body)
}

func TestUpsert(c *testing.T) {
	//s := `[{"id":0,"category_id":468,"user_id":118994,"title":"Z&amp,Z 运动内衣文胸背心女307","images":["http://file.lkyundong.com/productpics/1559091568942.jpg","http://file.lkyundong.com/productpics/1559091570202.jpg","http://file.lkyundong.com/productpics/1559091571465.jpg","http://file.lkyundong.com/productpics/1559091572665.jpg"],"descimgs":[{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/307_01.jpg","width":380,"height":675}],"type":0,"stock":600,"sold":0,"price":"73.8","rebat":"0.02","contact":{"contact_email":"a13338976405@163.com","contact_name":"周瀚","contact_phone":"13338976405"},"status":1,"publish_area":1,"skus":[{"id":0,"sku":null,"combines":{"尺寸":"S","颜色":"黑色"},"stock":600,"sold":0,"price":"73.8","img":"http://file.lkyundong.com/productpics/1559091568942.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"S","颜色":"粉红色"},"stock":600,"sold":0,"price":"73.8","img":"http://file.lkyundong.com/productpics/1559091568942.jpg"}]}]`
	//s := `[{"id":0,"category_id":468,"user_id":118994,"title":"优买会HKD代金券","images":["https://file.yunbay.com/upload/img/e1/11/e111c6014ea1b0923c8a2607604c2fa6936cab5b.png"],"descimgs":[{"path":"https://file.yunbay.com/upload/img/6d/5d/6d5d0565e667a521251a2227882220715e0a25a5.jpg","width":750,"height":600},{"path":"https://file.yunbay.com/upload/img/10/b2/10b24216f8f00322449f5c09c9e2a4aed8af40d8.jpg","width":750,"height":750}],"type":0,"stock":-1,"sold":0,"price":"100","rebat":"0.02","contact":{"contact_email":"a13338976405@163.com","contact_name":"周瀚","contact_phone":"13338976405"},"status":1,"publish_area":1,"skus":[{"id":0,"sku":null,"combines":{"券值":"100USD"},"stock":-1,"sold":0,"price":"100","img":"http://file.lkyundong.com/productpics/1559091568942.jpg","extinfo":{"voucher":{"type":4,"amount":100}}}]}]`
	//s := `[{"id":0,"category_id":466,"user_id":51869,"title":"运动上衣女夏季紧身短袖瑜伽服薄款性感韩版速干网红健身镂空罩衫FY059","images":["http://file.lkyundong.com/productpics/1561690369334.jpg","http://file.lkyundong.com/productpics/1561690371301.jpg","http://file.lkyundong.com/productpics/1561690373249.jpg","http://file.lkyundong.com/productpics/1561690375276.jpg","http://file.lkyundong.com/productpics/1561690377110.jpg"],"descimgs":[{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_01(1).jpg","width":380,"height":622},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_02(1).jpg","width":380,"height":621},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_03(1).jpg","width":380,"height":622},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_04(1).jpg","width":380,"height":622},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_05(1).jpg","width":380,"height":621},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_06(1).jpg","width":380,"height":622},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_07(1).jpg","width":380,"height":622},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_08(1).jpg","width":380,"height":622},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_09(1).jpg","width":380,"height":621},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_10(1).jpg","width":380,"height":622},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_11(1).jpg","width":380,"height":622},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_12(1).jpg","width":380,"height":621},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_13(1).jpg","width":380,"height":622},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_14(1).jpg","width":380,"height":622},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_15(1).jpg","width":380,"height":621},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/06/%E8%AF%A6%E6%83%85%E9%A1%B5_16(1).jpg","width":380,"height":622}],"type":0,"stock":1199,"sold":0,"cost_price":"44.04","price":"58","rebat":"0.02","contact":{"contact_email":"a13338976405@163.com","contact_name":"周瀚","contact_phone":"13338976405"},"status":1,"publish_area":1,"skus":[{"id":0,"sku":null,"combines":{"尺寸":"S","颜色":"灰色"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"S","颜色":"黑色"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"S","颜色":"白色"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"S","颜色":"玫红"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"M","颜色":"灰色"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"M","颜色":"黑色"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"M","颜色":"白色"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"M","颜色":"玫红"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"L","颜色":"灰色"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"L","颜色":"黑色"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"L","颜色":"白色"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"L","颜色":"玫红"},"stock":1199,"sold":0,"cost_price":"44.04","price":"58","img":"http://file.lkyundong.com/productpics/1561690369334.jpg"}]}]`
	s := `[{"id":0,"category_id":470,"user_id":51869,"title":"雷魅运动风衣男健身服开衫连帽外套LM1100271","images":["http://file.lkyundong.com/productpics/1558689996461.jpg","http://file.lkyundong.com/productpics/1558689998384.jpg","http://file.lkyundong.com/productpics/1558689999834.jpg","http://file.lkyundong.com/productpics/1558690001319.jpg"],"descimgs":[{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_02(1).jpg","width":380,"height":544},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_03(1).jpg","width":380,"height":539},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_04(1).jpg","width":380,"height":581},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_05(1).jpg","width":380,"height":388},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_06(1).jpg","width":380,"height":396},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_07(1).jpg","width":380,"height":497},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_08(1).jpg","width":380,"height":721},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_09(1).jpg","width":380,"height":352},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_10(1).jpg","width":380,"height":596},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_11(1).jpg","width":380,"height":596},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8A_12(1).jpg","width":380,"height":231},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8B_01(2).jpg","width":380,"height":603},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8B_02(2).jpg","width":380,"height":507},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8B_03(2).jpg","width":380,"height":502},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8B_04(3).jpg","width":380,"height":467},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8B_05(2).jpg","width":380,"height":472},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8B_06(2).jpg","width":380,"height":493},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8B_07(1).jpg","width":380,"height":628},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8B_08(1).jpg","width":380,"height":563},{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/%E4%B8%8B_09(1).jpg","width":380,"height":300}],"type":0,"stock":995,"sold":0,"cost_price":"65.55","price":"78","rebat":"0.02","contact":{"contact_email":"a13338976405@163.com","contact_name":"周瀚","contact_phone":"13338976405"},"status":1,"publish_area":1,"skus":[{"id":0,"sku":null,"combines":[{"尺寸":"M"},{"颜色":"黑蓝"}],"stock":995,"sold":0,"cost_price":"65.55","price":"78","img":"http://file.lkyundong.com/productpics/1558689996461.jpg"},{"id":0,"sku":null,"combines":[{"尺寸":"M"},{"颜色":"黑绿"}],"stock":995,"sold":0,"cost_price":"65.55","price":"78","img":"http://file.lkyundong.com/productpics/1558689996461.jpg"},{"id":0,"sku":null,"combines":[{"尺寸":"L"},{"颜色":"黑蓝"}],"stock":995,"sold":0,"cost_price":"65.55","price":"78","img":"http://file.lkyundong.com/productpics/1558689996461.jpg"},{"id":0,"sku":null,"combines":[{"尺寸":"L"},{"颜色":"黑绿"}],"stock":995,"sold":0,"cost_price":"65.55","price":"78","img":"http://file.lkyundong.com/productpics/1558689996461.jpg"},{"id":0,"sku":null,"combines":[{"尺寸":"XL"},{"颜色":"黑蓝"}],"stock":995,"sold":0,"cost_price":"65.55","price":"78","img":"http://file.lkyundong.com/productpics/1558689996461.jpg"},{"id":0,"sku":null,"combines":[{"尺寸":"XL"},{"颜色":"黑绿"}],"stock":995,"sold":0,"cost_price":"65.55","price":"78","img":"http://file.lkyundong.com/productpics/1558689996461.jpg"},{"id":0,"sku":null,"combines":[{"尺寸":"XXL"},{"颜色":"黑蓝"}],"stock":995,"sold":0,"cost_price":"65.55","price":"78","img":"http://file.lkyundong.com/productpics/1558689996461.jpg"},{"id":0,"sku":null,"combines":[{"尺寸":"XXL"},{"颜色":"黑绿"}],"stock":995,"sold":0,"cost_price":"65.55","price":"78","img":"http://file.lkyundong.com/productpics/1558689996461.jpg"},{"id":0,"sku":null,"combines":[{"尺寸":"3XL"},{"颜色":"黑蓝"}],"stock":995,"sold":0,"cost_price":"65.55","price":"78","img":"http://file.lkyundong.com/productpics/1558689996461.jpg"},{"id":0,"sku":null,"combines":[{"尺寸":"3XL"},{"颜色":"黑绿"}],"stock":995,"sold":0,"cost_price":"65.55","price":"78","img":"http://file.lkyundong.com/productpics/1558689996461.jpg"}]}]`
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(6187351))
	})
	router.POST("/test", GoodsUpsert)
	// RUN
	w := performPostRequest(router, "/test?ids=5", s)
	// TEST

	fmt.Println("body=", w.Body)
}

// 获取商品详情
func TestGoodsInfo(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(61871))
	})
	router.GET("/test", GoodsInfo)
	// RUN
	w := performGetRequest(router, "/test?id=7658&sku_id=3")
	// TEST
	//assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}

// 屏蔽商品
func TestGoodsHidOne(c *testing.T) {
	//s := `[{"id":0,"category_id":468,"user_id":118994,"title":"Z&amp,Z 运动内衣文胸背心女307","images":["http://file.lkyundong.com/productpics/1559091568942.jpg","http://file.lkyundong.com/productpics/1559091570202.jpg","http://file.lkyundong.com/productpics/1559091571465.jpg","http://file.lkyundong.com/productpics/1559091572665.jpg"],"descimgs":[{"path":"http://wechat.lkyundong.com/userfiles/2f39c36164974b4bb25cabe622d4fa05/images/product/2019/05/307_01.jpg","width":380,"height":675}],"type":0,"stock":600,"sold":0,"price":"73.8","rebat":"0.02","contact":{"contact_email":"a13338976405@163.com","contact_name":"周瀚","contact_phone":"13338976405"},"status":1,"publish_area":1,"skus":[{"id":0,"sku":null,"combines":{"尺寸":"S","颜色":"黑色"},"stock":600,"sold":0,"price":"73.8","img":"http://file.lkyundong.com/productpics/1559091568942.jpg"},{"id":0,"sku":null,"combines":{"尺寸":"S","颜色":"粉红色"},"stock":600,"sold":0,"price":"73.8","img":"http://file.lkyundong.com/productpics/1559091568942.jpg"}]}]`
	//s := `[{"id":0,"category_id":468,"user_id":118994,"title":"优买会HKD代金券","images":["https://file.yunbay.com/upload/img/e1/11/e111c6014ea1b0923c8a2607604c2fa6936cab5b.png"],"descimgs":[{"path":"https://file.yunbay.com/upload/img/6d/5d/6d5d0565e667a521251a2227882220715e0a25a5.jpg","width":750,"height":600},{"path":"https://file.yunbay.com/upload/img/10/b2/10b24216f8f00322449f5c09c9e2a4aed8af40d8.jpg","width":750,"height":750}],"type":0,"stock":-1,"sold":0,"price":"100","rebat":"0.02","contact":{"contact_email":"a13338976405@163.com","contact_name":"周瀚","contact_phone":"13338976405"},"status":1,"publish_area":1,"skus":[{"id":0,"sku":null,"combines":{"券值":"100USD"},"stock":-1,"sold":0,"price":"100","img":"http://file.lkyundong.com/productpics/1559091568942.jpg","extinfo":{"voucher":{"type":4,"amount":100}}}]}]`
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(6187351))
	})
	router.POST("/test", GoodsHidOne)
	// RUN
	w := performPostRequest(router, "/test?ids=5", `{"id":3412,"status":1,"reason":"23rwerf"}`)
	// TEST

	fmt.Println("body=", w.Body)
}

// 加减库存
func TestGoodsPlusQuantity(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.POST("/test", GoodsPlusQuantity)
	// RUN
	w := performPostRequest(router, "/test?ids=385", `[{"order_id":1605478,"product_id":1407,"product_sku_id":2009,"quantity":1}]`)

	fmt.Println("body=", w.Body)
}

func TestGoodsQuantitySet(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.POST("/test", GoodsQuantitySet)
	// RUN
	w := performPostRequest(router, "/test?ids=385", `{"product_id":7663,"sold":0, "stock":0}`)

	fmt.Println("body=", w.Body)
}

// 添加商品分类
func TestCategoryUpsert(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.POST("/test", CategoryUpsert)
	// RUN
	w := performPostRequest(router, "/test?ids=62,64,65", `{"title":"测试分类","parent_id":0, "is_show":1}`)
	fmt.Println("body=", w.Body)
}

// 删除商品分类
func TestCategoryDel(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(51890))
	})
	router.POST("/test", CategoryDel)
	// RUN
	w := performPostRequest(router, "/test?ids=62,64,65", `{"id":487}`)
	fmt.Println("body=", w.Body)
}
