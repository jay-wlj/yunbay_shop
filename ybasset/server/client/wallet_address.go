package client

import (
	"yunbay/ybasset/common"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

// 获取用户自己的充币地址
func Wallet_Address(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	txType, _ := base.CheckQueryIntDefaultField(c, "type", 1)
	str_type := common.GetCurrencyName(txType)
	if str_type == "" {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	// 先从库中查询
	var err error
	var address string
	if address, err = share.GetAndSaveUserAddress(user_id); err != nil {
		glog.Error("Wallet_Address fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// if address, err = dao.GetUserWalletAddress(user_id); err != nil {
	// 	glog.Error("Wallet_Address fail! err=", err)

	// }

	// if err == gorm.ErrRecordNotFound {
	// 	// 创建帐号相关的地址
	// 	switch txType {
	// 	case common.CURRENCY_YBT, common.CURRENCY_KT:
	// 	default:
	// 		yf.JSON_Fail(c, common.ERR_TYPE_NOT_SUPPORT)
	// 		return
	// 	}

	// 	address, err = util.GetUserWalletAddress(user_id, str_type)
	// 	if err != nil {
	// 		glog.Error("GetUserWalletAddress fail! err=", err)
	// 		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 		return
	// 	}
	// 	if err = share.AddUserWalletAddress(user_id, address, txType); err != nil {
	// 		glog.Error("Wallet_Address fail! err=", err)
	// 		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 		return
	// 	}
	// }
	yf.JSON_Ok(c, gin.H{"bind_address": address})
}

// 添加钱包地址
type walletaddress struct {
	Id       int64  `json:"id"`
	Type     uint16 `json:"type" `
	Name     string `json:"name" binding:"required"`
	Adddress string `json:"adddress" binding:"required"`
	Default  bool   `json:"default"`
}

func Wallet_Address_Upsert(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req walletaddress
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	if match, _ := regexp.MatchString("^0x[0-9a-zA-Z]{40}$", req.Adddress); !match {
		glog.Error("Wallet_Address_Upsert useraddress not match:", req.Adddress)
		yf.JSON_Fail(c, common.ERR_ADDRESS_INVALID)
		return
	}

	now := time.Now().Unix()
	v := common.WalletAddress{Id: req.Id, Type: req.Type, UserId: user_id, Name: req.Name, Adddress: req.Adddress, Default: req.Default, UpdateTime: now}
	if req.Id == 0 {
		v.CreateTime = now
	}
	db := db.GetTxDB(c)
	var selcount int = 0
	if err := db.Model(&common.WalletAddress{}).Where("type=? and user_id=? and \"default\"=?", req.Type, user_id, true).Count(&selcount).Error; err != nil {
		glog.Error("Wallet_Address_Upsert fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 设为默认地址
	if 0 == selcount {
		v.Default = true
	}

	if err := db.Save(&v).Error; err != nil {
		glog.Error("Wallet_Address_Upsert faiL! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 将其它选项置非勾选状态
	if selcount > 1 || (v.Default && (selcount != 0)) {
		if err := db.Model(&common.WalletAddress{}).Where("type=? and user_id=? and id <> ?", req.Type, user_id, v.Id).Updates(map[string]interface{}{"default": false}).Error; err != nil {
			glog.Error("Wallet_Address_Upsert fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}

	// 查找地址所属平台
	share.QueryAddressChannel(req.Adddress, int(req.Type))
	yf.JSON_Ok(c, gin.H{"id": v.Id})
}

// 添加钱包地址
type idSt struct {
	Id int64 `json:"id" binding:"required"`
}

func Wallet_Address_Del(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var req idSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	var address common.WalletAddress
	db := db.GetTxDB(c)
	if err := db.Find(&address, "id=? and user_id=?", req.Id, user_id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_ADDRESS_INVALID)
			return
		}
		glog.Error("Wallet_Address_Del faiL! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	}
	if err := db.Delete(&address).Error; err != nil {
		glog.Error("Wallet_Address_Upsert faiL! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if address.Default {
		// 设置最新的为默认值
		var v common.WalletAddress
		if err := db.Model(&common.WalletAddress{}).Last(&v, "type=? and user_id=?", address.Type, user_id).Error; err != nil && err != gorm.ErrRecordNotFound {
			glog.Error("Wallet_Address_Del fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
		if v.Id > 0 {
			if err := db.Model(&common.WalletAddress{}).Where("id=?", v.Id).Updates(map[string]interface{}{"default": true}).Error; err != nil {
				glog.Error("Wallet_Address_Del fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
		}
	}

	yf.JSON_Ok(c, gin.H{})
}

func Wallet_Address_List(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	type_, _ := base.CheckQueryIntDefaultField(c, "type", 0)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	if page <= 0 {
		page = 1
	}
	if page_size <= 0 {
		page_size = 10
	}
	vs := []common.WalletAddress{}
	if err := db.GetDB().ListPage(page, page_size).Order("\"default\" desc, create_time desc").Find(&vs, "type=? and user_id=?", type_, user_id).Error; err != nil {
		glog.Error("Wallet_Address_Upsert faiL! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
}
