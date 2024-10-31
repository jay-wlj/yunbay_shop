package share

import (
	"yunbay/ybsearch/conf"

	"github.com/jie123108/glog"
	"github.com/yunge/sphinx"
)

var (
	//sc *sx
	//host = "/var/run/searchd.sock"
	host  = "172.17.6.140"
	words = "手机"
)

type sx struct {
	*sphinx.Client
}

func GetSphinx() *sx {
	opts := sphinx.DefaultOptions
	opts.Host = conf.Config.Search.Host
	opts.SqlPort = 0
	sc := &sx{sphinx.NewClient(opts)}
	return sc
}

func (sc *sx) Search(keyword, sort, sequence string, page, page_size int) (total int, ids []uint64, err error) {
	// 设置翻页
	offset := (page - 1) * page_size
	sc.SetLimits(offset, page_size, sphinx.DefaultOptions.MaxMatches, 0)
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
	total = res.Total
	return
}

/*
	是否屏B商品
*/
func (sc *sx) Hid(id uint64, status int) (err error) {
	_, err = sc.UpdateAttributes("d_product", []string{"is_hid"}, [][]interface{}{[]interface{}{id, status}}, true)
	return
}

/*
	是否上下架商品
*/
func (sc *sx) Status(id uint64, status int) (err error) {
	_, err = sc.UpdateAttributes("d_product", []string{"status"}, [][]interface{}{[]interface{}{id, status}}, true)
	return
}
