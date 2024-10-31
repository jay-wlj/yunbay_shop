package share

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybapi/dao"
	"yunbay/ybapi/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func init() {

}

type Lotterys int64

type LotteryCh struct {
	notify chan struct{}
}

var g_lotterych *LotteryCh

var once sync.Once
var mu sync.Mutex

func GetLottery() (ch *LotteryCh) {
	once.Do(func() {
		g_lotterych = &LotteryCh{notify: make(chan struct{})}
	})

	return g_lotterych
}

// 有任务变更
func (t *LotteryCh) Notify() {
	t.notify <- struct{}{}
}

// 启动活动开始结束监控协程
func (t *LotteryCh) Start(exit chan struct{}) {
	db := db.GetDB()
	var err error
	for {
		var s_lotterys common.Lotterys
		var e_lotterys common.Lotterys
		s_duration := 72 * time.Hour

		// 查找最近一个待开始的计划
		if err = db.Select("id, start_time, end_time, status").Order("start_time asc").First(&s_lotterys, "status =0").Error; err != nil && err != gorm.ErrRecordNotFound {
			glog.Error("LotterysStart fail! err=", err)
			break
		}
		now := time.Now().Unix()
		var s_lotterys_id int64
		var e_lotterys_id int64
		// 查找最近开始和结束的时间
		if err != gorm.ErrRecordNotFound {
			s_duration = time.Duration(s_lotterys.StartTime-now) * time.Second
			s_lotterys_id = s_lotterys.Id
		}

		// 查找最近一个待结束的计划
		if err = db.Select("id, start_time, end_time, status").Order("end_time asc").First(&e_lotterys, "status =1").Error; err != nil && err != gorm.ErrRecordNotFound {
			glog.Error("LotterysStart fail! err=", err)
			break
		}
		e_duration := s_duration
		if err != gorm.ErrRecordNotFound {
			e_duration = time.Duration(e_lotterys.EndTime-now) * time.Second
			e_lotterys_id = e_lotterys.Id
		}

		select {
		case <-exit: // 退出
			break
		case <-t.notify: // 活动有变
		case <-time.After(s_duration): // 活动开始
			if s_lotterys_id > 0 {
				v := Lotterys(s_lotterys_id)
				if err = v.Begin(); err != nil {
					glog.Error("LotterysStart fail! err=", err)
					break
				}
			}

		case <-time.After(e_duration): // 活动结束
			if e_lotterys_id > 0 {
				v := Lotterys(e_lotterys_id)
				if err = v.Stop(); err != nil {
					glog.Error("LotterysStart fail! err=", err)
					break
				}
			}
		}
	}
	return
}

// 处理该计划下的所有记录是否都已上链
func (t Lotterys) HandleLotterys(db *db.PsqlDB) (err error) {
	// 查找该销售计划是否所有的交易记录都获取到了hash
	// 获取销售计划
	ly, e := (&dao.Lotterys{}).Get(int64(t))
	if err = e; err != nil {
		glog.Error("HandleLotterys fail! err=", err)
		return
	}
	// 参与人数没满
	if ly.Stock != ly.Sold {
		return
	}

	var total int
	if err = db.Model(&common.LotterysRecord{}).Where("lotterys_id=? and hash=''", t).Count(&total).Error; err != nil {
		glog.Error("HandleLotterys fail! err=", err)
		return
	}
	if 0 == total {
		// 颁奖逻辑 取numHash 从大到小前n个中奖
		vs := []common.LotterysRecord{}
		if err := db.Order("num_hash::numeric desc,create_time desc").Find(&vs, "lotterys_id=?", t).Error; err != nil {
			glog.Error("args")
		}

		sucess_ids := []int64{}
		for i := 0; i < ly.Num; i++ {
			sucess_ids = append(sucess_ids, vs[i].Id)
		}

		// 先全部设置为未中奖
		if err = db.Model(&common.LotterysRecord{}).Where("lotterys_id=?", t).Updates(base.Maps{"status": common.LOTTERYS_STATUS_NO}).Error; err != nil {
			glog.Error("HandleLotterys fail! err=", err)
			return
		}
		// 设置中奖的记录
		if err = db.Model(&common.LotterysRecord{}).Where("lotterys_id=? and id in(?)", t, sucess_ids).Updates(base.Maps{"status": common.LOTTERYS_STATUS_YES}).Error; err != nil {
			glog.Error("HandleLotterys fail! err=", err)
			return
		}

		// 发放ybt
		uids := []int64{}
		rs := []util.RewarSt{}
		for _, v := range vs {
			uids = append(uids, v.UserId)
			rs = append(rs, util.RewarSt{UserId: v.UserId, Amount: ly.RewardYbt})
		}

		if err = util.RewardYbt(rs); err != nil {
			glog.Error("HandleLotterys fail! err=", err)
			return
		}

		// 下发通知给所有参与抽奖的人
		db.AfterCommit(func() {
			// 刷新购买记录缓存
			cl := &dao.Lotterys{}
			cl.RefreshRecord(int64(t), 0) // 刷新抽奖下的购买记录缓存

			uids := base.UniqueInt64Slice(uids)
			for _, uid := range uids {
				cl.RefreshUserRecord(uid) // 刷新用户购买记录列表
			}
			t.notify_result()
		})
	}
	return
}

