package common



const (
	FLOW_TYPE_CHAIN int = 0
	FLOW_TYPE_YUNBAY int = 1
)

const (
	STATUS_FAIL = -1	// 失败状态
	STATUS_INIT = 0		// 初始状态
	STATUS_OK = 1		// 成功状态
)

const (
	USER_TYPE_NORMAL = 0  	// 正常用户
	USER_TYPE_BUSINESS = 1  // 商家用户
)

const (
	CURRENCY_YBT    = 0  //	ybt币种
	CURRENCY_KT		= 1  // kt币种 	
	CURRENCY_RMB	= 2	 // rmb

	CURRENCY_UNKNOW = 3 // 未知币种
)

const (
	CHANNEL_UNKNOW = -1			// 未知渠道
	KT_CHANNEL_CHAIN = 0		// kt充提官方渠道
	KT_CHANNEL_HOTCOIN = 1		// kt充提热币渠道
	KT_CHANNEL_YUNEX = 2		// kt充提云网渠道

	CHANNEL_ALIPAY = 10			// 支付宝
	CHANNEL_WEIXIN = 11			// 微信
)

const (
	KT_TRANSACTION_RECHARGE = 0 // 充值
	KT_TRANSACTION_PICKUP = 1 	// 提币
	KT_TRANSACTION_PROFIT = 2 	// 收益金
	KT_TRANSACTION_CONSUME = 3 	// 商品消费
	KT_TRANSACTION_SELLER = 4 	// 商品卖出
	KT_TRANSACTION_RETURND = 5 	// 退款
	KT_TRANSACTION_FEE = 6 		// 手续费用
	KT_TRANSACTION_PROJECT = 7 	// 项目分红
	//KT_TRANSACTION_TRANSFER = 8 // 内盘转帐
)

const (
	YBT_TRANSACTION_RECHARGE = 0 // 充值
	YBT_TRANSACTION_PICKUP = 1 // 提币
	YBT_TRANSACTION_CONSUME = 2 // 消费(挖矿)奖励
	YBT_TRANSACTION_SELLER = 3 // 商家奖励
	YBT_TRANSACTION_INVITE = 4 // 邀请奖励
	YBT_TRANSACTION_ACTIVITY = 5 // 活动奖励
	YBT_TRANSACTION_AIRDROP = 6 // 空投奖励
	YBT_TRANSACTION_PROJECT = 7	// 项目方奖励
	YBT_TRANSACTION_FEE = 8     // 手续费用
	YBT_TRANSACTION_BUY = 9 	// 商品消费
	YBT_TRANSACTION_RETURND = 10 // 退款
	//YBT_TRANSACTION_TRANSFER = 8 // 内盘转帐
)

const (	
	YBT_REWARD_AIRDROP = 0 		// ybt空投奖励
	YBT_REWARD_ACTIVITY = 1 	// ybt活动奖励
	YBT_REWARD_MING = 2 		// ybt挖矿奖励
	YBT_REWARD_PROJECT = 3		// ybt项目方奖励
)


const (
	ASSET_POOL_LOCK int = 0				// 用户交易资金平台冻结中
	ASSET_POOL_FINISH int = 1			// 用户交易完成，平台已向卖家打款
	ASSET_POOL_CANCEL int = 2			// 用户交易取消，平台已将款项向买家打回	
)

const (
	ORDER_STATUS_INIT int = 0			// 购物车
	ORDER_STATUS_UNPAY int = 1			// 待支付
	ORDER_STATUS_PAYED int = 2			// 已付款
	ORDER_STATUS_SHIPPED int = 3		// 已发货
	ORDER_STATUS_FINISH int = 4			// 已完成(已收货)	
	ORDER_STATUS_CANCEL int = 5			// 已取消	
	ORDER_STATUS_REFUND int = 6			// 已退款
)

const (
	SALE_STATUS_INIT int = 0 			// 未售后
	SALE_STATUS_ING int = 1 			// 售后中
	SALE_STATUS_END int = 2 			// 售后完成
)

const (
	ASSET_LOCK_AIRDROP int = 0				// 空投冻结
	ASSET_LOCK_FIX	int = 1					// 定期冻结
	ASSET_LOCK_FOREVER	int = 2				// 永结冻结
	ASSET_LOCK_WITHDRAW	int = 3				// 提币冻结
)

const (
	TX_STATUS_NOTPASS		int = -1 	// 审核不通过
	TX_STATUS_INIT 			int = 0		// 未审核
	TX_STATUS_CHECKPASS		int = 1		// 审核通过
	TX_STATUS_WAITING		int = 2		// 等待提交
	TX_STATUS_SUBMIT		int = 3		// 区块交易已提交
	TX_STATUS_CONFIRM		int = 4		// 区块交易确认中
	TX_STATUS_FAILED		int = 5		// 区块交易失败
	TX_STATUS_SUCCESS		int = 6		// 区块交易成功
)

const (
	PROJECT_TYPE_TEAMS 			int = 0			// 团队激励
	PROJECT_TYPE_DEVELOPMENT 	int = 1			// 项目研发
	PROJECT_TYPE_REPURCHASE 	int = 2			// YBT回购
	PROJECT_TYPE_INVESTORS 		int = 3			// 战略投资
)



func GetCurrencyName(currency_type int)string{
	switch currency_type {
	case CURRENCY_YBT:
		return "ybt"
	case CURRENCY_KT:
		return "kt"
	case CURRENCY_RMB:
		return "rmb"
	default:
		return ""
	}
}

