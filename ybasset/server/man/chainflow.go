package man

import (
	"yunbay/ybasset/common"
	"yunbay/ybasset/dao"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func ChainFlow_List(c *gin.Context) {
	id, _ := base.CheckQueryInt64DefaultField(c, "id", 0)
	channel, _ := base.CheckQueryIntDefaultField(c, "channel", -2)
	user_id, _ := base.CheckQueryInt64DefaultField(c, "user_id", -1)
	txType, _ := base.CheckQueryIntDefaultField(c, "type", -1)
	status, _ := base.CheckQueryIntDefaultField(c, "status", -2)
	begin_date, _ := base.CheckQueryStringField(c, "begin_date")
	end_date, _ := base.CheckQueryStringField(c, "end_date")
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)
	flow_type, _ := base.CheckQueryIntDefaultField(c, "flow_type", -1)
	man_draw, _ := base.CheckQueryBoolField(c, "mandraw")
	country, _ := base.CheckQueryIntDefaultField(c, "country", -1)

	db := db.GetDB()
	if id > 0 {
		db.DB = db.Where("id=?", id)
	}
	if channel > -2 {
		db.DB = db.Where("channel=?", channel)
	}
	if txType > -1 {
		db.DB = db.Where("tx_type=?", txType)
	}
	if status > -2 {
		db.DB = db.Where("status=?", status)
	}
	if user_id > -1 {
		db.DB = db.Where("user_id=?", user_id)
	}
	if flow_type > -1 {
		db.DB = db.Where("flow_type=?", flow_type)
	}
	if begin_date != "" {
		db.DB = db.Where("date >= ?", begin_date)
	}
	if end_date != "" {
		db.DB = db.Where("date <= ?", end_date)
	}
	if man_draw {
		db.DB = db.Where("user_id < 50000")
	}
	if country > -1 {
		db.DB = db.Where("country=?", country)
	}

	var total int64 = 0
	if err := db.Model(&common.WithdrawFlow{}).Count(&total).Error; err != nil {
		glog.Error("TradeFlow_List fail! count err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 获取提币数及手续费
	info := []dao.TotalWithDraw{}
	if err := db.Model(&common.WithdrawFlow{}).Select("tx_type, sum(amount) as amount, sum(fee) as fee, count(*) as count").Group("tx_type").Scan(&info).Error; err != nil {
		glog.Error("TradeFlow_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	db.DB = db.ListPage(page, page_size)
	vs := []common.WithdrawFlow{}
	if err := db.Order("id desc").Find(&vs).Error; err != nil && err != gorm.ErrRecordNotFound {
		glog.Error("TradeFlow_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}

	// v, err := dao.GetTotalWithdraw()
	// if err != nil {
	// 	glog.Error("TradeFlow_List fail! err=",err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended, "total": total, "info": info})
}

type checkParams struct {
	Id   int64 `json:"id"`
	Pass bool  `json:"pass"`
}

// 提币申请审核接口
func ChainFlow_Check(c *gin.Context) {
	maner, err := util.GetHeaderString(c, "X-Yf-Maner")
	if maner == "" || err != nil {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var req checkParams
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	status := common.TX_STATUS_NOTPASS
	if req.Pass {
		status = common.TX_STATUS_CHECKPASS
	}
	now := time.Now().Unix()
	db := db.GetTxDB(c)

	var v common.WithdrawFlow
	if err := db.Find(&v, "id=? and status in(?)", req.Id, []int{common.TX_STATUS_INIT, common.TX_STATUS_CHECKPASS}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_WITHDRAW_FORBIDDEN_MODIFY)
			return
		}
		glog.Error("ChainFlow_Check fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 审核通过 如是平台内地址互转 直接内盘操作
	if status == common.TX_STATUS_CHECKPASS && v.Channel == common.CHANNEL_CHAIN {
		if _, err := share.WalletTransfer(db, v); err != nil {
			glog.Error("Chain_Withdraw_Callback AssetLock save fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		} else {
			if err := db.Model(&v).Updates(map[string]interface{}{"status": common.TX_STATUS_SUCCESS, "maner": maner, "check_time": now, "update_time": now}).Error; err != nil {
				glog.Error("Chain_Withdraw_Callback update fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
		}
	} else {
		rowaffect := db.Model(&v).Updates(map[string]interface{}{"status": status, "maner": maner, "check_time": now, "update_time": now}).RowsAffected

		// 如果审核通过，则需调用提币接口
		if rowaffect == 1 {
			if status == common.TX_STATUS_NOTPASS { // 审核不通过 需要解冻用户提币冻结资产
				// 解冻相应资产id
				if err := db.Model(&common.AssetLock{}).Where("id=?", v.LockAssetId).Updates(map[string]interface{}{"lock_amount": 0, "update_time": now}).Error; err != nil {
					glog.Error("Chain_Withdraw_Callback AssetLock save fail! err=", err)
					yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
					return
				}
				dao.RefleshUserDayWidthDraw(v.UserId, v.TxType)
			}
		}
	}

	if status == 1 {
		// go PrepareDraw(req.Id)  // 放定时器里执行 避免并发问题
	}
	yf.JSON_Ok(c, gin.H{})
}

func withdraw(db *gorm.DB, id int64) (err error) {
	var v common.WithdrawFlow
	if err = db.Find(&v, "id=?", id).Error; err != nil {
		glog.Error("prepare_withdraw fail! err=", err)
		return
	}

	// 如果是审核状态 则调用提币接口
	if v.Status != common.TX_STATUS_CHECKPASS {
		glog.Error("prepare_withdraw fail! status is not TX_STATUS_WAITING")
		return
	}

	// 置等待提交状态
	v.Status = common.TX_STATUS_WAITING
	now := time.Now().Unix()
	if err = db.Model(&v).Updates(map[string]interface{}{"status": v.Status, "update_time": now}).Error; err != nil {
		glog.Error("prepare_withdraw fail! err=", err)
		return
	}

	// 手续费不用传
	req := util.WithDraw{Symbol: common.GetCurrencyName(v.TxType), UserId: v.UserId, OrderId: v.Id, Address: v.Address, Amount: v.Amount /*Fee:v.Fee*/}
	if err = util.ChainWithdrawWallet(req); err != nil {
		glog.Error("ChainWithdrawWallet fail! err=", err)
		return
	}
	return
}

func PrepareDraw(id int64) (err error) {

	db := db.GetTxDB(nil)
	err = withdraw(db.DB, id)
	if err != nil {
		glog.Error("PrepareDraw fail! err=", err)
		db.Rollback()
		return
	}
	db.Commit()
	return
}

type drawRetSt struct {
	Id      int64  `json:"id"`
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
}

// 提现结果设置
func ChainFlow_DrawSet(c *gin.Context) {
	country := util.GetCountry(c)
	if 1 != country {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var req drawRetSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	status := "failed"
	if req.Success {
		status = "success"
	}

	db := db.GetTxDB(c)
	var v common.WithdrawFlow
	if err := db.Find(&v, "id=? and channel>=?", req.Id, common.CHANNEL_ALIPAY).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Fail(c, common.ERR_ORDER_NOT_EXIST)
			return
		}
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	r := share.WithDrawSt{OrderId: req.Id, Reason: req.Reason, Status: status}
	if reason, err := share.WithDrawCallbackHandle(db, r); err != nil {
		if reason != "" {
			yf.JSON_Fail(c, reason)
		} else {
			yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		}
		glog.Error("Wallet_Withdraw_Callback fail! err=", err)
		return
	}
	yf.JSON_Ok(c, gin.H{})
	return
}
