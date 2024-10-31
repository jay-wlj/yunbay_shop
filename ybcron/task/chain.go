package task

import (
	"fmt"
	"strings"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybcron/util"

	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

var m_drawstatus map[string]int

func init() {
	m_drawstatus = make(map[string]int)
	m_drawstatus["waiting"] = common.TX_STATUS_WAITING
	m_drawstatus["submitted"] = common.TX_STATUS_SUBMIT
	m_drawstatus["confirming"] = common.TX_STATUS_CONFIRM
	m_drawstatus["failed"] = common.TX_STATUS_FAILED
	m_drawstatus["success"] = common.TX_STATUS_SUCCESS
}

func getstatustxtbyint(status int) string {
	for k, v := range m_drawstatus {
		if v == status {
			return k
		}
	}
	return ""
}

// 调用区块接口进行提币操作
func Chain_WithDraw() {
	var count int
	var err error
	for {
		db := db.GetTxDB(nil)
		err = withdraw(db.DB)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				glog.Error("Chain_WithDraw fail! err=", err)
			}
			db.Rollback()
			return
		}
		db.Commit()
		count += 1
		// 一次定时调用最多50次调用提交
		if count > 50 {
			break
		}
	}
}

func withdraw(db *gorm.DB) (err error) {
	// 查询一个审核通过还未进行提交的申请
	var v common.WithdrawFlow
	if err = db.First(&v, "status=? and channel<?", common.TX_STATUS_CHECKPASS, common.CHANNEL_ALIPAY).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			glog.Error("withdraw fail! err=", err)
		}
		return
	}
	// 置等待提交状态
	v.Status = common.TX_STATUS_WAITING

	// 走平台与合作平台互转
	if v.FlowType == common.FLOW_TYPE_YUNBAY {
		switch v.Channel {
		case common.CHANNEL_CHAIN:
			if err = util.WithdrawCheck(v.Id, true); err != nil {
				glog.Error("withdraw fail! update withdarw_flow error! err=", err)
				return
			}
			break
		case common.CHANNEL_HOTCOIN:

			// 标记走合约转帐
			if err = db.Model(&v).Updates(map[string]interface{}{"flow_type": common.FLOW_TYPE_CHAIN, "update_time": time.Now().Unix()}).Error; err != nil {
				glog.Error("withdraw fail! update withdarw_flow error! id:", v.Id)
				return
			}

			// v.Status = common.TX_STATUS_SUCCESS
			// if err = withdraw_hotcoin(&v); err != nil {
			// 	v.Reason = err.Error()
			// 	if v.Reason == "INSUFFICIENT_FUNDS" {
			// 		// 余额不够 标记走合约转帐
			// 		if err = db.Model(&v).Updates(map[string]interface{}{"flow_type": common.FLOW_TYPE_CHAIN, "update_time": time.Now().Unix()}).Error; err != nil {
			// 			glog.Error("withdraw fail! update withdarw_flow error! id:", v.Id)
			// 			return
			// 		}
			// 		glog.Info("withdraw id:", v.Id, " amount:", v.Amount, v.Reason, " turn FLOW_TYPE_YUNBAY to FLOW_TYPE_CHAIN")
			// 		return
			// 	}
			// 	if v.Reason != "ORDER_ID_EXIST" {
			// 		v.Status = common.TX_STATUS_FAILED
			// 	}
			// }

			// draw := util.WithDrawSt{OrderId: v.Id, TxHash: v.Txhash, Status: getstatustxtbyint(v.Status), Reason: v.Reason, Channel: v.Channel}
			// if err = util.WithdrawCallback(draw); err != nil {
			// 	glog.Error("withdraw fail! err=", err)
			// 	return
			// }

			break
		case common.CHANNEL_YUNEX:
			// 标记走合约转帐
			if err = db.Model(&v).Updates(map[string]interface{}{"flow_type": common.FLOW_TYPE_CHAIN, "update_time": time.Now().Unix()}).Error; err != nil {
				glog.Error("withdraw fail! update withdarw_flow error! id:", v.Id)
				return
			}
			// v.Status = common.TX_STATUS_SUCCESS
			// if err = withdraw_yunex(db, &v); err != nil {
			// 	v.Reason = err.Error()
			// 	if v.Reason == "BALANCE_NOT_ENOUGH" {
			// 		// 余额不够 标记走合约转帐
			// 		if err = db.Model(&v).Updates(map[string]interface{}{"flow_type": common.FLOW_TYPE_CHAIN, "update_time": time.Now().Unix()}).Error; err != nil {
			// 			glog.Error("withdraw fail! update withdarw_flow error! id:", v.Id)
			// 			return
			// 		}
			// 		glog.Info("withdraw id:",v.Id, " amount:", v.Amount, v.Reason, " turn FLOW_TYPE_YUNBAY to FLOW_TYPE_CHAIN")
			// 		return
			// 	}
			// 	if v.Reason != "ORDER_ID_EXIST" {
			// 		v.Status = common.TX_STATUS_FAILED
			// 	}
			// }

			// draw := util.WithDrawSt{OrderId:v.Id, TxHash:v.Txhash, Status:getstatustxtbyint(v.Status), Reason:v.Reason, Channel:v.Channel}
			// if err = util.WithdrawCallback(draw); err != nil {
			// 	glog.Error("withdraw fail! err=", err)
			// 	return
			// }

			break
			// case common.CHANNEL_ALIPAY:		// ali目前不支持接口转帐操作
			// 	v.Status = common.TX_STATUS_SUCCESS
			// 	if err = withdraw_alipay(&v); err != nil {
			// 		v.Reason = err.Error()

			// 		v.Status = common.TX_STATUS_FAILED
			// 	}

			// 	draw := util.WithDrawSt{OrderId:v.Id, TxHash:v.Txhash, Status:getstatustxtbyint(v.Status), Reason:v.Reason, Channel:v.Channel}
			// 	if err = util.WithdrawCallback(draw); err != nil {
			// 		glog.Error("withdraw fail! err=", err)
			// 		return
			// 	}
		}
	} else {
		// 走合约转帐 手续费不用传
		v.Status = common.TX_STATUS_SUCCESS
		if err = withdraw_chain(&v); err != nil {
			v.Reason = err.Error()
			if v.Reason != "ORDER_ID_EXIST" {
				v.Status = common.TX_STATUS_FAILED
			}
		}

		draw := util.WithDrawSt{OrderId: v.Id, TxHash: v.Txhash, Status: getstatustxtbyint(v.Status), Reason: v.Reason, Channel: v.Channel}
		if err = util.WithdrawCallback(draw); err != nil {
			glog.Error("withdraw fail! err=", err)
			return
		}
	}

	//glog.Infof("withdraw success id=", v.Id)
	return
}

