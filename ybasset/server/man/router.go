package man

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {
	//评论相关操作
	routers = []yf.RouterInfo{
		// 资产相关
		{yf.HTTP_POST, "/ybasset/info", true, false, UserAsset_Add},
		{yf.HTTP_GET, "/ybasset/list", true, false, YBAsset_List},     // 获取平台资产列表
		{yf.HTTP_POST, "/asset/user/lock", true, false, Asset_Lock},   // 冻结用户资产(不可提现)
		{yf.HTTP_POST, "/user/asset/add", true, false, UserAsset_Add}, // 添加用户资产信息
		{yf.HTTP_POST, "/user/wallet/draw", true, false, Wallet_Draw}, // 系统帐号提币

		// 支付接口
		{yf.HTTP_POST, "/asset/pay", true, true, Asset_Pay},
		{yf.HTTP_POST, "/asset/pay_by_userid", true, false, Asset_PayByUserId}, // 非token接口
		{yf.HTTP_POST, "/asset/payset", true, false, Asset_SetStatus},          // 支付成功or失败
		{yf.HTTP_POST, "/asset/pay/rebat", true, false, ManPayRebat},           // 订单折扣退款接口

		{yf.HTTP_GET, "/tradeflow/list", true, false, TradeFlow_List},
		{yf.HTTP_GET, "/rebat/order/list", true, false, BonusOrders_List}, // 订单消费挖矿分红列表
		{yf.HTTP_GET, "/rebat/order/info", true, false, BonusOrders_Info}, // 订单消费挖矿分红

		// 用户资产相关接口
		{yf.HTTP_GET, "/user/asset/list", true, false, Man_UserAssetList}, // 用户资产列表
		{yf.HTTP_GET, "/user/asset/info", true, false, Man_UserAssetInfo}, // 用户资产详情

		{yf.HTTP_GET, "/ybt/reward/list", true, false, Man_YbtRewardList},               // ybt挖矿释放列表
		{yf.HTTP_GET, "/kt/reward/list", true, false, Man_KtRewardList},                 // kt分红列表
		{yf.HTTP_POST, "/ybt/reward/check", true, false, Man_YbtRewardCheck},            // ybt发放审核
		{yf.HTTP_POST, "/kt/reward/check", true, false, Man_KtRewardCheck},              // kt分红审核
		{yf.HTTP_POST, "/kt/reward/seller/check", true, false, Man_KtRewardSellerCheck}, // 完成订单后商家所的ybt发放
		{yf.HTTP_POST, "/ybt/reward/activity", true, false, Man_YbtActivityReward},      // 发放活动奖励ybt
		{yf.HTTP_GET, "/reward/gift/list", true, false, Man_GiftRewardList},             //ybt奖励记录列表

		// 提币审核相关接口
		{yf.HTTP_POST, "/wallet/draw/check", true, false, ChainFlow_Check}, // 提币审核
		{yf.HTTP_GET, "/wallet/draw/list", true, false, ChainFlow_List},    // 提币审核列表
		{yf.HTTP_POST, "/wallet/draw/set", true, false, ChainFlow_DrawSet}, // 国内提现帐号结果

		// 充提币接口回调
		{yf.HTTP_GET, "/wallet/balance/byaddress", true, false, Man_BalanceByAddress},       // 通过地址查询内盘用户的资产信息等
		{yf.HTTP_POST, "/wallet/address/info", true, false, Wallet_Address},                 // 获取指定用户的充值地址
		{yf.HTTP_POST, "/wallet/recharge/callback", false, false, Wallet_Charge_Callback},   // 充值回调
		{yf.HTTP_POST, "/wallet/withdraw/callback", false, false, Wallet_Withdraw_Callback}, // 提币回调

		// 团队激励资产相关接口
		//{yf.HTTP_GET, "/project/asset/teams", false, false, true, Project_TeamsInfo},
		{yf.HTTP_POST, "/project/asset/reward", false, false, Project_Reward}, // 从团队激励帐户中扣除相应数量ybt分配到项目人员的帐户

		// yunex分红相关接口
		//{yf.HTTP_POST, "/third/bonus/deliver", true, false, Third_BonusKt},// 发放第三方平台yunex分红接口(已弃用)

		{yf.HTTP_POST, "/reward/unlock", true, false, Man_UnlockReward}, // 设置邀请奖励是否发放

		// 设置货币兑换比例
		//{yf.HTTP_POST, "/currency/rmbratio/set", true, false, CurrencyRmbRatioSet},
		{yf.HTTP_POST, "/currency/ratio/set", true, false, CurrencyRatioSet}, // 币种汇率设置
		{yf.HTTP_GET, "/currency/ratio", true, false, CurrencyRatio},         // 获取币种汇率

		{yf.HTTP_POST, "/rmb/recharge/notify", true, false, RmbRechargeNotify}, // rmb充值成功接口

		{yf.HTTP_POST, "/voucher/recharge", true, false, VoucherRecharge},          // 代金券充值
		{yf.HTTP_POST, "/voucher/upsert", true, false, VoucherInfoUpsert},          // 设置代金券详情信息
		{yf.HTTP_GET, "/voucher/list", true, false, VoucherInfoList},               // 代金券列表
		{yf.HTTP_GET, "/voucher/info", true, false, VoucherInfo},                   // 代金券详情
		{yf.HTTP_POST, "/voucher/record/update", true, false, VoucherRecordUpdate}, // 代金券消费记录更新

		// 抽奖相关接口
		{yf.HTTP_POST, "/lotterys/transfer", true, false, LotterysPay},         // 用户支付给抽奖帐号
		{yf.HTTP_POST, "/assert/transfer/refund", true, false, TransferRefund}, // 退还用户支付给抽奖帐号的资产
		{yf.HTTP_POST, "/lotterys/order/pay", true, false, LotterysOrderPay},   // 将抽奖池里的该计划 根据订单id下单 生成订单流水记录

	}
	return
}
