package common

const (
	STATUS_FAIL = -1 // 失败状态
	STATUS_INIT = 0  // 初始状态
	STATUS_OK   = 1  // 成功状态
)

const (
	USER_TYPE_NORMAL   = iota // 正常用户
	USER_TYPE_BUSINESS        // 商家用户
)

const (
	RedisPub   = "pub"
	RedisApi   = "api"
	RedisGoods = "api"
)

const (
	GOODS_TYPE_PHYSICAL     = iota // 实体商品
	GOODS_TYPE_TEL_RECHARGE        // 话费充值
)

const (
	CURRENCY_YBT  = iota //	ybt币种
	CURRENCY_KT          // kt币种
	CURRENCY_RMB         // cny 人民币
	CURRENCY_SNET        // snet
	CURRENCY_USD         // usd 美金
	CURRENCY_HKD         // 港币

	CURRENCY_UNKNOW // 未知币种
)

const (
	KT_TRANSACTION_RECHARGE = 0 // 充值
	KT_TRANSACTION_PICKUP   = 1 // 提币
	KT_TRANSACTION_PROFIT   = 2 // 收益金
	KT_TRANSACTION_CONSUME  = 3 // 商品消费
	KT_TRANSACTION_SELLER   = 4 // 商品卖出
	KT_TRANSACTION_RETURND  = 5 // 退款
)

const (
	YBT_TRANSACTION_RECHARGE = 0 // 充值
	YBT_TRANSACTION_PICKUP   = 1 // 提币
	YBT_TRANSACTION_CONSUME  = 2 // 消费(挖矿)奖励
	YBT_TRANSACTION_SELLER   = 3 // 商家奖励
	YBT_TRANSACTION_INVITE   = 4 // 邀请奖励
	YBT_TRANSACTION_ACTIVITY = 5 // 活动奖励
)

const (
	ORDER_STATUS_INIT    int = 0 // 购物车
	ORDER_STATUS_UNPAY   int = 1 // 待支付
	ORDER_STATUS_PAYED   int = 2 // 已付款
	ORDER_STATUS_SHIPPED int = 3 // 已发货
	ORDER_STATUS_FINISH  int = 4 // 已完成(已收货)
	ORDER_STATUS_CANCEL  int = 5 // 已取消
	ORDER_STATUS_REFUND  int = 6 // 已退款
)

const (
	SALE_STATUS_INIT int = 0 // 未售后
	SALE_STATUS_ING  int = 1 // 售后中
	SALE_STATUS_END  int = 2 // 售后完成
)

const (
	ASSET_POOL_LOCK   int = 0 // 用户交易资金平台冻结中
	ASSET_POOL_FINISH int = 1 // 用户交易完成，平台已向卖家打款
	ASSET_POOL_CANCEL int = 2 // 用户交易取消，平台已将款项向买家打回
)

const (
	PUBLISH_AREA_COMMON = 0 // 代币销售专区
	PUBLISH_AREA_KT     = 1 // kt销售专区
	PUBLISH_AREA_REBAT  = 2 // 折扣销售专区
)

func GetCurrencyName(currency_type int) string {
	switch currency_type {
	case CURRENCY_YBT:
		return "ybt"
	case CURRENCY_KT:
		return "kt"
	case CURRENCY_RMB:
		return "cny"
	case CURRENCY_SNET:
		return "snet"
	case CURRENCY_USD:
		return "usd"
	case CURRENCY_HKD:
		return "hkd"
	default:
		return ""
	}
}

func GetOrderStatusTxt(status int) string {
	switch status {
	case ORDER_STATUS_UNPAY:
		return "待支付"
	case ORDER_STATUS_PAYED:
		return "待发货"
	case ORDER_STATUS_SHIPPED:
		return "已发货"
	case ORDER_STATUS_FINISH:
		return "已完成"
	case ORDER_STATUS_CANCEL:
		return "已取消"
	case ORDER_STATUS_REFUND:
		return "已退款"
	}
	return ""
}
