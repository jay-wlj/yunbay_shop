package models

type AccountSMSSendReq struct {
	TelInfo
	Devid   string `json:"devid"`
	Type    int    `json:"type"`    //短信类型，默认是0： 0: 注册新用户，1：重设密码
	ImgKey  string `json:"imgkey"`  // 图片key
	ImgCode string `json:"imgcode"` // 图片验证码
}

type AccountSMSCheckReq struct {
	TelInfo
	Code string `json:"code" valid:"Required"` // 验证码
}

type AccountRegReq struct {
	TelInfo
	Password     string `json:"password"`              //密码(hash值)
	ZJPassword   string `json:"zjpassword`             // 资金密码(hash值)
	Code         string `json:"code" valid:"Required"` // 验证码
	Type         int    `json:"type" `                 //密码类型:0,注册 1,重设登陆密码
	FromInviteId int64  `json:"from_inviteid"`         // 邀请人id
}

type AccountLoginReq struct {
	TelInfo
	Devid    string `json:"devid"`
	Password string `json:"password" valid:"Required"` //密码(hash值)
	ImgKey   string `json:"imgkey"`                    // 图片key
	ImgCode  string `json:"imgcode"`                   // 图片验证码
}

type AccountUserinfoSetUsernameReq struct {
	Username string `json:"username" valid:"Required"` //用户名
}

type AccountUserinfoSetReq struct {
	Avatar string `json:"avatar"` //头像URL
	// Username   string `json:"username"`                //用户名
	Username string `json:"username"`
	//Sex        int16  `json:"sex"`       //性别: 0：女. 1：男. 2: 未知
	Birthday string `json:"birthday" ` //生日('1988-10-11')
	//Motto      string `json:"motto" `    //个人签名/简介
	UpdateTime int64 `json:"-" `
}

type AccountUserinfoAuthReq struct {
	CardCountry string  `json:"card_country"` // 国家
	CardName    string  `json:"card_name"`    // 用户实名
	CardId      string  `json:"card_id"`      // 身份证号
	Imgs        CartImg `json:"card_imgs"`    // 认证图片数组
	UpdateTime  int64   `json:"-" `
}

type CartImg struct {
	FrontImg string `json:"front_img" valid:"Required"`
	BackImg  string `json:"back_img" valid:"Required"`
	HandImg  string `json:"hand_img" valid:"Required"`
}

type AccountZFAuthReq struct {
	ZJPassword string `json:"zjpassword"` // 资金密码
}

type AccountZJCheckReq struct {
	AccountZFAuthReq
	Code string `json:"code" valid:"Required"` // 验证码
}

type AccountLoginPwdResetCheckReq struct {
	OldPwd string `json:"old_password" valid:"Required"` // 原密码
	NewPwd string `json:"new_password" valid:"Required"` // 新密码
}