// 下发通知结果
func (t *Lotterys) notify_result() (err error) {
	id := int64(*t)
	db := db.GetDB()
	vs := []common.LotterysRecord{}
	if err = db.Select("user_id, status").Find(&vs, "lotterys_id=?", id).Error; err != nil {
		glog.Error("notify_result fail! err=", err)
		return
	}

	ok_uids := []int64{}   // 中奖者
	fail_uids := []int64{} // 未中奖者

	for _, v := range vs {
		if common.LOTTERYS_STATUS_YES == v.Status {
			ok_uids = append(ok_uids, v.UserId)
			continue
		}
		fail_uids = append(fail_uids, v.UserId)
	}
	ok_uids = base.UniqueInt64Slice(ok_uids)
	fail_uids = base.UniqueInt64Slice(fail_uids)

	if len(ok_uids) == 0 && len(fail_uids) == 0 {
		return
	}

	ch := &dao.Lotterys{}
	lott, e := ch.Get(id) // 获取抽奖信息
	if err = e; err != nil {
		glog.Error("notify_result fail! err=", err)
		return
	}
	p, e := util.GetProductInfo(lott.PId, 0) // 获取商品信息
	if err = e; err != nil {
		glog.Error("notify_result fail! err=", err)
		return
	}

	// 下发通知消息
	if len(ok_uids) > 0 {
		notify := &util.LotterysNotify{}
		if err = notify.NotifyRetMsg(id, common.LOTTERYS_STATUS_YES, p.Title, ok_uids); err != nil {
			glog.Error("notify_result fail! ok_uids=", ok_uids, " err=", err)
		}
	}

	return
}

// 计划开始
func (t Lotterys) Begin() (err error) {
	db := db.GetDB()
	if err = db.Model(&common.Lotterys{}).Where("id=? and status=?", t, common.ACTIVITY_STATUS_INIT).Updates(base.Maps{"status": common.ACTIVITY_STATUS_RUNNING}).Error; err != nil {
		glog.Error("Lotterys Stop fail! err=", err)
		return
	}
	(&dao.Lotterys{}).Refresh(int64(t)) // 刷新缓存

	return
}

func (t Lotterys) _stop(db *db.PsqlDB) (err error) {
	v, e := (&dao.Lotterys{}).Get(int64(t))
	if err = e; err != nil {
		glog.Error("Lotterys Stop fail! err=", err)
		return
	}

	status := common.ACTIVITY_STATUS_END
	// if v.Sold < v.Stock {
	// 	status = common.ACTIVITY_STATUS_FAIL // 活动参与次数不够 失效处理
	// }
	if err = db.Model(&common.Lotterys{}).Where("id=? and status=?", t, common.ACTIVITY_STATUS_RUNNING).Updates(base.Maps{"status": status}).Error; err != nil {
		glog.Error("Lotterys Stop fail! err=", err)
		return
	}

	if v.Stock == v.Sold {

	} else {
		// 设置为已退还状态
		if err = db.Model(&common.LotterysRecord{}).Where("lotterys_id=?", t).Updates(base.Maps{"status": common.LOTTERYS_STATUS_FAIL}).Error; err != nil {
			glog.Error("HandleLotterys fail! err=", err)
			return
		}
		// 需要装此计划中所有支付记录退回
		key := fmt.Sprintf("lotterys_%v", t) // 设置支付key
		if err = util.YBAsset_LotterysRefund(key); err != nil {
			glog.Error("HandleLotterys fail! YBAsset_LotterysRefund err=", err)
			return
		}
	}
	return
}

// 计划已结束
func (t Lotterys) Stop() (err error) {
	db := db.GetTxDB(nil)
	if err = t._stop(db); err != nil {
		db.Rollback()
		glog.Error("Lotterys Stop faiL! err=", err)
	}
	db.Commit()
	(&dao.Lotterys{}).Refresh(int64(t)) // 刷新缓存
	return
}

