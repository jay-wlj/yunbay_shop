package client

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {
	//评论相关操作
	routers = []yf.RouterInfo{
		// 获取货币兑换比例
		{yf.HTTP_GET, "/currency/ratio", true, false, CurrencyRatio},
		{yf.HTTP_GET, "/currency/ratios", true, false, CurrencyAllRatio},

		// 用户资产相关接口
		{yf.HTTP_GET, "/user/asset/info", true, true, UserAsset_Get},                   // 获取用户资产信息
		{yf.HTTP_GET, "/user/asset/all", true, true, UserAsset_All},                    // 获取用户所有资产列表
		{yf.HTTP_GET, "/user/asset/list", true, true, UserAssetDetail_List},            //获取用户资产列表(条件筛选)
		{yf.HTTP_GET, "/user/asset/bonus/info", true, true, UserAssetDetail_BonusInfo}, // 获取用户昨日的kt收益金及ybt奖励
		{yf.HTTP_GET, "/user/asset/bonus/ybt", true, true, UserAssetDetail_YbtBonus},   // 获取用户往日ybt分红记录
		{yf.HTTP_GET, "/user/asset/bonus/kt", true, true, UserAssetDetail_KtBonus},     // 获取用户往日kt分红记录
		{yf.HTTP_GET, "/user/asset/ybt/info", true, true, UserAssetDetail_YbtInfo},     // 获取用户邀请的ybt等

		// 平台资产相关接口
		{yf.HTTP_GET, "/ybasset/detail/info", true, false, YBAssetDetail_Get},    // 获取昨天(某天)平台资产详细信息
		{yf.HTTP_GET, "/ybasset/detail/list", true, false, YBAssetDetail_List},   // 获取平台资产列表
		{yf.HTTP_GET, "/ybasset/info", true, false, YBAsset_Get},                 // 获取昨天(某天)平台资产信息
		{yf.HTTP_GET, "/ybasset/list", true, false, YBAsset_List},                // 按日获取平台帐户资产列表信息
		{yf.HTTP_GET, "/ybasset/release/info", true, false, YBAsset_ReleaseInfo}, // 获取某日挖矿释放信息
		{yf.HTTP_GET, "/ybasset/difficult", true, false, YBAsset_Diffcult},       // 获取某日挖矿难度

		// 提币接口
		{yf.HTTP_GET, "/user/wallet/info", true, true, Wallet_Address},                   // 查询用户钱包地址
		{yf.HTTP_POST, "/user/wallet/address/upsert", true, true, Wallet_Address_Upsert}, // 用户添加转帐地址
		{yf.HTTP_POST, "/user/wallet/address/del", true, true, Wallet_Address_Del},       // 用户删除地址
		{yf.HTTP_GET, "/user/wallet/address/list", true, true, Wallet_Address_List},      // 用户查询转帐地址列表
		{yf.HTTP_GET, "/user/wallet/draw/fee", true, true, Wallet_Fee},                   // 查询提币手续费用
		{yf.HTTP_POST, "/user/wallet/draw", true, true, Wallet_Draw},                     // 提币操作
		{yf.HTTP_GET, "/user/wallet/switch", true, false, Wallet_ChargeSwitch},           // 获取提币开关配置
		{yf.HTTP_POST, "/user/wallet/transfer", true, true, Wallet_Transfer},             // 平台内部转帐接口

		{yf.HTTP_POST, "/user/wallet/draw/alipay", true, true, Wallet_DrawAlipay}, // 提币到支付宝微信等接口
		{yf.HTTP_POST, "/user/wallet/draw/bank", true, true, Wallet_DrawBank},     // 提币到银行

		// 充提记录
		{yf.HTTP_GET, "/user/wallet/draw/list", true, true, Wallet_Withdraw_List},
		{yf.HTTP_GET, "/user/wallet/recharge/list", true, true, Wallet_Recharge_List},

		// 热币相关接口
		{yf.HTTP_GET, "/user/wallet/hotcoin/token", true, true, Wallet_HotCoinTokenGet},  // 此接口app调用
		{yf.HTTP_GET, "/user/wallet/hotcoin/address", true, true, Wallet_HotCoinAddress}, // 此接口app调用
		{yf.HTTP_GET, "/wallet/hotcoin/token", false, false, HotCoinToken},               // 通过token获取用户id及电话
		{yf.HTTP_GET, "/wallet/hotcoin/address", false, false, Wallet_AddressQuery},      // 通过用户充值地址查询用户id
		//{yf.HTTP_POST, "/wallet/hotcoin/recharge", false, false, HotCoin_Charge_Callback}, // 热币充值回调(老板决定关闭yunbay与热币的内盘互转功能)

		// yunex相关接口
		{yf.HTTP_GET, "/user/wallet/yunex/address", false, false, Wallet_YunexAddressQuery}, // 此接口app调用
		//{yf.HTTP_POST, "/wallet/yunex/recharge", false, false, Yunex_Charge_Callback},       // yunex充值回调(老板决定关闭yunbay与yunex的内盘互转功能)
		{yf.HTTP_GET, "/wallet/yunex/address", false, false, Yunex_AddressQuery},          // yunex地址查询
		{yf.HTTP_GET, "/wallet/yunex/balance", false, false, Yunex_BalanceQuery},          // // 查询yunex帐户在yunbay平台里的余额信息
		{yf.HTTP_POST, "/wallet/yunex/deposit/notify", false, false, Yunex_DepositNotify}, // 提币回调接口

		// llt相关接口
		{yf.HTTP_POST, "/wallet/snet/recharge", false, false, LLT_Recharge},      // llt充值接口
		{yf.HTTP_GET, "/wallet/snet/recharge/query", false, false, LLT_Recharge}, // llt充值查询接口

		// 支付扫码识别接口
		{yf.HTTP_GET, "/qrcode/query", false, false, Qrcode_Query},

		// 代金券相关接口(已弃用，之前购买此商品后同步到优买会资产中)
		{yf.HTTP_POST, "/user/voucher/pay", true, true, Voucher_Pay},
		{yf.HTTP_GET, "/user/voucher/list", true, true, Voucher_List},
		{yf.HTTP_GET, "/user/voucher/record/list", true, true, Voucher_RecordList},
		{yf.HTTP_GET, "/user/voucher/info", true, false, Voucher_Info},
	}
	return
}