func withdraw_chain(v *common.WithdrawFlow) (err error) {
	// 这里需要事务处理 先修改状态为等待提交 以防多次调用第三方接口
	tx := db.GetTxDB(nil)
	if err = tx.Model(&v).Updates(map[string]interface{}{"status": common.TX_STATUS_WAITING, "update_time": time.Now().Unix()}).Error; err != nil {
		glog.Error("prepare_withdraw fail! err=", err)
		tx.Rollback()
		return
	}

	req := util.WithDraw{UserId: v.UserId, OrderId: v.Id, Address: v.Address, Amount: v.Amount, Symbol: common.GetCurrencyName(v.TxType) /*Fee:v.Fee*/}
	if err = util.ChainWithdrawWallet(req); err != nil {
		glog.Error("ChainWithdrawWallet fail! err=", err)
		tx.Rollback()
		return
	}
	tx.Commit()

	return
}

func withdraw_hotcoin(v *common.WithdrawFlow) (err error) {
	if v.Channel != common.CHANNEL_HOTCOIN {
		return
	}

	// 这里需要事务处理 先修改状态为等待提交 以防多次调用第三方接口
	tx := db.GetTxDB(nil)
	if err = tx.Model(&v).Updates(map[string]interface{}{"status": common.TX_STATUS_WAITING, "update_time": time.Now().Unix()}).Error; err != nil {
		glog.Error("prepare_withdraw fail! err=", err)
		tx.Rollback()
		return
	}
	// 调用热币提币接口
	m := util.HotCoinWithDraw{Coin: "KT", Address: v.Address, OrderId: fmt.Sprintf("%v", v.Id), Amount: v.Amount, Platform: "YunBay"}
	var txhash string
	txhash, err = util.HotCoinWithdrawWallet(m)
	if err != nil {
		tx.Rollback()
		glog.Error("withdraw_hotcoin fail! err=", err)
		return
	}
	tx.Commit()

	if err == nil {
		v.Status = common.TX_STATUS_SUCCESS
	}
	v.Txhash = txhash
	return
}

