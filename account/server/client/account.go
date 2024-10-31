package client

import (
	"fmt"
	"regexp"
	"time"
	"yunbay/account/common"
	"yunbay/account/dao"
	"yunbay/account/models"
	"yunbay/account/server/share"
	"yunbay/account/util"

	ydb "yunbay/account/db"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/jinzhu/gorm"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

type Account struct{}

// 帐号注册
func (t *Account) Reg(c *gin.Context) {
	var req models.AccountRegReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	if !req.Valid() {
		yf.JSON_Fail(c, common.ERR_TEL_INVALID)
		return
	}
	var id int64
	platform, version := util.GetPlatformVersionByContext(c)
	country := util.GetCountry(c)

	//country =
	now := time.Now().Unix()
	// TODO: 验证码出错次数验证.
	account, err := dao.GetAccountByTel(req.Cc, req.Tel)
	if err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetAccountByTel(", req.Cc, ",", req.Tel, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	tel_full := req.FullTel()

	if req.Password == "" || (req.Type == 0 && req.ZJPassword == "") {
		glog.Error("the password is empty!")
		yf.JSON_Fail(c, common.ERR_ARGS_INVALID)
		return
	}
	if req.Password == req.ZJPassword {
		glog.Error("the login pwd and zjpwd is same!")
		yf.JSON_Fail(c, common.ERR_PASSWORD_SAME)
		return
	}

	db := ydb.GetTxDB(c)
	if req.Type > 0 { //重设登陆密码
		// 重设登陆密码，但是手机号并不存在。(需要提示注册)
		if account == nil {
			glog.Error("tel [", req.FullTel(), "] is not exist")
			yf.JSON_Fail(c, common.ERR_TEL_NOT_EXIST)
			return
		}
		id = int64(account.Id)

		//校验验证码
		ok, reason := share.CheckSmsCode(tel_full, req.Code)
		if !ok {
			yf.JSON_Fail(c, reason)
			return
		}
		// 校验登录与资金密码不能相同
		if util.CheckPwd(req.Password, account.ZJPassword) {
			glog.Error("the login pwd and zjpwd is same!")
			yf.JSON_Fail(c, common.ERR_PASSWORD_SAME)
			return
		}
		account.UpdateTime = time.Now().Unix()
		enc_pwd := util.MakePwd(req.Password)
		account.Password = enc_pwd
		err = db.Model(account).Updates(map[string]interface{}{"password": account.Password, "update_time": account.UpdateTime}).Error

		if err != nil {
			glog.Error("UpdateAccountById(", account, ") failed! err:", err)
			yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
			return
		}
	} else { //新注册
		//注册，但是手机号已经存在。
		if account != nil {
			glog.Error("tel [", req.FullTel(), "] is exist")
			yf.JSON_Fail(c, common.ERR_TEL_EXIST)
			return
		}

		//校验验证码
		ok, reason := share.CheckSmsCode(tel_full, req.Code)
		if !ok {
			yf.JSON_Fail(c, reason)
			return
		}
		enc_pwd := util.MakePwd(req.Password)
		enc_zjpwd := util.MakePwd(req.ZJPassword)
		ip := c.ClientIP()
		date := time.Now().Format("2006-01-02")
		//username := util.RandomSample("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", 6)	// 随机分配6字符的用户名
		//did := c.Ctx.Request.Header.Get("X-Yf-Devid")
		did, _ := util.GetDevId(c)
		account = &common.Account{
			Cc:         req.Cc,
			Tel:        req.Tel,
			Password:   enc_pwd,
			ZJPassword: enc_zjpwd,
			UserType:   0,
			//Username: username,
			Platform:   platform,
			Version:    version,
			DeviceId:   did,
			CertStatus: -1,
			Ip:         ip,
			Country:    country,
			Date:       date,
			UpdateTime: now,
			CreateTime: now,
		}
		if err = db.Save(&account).Error; err != nil {
			glog.Error("AddAccount(", account, ") failed! err:", err)
			yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
			return
		}
		id = account.Id
		glog.Info("req.Type=", req.Type, " user_id:", id)
		// 注册用户邀请人id入库
		// 先查看邀请人id是否已存在
		if req.FromInviteId > 0 {
			if v, err := dao.GetAccountById(req.FromInviteId); err != nil || v == nil {
				glog.Error("user ret inviteid is not exist! invite_id:", req.FromInviteId)
				req.FromInviteId = 0
			}
		} else if req.FromInviteId < 0 {
			req.FromInviteId = 0
		}

		u := common.MQUserAdd{UserId: id, RecommendUserId: req.FromInviteId, Tel: req.Tel}
		if err = util.PublishMsg(u); err != nil {
			glog.Error("PublishMsg fail! err=", err)
			yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
			return
		}

		// 注册im
		mq := common.MQUrl{Methond: "POST", Uri: "/man/user/register", AppKey: "ybim", Data: u, MaxTrys: -1}
		util.PublishMsg(mq)
		// func RegisterIM(user_id int64) (err error) {
		// 	uri := "/man/user/register"
		// 	v := userTypeSt{UserId: user_id}
		// 	err = post_info(uri, "ybim", nil, v, "", nil, "", false, EXPIRE_RES_INFO)
		// 	if err != nil {
		// 		glog.Error("RegisterIM fail! err=", err)
		// 		return
		// 	}
		// 	return
		// }

	}

	var token string
	token, err = share.Login_internal(id, account.UserType)
	if err != nil {
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	glog.Info("user [", req.Cc, "-", req.Tel, "] user_id:", id, " login success!")

	yf.JSON_Ok(c, gin.H{"token": token, "user_id": id})
}

// 重置登陆密码
func (t *Account) ResetLoginPwd(c *gin.Context) {
	var req models.AccountLoginPwdResetCheckReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}

	account, err := dao.GetAccountById(user_id)
	if err != nil {
		glog.Error("GetAccountById(", user_id, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	if account == nil {
		yf.JSON_Fail(c, common.ERR_USER_NOT_EXIST)
		return
	}
	// 校验原密码
	if !util.CheckPwd(req.OldPwd, account.Password) {
		glog.Error("CheckPwd(", user_id, ") failed!")
		yf.JSON_Fail(c, common.ERR_PASSWORD_ERR)
		return
	}
	// 校验登录与资金密码不能相同
	if util.CheckPwd(req.NewPwd, account.ZJPassword) {
		glog.Error("the login pwd and zjpwd is same!")
		yf.JSON_Fail(c, common.ERR_PASSWORD_SAME)
		return
	}
	// 重置登陆密码.
	enc_pwd := util.MakePwd(req.NewPwd)
	db := ydb.GetTxDB(c)
	err = db.Model(account).Updates(map[string]interface{}{"password": enc_pwd, "update_time": time.Now().Unix()}).Error
	if err != nil {
		glog.Error("ResetLoginPwd(", account, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

// 重置密码
func (t *Account) ZJReset(c *gin.Context) {
	var req models.AccountZJCheckReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	user_info := c.MustGet("user_info").(util.TokenInfo)

	account, err := dao.GetAccountById(user_info.UserId)
	if err != nil {
		glog.Error("GetAccountById(", user_info.UserId, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	if account == nil {
		yf.JSON_Fail(c, common.ERR_USER_NOT_EXIST)
		return
	}
	tel := models.TelInfo{account.Cc, account.Tel}
	//校验验证码
	ok, reason := share.CheckSmsCode(tel.FullTel(), req.Code)
	if !ok {
		yf.JSON_Fail(c, reason)
		return
	}
	// 校验登录与资金密码不能相同
	if util.CheckPwd(req.ZJPassword, account.Password) {
		glog.Error("the login pwd and zjpwd is same!")
		yf.JSON_Fail(c, common.ERR_PASSWORD_SAME)
		return
	}
	// 重置资金密码.
	enc_pwd := util.MakePwd(req.ZJPassword)
	db := ydb.GetTxDB(c)
	err = db.Model(account).Updates(map[string]interface{}{"zjpassword": enc_pwd, "update_time": time.Now().Unix()}).Error
	if err != nil {
		glog.Error("UpdateAccountById(", account, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

type userAdd struct {
	UserId          int64  `json:"user_id"`
	Tel             string `json:"tel"`
	RecommendUserId int64  `json:"from_inviteid"`
}

// 登录
func (t *Account) Login(c *gin.Context) {
	var req models.AccountLoginReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	if !req.Valid() {
		yf.JSON_Fail(c, common.ERR_TEL_INVALID)
		return
	}
	country := util.GetCountry(c)

	var id int64
	platform, _ := util.GetPlatformVersionByContext(c)
	if platform == "web" {
		if ok := ImgCodeCheck(req.ImgKey, req.ImgCode); !ok {
			glog.Error("reg.ImgCode check fail!")
			yf.JSON_Fail(c, common.ERR_IMGCODE_INVALID)
			return
		}
	}
	account, err := dao.GetAccountByTel(req.Cc, req.Tel)
	if err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("GetAccountByTel(", req.Cc, ",", req.Tel, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	tel_full := req.FullTel()
	if account == nil {
		glog.Info("user [", tel_full, "] not exists!")
		yf.JSON_Fail(c, common.ERR_TEL_NOT_EXIST)
		return
	}

	//校验密码.
	valid := util.CheckPwd(req.Password, account.Password)
	if !valid {
		glog.Info("user [", tel_full, "] password error!")
		yf.JSON_Fail(c, common.ERR_PASSWORD_ERR)
		return
	}

	id = account.Id

	// yf.JSON_Ok(c, gin.H{})
	var token string
	token, err = share.Login_internal(id, account.UserType)
	if err != nil {
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	glog.Info("user [", req.Cc, "-", req.Tel, "] user_id:", id, " login success!")

	// 登陆成功入库
	now := time.Now().Unix()
	r := common.LoginRecord{UserId: id, Ip: c.ClientIP(), Country: country, CreateTime: now, UpdateTime: now}
	db := ydb.GetTxDB(c)
	if err = db.Save(&r).Error; err != nil {
		glog.Error("InsertLoginRecord fail! user_id:", id, " err:", err)
	}
	// 获取用户的IM token
	var imtoken common.ImToken
	// if err = db.Find(&imtoken, "user_id=?", id).Error; err != nil && err != gorm.ErrRecordNotFound {
	// 	glog.Error("imtoken find err")
	// 	return
	// }

	yf.JSON_Ok(c, gin.H{"token": token, "user_id": id, "accid": imtoken.Imid, "im_token": imtoken.Token})
}

// 退出登录
func (t *Account) Logout(c *gin.Context) {
	user_info := c.MustGet("user_info").(util.TokenInfo)
	err := share.Token_delete(user_info.Token, user_info.UserId)
	if err != nil {
		glog.Info("token_delete [", user_info.Token, "] failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

// 检测token
func (t *Account) TokenCheck(c *gin.Context) {
	user_info := c.MustGet("user_info").(util.TokenInfo)
	yf.JSON_Ok(c, gin.H{"user_id": user_info.UserId, "user_type": user_info.UserType, "expire_time": user_info.ExpireTime})
}

// 密码验证
func (t *Account) ZFAuth(c *gin.Context) {
	var req models.AccountZFAuthReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	user_info := c.MustGet("user_info").(util.TokenInfo)

	account, err := dao.GetAccountById(user_info.UserId)
	if err != nil {
		glog.Error("GetAccountById(", user_info.UserId, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	if account == nil {
		yf.JSON_Fail(c, common.ERR_USER_NOT_EXIST)
		return
	}
	//校验密码.
	valid := util.CheckPwd(req.ZJPassword, account.ZJPassword)
	if !valid {
		glog.Info("user [", user_info.UserId, "] password error!")
		yf.JSON_Fail(c, common.ERR_PASSWORD_ERR)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

// 检测用户名是否已被使用
func (t *Account) CheckUsername(c *gin.Context) {
	username := c.GetString("username")
	valid := util.UsernameCheck(username)
	if !valid {
		yf.JSON_Fail(c, common.ERR_USERNAME_INVALID)
		return
	}
	exist, err := dao.CheckUsernameExist(username)
	if err != nil {
		glog.Error("CheckUsernameExist(", username, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	if exist {
		yf.JSON_Fail(c, common.ERR_USERNAME_EXIST)
	} else {
		yf.JSON_Ok(c, gin.H{})
	}
}

// 设置用户名
func (t *Account) UserinfoSetUsername(c *gin.Context) {
	var req models.AccountUserinfoSetUsernameReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	user_info := c.MustGet("user_info").(util.TokenInfo)

	valid := util.UsernameCheck(req.Username)
	if !valid {
		glog.Error("username [", req.Username, "] is invalid!")
		yf.JSON_Fail(c, common.ERR_USERNAME_INVALID)
		return
	}

	// 检查用户名是否已经使用了.
	exist, err := dao.CheckUsernameExist(req.Username)
	if err != nil {
		glog.Error("CheckUsernameExist(", req.Username, ",", user_info.UserId, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	if exist { //用户名已经存在
		glog.Error("username [", req.Username, "] is exist!")
		yf.JSON_Fail(c, common.ERR_USERNAME_EXIST)
		return
	}

	account := &common.Account{Id: user_info.UserId}

	db := ydb.GetTxDB(c)
	err = db.Model(account).Updates(map[string]interface{}{"username": req.Username, "update_time": time.Now().Unix()}).Error
	if err != nil {
		glog.Error("UpdateAccountById(", req, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

// 帐号实名
func (t *Account) UserinfoAuth(c *gin.Context) {
	var req models.AccountUserinfoAuthReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	if req.CardCountry == "中国" {
		if match, _ := regexp.MatchString("^\\d{15}$|^\\d{18}$|^\\d{17}[xX]$", req.CardId); !match {
			glog.Error("CardId auth failed! cardId:", req.CardId)
			yf.JSON_Fail(c, common.ERR_CERT_CARDID_ERROR)
			return
		}
	}
	user_info := c.MustGet("user_info").(util.TokenInfo)

	// 判断该身份证号是否已有记录
	if exist, err := dao.ChecCardIdExist(req.CardId); err != nil || exist {
		if err != nil {
			glog.Error("GetCertByCardId failed! err:", err)
			yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
			return
		}
		yf.JSON_Fail(c, common.ERR_CARDID_EXIST)
		return
	}
	country := util.GetCountry(c)

	now := time.Now().Unix()
	//v := &common.Cert{UserId: user_info.UserId, CardCountry: req.CardCountry, CardName: req.CardName, CardId: req.CardId, CardImgs: base.StructToMap(req.Imgs), Country: country, CreateTime: now, UpdateTime: now}
	db := ydb.GetTxDB(c)

	var v common.Cert
	if err := db.Find(&v, "user_id=?", user_info.UserId).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("UpsertCertification failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	v.UserId = user_info.UserId
	v.CardCountry = req.CardCountry
	v.CardName = req.CardName
	v.CardId = req.CardId
	v.CardImgs = base.StructToMap(req.Imgs)
	v.Country = country
	v.Status = common.STATUS_INIT
	v.UpdateTime = now
	if 0 == v.Id {
		v.CreateTime = now
	}

	err := db.Save(&v).Error
	if err != nil {
		glog.Error("UpsertCertification failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	// 保存用户信息里的实名状态
	a, err := dao.GetAccountById(user_info.UserId)
	if err != nil {
		glog.Error("UpsertCertification failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	if a == nil {
		yf.JSON_Fail(c, common.ERR_USER_NOT_EXIST)
		return
	}
	a.CertStatus = 0
	if err = db.Model(a).Updates(map[string]interface{}{"cert_status": a.CertStatus}).Error; err != nil {
		glog.Error("UpsertCertification failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

type uidSt struct {
	UserId int64 `json:"user_id"`
}

// 设置本人帐号信息
func (t *Account) UserinfoSet(c *gin.Context) {
	var req models.AccountUserinfoSetReq
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	user_info := c.MustGet("user_info").(util.TokenInfo)

	ms := make(map[string]interface{})
	account := &common.Account{}
	account.Id = user_info.UserId
	if req.Avatar != "" {
		ms["avatar"] = req.Avatar
	}
	if req.Username != "" {
		ms["username"] = req.Username
	}
	if req.Birthday != "" {
		ms["birthday"] = req.Birthday
	}

	if len(ms) > 1 {
		ms["update_time"] = time.Now().Unix()
		db := ydb.GetTxDB(c)
		err := db.Model(account).Updates(ms).Error
		if err != nil {
			glog.Error("UpdateAccountById(", req, ") failed! err:", err)
			yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
			return
		}

		// 异步更新im的用户信息
		v := uidSt{user_info.UserId}
		im := common.MQUrl{Methond: "POST", AppKey: "ybim", Uri: "/man/user/info/update", Data: v}
		if err = util.PublishMsg(im); err != nil {
			glog.Error("UserinfoSet PublishMsg fail! err=", err)
			yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
			return
		}
	}

	yf.JSON_Ok(c, gin.H{})
}

// 获取本人帐号信息
func (t *Account) UserinfoGet(c *gin.Context) {
	user_info := c.MustGet("user_info").(util.TokenInfo)
	var v common.Account
	err := ydb.GetDB().Preload("Cert").Find(&v, "user_id=?", user_info.UserId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_USER_NOT_EXIST)
			return
		}
		glog.Error("GetAccountById(", user_info.UserId, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}

	//a := base.SelectStructView(v, "not man")
	a := base.FilterStruct(v, false, "platform", "version", "did", "ip", "date", "country")
	yf.JSON_Ok(c, a)
}

// 获取其它人的帐号信息
func (t *Account) UserinfoOther(c *gin.Context) {
	user_id, _ := base.CheckQueryInt64DefaultField(c, "user_id", -1)
	if user_id < 0 {
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	account, err := dao.GetAccountById(user_id)
	if err != nil {
		glog.Error("UserinfoGetOther(", user_id, ") failed! err:", err)
		yf.JSON_Fail(c, common.ERR_SERVER_ERROR)
		return
	}
	if account == nil {
		yf.JSON_Fail(c, common.ERR_USER_NOT_EXIST)
		return
	}
	// 用户名没有设置 则默认显示处理的手机号
	if account.Username == "" {
		num := len(account.Tel)
		if num > 11 {
			account.Username = account.Tel[:3] + "****" + account.Tel[7:]
		} else if num > 5 {
			account.Username = account.Tel[:2] + "***" + account.Tel[5:]
		} else {
			account.Username = fmt.Sprintf("%v", account.Id)
		}
	}
	//v := base.SelectStructView(account, "other")
	v := base.FilterStruct(account, true, "usernmae", "avatar")
	yf.JSON_Ok(c, v)
}

// 登录历史记录
func (t *Account) LoginRecord(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	user_info := c.MustGet("user_info").(util.TokenInfo)

	vs := []common.LoginRecord{}
	db := ydb.GetDB().ListPage(page, page_size)
	if err := db.Find(&vs, "user_id=?", user_info.UserId).Error; err != nil {
		glog.Error("LoginRecord fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if len(vs) == page_size {
		list_ended = true
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
}

func (t *Account) Third_YoubuyAccount(c *gin.Context) {
	user_info := c.MustGet("user_info").(util.TokenInfo)

	ret, err := share.GetYoubuyAccount(user_info.UserId)
	if err != nil {
		yf.JSON_Fail(c, err.Error())
		return
	}
	yf.JSON_Ok(c, ret)
}
