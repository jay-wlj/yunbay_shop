package share

import (
	"yunbay/ybpay/common"
	"yunbay/ybpay/conf"
	"yunbay/ybpay/util"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

const (
	YBPAY_PAY_URL string = "ybpay_url"
)

type OrderSt struct {
	Id          int64           `json:"-"`
	Channel     int             `json:"-"`
	OrderIds    []int64         `json:"order_ids"`
	Subject     string          `json:"subject"`
	Amount      decimal.Decimal `json:"amount"`
	ProductCode string          `json:"product_code"`
	UserId      int64           `json:"user_id"`
	OverTime    int64
	RemoteIp    string
}

// 统一下单
func (req *OrderSt) CreatePay(db *db.PsqlDB) (sign, reason string, err error) {
	user_id := req.UserId

	// 查看交易id记录是否存在
	var r common.RmbRecharge
	if len(req.OrderIds) == 1 {
		if err = db.Find(&r, "user_id=? and channel=? and status>=? and amount=? and order_ids[1]=?", user_id, req.Channel, common.STATUS_INIT, req.Amount, req.OrderIds[0]).Error; err != nil && err != gorm.ErrRecordNotFound {
			glog.Error("CreateAliPay fail! err=", err)
			return
		}
	}

	if r.Status == common.STATUS_OK {
		err = fmt.Errorf(yf.ERR_SERVER_ERROR)
		reason = common.ERR_ORDER_HASPAYED
		return
	}

	var duration time.Duration
	str_over_time := conf.Config.Server.Ext["order_over_time"].(string)
	duration, _ = time.ParseDuration(str_over_time)
	req.OverTime = time.Now().Unix() + int64(duration.Seconds())

	if 0 == r.Id {
		// 设置订单锁定时间(待支付到已支付时间,到期需自动关闭)
		if len(req.OrderIds) > 0 {
			// 查询订单标题
			vs, err1 := util.GetOrderByIds(req.OrderIds)
			if err1 != nil {
				err = err1
				glog.Error("CreateAliPay fail! GetOrderByIds err=", err)
				return
			}
			titles := []string{}
			var amount decimal.Decimal
			for _, v := range vs {
				if t, ok := v.Product["title"]; ok {
					if title, ok := t.(string); ok {
						titles = append(titles, title)
					}
				}
				amount = amount.Add(v.TotalAmount)
				if v.AutoCancelTime < req.OverTime {
					req.OverTime = v.AutoCancelTime
				}
			}
			// 效验充值金额
			//if !base.IsEqual(req.Amount, amount) {
			if !req.Amount.Equal(amount) {
				reason = common.ERR_AMOUNT_INVALID
				err = fmt.Errorf(reason)
				glog.Error("CreateAliPay req.Amount:", req.Amount, " amount:", amount, " cmp=", req.Amount.Cmp(amount))
				return
			}
			req.Subject = base.StringSliceToString(titles, ",")
			if len(req.Subject) > 255 {
				req.Subject = req.Subject[:255]
			}
		}
		if req.Subject == "" {
			req.Subject = "云贝商城"
		}
		now := time.Now().Unix()
		r = common.RmbRecharge{Channel: req.Channel, UserId: user_id, OrderIds: req.OrderIds, Subject: req.Subject, TxType: common.CURRENCY_RMB, Amount: req.Amount, OverTime: req.OverTime, CreateTime: now, UpdateTime: now}
		if err = db.Save(&r).Error; err != nil {
			glog.Error("CreateAliPay fail! err=", err)
			return
		}
	} else {
		// 从缓存中取出签名串
		sign, err = get_pay_url_cache(r.Id)
		if err == nil {
			return
		}
	}

	// 订单超时
	// if r.OverTime <= time.Now().Unix() {
	// 	reason = common.ERR_ORDER_NOT_EXIST
	// 	err = fmt.Errorf(reason)
	// 	glog.Error("pay order", r.Id, " over time")
	// 	return
	// }
	req.Id = r.Id
	req.Subject = r.Subject

	switch req.Channel {
	case common.CHANNEL_ALIPAY:
		sign, reason, err = GetAliPay().TradeAppPay(req)
		if err != nil {
			glog.Error("Alipay_pay faiL! err=", err)
			return
		}
	case common.CHANNEL_WEIXIN:
		sign, reason, err = GetWeixin().TradeAppPay(req)
		if err != nil {
			glog.Error("GetWeixin faiL! err=", err)
			return
		}
	}

	set_pay_url_cache(r.Id, sign, duration)
	return
}

func get_pay_url_cache(id int64) (string, error) {
	cache, e := cache.GetReader("asset")
	if e != nil {
		return "", e
	}
	return cache.HGet(YBPAY_PAY_URL, strconv.FormatInt(id, 10))
}

func set_pay_url_cache(id int64, pay_url string, exptime time.Duration) error {
	cache, e := cache.GetWriter("asset")
	if e != nil {
		return e
	}
	return cache.HSet(YBPAY_PAY_URL, strconv.FormatInt(id, 10), pay_url, exptime)
}

type TxParms struct {
	Channel int
	Status  int    `json:"status"`
	Reason  string `json:"reason"`
	TradeId string `json:"trade_id"`
	TxHash  string `json:"tx_hash"`
	Account string `json:"account"`
	Amount  string `json:"amount"`
}

func (t *TxParms) UpdateRmbRecharge() (err error) {
	trade_id, _ := base.StringToInt64(t.TradeId)
	// txHash := noti.TradeNo
	// account := noti.SellerId
	now := time.Now().Unix()
	amount, _ := decimal.NewFromString(t.Amount)

	if t.Channel == common.CHANNEL_WEIXIN {
		amount = amount.Div(decimal.New(int64(100), 0))
	}
	var v common.RmbRecharge
	db := db.GetDB()

	if err = db.Find(&v, "id=? and status=?", trade_id, common.STATUS_INIT).Error; err != nil {
		glog.Error("updateRmbRecharge id=", trade_id, " err=", err)
		return
	}
	//if !base.IsEqual(v.Amount, amount) {
	if !v.Amount.Equal(amount) {
		s := fmt.Sprintln("updateRmbRecharge fail! req.TotalAmount=", t.Amount, " origin amount:", v.Amount)
		glog.Error(s)
		err = fmt.Errorf(s)
		return
	}

	// 事务更新
	res := db.Model(&common.RmbRecharge{}).Where("id=? and status=?", trade_id, common.STATUS_INIT).Updates(map[string]interface{}{"status": t.Status, "txhash": t.TxHash, "address": t.Account, "update_time": now})
	if err = res.Error; err != nil {
		glog.Error("updateRmbRecharge fail! err=", err)
		return
	}

	// 交易成功
	if t.Status == common.STATUS_OK {
		// 验证其它参数
		if res.RowsAffected > 0 {
			// 通知转帐相应的kt给该用户
			go RechargeNotfiyAsset(v.Id)
			// if er := util.RechargeNotify(v.Id); er != nil {
			// 	// 改用异步调用
			// 	if err1 := util.AsyncRechargeNotify(v.Id); err1 != nil {
			// 		util.SendDingTextTalkToMe(fmt.Sprintf("Alipay_Notify fail! RechargeNotify id=%v", v.Id))
			// 		glog.Error("updateRmbRecharge AsyncRechargeNotify fail! err=", err, " id=", v.Id)
			// 		return
			// 	}
			// }
		}
	}

	return
}

func RechargeNotfiyAsset(id int64) {
	if er := util.RechargeNotify(id); er != nil {
		// 改用异步调用
		if err1 := util.AsyncRechargeNotify(id); err1 != nil {
			util.SendDingTextTalkToMe(fmt.Sprintf("Alipay_Notify fail! RechargeNotify id=%v", id))
			glog.Error("updateRmbRecharge AsyncRechargeNotify fail! err=", err1, " id=", id)
			return
		}
	}
}

type RefundNoti struct {
	common.RmbRefund
}

func (t *RefundNoti) Notify() (err error) {
	var v common.RmbRefund
	db := db.GetDB()

	if err = db.Find(&v, "id=? and status=?", t.Id, common.STATUS_INIT).Error; err != nil {
		glog.Error("updateRmbRecharge id=", t.Id, " err=", err)
		return
	}
	if !v.TotalFee.Equal(t.TotalFee) {
		s := fmt.Sprintln("updateRmbRecharge fail! req.TotalAmount=", t.TotalFee, " origin amount:", v.TotalFee)
		glog.Error(s)
		err = fmt.Errorf(s)
		return
	}

	// 事务更新
	if err = db.Save(t).Error; err != nil {
		glog.Error("RefundNoti notify fail! err=", err)
		return
	}

	// 交易成功
	if t.Status == common.STATUS_OK {
		// 验证其它参数

	}

	return
}