// 该计划已失效 退回相应资产
func (t Lotterys) Refund() (err error) {
	db := db.GetTxDB(nil)
	// 设置为已退还状态
	if err = db.Model(&common.LotterysRecord{}).Where("lotterys_id=?", t).Updates(base.Maps{"status": common.LOTTERYS_STATUS_FAIL}).Error; err != nil {
		glog.Error("HandleLotterys fail! err=", err)
		db.Rollback()
		return
	}
	// 确定此销售记录的所有订单资产状态
	if err = util.YBAsset_PayStatus(util.YBAssetStatus{PublishArea: common.PUBLISH_AREA_LOTTERYS, OrderIds: []int64{int64(t)}, Status: common.STATUS_FAIL}); err != nil {
		glog.Error("Refund fail! err=", err)
		db.Rollback()
		return
	}
	db.Commit()
	(&dao.Lotterys{}).Refresh(int64(t)) // 刷新缓存
	return
}

type LotterysPayParmas struct {
	LotteryId int64  `json:"lotterys_id" binding:"gt=0"`
	Memo      string `json:"memo"`
	util.TransferPaySt
}

// 抽奖支付逻辑
func (t *Lotterys) Pay(db *db.PsqlDB, req LotterysPayParmas) (lotterys_record_id int64, err error) {
	err = fmt.Errorf("ERR_SERVER_ERROR")
	// 是否可售状态
	var ld dao.Lotterys
	v, e := ld.Get(req.LotteryId)
	if err = e; err != nil {
		glog.Error("LotterysPay fail! err=", err)
		return
	}

	// 非可售状态
	switch v.Status {
	case common.ACTIVITY_STATUS_END:
		err = errors.New(common.ERR_LOTTERYS_OVER)
		return
	case common.ACTIVITY_STATUS_INIT:
		err = errors.New(common.ERR_LOTTERYS_NOSTART)
		return
	}

	// 判断订单金额是否一致
	if v.Coin != req.CoinType || !v.Amount.Equal(req.Amount) {
		glog.Errorf("pay coin:%v amount:%v order coin:%v amount:%v", req.CoinType, req.Amount, v.Coin, v.Amount)
		reason := common.ERR_AMOUNT_INVALID
		err = fmt.Errorf(reason)
		return
	}

	// 判断限制参与次数
	if v.Pertimes > 0 {
		var count int
		if err = db.Model(&common.LotterysRecord{}).Where("lotterys_id=? and  user_id=?", req.LotteryId, req.From).Count(&count).Error; err != nil {
			return
		}
		// 超过抽奖次数 TODO 优化 放缓存处理
		if count >= v.Pertimes {
			err = errors.New(common.ERR_EXCEED_TIMES)
			return
		}
	}
	// 先更新数量
	res := db.Model(&common.Lotterys{}).Where("id=? and status=? and sold<stock", req.LotteryId, common.ACTIVITY_STATUS_RUNNING).Updates(base.Maps{"sold": gorm.Expr("sold+1")})
	if err = res.Error; err != nil {
		glog.Error("LotterysPay fail! err=", err)
		return
	}
	if 0 == res.RowsAffected {
		//err = errors.New(common.ERR_)
		err = errors.New(common.ERR_LOTTERYS_OVER)
		return
	}

	// 保存记录
	r := common.LotterysRecord{LotterysId: req.LotteryId, UserId: req.From, Amount: req.Amount, Memo: req.Memo}
	if err = db.Save(&r).Error; err != nil {
		glog.Error("LotterysPay fail! err=", err)
		return
	}

	// 已全部售完,需设置状态
	res = db.Model(&common.Lotterys{}).Where("id=? and status=? and sold=stock", req.LotteryId, common.ACTIVITY_STATUS_RUNNING).Updates(base.Maps{"status": common.ACTIVITY_STATUS_END})
	if err = res.Error; err != nil {
		glog.Error("LotterysPay fail! err=", err)
		return
	}

	// 添加平台交易资金池记录
	req.Key = fmt.Sprintf("lotterys_%v", req.LotteryId) // 设置支付key
	err = util.YBAsset_LotterysPay(req.TransferPaySt)

	if err != nil {
		// 支付失败 需相应减少已售数量
		return
	}

	lotterys_record_id = r.Id
	// 提交后刷新订单相关缓存
	db.AfterCommit(func() {
		// 添加到唯一key缓存中
		dao.SaveUniqueKey(req.Memo)

		ld.Refresh(req.LotteryId)                                      // 刷新活动参与人数记录
		ld.RefreshRecord(req.LotteryId, req.From)                      // 刷新购买记录缓存
		util.AsynGenerateEosLotterysHash(lotterys_record_id, req.Memo) // 购买记录上链处理
		if res.RowsAffected > 0 {
			//	 异步处理
		}
		if v, e := ld.Get(req.LotteryId); e == nil {
			notify := &util.LotterysNotify{}
			notify.NotifyStatus(req.LotteryId, v.Sold, v.Status) // 通知前端更新该活动状态及已售量
		}

	})

	err = nil // 成功
	return
}
