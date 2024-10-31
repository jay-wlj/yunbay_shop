package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"yunbay/ybsearch/conf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
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

func performPostRequest(r http.Handler, path, body string) *httptest.ResponseRecorder {
	reader := bytes.NewBufferString(body)
	req, _ := http.NewRequest("POST", path, reader)
	req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	req.Header["X-Yf-Third"] = []string{"1"}
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
		return
	}
	// if _, err := db.InitPsqlDb(conf.Server.PSQLUrl, conf.Server.Debug); err != nil {
	// 	panic(err.Error())
	// }
	// cache.InitRedis(conf.Redis)

	opts := sphinx.DefaultOptions
	opts.Host = host
	opts.SqlPort = 0
	// opts := &sphinx.Options{
	// 	Host:       host,
	// 	Port:       9312,
	// 	Timeout:    5000,
	// 	MaxMatches: 1,
	// }

	sc = sphinx.NewClient(opts)
	if err := sc.Error(); err != nil {
		println("Init sphinx client > %v\n", err)
	}

	if err := sc.Open(); err != nil {
		println("Init sphinx client > %v\n", err)
	}

	status, err := sc.Status()
	if err != nil {
		println("Error: %s\n", err)
	}

	for _, row := range status {
		println("%20s:\t%s\n", row[0], row[1])
	}
}

func search(keyword, sort, sequence string, page, page_size int) (ids []uint64, err error) {
	// 设置翻页
	offset := (page - 1) * page_size
	sc.SetLimits(offset, page_size, 1000, 0)
	sc.SetRankingMode(sphinx.SPH_RANK_FIELDMASK) // 设置等级模式，这里为：设置评分模式
	// 设置排序
	sc.SetSortMode(sphinx.SPH_SORT_ATTR_DESC, "sold")
	// 设置排序
	var sortBy string
	switch sort {
	case "total_sold_quantity":
		sortBy = "sold"
	case "rebat":
		sortBy = "rebat"
	case "sale_price":
		sortBy = "price"
	default:
		// 综合排序（匹配权重进行排序）
		mWights := make(map[string]int)
		mWights["title"] = 5
		mWights["type_name"] = 3
		sc.SetFieldWeights(mWights) // 设置字段的权重，如果title命中，那么权重*5
		// $this->_client->SetSortMode('SPH_SORT_RELEVANCE', '@weight'); // 按照权重排序。 特别注意，这里第一个参数传入的是字符串（很奇怪，其实不奇怪，第一个参数为空就可以了）
		sc.SetSortMode(sphinx.SPH_SORT_RELEVANCE, "@weight DESC, sold DESC")
	}
	if sortBy != "" {
		if sequence == "desc" {
			sc.SetSortMode(sphinx.SPH_SORT_ATTR_DESC, sortBy) // 排序，按照销量的倒序排。这里的第一个参数是整形
		} else {
			sc.SetSortMode(sphinx.SPH_SORT_ATTR_ASC, sortBy)
		}
	}

	/*SPH_SORT_RELEVANCE     : 按相关度降序排列（最好的匹配排在最前面）
	         SPH_SORT_ATTR_DESC     : 按属性降序排列 （属性值越大的越是排在前面）
	         SPH_SORT_ATTR_ASC      : 按属性升序排列（属性值越小的越是排在前面）
	         SPH_SORT_TIME_SEGMENTS : 先按时间段（最近一小时/天/周/月）降序，再按相关度降序
	         SPH_SORT_EXTENDED      : 按一种类似SQL的方式将列组合起来，升序或降序排列。
	         可以指定一个类似SQL的排序表达式，但涉及的属性（包括内部属性）不能超过5个，如： @relevance DESC, price ASC, @id DESC
	         使用@开头的为内部属性，用户属性按原样使用就行。
	         内置属性有：
	         @id (匹配文档的 ID)
	         @weight (匹配权值) 【@rank 和 @relevance 只是 @weight 的别名】 匹配权重
	         @rank (等同 weight)
	         @relevance (等同 weight)
			 @random (随机顺序返回结果)
	*/

	//sc.AddQuery(words, index, "")
	//sc.AddQuery("出来", "d_mall", "")
	// 设置过滤条件
	sc.SetFilter("status", []uint64{1}, false) // 已上架
	//sc.SetFilter('is_hid', array(self::IS_HID_NO)); // 未屏蔽
	res, e := sc.Query(keyword, "d_product", "")
	if err = e; err != nil {
		glog.Error("search fail! err=", err)
		return
	}
	ids = []uint64{}
	for _, v := range res.Matches {
		ids = append(ids, v.DocId)
	}
	return
}

func TestQuery(t *testing.T) {
	ids, err := search("手机", "", "", 1, 10)
	if err != nil {
		glog.Error("TestParallelQuery fail! err=", err)
	}
	fmt.Println(ids)
}

func TestParallelQuery(t *testing.T) {
	fmt.Println("Running parallel Query() test...")
	f := func(i int) {

		//sc.SetFilterFloatRange("rate", 0.0, 1.0, false)
		//sc.SetFilter("status", []uint64{1}, false)
		//sc.SetGroupBy("rate", 0, "@group asc")

		sc.SetRankingMode(sphinx.SPH_RANK_FIELDMASK) // 设置等级模式，这里为：设置评分模式
		// 设置排序
		sc.SetSortMode(sphinx.SPH_SORT_ATTR_DESC, "sold")

		sc.AddQuery(words, index, "")
		//sc.AddQuery("出来", "d_mall", "")
		// 设置过滤条件
		sc.SetFilter("status", []uint64{1}, false)

		res, err := sc.RunQueries()
		if err != nil {
			t.Fatalf("Parallel %d > %s\n", i, err)
		}

		// if res.Total != 4 || res.TotalFound != 4 {
		// 	t.Fatalf("Parallel %d > res.Total: %d\tres.TotalFound: %d\n", i, res.Total, res.TotalFound)
		// }

		for _, r := range res {
			for _, v := range r.Matches {
				body, _ := json.Marshal(v.AttrValues)
				fmt.Println("body=", string(body))
			}
		}

		if sc.GetLastWarning() != "" {
			fmt.Printf("Parallel %d warning: %s\n", i, sc.GetLastWarning())
		}

		if err := sc.Close(); err != nil {
			t.Fatalf("Parallel %d > %s\n", i, err)
		}
	}

	//Please use fork mode for "workers" setting of searchd in sphinx.conf, there are some concurrent issues in prefork mode now.
	for i := 1; i <= 30; i++ {
		//go f(i)
		f(i)
		if i%10 == 0 {
			fmt.Printf("Already start %d goroutines...\n", i)
		}
	}
}

// 搜索商品
func TestSearchGoods(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(61871))
	})
	router.GET("/test", Index)
	// RUN
	w := performGetRequest(router, "/test?keyword=,&page=1&page_size=12")
	// TEST
	assert.Equal(t, w.Code, 200)

	fmt.Println("body=", w.Body)
}
