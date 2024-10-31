package client

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {

	routers = []yf.RouterInfo{

		{yf.HTTP_GET, "/conf", true, false, Pub_GetConf}, // app端配置信息接口

		{yf.HTTP_GET, "/notice/recommend", true, false, Notice_Recommend}, // 平台公告相关接口
		{yf.HTTP_GET, "/notice/list", true, false, Notice_List},           // 平台公告相关接口
		{yf.HTTP_GET, "/notice/info", true, false, Notice_Info},           // 平台公告相关接口

		{yf.HTTP_GET, "/upgrade/check", true, false, Upgrade_Check}, // 升级检测接口
		{yf.HTTP_GET, "/app/download", true, false, App_Download},   // 获取app端最新的安装包

		{yf.HTTP_POST, "/feedback/add", true, false, Feedback_Add}, // 平台反馈

		// 商家认证信息接口
		{yf.HTTP_POST, "/business/upsert", true, true, Business_Upsert},     // 商家资料提交
		{yf.HTTP_GET, "/business/info", true, true, Business_Info},          // 商家详情
		{yf.HTTP_GET, "/business/ratio/set", true, true, Business_RatioSet}, //商家汇率设置(已弃用)

		// 邀请列表查询
		{yf.HTTP_GET, "/invite/list", true, true, Invite_List},
		{yf.HTTP_GET, "/invite/count", true, true, Invite_Count}, // 邀请人数

		{yf.HTTP_GET, "/paid/affiche/list", true, false, RebatePaidAfficheList}, // 折扣专区购买公告

		// 收货地址相关接口
		{yf.HTTP_GET, "/user/address/list", true, true, Address_List},
		{yf.HTTP_POST, "/user/address/upsert", true, true, Address_Upsert},
		{yf.HTTP_POST, "/user/address/del", true, true, Address_Del},

		// 首页推荐接口
		{yf.HTTP_GET, "/index/recom", true, false, Recommend_Index},
		{yf.HTTP_GET, "/index/recom_list", true, false, Recommend_List},

		{yf.HTTP_POST, "/user/logistics/upsert", true, true, Logistics_Upsert}, // 添加物流单号
		{yf.HTTP_GET, "/user/logistics/info", true, false, Logistics_Info},     // 查看物流单号

		// 抽奖活动
		{yf.HTTP_GET, "/lotterys/list", true, false, Lotterys},                  // 抽奖列表
		{yf.HTTP_GET, "/lotterys/info", true, false, Lotterys_Info},             // 抽奖详情
		{yf.HTTP_GET, "/lotterys/record", true, false, Lotterys_Record},         // 抽奖记录列表
		{yf.HTTP_GET, "/lotterys/self/record", true, true, Lotterys_SelfRecord}, // 自己的投奖记录列表
		{yf.HTTP_POST, "/lotterys/pay", true, true, Lotterys_Pay},               // 立即抽奖
		{yf.HTTP_POST, "/lotterys/confirm", true, true, Lotterys_Confirm},       // 确认订单
		{yf.HTTP_GET, "/lotterys/key", true, true, Lotterys_Key},                // 获取memo
	}
	return
}
