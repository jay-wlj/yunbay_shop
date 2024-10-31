package share

import (
	"yunbay/account/common"
	"yunbay/account/dao"
	"yunbay/account/util"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"time"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func GetYoubuyAccount(user_id int64) (ret *common.ThirdAccount, err error) {

	account, err := dao.GetAccountById(user_id)
	if err != nil {
		glog.Error("Third_YoubuyAccount fail! err=", err)
		err = fmt.Errorf(yf.ERR_SERVER_ERROR)
		return
	}
	if account == nil {
		err = fmt.Errorf(yf.ERR_NOT_FOUND)
		return
	}

	// 先从本地查询
	var r common.ThirdAccount
	if err = db.GetDB().Find(&r, "user_id=?", user_id).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("Third_YoubuyAccount fail! err=", err)
		err = fmt.Errorf(yf.ERR_SERVER_ERROR)
		return
	}
	if err == nil {
		ret = &r
		return
	}

	// 根据手机号查询优买会帐号id及呢称等
	var youbuy_account *util.Account
	if youbuy_account, err = util.GetYoubuyAccount(account.Cc, account.Tel); err != nil {
		glog.Error("Third_YoubuyAccount fail! GetYoubuyAccount err=", err)
		if err.Error() == "ERR_USER_NOT_EXIST" {
			err = fmt.Errorf(common.ERR_YOUBUY_ACCOUNT_NOT_FOUND)
		}
		return
	}

	now := time.Now().Unix()
	ret = &common.ThirdAccount{UserId: user_id, ThirdName: "youbuy", ThirdId: youbuy_account.UserId, ThirdAccount: base.StructToMap(youbuy_account), CreateTime: now, UpdateTime: now}
	db := db.GetDB()
	if err = db.Save(ret).Error; err != nil {
		glog.Error("Third_YoubuyAccount fail! Save err=", err)
		err = fmt.Errorf(yf.ERR_SERVER_ERROR)
		return
	}
	return
}
