package common

import (
	"errors"
)

const (
	ERR_FORBIDDEN_COMMENT      string = "ERR_FORBIDDEN_COMMENT"
	ERR_FORBIDDEN_FILM_ARTICLE string = "ERR_FORBIDDEN_FILM_ARTICLE"
	ERR_FORBIDDEN_FILM_LIST    string = "ERR_FORBIDDEN_FILM_LIST"
	ERR_FORBIDDEN_WEMEDIA      string = "ERR_FORBIDDEN_WEMEDIA"
	ERR_FORBIDDEN_COMMENTARY   string = "ERR_FORBIDDEN_COMMENTARY"
	ERR_FORBIDDEN_FILM_DISCUSS string = "ERR_FORBIDDEN_FILM_DISCUSS"
	ERR_MQ_NAME_NOT_FOUND      string = "ERR_NAME_NOT_FOUND"
	ERR_ENCRYPT_FAILED         string = "ERR_ENCRYPT_FAILED"
	ERR_DECRYPT_FAILED         string = "ERR_DECRYPT_FAILED"

	ERR_USERASSET_LOCK                   string = "ERR_USERASSET_LOCK"                   // 用户资产被锁定
	ERR_WALLET_ADDRESS_NOT_FOUND         string = "ERR_WALLET_ADDRESS_NOT_FOUND"         // 用户钱包地址不慧
	ERR_WALLET_ADDRESS_WITHDRAW_NOTOWNER string = "ERR_WALLET_ADDRESS_WITHDRAW_NOTOWNER" // 用户的提币地址不能是自己的充值地址
	ERR_FORBIDDEN_WITHDRAW               string = "ERR_FORBIDDEN_WITHDRAW"               // 禁止提现
	ERR_ZJPASSWORD_INVALID               string = "ERR_ZJPASSWORD_INVALID"
	ERR_PRODUCT_NOT_MORE                 string = "ERR_PRODUCT_NOT_MORE"
	ERR_MONEY_NOT_MORE                   string = "ERR_MONEY_NOT_MORE"
	ERR_AMOUNT_EXCEED                    string = "ERR_AMOUNT_EXCEED" // 超过可提数量
	ERR_AMOUNT_INVALID                   string = "ERR_AMOUNT_INVALID"
	ERR_TYPE_NOT_SUPPORT                 string = "ERR_TYPE_NOT_SUPPORT" // 类型不支持
	ERR_CURRENCY_TYPE_NOT_SUPPORT        string = "ERR_CURRENCY_TYPE_NOT_SUPPORT"
	ERR_ADDRESS_INVALID                  string = "ERR_ADDRESS_INVALID"           // 地址不合法
	ERR_WITHDRAW_FORBIDDEN_MODIFY        string = "ERR_WITHDRAW_FORBIDDEN_MODIFY" // 提币禁止修改

	ERR_ORDER_NOT_EXIST         string = "ERR_ORDERS_NOT_EXIST"
	ERR_ORDER_AMOUNT_INVALID    string = "ERR_ORDER_AMOUNT_INVALID"
	ERR_ORDER_FORBIDDENT_MODIFY string = "ERR_ORDER_FORBIDDENT_MODIFY" // 订单不允许修改(已经支付)
	ERR_ORDER_HASPAYED          string = "ERR_ORDERS_HASPAYED"         // 订单已经被支付
	ERR_ORDER_FORBIDDEN_CANCEL  string = "ERR_ORDER_FORBIDDEN_CANCEL"  // 订单禁止取消
	ERR_ORDER_FINISH_FAILED     string = "ERR_ORDERS_FINISH_FAILED"    // 订单完成失败
	ERR_ORDER_FORBIDDEN_DEL     string = "ERR_ORDER_FORBIDDEN_DEL"     // 订单禁止删除

	ERR_YBT_NOT_MORE                 string = "ERR_YBT_NOT_MORE"                 // 可用ybt金额不够
	ERR_KT_NOT_MORE                  string = "ERR_KT_NOT_MORE"                  // 项目用户中可用kt金额不够
	ERR_YBT_USER_REWARD_NOT_MORE     string = "ERR_REWARD_ACTIVITY_NOT_MORE"     // 用户奖励池中可释放金额不够
	ERR_YBT_PROJECT_REWARD_NOT_MORE  string = "ERR_YBT_PROJECT_REWARD_NOT_MORE"  // 项目池中可释放金额不够
	ERR_YBT_MINEPOOL_REWARD_NOT_MORE string = "ERR_YBT_MINEPOOL_REWARD_NOT_MORE" // 矿池中可释放金额不够

	ERR_HOTCOIN_USER_NOT_FOUND string = "ERR_HOTCOIN_USER_NOT_FOUND" // 热币用户不存在
	ERR_YUNEX_USER_NOT_FOUND   string = "ERR_YUNEX_USER_NOT_FOUND"   // yunex用户不存在

	ERR_BANK_CARDID_ERROR string = "ERR_BANK_CARDID_ERROR" // 银行卡号错误
	ERR_BANK_NAME_ERROR   string = "ERR_BANK_NAME_ERROR"   // 开户行错误

	ERR_QRCODE_NOT_SUPPORT        string = "ERR_QRCODE_NOT_SUPPORT"        // 不支持的二维码
	ERR_VOUCHER_NOT_EXIST         string = "ERR_VOUCHER_NOT_EXIST"         // 消费券不存在
	ERR_FORBIDDEN_TRANSFER_TO_OWN string = "ERR_FORBIDDEN_TRANSFER_TO_OWN" // 不能转给自己
)

var ( 
	ErrNoRowsInSet error = errors.New("sql: no rows in result set")
)