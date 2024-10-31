package man

import (
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"strings"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"
	"yunbay/ybasset/dao"
	"yunbay/ybasset/util"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type payConfirm struct {
	UserId   int64         `json:"user_id"`
	OrderIds pq.Int64Array `json:"order_ids"`
	Amount   float64       `json:"amount" binding:"required"`
}

// 支付逻辑
func pay(db *db.PsqlDB, user_id int64, pools []common.YBAssetPool) (reason string, err error) {
	err = fmt.Errorf("ERR_SERVER_ERROR")
	reason = yf.ERR_SERVER_ERROR

	// 验证用户帐户是否冻结状态,否则禁止提现和交易
	var lock bool = false
	if lock, err = dao.UserAssetIsLocked(user_id); err != nil || lock {
		if lock {
			reason = common.ERR_USERASSET_LOCK
			err = fmt.Errorf("ERR_SERVER_ERROR")
			return
		} else {
			glog.Error("Asset_Pay fail! err=", err)
			return
		}
		return
	}
	var userasset common.UserAsset
	if userasset, err = dao.GetUserAsset(user_id, db.DB); err != nil {
		glog.Error("Asset_Pay fail! err=", err)
		return
	}
	mAmounts := make(map[int]float64)
	for _, v := range pools {
		mAmounts[v.CurrencyType] += v.PayAmount
	}
	for k, v := range mAmounts {
		switch k {
		case common.CURRENCY_YBT:
			if userasset.NormalYbt < v {
				glog.Errorf("money not more amount normal_ybt:%v order amount:%v", userasset.NormalYbt, v)
				reason = common.ERR_MONEY_NOT_MORE
				err = fmt.Errorf(reason)
				return
			}
		case common.CURRENCY_KT:
			if userasset.NormalKt < v {
				glog.Errorf("money not more amount normal_kt:%v order amount:%v", userasset.NormalKt, v)
				//yf.JSON_FailEx(c, common.ERR_MONEY_NOT_MORE, gin.H{"amount":userasset.NormalKt})
				reason = common.ERR_MONEY_NOT_MORE
				err = fmt.Errorf(reason)
				return
			}
		case common.CURRENCY_SNET:
			if userasset.NormalSnet < v {
				glog.Errorf("money not more amount normal_snet:%v order amount:%v", userasset.NormalSnet, v)
				//yf.JSON_FailEx(c, common.ERR_MONEY_NOT_MORE, gin.H{"amount":userasset.NormalKt})
				reason = common.ERR_MONEY_NOT_MORE
				err = fmt.Errorf(reason)
				return
			}
		default:
			glog.Error("Asset_Pay fail!")
			reason = yf.ERR_SERVER_ERROR
			err = fmt.Errorf(reason)
			return
		}
	}

	today := time.Now().Format("2006-01-02")
	bs := []util.BusinessSt{}
	us := []common.UserAssetDetail{}
	for _, v := range pools {
		if v.OrderId == 0 || v.PayerUserId == 0 || !base.IsEqual(v.PayAmount, v.SellerAmount+v.RebatAmount) {
			glog.Error("Asset_Pay args invalid! v=", v)
			//yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
			reason = yf.ERR_ARGS_INVALID
			err = fmt.Errorf(reason)
			return
		}

		now := time.Now().Unix()
		v.Id = 0
		v.Date = time.Now().Format("2006-01-02")
		v.Status = 0
		v.CreateTime = now
		v.UpdateTime = now

		// 商家的交易额只算kt
		if v.PublishArea != common.PUBLISH_AREA_REBAT { // 折扣专区 在折扣出来后再进行商家交易额更新
			bs = append(bs, util.BusinessSt{UserId: v.SellerUserId, Type: v.CurrencyType, Amount: v.PayAmount, Rebat: v.RebatAmount})
		}

		// 添加平台交易资金池记录
		if err = db.Create(&v).Error; err != nil {
			glog.Error("Asset_Pay fail! err=", err)
			if strings.Index(err.Error(), "cannot be negtive") > 0 {
				reason = common.ERR_MONEY_NOT_MORE
				err = fmt.Errorf(reason)
			}
			//yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}

		// ybt的资金划入公共帐号
		var system_acount int64 = conf.Config.SystemAccounts["zone_account"]
		switch v.CurrencyType {
		case common.CURRENCY_YBT:
			// 扣除用户帐户的ybt到公共帐号里
			us = append(us, common.UserAssetDetail{UserId: v.PayerUserId, Type: v.CurrencyType, TransactionType: common.YBT_TRANSACTION_BUY, Amount: -v.PayAmount, Date: today})
			us = append(us, common.UserAssetDetail{UserId: system_acount, Type: v.CurrencyType, TransactionType: common.YBT_TRANSACTION_BUY, Amount: v.PayAmount, Date: today})
		case common.CURRENCY_SNET:
			// 扣除用户的snet到公共帐号里
			system_acount = conf.Config.SystemAccounts["snet_zone_account"]
			if common.PUBLISH_AREA_LOTTERYS == v.PublishArea {
				// 积分抽奖帐号
				system_acount = conf.Config.SystemAccounts["lotterys_account"]
			}
			us = append(us, common.UserAssetDetail{UserId: v.PayerUserId, Type: v.CurrencyType, TransactionType: common.TRANSACTION_CONSUME, Amount: -v.PayAmount, Date: today})
			us = append(us, common.UserAssetDetail{UserId: system_acount, Type: v.CurrencyType, TransactionType: common.TRANSACTION_CONSUME, Amount: v.PayAmount, Date: today})

		case common.CURRENCY_KT:
			if v.PublishArea != common.CURRENCY_KT { // 非KT专区购买
				us = append(us, common.UserAssetDetail{UserId: v.PayerUserId, Type: v.CurrencyType, TransactionType: common.KT_TRANSACTION_CONSUME, Amount: -v.PayAmount, Date: today})
				us = append(us, common.UserAssetDetail{UserId: system_acount, Type: v.CurrencyType, TransactionType: common.KT_TRANSACTION_CONSUME, Amount: v.PayAmount, Date: today})
			} else { // 目前在触发器中处理

			}
		}
	}

	// 增量修改
	if len(us) > 0 {
		for _, v := range us {
			if err = db.Save(&v).Error; err != nil {
				glog.Error("Asset_Pay fail! err=", err)
				if strings.Index(err.Error(), "cannot be negtive") > 0 {
					reason = common.ERR_MONEY_NOT_MORE
					err = fmt.Errorf(reason)
				}
				//yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
		}
	}

	// 增量修改商家交易额信息
	if len(bs) > 0 {
		err = util.PublishMsg(common.MQUrl{Methond: "POST", AppKey: "ybapi", Uri: "/man/business/amount/update", Data: bs})
		if err != nil {
			glog.Error("Asset_Pay fail! PublishMqurl err=", err)
			//yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}

	err = nil
	reason = ""
	return
}

// 订单支付接口 需登录接口
func Asset_Pay(c *gin.Context) {
	user_id, ok := util.GetUid(c)
	if !ok {
		return
	}
	var args []common.YBAssetPool
	if ok := util.UnmarshalReq(c, &args); !ok {
		glog.Error("Asset_Pay fail! args invalid!", args)
		return
	}

	db := db.GetTxDB(c)
	if reason, err := pay(db, user_id, args); err != nil {
		glog.Error("Asset_Pay fail! err=", err, " reason:", reason)
		yf.JSON_Fail(c, reason)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

type paynotokenSt struct {
	Pools  []common.YBAssetPool `json:"pools"`
	UserId int64                `json:"user_id"`
}

// 通过用户id支付 无登录接口
func Asset_PayByUserId(c *gin.Context) {
	var args paynotokenSt
	if ok := util.UnmarshalReq(c, &args); !ok {
		glog.Error("Asset_Pay fail! args invalid!", args)
		return
	}

	db := db.GetTxDB(c)
	if reason, err := pay(db, args.UserId, args.Pools); err != nil {
		glog.Error("Asset_PayNoToken fail! err=", err, " reason:", reason)
		yf.JSON_Fail(c, reason)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

type orderSt struct {
	OrderIds []int64 `json:"order_ids" binding:"required"`
	Status   int     `json:"status" binding:"required"`
}

var nSendSnetEmailTime int64

// 修改订单id状态
func Asset_SetStatus(c *gin.Context) {
	// user_id, ok := util.GetUid(c)
	// if !ok {
	// 	return
	// }
	var args orderSt
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	args.OrderIds = base.UniqueInt64Slice(args.OrderIds) // 去重
	db := db.GetTxDB(c)
	var vs []common.YBAssetPool
	// 查找处于冻结状态的订单
	if err := db.Where("status in(?) and order_id in(?)", []int{common.ASSET_POOL_LOCK, args.Status}, args.OrderIds).Find(&vs).Error; err != nil {
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 订单数量不正确
	if len(vs) != len(args.OrderIds) {
		ids := []int64{}
		for _, v := range vs {
			ids = append(ids, v.OrderId)
		}
		glog.Errorf("Asset_SetStatus fail! args.OrderIds:%v vs:%v", args.OrderIds, ids)
		yf.JSON_Fail(c, common.ERR_ORDER_NOT_EXIST)
		return
	}

	// 踢除掉已是取消或退款的订单
	pools := []common.YBAssetPool{}
	for i, v := range vs {
		if v.Status == common.ASSET_POOL_LOCK {
			pools = append(pools, vs[i])
		}
	}
	vs = pools

	if common.ASSET_POOL_FINISH == args.Status || common.ASSET_POOL_CANCEL == args.Status {
		// 订单完成	或 退款
		now := time.Now().Unix()
		db1 := db.Model(&common.YBAssetPool{}).Where("order_id in(?)", args.OrderIds).Updates(map[string]interface{}{"status": args.Status, "update_time": now})
		if db1.Error != nil || db1.RowsAffected != int64(len(args.OrderIds)) {
			glog.Error("Asset_SetStatus fail! err", db1.Error)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}

		// ybt的支付订单 将从系统帐号划扣相应的
		today := time.Now().Format("2006-01-02")
		us := []common.UserAssetDetail{}
		for _, v := range vs {
			if v.PublishArea != common.CURRENCY_KT { // 非KT专区
				if common.ASSET_POOL_FINISH == args.Status {
					if v.SellerKt > 0 {
						// 完成  将相应的kt从系统帐号转给商家用户
						ktAccount := "kt_account"
						if v.CurrencyType == common.CURRENCY_SNET {
							ktAccount = "snet_zone_account"
						}
						us = append(us, common.UserAssetDetail{UserId: conf.Config.SystemAccounts[ktAccount], Type: common.CURRENCY_KT, TransactionType: common.KT_TRANSACTION_SELLER, Amount: -v.SellerKt, Date: today})
						us = append(us, common.UserAssetDetail{UserId: v.SellerUserId, Type: common.CURRENCY_KT, TransactionType: common.KT_TRANSACTION_SELLER, Amount: v.SellerKt, Date: today})
					}
				} else {
					// 退款操作
					var system_acount int64 = conf.Config.SystemAccounts["zone_account"]
					switch v.CurrencyType {
					case common.CURRENCY_YBT:
						us = append(us, common.UserAssetDetail{UserId: system_acount, Type: v.CurrencyType, TransactionType: common.YBT_TRANSACTION_RETURND, Amount: -v.PayAmount, Date: today})
						us = append(us, common.UserAssetDetail{UserId: v.PayerUserId, Type: v.CurrencyType, TransactionType: common.YBT_TRANSACTION_RETURND, Amount: v.PayAmount, Date: today})
					case common.CURRENCY_KT:
						us = append(us, common.UserAssetDetail{UserId: system_acount, Type: v.CurrencyType, TransactionType: common.KT_TRANSACTION_RETURND, Amount: -v.PayAmount, Date: today})
						us = append(us, common.UserAssetDetail{UserId: v.PayerUserId, Type: v.CurrencyType, TransactionType: common.KT_TRANSACTION_RETURND, Amount: v.PayAmount, Date: today})
					case common.CURRENCY_SNET:
						system_acount = conf.Config.SystemAccounts["snet_zone_account"]
						if common.PUBLISH_AREA_LOTTERYS == v.PublishArea {
							// 积分抽奖帐号
							system_acount = conf.Config.SystemAccounts["lotterys_account"]
						}
						us = append(us, common.UserAssetDetail{UserId: system_acount, Type: v.CurrencyType, TransactionType: common.TRANSACTION_RETURND, Amount: -v.PayAmount, Date: today})
						us = append(us, common.UserAssetDetail{UserId: v.PayerUserId, Type: v.CurrencyType, TransactionType: common.TRANSACTION_RETURND, Amount: v.PayAmount, Date: today})
					}
				}
			}
		}

		for _, v := range us {
			if err := db.Save(&v).Error; err != nil {
				glog.Error("Asset_SetStatus fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)

				s := "成交"
				if args.Status == common.ASSET_POOL_CANCEL {
					s = "退款"
				}
				util.SendDingTextTalkToMe(fmt.Sprintf("订单支付状态更新失败,状态:%v 帐号id:%v 币种:%v err=%v", s, v.UserId, v.Type, err))

				// 若是snet帐号kt不足 则发邮件通知miner
				if v.UserId == conf.Config.SystemAccounts["snet_zone_account"] {
					now := time.Now().Unix()
					if nSendSnetEmailTime < now {
						nSendSnetEmailTime = now + 10*int64(time.Minute) // 每隔10分钟发了一次邮件
						if third, ok := conf.Config.ThirdAccount["miner"]; ok {
							// 第三方帐户余额不足 及时通知
							if alarm_email, ok := third.Ext["alarm_email"]; ok {
								s := fmt.Sprintf("您帐号[%v]余额已不足，请及时充值！本次用户兑换需划转:%v %v", v.UserId, v.Amount, common.GetCurrencyName(v.Type))
								util.PublishMsg(common.MQMail{Receiver: []string{alarm_email}, Subject: "Yunbay商城", Content: s})
							}
						}
					}
				}

				return
			}
		}
	} else {
		glog.Error("Asset_SetStatus fail! type:", args.Status)
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

type OrderRebatSt struct {
	OrderId int64   `json:"order_id" binding:"required"`
	Rebat   float64 `json:"rebat" binding:"required,min=0,max=1"`
	//TxHash string `json:"tx_hash"`
}

func ManPayRebat(c *gin.Context) {
	var req OrderRebatSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	db := db.GetTxDB(c)
	var v common.YBAssetPool
	err := db.Find(&v, "order_id=? and publish_area=? and status=? and extinfos->'rebat' is null", req.OrderId, common.PUBLISH_AREA_REBAT, common.STATUS_INIT).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("ManPayRebat fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	// 计算需要退款的金额及实际付款金额等
	refundAmount := v.PayAmount - v.PayAmount*req.Rebat
	v.PayAmount *= req.Rebat
	v.SellerAmount *= req.Rebat
	v.RebatAmount *= req.Rebat

	// 更新订单折扣价格
	res := db.Model(&common.YBAssetPool{}).Where("order_id=? and publish_area=? and status=? and extinfos->'rebat' is null", req.OrderId, common.PUBLISH_AREA_REBAT, common.STATUS_INIT).
		Updates(map[string]interface{}{"pay_amount": v.PayAmount, "seller_amount": v.SellerAmount, "rebat_amount": v.RebatAmount, "extinfos": gorm.Expr("extinfos || ?", fmt.Sprintf("{\"rebat\":%v}", req.Rebat))})

	if res.Error != nil {
		glog.Error("ManPayRebat fail! err=", res.Error)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if 0 == res.RowsAffected {
		glog.Error("ManPayRebat no Rowsaffected! order_id:", req.OrderId)
		yf.JSON_Ok(c, gin.H{})
		return
	}

	today := time.Now().Format("2006-01-02")

	us := []common.UserAssetDetail{}
	// 执行退款操作
	switch v.CurrencyType {
	case common.CURRENCY_YBT:
		us = append(us, common.UserAssetDetail{UserId: conf.Config.SystemAccounts["zone_account"], Type: v.CurrencyType, TransactionType: common.YBT_TRANSACTION_RETURND, Amount: -refundAmount, Date: today})
		us = append(us, common.UserAssetDetail{UserId: v.PayerUserId, Type: v.CurrencyType, TransactionType: common.YBT_TRANSACTION_RETURND, Amount: refundAmount, Date: today})
	case common.CURRENCY_KT:
		us = append(us, common.UserAssetDetail{UserId: conf.Config.SystemAccounts["zone_account"], Type: v.CurrencyType, TransactionType: common.KT_TRANSACTION_RETURND, Amount: -refundAmount, Date: today})
		us = append(us, common.UserAssetDetail{UserId: v.PayerUserId, Type: v.CurrencyType, TransactionType: common.KT_TRANSACTION_RETURND, Amount: refundAmount, Date: today})
	case common.CURRENCY_SNET:
		us = append(us, common.UserAssetDetail{UserId: conf.Config.SystemAccounts["snet_zone_account"], Type: v.CurrencyType, TransactionType: common.TRANSACTION_RETURND, Amount: -refundAmount, Date: today})
		us = append(us, common.UserAssetDetail{UserId: v.PayerUserId, Type: v.CurrencyType, TransactionType: common.TRANSACTION_RETURND, Amount: refundAmount, Date: today})
	default:

	}

	for _, u := range us {
		if err = db.Save(&u).Error; err != nil {
			glog.Error("ManPayRebat fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}

	bs := []util.BusinessSt{}
	bs = append(bs, util.BusinessSt{UserId: v.SellerUserId, Type: v.CurrencyType, Amount: v.PayAmount, Rebat: v.RebatAmount})
	// 增量修改商家交易额信息
	if len(bs) > 0 {
		db.AfterCommit(func() {
			err = util.PublishMsg(common.MQUrl{Methond: "POST", AppKey: "ybapi", Uri: "/man/business/amount/update", Data: bs})
			if err != nil {
				glog.Error("Asset_Pay fail! PublishMqurl err=", err)
				//yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
		})
	}

	yf.JSON_Ok(c, gin.H{})
}
