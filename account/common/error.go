package common

import (
	"errors"
)

const (
	ERR_ARGS_INVALID       string = "ERR_ARGS_INVALID"       //请求参数错误：缺少字段，或字段值不对。
	ERR_SERVER_ERROR       string = "ERR_SERVER_ERROR"       //服务器内部错误：如访问数据库出错了
	ERR_SIGN_ERROR         string = "ERR_SIGN_ERROR"         //签名错误
	ERR_TOKEN_INVALID      string = "ERR_TOKEN_INVALID"      //Token错误,或失效
	ERR_TOKEN_EXPIRED      string = "ERR_TOKEN_EXPIRED"      //Token已经过期.
	ERR_IMGCODE_INVALID    string = "ERR_IMGCODE_INVALID"    //图片验证码错误.
	ERR_CODE_INVALID       string = "ERR_CODE_INVALID"       //验证码错误.
	ERR_TEL_INVALID        string = "ERR_TEL_INVALID"        //手机号非法
	ERR_TEL_NOT_EXIST      string = "ERR_TEL_NOT_EXIST"      //手机号不存在/还未注册
	ERR_TEL_EXIST          string = "ERR_TEL_EXIST"          //手机号已经存在
	ERR_PASSWORD_ERR       string = "ERR_PASSWORD_ERR"       //密码错误
	ERR_USERNAME_EXIST     string = "ERR_USERNAME_EXIST"     //用户名已经存在
	ERR_USERNAME_INVALID   string = "ERR_USERNAME_INVALID"   //用户名非法
	ERR_USER_NOT_EXIST     string = "ERR_USER_NOT_EXIST"     //用户不存在
	ERR_OBJECT_NOT_EXIST   string = "ERR_OBJECT_NOT_EXIST"   //对象不存在
	ERR_PASSWORD_SAME      string = "ERR_PASSWORD_SAME"      //登录与资金密码相同
	ERR_CERT_CARDID_ERROR  string = "ERR_CERT_CARDID_ERROR"  //校验身份证id失败
	ERR_CARDID_EXIST       string = "ERR_CARDID_EXIST"       //身份证id已存在
	ERR_ZJPASSWORD_INVALID       string = "ERR_ZJPASSWORD_INVALID"       //资金密码错误
	ERR_YOUBUY_ACCOUNT_NOT_FOUND string = "ERR_YOUBUY_ACCOUNT_NOT_FOUND" // 优买会帐号不存在
)

var (
	ErrNoRowsInSet error = errors.New("sql: no rows in result set")
)
