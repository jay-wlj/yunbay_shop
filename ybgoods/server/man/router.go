package man

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {
	//评论相关操作
	routers = []yf.RouterInfo{
		{yf.HTTP_POST, "/category/upsert", true, false, CategoryUpsert}, // 发布商品分类
		{yf.HTTP_POST, "/category/del", true, false, CategoryDel},       // 删除商品分类

		{yf.HTTP_GET, "/list", true, false, GoodsList},                // 商品列表
		{yf.HTTP_GET, "/info", true, false, GoodsInfo},                // 商品详情
		{yf.HTTP_GET, "/backend/info", true, false, GoodsBackendInfo}, // 管理后台商品详情
		{yf.HTTP_GET, "/contact", true, false, GoodsContact},          // 商品售后联系

		{yf.HTTP_GET, "/list-detail", true, false, GoodsListDetail}, // 批量获取指定规格的商品列表

		{yf.HTTP_GET, "/list_by_ids", true, false, GoodsListByIds},            // 搜索商品(搜索服务调用)
		{yf.HTTP_POST, "/recommend/upsert", true, false, RecommendUpsert},     // 推荐最新最精商品
		{yf.HTTP_GET, "/recommend", true, false, RecommendList},               // 最新最精商品
		{yf.HTTP_POST, "/upsert", true, false, GoodsUpsert},                   // 批量添加商品
		{yf.HTTP_POST, "/quantity/add", true, false, GoodsPlusQuantity},       // 商品库存添加或修改
		{yf.HTTP_POST, "/quantity/set", true, false, GoodsQuantitySet},        // 设置商品库存
		{yf.HTTP_GET, "/price/list_by_ids", true, false, GoodsPriceListByIds}, // 获取商品价格(搜索服务调用)

		{yf.HTTP_POST, "/hid/one", true, false, GoodsHidOne},          // 上下架商品
		{yf.HTTP_POST, "/hid", true, false, GoodsHid},                 // 上下架商品
		{yf.HTTP_POST, "/cache/reload", true, false, GoodsRedisReset}, // 上下架商品

		//{yf.HTTP_POST, "/check_status", true, false, GoodsCheckStatus}, // 审核商品

	}
	return
}
