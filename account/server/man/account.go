package man

import (
	"github.com/jie123108/glog"
	//"database/sql"
	"fmt"
	"github.com/jay-wlj/gobaselib/yf"

	//"strconv"
	"encoding/json"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"strings"
	"time"
	"yunbay/account/common"
	"yunbay/account/dao"
	"yunbay/account/models"
	"yunbay/account/server/share"
	"yunbay/account/util"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type ManAccountController struct{}

type useridSt struct {
	UserId int64 `json:"user_id" valid:"Required"`
}

// @router /logout [post]
func (t *ManAccountController) Logout(c *gin.Context) {
	var req useridSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	share.TokenDelByUserid(req.UserId)
	yf.JSON_Ok(c, gin.H{})
}

type smsSt struct {
	UserIds []int64 `json:"user_ids"`
	Content string  `json:"content"`
}

// @router /sms/send [post]
func (t *ManAccountController) SendSms(c *gin.Context) {
	var req smsSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	fail_ids := []int64{}
	for _, v := range req.UserIds {
		if err := sendSms(v, req.Content); err != nil {
			fail_ids = append(fail_ids, v)
		}
	}
	yf.JSON_Ok(c, gin.H{"fail_ids": fail_ids})
}

type smsTelSt struct {
	Tels    []models.TelInfo `json:"tels"`
	Content string           `json:"content"`
}

func (t *ManAccountController) SendSmsByTels(c *gin.Context) {
	var req smsTelSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	fail_ids := []models.TelInfo{}
	for _, v := range req.Tels {
		if err := util.SendSms(v.Cc, v.Tel, req.Content); err != nil {
			fail_ids = append(fail_ids, v)
		}
	}
	yf.JSON_Ok(c, gin.H{"fail_tels": fail_ids})
}

func sendSms(user_id int64, content string) (err error) {
	account, err := dao.GetAccountById(user_id)
	if err != nil {
		glog.Error("GetAccountById(", user_id, ") failed! err:", err)
		return
	}
	if account == nil {
		return
	}
	err = util.SendSms(account.Cc, account.Tel, content)
	return
}

// @router /token/check [get]
func (t *ManAccountController) TokenCheck(c *gin.Context) {
	user_info := c.MustGet("user_info").(util.TokenInfo)
	yf.JSON_Ok(c, gin.H{"user_id": user_info.UserId, "user_type": user_info.UserType, "expire_time": user_info.ExpireTime})
}

type userTypeSet struct {
	UserId   int64 `json:"user_id" valid:"Required"`
	UserType int16 `json:"user_type" valid:"Required"`
}

// @router /usertype [post]
func (t *ManAccountController) SetUserType(c *gin.Context) {
	var req userTypeSet
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	account := &common.Account{Id: req.UserId, UserType: req.UserType, UpdateTime: time.Now().Unix()}
	d := db.GetTxDB(c)
	err := d.Model(account).Updates(map[string]interface{}{"user_type": req.UserType, "update_time": account.UpdateTime}).Error

	if err != nil {
		glog.Error("UpdateAccountById(", req, ") failed! err:", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 用户类型改变 需要删除用户id对应的token以便让用户重新登录
	share.TokenDelByUserid(req.UserId)

	yf.JSON_Ok(c, gin.H{})
}

// @router /userinfo/get [get]
func (t *ManAccountController) UserinfoGet(c *gin.Context) {
	str_ids, _ := base.CheckQueryStringField(c, "user_ids")
	user_ids := base.StringToInt64Slice(str_ids, ",")

	vs := []common.Account{}
	var err error
	if len(user_ids) > 0 {
		vs, err = dao.GetAccountByIds(user_ids)
		if err != nil {
			glog.Error("GetAccountByIds fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
		// for i, v := range vs {
		// 	vs[i].CertInfos, _ = dao.GetCertByUserId(v.Id)
		// }
	}

	yf.JSON_Ok(c, gin.H{"list": vs})
}

// @router /userinfo/search [get]
func (t *ManAccountController) UserinfoSearch(c *gin.Context) {
	cc := c.GetString("cc")
	tel := c.GetString("tel")
	username := c.GetString("username")

	var v *common.Account
	var err error
	if cc != "" || tel != "" {
		v, err = dao.GetAccountByTel(cc, tel)
		if err != nil && err != gorm.ErrRecordNotFound {
			glog.Error("GetAccountByTel fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	} else if username != "" {
		v, err = dao.GetAccountByUsername(username)
		if err != nil {
			glog.Error("GetAccountByUsername fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}
	// 获取认证信息
	if v.Id > 0 {
		v.Cert, _ = dao.GetCertByUserId(v.Id)
	}

	yf.JSON_Ok(c, v)
}

type userSmsCode struct {
	Code string `json:"code" valid:"Required"`
}

// @router /sms/code/check [post]
func (t *ManAccountController) CodeCheck(c *gin.Context) {
	user_info := c.MustGet("user_info").(util.TokenInfo)
	var req userSmsCode
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	v, err := dao.GetAccountById(user_info.UserId)
	if err != nil {
		glog.Error("CodeCheck fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if v == nil {
		yf.JSON_Fail(c, common.ERR_USER_NOT_EXIST)
		return
	}
	tel := models.TelInfo{Cc: v.Cc, Tel: v.Tel}
	tel_full := tel.FullTel()
	ok, reason := share.CheckSmsCode(tel_full, req.Code)
	if !ok {
		yf.JSON_Fail(c, reason)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

type smspwdCode struct {
	Code       string `json:"code" valid:"Required"`
	ZJPassword string `json:"zjpassword" valid:"Required"`
}

// 验证资金密码及手机验证码
// @router /smspwd/check [post]
func (t *ManAccountController) SmsPwdCheck(c *gin.Context) {
	user_info := c.MustGet("user_info").(util.TokenInfo)
	var req smspwdCode
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	v, err := dao.GetAccountById(user_info.UserId)
	if err != nil {
		glog.Error("CodeCheck fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if v == nil {
		yf.JSON_Fail(c, common.ERR_USER_NOT_EXIST)
		return
	}
	tel := models.TelInfo{Cc: v.Cc, Tel: v.Tel}
	tel_full := tel.FullTel()
	ok, reason := share.CheckSmsCode(tel_full, req.Code)
	if !ok {
		yf.JSON_Fail(c, reason)
		return
	}

	//校验密码.
	valid := util.CheckPwd(req.ZJPassword, v.ZJPassword)
	if !valid {
		glog.Info("user [", user_info.UserId, "] password error!")
		yf.JSON_Fail(c, common.ERR_ZJPASSWORD_INVALID)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}

type userauthSt struct {
	UserId int64  `json:"user_id" valid:"Required"`
	Status int    `json:"status" valid:"Required"`
	Reason string `json:"reason"`
}

type uidSt struct {
	UserId int64 `json:"user_id"`
}

// 验证资金密码及手机验证码
// @router /userinfo/auth/check [post]
func (t *ManAccountController) UserinfoAuthCheck(c *gin.Context) {
	//user_info := c.MustGet("user_info").(util.TokenInfo)
	var req userauthSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	maner, _ := util.GetHeaderString(c, "X-Yf-Maner")
	if maner == "" {
		h, _ := json.Marshal(c.Request.Header)
		glog.Error("UserinfoAuthCheck fail! headers:", string(h))
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	v, err := dao.GetCertByUserId(req.UserId)
	if err != nil || v == nil {
		glog.Error("UserinfoAuthCheck fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 查找是否身份证号码是否已被采用

	v.Status = req.Status
	v.Reason = req.Reason
	v.Maner = maner
	v.UpdateTime = time.Now().Unix()
	db := db.GetTxDB(c)
	err = db.Model(v).Updates(map[string]interface{}{"status": v.Status, "reason": v.Reason, "maner": v.Maner, "update_time": v.UpdateTime}).Error
	if err != nil {
		glog.Error("UserinfoAuthCheck fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 保存用户信息里的实名状态
	a, err := dao.GetAccountById(req.UserId)
	if err != nil {
		glog.Error("UpsertCertification failed! err:", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if a == nil {
		yf.JSON_Fail(c, common.ERR_USER_NOT_EXIST)
		return
	}
	a.CertStatus = req.Status
	if err = db.Model(a).Updates(map[string]interface{}{"cert_status": a.CertStatus}).Error; err != nil {
		glog.Error("UpsertCertification failed! err:", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if req.Status == common.STATUS_OK {
		// // 实名通过 发放邀请空投奖励
		v := uidSt{req.UserId}
		im := common.MQUrl{Methond: "POST", AppKey: "ybasset", Uri: "/man/reward/unlock", Data: v}
		if err = util.PublishMsg(im); err != nil {
			glog.Error("UserinfoAuthCheck PublishMsg fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}
	yf.JSON_Ok(c, gin.H{})
}

// 获取用户实名认证列表
// @router /userinfo/auth/list [get]
func (t *ManAccountController) UserinfoAuthList(c *gin.Context) {
	//user_info := c.MustGet("user_info").(util.TokenInfo)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	user_id, _ := base.CheckQueryIntDefaultField(c, "user_id", -1)
	card_name, _ := base.CheckQueryStringField(c, "card_name")
	card_id, _ := base.CheckQueryStringField(c, "card_id")
	status, _ := base.CheckQueryIntDefaultField(c, "status", -1)
	country, _ := base.CheckQueryIntDefaultField(c, "country", -1)

	db := db.GetDB()

	if user_id > 0 {
		db.DB = db.Where("user_id=?", user_id)
	}
	if card_name != "" {
		db.DB = db.Where("card_name=?", card_name)
	}
	if card_id != "" {
		db.DB = db.Where("card_id=?", card_id)
	}
	if status > -1 {
		db.DB = db.Where("status=?", status)
	}
	if country > -1 {
		db.DB = db.Where("country=?", country)
	}

	var total int64 = 0
	if err := db.Model(&common.Cert{}).Count(&total).Error; err != nil {
		glog.Error("UserinfoAuthList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	db.DB = db.Order("status asc, create_time desc")
	vs := []common.Cert{}
	if err := db.ListPage(page, page_size).Find(&vs).Error; err != nil {
		glog.Error("UserinfoAuthList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}

	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended, "total": total})
}

type lParams struct {
	Page       int
	PageSize   int
	UserId     int64
	Tel        string
	CertStatus int
	UserType   int
	Country    int
	BeginDate  string
	EndDate    string
	Sorts      []string
	Orders     []string
}

// @router /userinfo/list [get]
func (t *ManAccountController) List(c *gin.Context) {
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	user_id, _ := base.CheckQueryIntDefaultField(c, "user_id", -1)
	tel, _ := base.CheckQueryStringField(c, "tel")
	cert_status, _ := base.CheckQueryIntDefaultField(c, "cert_status", -2)
	user_type, _ := base.CheckQueryIntDefaultField(c, "user_type", -1)
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")
	sort_str, _ := base.CheckQueryStringField(c, "sorts")
	order_str, _ := base.CheckQueryStringField(c, "orders")
	country, _ := base.CheckQueryIntDefaultField(c, "country", -1)

	sorts := strings.Split(sort_str, ",")
	orders := strings.Split(order_str, ",")

	db := db.GetDB()
	if user_id > -1 {
		db.DB = db.Where("user_id=?", user_id)
	}
	if user_type > -1 {
		db.DB = db.Where("user_type=?", user_type)
	}
	if country > -1 {
		db.DB = db.Where("country=?", country)
	}
	if tel != "" {
		db.DB = db.Where("tel=?", tel)
	}
	if cert_status > -2 {
		db.DB = db.Where("cert_status=?", cert_status)
	}
	if begin_date != "" {
		db.DB = db.Where("date>=?", begin_date)
	}
	if end_date != "" {
		db.DB = db.Where("date<=?", end_date)
	}

	var total int64 = 0
	if err := db.Model(&common.Account{}).Count(&total).Error; err != nil {
		glog.Error("List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// order
	for i, val := range sorts {
		order := "desc"
		if len(orders) > (i+1) && (orders[i] == "asc" || orders[i] == "desc") {
			order = orders[i]
		}
		if strings.TrimSpace(val) != "" {
			db.DB = db.Order(fmt.Sprintf("%v %v", val, order))
		}
	}
	db.DB = db.Order("create_time desc")
	vs := []common.Account{}
	if err := db.ListPage(page, page_size).Find(&vs).Error; err != nil {
		glog.Error("List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended, "total": total})
}

func (t *ManAccountController) Third_YoubuyAccount(c *gin.Context) {
	user_id, _ := base.CheckQueryInt64DefaultField(c, "user_id", -1)

	if user_id <= 0 {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}

	ret, err := share.GetYoubuyAccount(user_id)
	if err != nil {
		yf.JSON_Fail(c, err.Error())
		return
	}
	yf.JSON_Ok(c, ret)
}
