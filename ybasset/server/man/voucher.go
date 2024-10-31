package man

import (
	"fmt"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybasset/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/jinzhu/gorm"

	"github.com/jie123108/glog"

	"github.com/gin-gonic/gin"
)

type voucherSt struct {
	UserId     int64   `json:"user_id" binding:"required"`
	Type       int     `json:"type"`
	Amount     float64 `json:"amount" binding:"gt=0"`
	Title      string  `json:"title"`
	UnlockTime int64   `json:"unlock_time"`
}

func VoucherRecharge(c *gin.Context) {
	var req voucherSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	v := common.Voucher{UserId: req.UserId, Type: req.Type, UnlockTime: req.UnlockTime}
	now := time.Now().Unix()

	db := db.GetTxDB(c)
	db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (user_id, type) DO update set update_time=%v", now))
	if err := db.Save(&v).Error; err != nil {
		glog.Error("VoucherAdd fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	vr := common.VoucherRecord{VoucherId: v.Id, Amount: req.Amount, CreateTime: now, UpdateTime: now}
	db.DB = db.Set("gorm:insert_option", "")
	if err := db.Save(&vr).Error; err != nil {
		glog.Error("VoucherAdd fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"order_id": vr.Id})
}

// 设置代金券信息
func VoucherInfoUpsert(c *gin.Context) {
	var req common.VoucherInfo
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	req.CreateTime = time.Now().Unix()
	req.UpdateTime = req.CreateTime
	db := db.GetTxDB(c)
	db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (type) DO update set title='%v', context='%v', update_time=%v", req.Title, req.Context, req.UpdateTime))
	if err := db.Save(&req).Error; err != nil {
		glog.Error("VoucherInfoSet fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"id": req.Id})
}

type totalSt struct {
	Total int
}

// 代金券列表信息
func VoucherInfoList(c *gin.Context) {
	_type, _ := base.CheckQueryIntDefaultField(c, "type", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	vs := []common.VoucherInfo{}
	db := db.GetDB()
	if _type > -1 {
		db.DB = db.Where("type=?", _type)
	}
	var total totalSt
	var err error
	if err = db.Model(&common.VoucherInfo{}).Select("count(*) as total").Scan(&total).Error; err != nil {
		glog.Error("VoucherInfoList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if err = db.Order("type asc").Find(&vs).Error; err != nil {
		glog.Error("VoucherInfoList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": base.IsListEnded(page, page_size, len(vs), total.Total), "total": total.Total})
}

// 代金券详情
func VoucherInfo(c *gin.Context) {
	id, _ := base.CheckQueryIntField(c, "id")

	v := common.VoucherInfo{}
	db := db.GetDB()
	if err := db.Find(&v, "id=?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, yf.ERR_NOT_FOUND)
			return
		}
		glog.Error("VoucherInfoList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, v)
}

type voucheridSt struct {
	Id int64 `json:"id"`
}

// 更新代金券消费记录的信息
func VoucherRecordUpdate(c *gin.Context) {
	var req voucheridSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	var err error
	var v common.VoucherRecord
	db := db.GetTxDB(c)
	if err = db.Find(&v, "id=?", req.Id).Error; err != nil {
		glog.Error("VoucherRecordUpdate fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_NOT_FOUND)
		return
	}

	account, err1 := util.UserInfoByUid(v.ToUid)
	if err1 != nil {
		err = err1
		glog.Error("updateVoucherRecord fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	summary := account.Tel
	if account.Username != "" {
		summary = account.Username
	}

	if err = db.Model(&v).Updates(map[string]interface{}{"summary": summary}).Error; err != nil {
		glog.Error("updateVoucherRecord fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}
