package client

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {

	routers = []yf.RouterInfo{

		{yf.HTTP_GET, "/category-first", true, false, CategoryFirst},               // 获取商品大分类
		{yf.HTTP_GET, "/category-list", true, false, CategoryList},                 // 商品列表
		{yf.HTTP_GET, "/list-by-category", true, false, GoodsByCategory},           // 获取分类商品列表
		{yf.HTTP_GET, "/list-by-firstcategory", true, false, GoodsByFirstCategory}, // 获取大分类下的商品列表

		{yf.HTTP_GET, "/index-recommend", true, false, GoodsIndexRecommend}, // 获取首页平台优选和最新上线推荐商品数据接口
		{yf.HTTP_GET, "/recommend", true, false, GoodsIndexRecommendMore},   // 获取首页平台优选和最新上线推荐商品数据接口

		{yf.HTTP_GET, "/ybt-arrondi", true, false, GoodsYbt}, // YBT购买专区接口

		{yf.HTTP_GET, "/discount-list", true, false, GoodsDiscountList}, // 随机折扣专区列表接口

		{yf.HTTP_GET, "/detail", true, false, GoodsInfo},              // 商品列表
		{yf.HTTP_GET, "/list-highest", true, false, GoodsListHighest}, // 获取人气最高的商品列表接口
		{yf.HTTP_GET, "/info", true, false, GoodsInfo},                // 商品详情
		{yf.HTTP_GET, "/contact", true, false, GoodsContact},          // 商品售后联系

		{yf.HTTP_POST, "/self/duplicate", true, true, GoodsDuplicate}, // 复制商品
		{yf.HTTP_POST, "/self/sku/add", true, true, GoodsAddSku},      // 添加属性
		{yf.HTTP_POST, "/self/upsert", true, true, GoodsUpsert},       // 发布商品
		//{yf.HTTP_GET, "/self/modify-status", true, true, GoodsOffine},       // 下线商品
		{yf.HTTP_POST, "/self/status/one", true, true, GoodsStatusOne}, // 上下架商品
		{yf.HTTP_POST, "/self/status", true, true, GoodsStatus},        // 批量上下架商品
		{yf.HTTP_GET, "/self/list", true, true, GoodsSelfList},         // 商家获取自己的商品列表
		{yf.HTTP_GET, "/self/info", true, true, GoodsSelfInfo},         // 商家获取自己的商品信息

	}
	return
}
