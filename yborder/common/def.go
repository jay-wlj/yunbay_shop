package common

const (
	RedisPub   = "pub"
	RedisApi   = "api"
	RedisGoods = "api"
)

const (
	GOODS_TYPE_PHYSICAL     = 0  // 实体商品
	GOODS_TYPE_TEL_RECHARGE = 1  // 话费充值
	GOODS_TYPE_CARD         = 2  // 卡密商品
	GOODS_TYPE_YOUBUY       = 10 // 优买会代金券
)
const (
	USER_TYPE_COMMON = 0 // 普通用户
	USER_TYPE_SELLER = 1 // 商家用户
	USER_TYPE_SYSTEM = 2 // 系统用户
)

const (
	CURRENCY_YBT  = 0 // ybt币种
	CURRENCY_KT   = 1 // kt币种
	CURRENCY_RMB  = 2 // rmb
	CURRENCY_SNET = 3 // snet
	CURRENCY_USD  = 4 // 美金
	CURRENCY_HKD  = 5 // 港币
)

const (
	STATUS_FAIL = -1 // 失败状态
	STATUS_INIT = 0  // 初始状态
	STATUS_OK   = 1  // 成功状态
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
	PUBLISH_AREA_COMMON   = 0 // 代币销售专区
	PUBLISH_AREA_KT       = 1 // kt销售专区
	PUBLISH_AREA_REBAT    = 2 // 折扣销售专区
	PUBLISH_AREA_LOTTERYS = 3 // 抽奖专区
)

const (
	LOTTERYS_STATUS_FAIL = -1 // 已返还
	LOTTERYS_STATUS_INIT = 0  // 待开奖
	LOTTERYS_STATUS_NO   = 1  // 未中奖
	LOTTERYS_STATUS_YES  = 2  // 已中奖
)

const (
	ACTIVITY_STATUS_INIT    = iota // 待开启
	ACTIVITY_STATUS_RUNNING        // 开启中
	ACTIVITY_STATUS_END            // 结束
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
