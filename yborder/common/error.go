package common

const (
	ERR_TEL_INVALID              string = "ERR_TEL_INVALID"              // 手机号不正确
	ERR_TEL_NOT_SUPPORT_RECHARGE string = "ERR_TEL_NOT_SUPPORT_RECHARGE" // 不支持此手机号充值
	ERR_FORBIDDEN_COMMENT        string = "ERR_FORBIDDEN_COMMENT"
	ERR_FORBIDDEN_FILM_ARTICLE   string = "ERR_FORBIDDEN_FILM_ARTICLE"
	ERR_FORBIDDEN_FILM_LIST      string = "ERR_FORBIDDEN_FILM_LIST"
	ERR_FORBIDDEN_WEMEDIA        string = "ERR_FORBIDDEN_WEMEDIA"
	ERR_FORBIDDEN_COMMENTARY     string = "ERR_FORBIDDEN_COMMENTARY"
	ERR_FORBIDDEN_FILM_DISCUSS   string = "ERR_FORBIDDEN_FILM_DISCUSS"
	ERR_MQ_NAME_NOT_FOUND        string = "ERR_NAME_NOT_FOUND"
	ERR_ENCRYPT_FAILED           string = "ERR_ENCRYPT_FAILED"
	ERR_DECRYPT_FAILED           string = "ERR_DECRYPT_FAILED"

	ERR_ZJPASSWORD_INVALID string = "ERR_ZJPASSWORD_INVALID"
	ERR_PRODUCT_NOT_MORE   string = "ERR_PRODUCT_NOT_MORE"
	ERR_MONEY_NOT_MORE     string = "ERR_MONEY_NOT_MORE"

	ERR_CURRENCY_TYPE_NOT_SUPPORT string = "ERR_CURRENCY_TYPE_NOT_SUPPORT"

	ERR_ADDRESS_NOT_EXIST              string = "ERR_ADDRESS_NOT_EXIST" // 地址不存在
	ERR_CARTID_NOT_EXIST               string = "ERR_CARTID_NOT_EXIST"
	ERR_FORBIDDEN_BUY_OWNGOODS         string = "ERR_FORBIDDEN_BUY_OWNGOODS" // 不能购物自己的商品
	ERR_USER_TYPE_NOT_SELLER           string = "ERR_USER_TYPE_NOT_SELLER"   // 非卖家用户类型
	ERR_ORDER_NOT_EXIST                string = "ERR_ORDER_NOT_EXIST"
	ERR_ORDER_AMOUNT_INVALID           string = "ERR_ORDER_AMOUNT_INVALID"
	ERR_ORDER_FORBIDDENT_MODIFY        string = "ERR_ORDER_FORBIDDENT_MODIFY"        // 订单不允许修改(已经支付)
	ERR_ORDER_FORBIDDEN_PAY            string = "ERR_ORDER_FORBIDDEN_PAY"            // 订单禁止支付
	ERR_ORDER_HASPAYED                 string = "ERR_ORDER_HASPAYED"                 // 订单已经被支付
	ERR_ORDER_FORBIDDEN_CANCEL         string = "ERR_ORDER_FORBIDDEN_CANCEL"         // 订单禁止取消
	ERR_ORDER_FINISH_FAILED            string = "ERR_ORDER_FINISH_FAILED"            // 订单完成失败
	ERR_ORDER_FORBIDDEN_DEL            string = "ERR_ORDER_FORBIDDEN_DEL"            // 订单禁止删除
	ERR_VIRTUAL_ORDER_FORBIDDEN_CANCEL string = "ERR_VIRTUAL_ORDER_FORBIDDEN_CANCEL" // 虚拟订单不能取消
	ERR_FORBIDDENT_MODIFY              string = "ERR_FORBIDDENT_MODIFY"              // 记录不允许修改

	ERR_AMOUNT_INVALID           string = "ERR_AMOUNT_INVALID"           // 数量错误
	ERR_YOUBUY_ACCOUNT_NOT_FOUND string = "ERR_YOUBUY_ACCOUNT_NOT_FOUND" // 优买会帐号不存在
	ERR_LOTTERYS_NOSTART         string = "ERR_LOTTERYS_NOSTART"         // 抽奖未开始
	ERR_LOTTERYS_OVER            string = "ERR_LOTTERYS_OVER"            // 抽奖已结束
	ERR_EXCEED_TIMES             string = "ERR_EXCEED_TIMES"             // 超过次数
)