func withdraw_yunex(v *common.WithdrawFlow) (err error) {
	if v.Channel != common.CHANNEL_YUNEX {
		return
	}
	coin := strings.ToUpper(common.GetCurrencyName(v.TxType))

	// 这里需要事务处理 先修改状态为等待提交 已以防多次调用第三方接口
	tx := db.GetTxDB(nil)
	if err = tx.Model(&v).Updates(map[string]interface{}{"status": common.TX_STATUS_WAITING, "update_time": time.Now().Unix()}).Error; err != nil {
		glog.Error("prepare_withdraw fail! err=", err)
		tx.Rollback()
		return
	}

	m := util.YunexWithDraw{Plat: "yunbay", Coin: coin, Address: v.Address, OrderId: fmt.Sprintf("%v", v.Id), Amount: v.Amount}
	ret, err1 := util.YunexWithdrawWallet(m)
	if err1 != nil {
		err = err1
		tx.Rollback()
		glog.Error("withdraw_yunex fail! err=", err)
		return
	}
	tx.Commit()

	switch ret.Status {
	case 1:
		v.Status = common.TX_STATUS_SUCCESS // 直接到帐
	case 3:
		v.Status = common.TX_STATUS_CONFIRM // 审核中
	default:
		v.Status = common.TX_STATUS_FAILED // 其它视为失败
		glog.Error("withdraw_yunex fail! ret.Status=", ret.Status)
	}

	v.Reason = ret.Reason
	v.Txhash = ret.TxHash
	return
}

// 调用区块查询接口进行确认操作
func Chain_WithDrawQuery() {
	db := db.GetDB()
	if err := withdrawquery(db.DB); err != nil {
		glog.Error("Chain_WithDraw fail! err=", err)
		return
	}
}

func withdrawquery(db *gorm.DB) (err error) {
	// 查询一个审核通过还未进行提交的申请
	var vs []common.WithdrawFlow
	if err = db.Limit(50).Order("create_time asc").Find(&vs, "flow_type=? and status>=? and status<?", common.FLOW_TYPE_CHAIN, common.TX_STATUS_WAITING, common.TX_STATUS_FAILED).Error; err != nil {
		return
	}
	for _, v := range vs {
		// 查询该笔提取状态
		var ret util.TxStatus
		if ret, err = util.ChainTxQuery(v.Id); err != nil {
			glog.Error("withdrawquery fail! err=", err)
			return
		}
		status := m_drawstatus[ret.Status]
		if v.Status != status { // 状态有变化才去通知
			draw := util.WithDrawSt{OrderId: v.Id, TxHash: ret.TxHash, Status: ret.Status, FeeInEther: ret.FeeInEther, Reason: ret.Reason}
			if err = util.WithdrawCallback(draw); err != nil {
				glog.Error("withdrawquery fail! err=", err)
				return
			}
		}
	}

	return
}

func withdraw_alipay(v *common.WithdrawFlow) (err error) {
	if v.Channel != common.CHANNEL_ALIPAY && v.TxType != common.CURRENCY_KT {
		return
	}

	// 这里需要事务处理 先修改状态为等待提交 已以防多次调用第三方接口
	tx := db.GetTxDB(nil)
	if err = tx.Model(&v).Updates(map[string]interface{}{"status": common.TX_STATUS_WAITING, "update_time": time.Now().Unix()}).Error; err != nil {
		glog.Error("prepare_withdraw fail! err=", err)
		tx.Rollback()
		return
	}

	//m := util.YunexWithDraw{Plat:"yunbay", Coin:coin, Address:v.Address, OrderId:fmt.Sprintf("%v", v.Id), Amount:v.Amount}
	ret, err1 := util.Ali_Withdraw(v.Id, v.Amount, v.Address)
	if err1 != nil {
		err = err1
		tx.Rollback()
		glog.Error("withdraw_yunex fail! err=", err)
		return
	}
	tx.Commit()

	v.Status = m_drawstatus[ret.Status]
	v.Reason = ret.Reason
	v.Txhash = ret.TxHash
	v.FeeInEther = ret.FeeInEther

	return

}
