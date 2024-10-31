package share

import (
	"errors"
	"fmt"
	"yunbay/yborder/common"
	"yunbay/yborder/dao"
	"yunbay/yborder/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
)

// 欧飞订单号结果通知
func OfRetNotify(ydb *db.PsqlDB, v *common.OfOrder) (err error) {

	defer func() {
		if err != nil {
			util.SendDingTextTalkToMe(err.Error())
		}
	}()

	GameCode := v.GameState
	order_id := v.OrderId
	err_msg := v.ErrMsg

	switch v.GameState {
	case -1: // -1 表示查不到此订单，此时不能作为失败处理该订单，需要联系欧飞人工核实
		err = errors.New(fmt.Sprintf("开放商城欧飞充值查询不到此订单号:%v", order_id))
		glog.Error("OfRetNotify retCode:", GameCode, " order_id:", order_id)

		return
	case 0: // 充值中
		return
	}

	if ydb == nil {
		ydb = db.GetTxDB(nil)
	}

	defer func() {
		if err != nil {
			glog.Error("OfRetNotify fail! err=", err)
			ydb.Rollback()
		}
	}()

	res := ydb.Model(&common.OfOrder{}).Where("order_id=? and game_state=?", order_id, common.STATUS_INIT).Updates(base.Maps{"game_state": GameCode, "reason": err_msg})
	if err = res.Error; err != nil {
		return
	}

	if 0 == res.RowsAffected {
		ydb.Rollback()
		return
	}

	status := 0

	switch GameCode {
	case util.GAME_STATE_FAIL:
		// 如果充值失败 则需要退款给用户
		o := Order{}
		if err = o.RefundById(ydb, order_id); err != nil {
			glog.Error("Orders_AutoDeiver fail! Refund err=", err)
			return
		}
	case util.GAME_STATE_OK:
		status = 1
		// 有卡密信息保存
		if len(v.Cards) > 0 {
			// data, _ := json.Marshal(v.Cards)
			// if err = ydb.Model(&common.Orders{}).Where("id=?", order_id).Updates(base.Maps{"extinfos": gorm.Expr(`jsonb_set(extinfos, '{card}', ?::jsonb)`, string(data))}).Error; err != nil {
			// 	glog.Error("OfRetNotify fail! err=", err)
			// 	ydb.Rollback()
			// 	return
			// }
			for i := range v.Cards {
				v.Cards[i].OrderId = order_id
				if err = ydb.Save(&v.Cards[i]).Error; err != nil {
					glog.Error("OfRetNotify fail! err=", err)
					ydb.Rollback()
					return
				}
			}
		}

		// 充值成功
		// 完成此笔订单
		v := util.YBAssetStatus{OrderIds: []int64{order_id}, Status: common.ASSET_POOL_FINISH}
		mq := common.MQUrl{Methond: "post", Uri: "/man/asset/payset", AppKey: "ybasset", Data: v, MaxTrys: -1}
		if err = util.PublishMsg(mq); err != nil {
			glog.Error("Orders_AutoDeiver fail! err=", err)
			return
		}
	}

	// 推送消息
	ydb.AfterCommit(func() {
		noti := util.Notify{}

		// 获取订单买家id TODO 有待优化
		o := &dao.Orders{}
		var user_id int64
		if user_id, _, err = o.GetUserById(order_id); err != nil {
			glog.Error("Orders_AutoDeiver fail! err=", err)
			return
		}
		if err = noti.NotifyOfCardStatus(order_id, user_id, v.Cards, status); err != nil {
			glog.Error("Orders_AutoDeiver fail! err=", err)
			return
		}
	})

	ydb.Commit()

	return
}
