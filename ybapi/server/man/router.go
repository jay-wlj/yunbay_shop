package man

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {
	//评论相关操作
	routers = []yf.RouterInfo{
		// 平台公告相关接口
		{yf.HTTP_POST, "/notice/upsert", true, false, Notice_Upsert},
		{yf.HTTP_POST, "/notice/del", true, false, Notice_Del},
		{yf.HTTP_GET, "/notice/list", true, false, Notice_List},
		{yf.HTTP_POST, "/notice/recommend", true, false, Notice_RecommendUpsert},

		// 平台反馈
		{yf.HTTP_GET, "/feedback/list", true, false, ManFeedback_List},

		{yf.HTTP_GET, "/invite/beinvites", true, false, ManInvite_BeInvites}, // 获取推荐人层级
		{yf.HTTP_GET, "/invite/beinvite", true, false, ManInvite_BeInvite},   // 获取一批用户的直接邀请者
		{yf.HTTP_POST, "/invite/add", true, false, ManInvite_Add},            // 添加邀请人

		// 商家认证相关接口
		{yf.HTTP_GET, "/business/status", true, false, Business_Status},
		{yf.HTTP_GET, "/business/list", true, false, Business_List},
		{yf.HTTP_POST, "/business/amount/update", true, false, Business_AmountUpdate},

		{yf.HTTP_POST, "/index/recom/upsert", true, false, RecommendUpsert}, // 推荐数据更新

		// 抽奖相关接口
		{yf.HTTP_POST, "/lotterys/hid", true, false, Lotterys_Hid},                // 删除抽奖任务
		{yf.HTTP_POST, "/lotterys/upsert", true, false, Lotterys_Upsert},          // 添加抽奖任务
		{yf.HTTP_GET, "/lotterys/list", true, false, Lotterys_List},               // 抽奖任务列表
		{yf.HTTP_GET, "/lotterys/info", true, false, Lotterys_Info},               // 抽奖任务详情
		{yf.HTTP_GET, "/lotterys/record", true, false, Lotterys_Record},           // 单个抽奖任务参与记录
		{yf.HTTP_POST, "/lotterys/record/hash", true, false, Lotterys_RecordHash}, // 更新抽奖记录的链hash

	}
	return
}
