package man

import (
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"math"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybasset/conf"
	"yunbay/ybasset/dao"
	"yunbay/ybasset/util"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func Man_KtRewardList(c *gin.Context) {
	id, _ := base.CheckQueryInt64DefaultField(c, "id", 0)
	date, _ := base.CheckQueryStringField(c, "date")
	user_id, _ := base.CheckQueryInt64DefaultField(c, "user_id", -1)
	check_status, _ := base.CheckQueryIntDefaultField(c, "status", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	db := db.GetDB()
	db.DB = db.Where("(bonus_ybt>0 or kt_bonus>0)")
	if id > 0 {
		db.DB = db.Where("id=?", id)
	}
	if user_id > -1 {
		db.DB = db.Where("user_id=?", user_id)
	}
	if date != "" {
		db.DB = db.Where("date=?", date)
	}
	if check_status > -1 {
		db.DB = db.Where("check_status=?", check_status)
	}
	var total int = 0
	if err := db.Model(&common.KtBonusDetail{}).Count(&total).Error; err != nil {
		glog.Error("BonusOrders_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	db.DB = db.ListPage(page, page_size)
	vs := []common.KtBonusDetail{}
	if err := db.Order("id desc").Find(&vs).Error; err != nil {
		glog.Error("Man_YbtRewardList fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": vs, "list_page": base.IsListEnded(page, page_size, len(vs), total), "total": total})
}

func Man_KtRewardCheck(c *gin.Context) {
	checker_name, err := util.GetHeaderString(c, "X-Yf-Maner")
	if checker_name == "" || err != nil {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var args datecheck
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	now := time.Now()
	// 释放ybt
	db := db.GetTxDB(c)
	if err := release_kt(db, args.Date); err != nil {
		s := fmt.Sprintf("Man_KtRewardCheck fail! err=%v", err)
		glog.Error(s)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		util.PublishMsg(common.MQMail{Receiver: []string{"305898636@qq.com"}, Subject: "Yunbay Error", Content: s})
		util.SendDingTextTalk(s, []string{"15818717950"})
		return
	}
	glog.Infof("Man_KtRewardCheck ok! tick=%v", time.Since(now).String())
	util.SendDingTextTalk(fmt.Sprintf("kt:%v 发放完毕 共耗时:%v", args.Date, time.Since(now).String()), nil)
	yf.JSON_Ok(c, gin.H{})
}

// 发放当天kt收益金
func release_kt(db *db.PsqlDB, date string) (err error) {
	// 获取昨日平台交易记录
	now := time.Now().Unix()
	db_asset := db.Model(&common.YBAssetDetail{}).Where("date=? and kt_status=?", date, common.STATUS_INIT).Updates(map[string]interface{}{"kt_status": common.STATUS_OK, "update_time": now})
	if db_asset.Error != nil {
		glog.Error("release_ybt fail! YBAssetDetail update err=", err)
		return
	}
	if db_asset.RowsAffected == 0 {
		glog.Error("release_ybt has released! return")
		return
	}

	vs := []common.KtBonusDetail{}
	if err = db.Find(&vs, "date=? and check_status=?", date, common.STATUS_INIT).Error; err != nil {
		glog.Error("release_ybt fail! err=", err)
		return
	}

	// 将状态置为已发放
	if err = db.Model(&common.KtBonusDetail{}).Where("date=? and check_status=?", date, common.STATUS_INIT).Updates(map[string]interface{}{"check_status": common.STATUS_OK, "update_time": now}).Error; err != nil {
		glog.Error("release_kt faiL! err=", err)
		return
	}
	yester_day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	if len(vs) == 0 {
		go async_deliver_seller(yester_day)
		glog.Error("release_kt no rows")
		return
	}

	today := time.Now().Format("2006-01-02")
	uids := []uIds{}
	if err = db.Model(&common.UserAsset{}).Select("user_id").Group("user_id").Scan(&uids).Error; err != nil {
		glog.Error("get all user_ids fail! err=", err)
		return
	}
	mks := make(map[int64]*common.KtBonusDetail)
	for i, v := range vs {
		mks[v.UserId] = &vs[i]
	}

	us := []common.UserAssetDetail{}
	bks := []common.BonusKtDetail{}
	for _, u := range uids {
		//for _, v := range vs {
		if v := mks[u.UserId]; v != nil {
			if v.KtBonus > 0 {
				// 如果是第三方平台分红 需走接口调用
				if v.ThirdBonus > 0 {
					dt := thirdbonus{ThirdBonusId: v.UserId, Date: date}
					mq := common.MQUrl{Methond: "post", AppKey: "ybasset", Uri: "/man/third/bonus/deliver", Data: dt}
					if err = util.PublishMsg(mq); err != nil {
						glog.Error("release_kt third bonus fail! date err=", err)
						err = nil
					}
					if err = bonus_yunex(db.DB, v.UserId, yester_day); err != nil {
						glog.Error("release_kt faiL! UserAssetDetail add err=", err)
						err = nil
					}
				} else {
					u := common.UserAssetDetail{UserId: v.UserId, Type: common.CURRENCY_KT, TransactionType: common.KT_TRANSACTION_PROFIT, Amount: v.KtBonus, Date: today}
					us = append(us, u)
				}
			}
			bks = append(bks, common.BonusKtDetail{UserId: u.UserId, Ybt: v.BonusYbt, Kt: v.KtBonus, Date: yester_day})
		} else {
			bks = append(bks, common.BonusKtDetail{UserId: u.UserId, Ybt: 0, Kt: 0, Date: yester_day})
		}
	}

	// 生成入帐记录
	for _, v := range us {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("release_kt faiL! UserAssetDetail add err=", err)
			return
		}
	}

	// 生成用户分红记录
	for _, v := range bks {
		if err = db.Save(&v).Error; err != nil {
			glog.Error("release_kt faiL! BonusKtDetail add err=", err)
			return
		}
	}
	dao.RefrenshUserBonus(yester_day)

	// 异步释放商家奖励的ybt及kt收益金
	//go async_deliver_seller(yester_day)
	if err = deliverSellerReward(db, yester_day); err != nil {
		glog.Error("release_kt faiL! deliverSellerReward err=", err)
		return
	}
	return
}

func async_deliver_seller(date string) {
	// 异步释放商家奖励的ybt及kt收益金
	data := datecheck{Date: date}
	headers := make(map[string]string)
	headers["X-Yf-Maner"] = "system"
	mq := common.MQUrl{Methond: "POST", AppKey: "ybasset", Uri: "/man/kt/reward/seller/check", Data: data, Headers: headers}
	if e := util.PublishMsg(mq); e != nil {
		glog.Error("async_deliver_seller fail! e=", e)
		util.PublishMsg(common.MQMail{Receiver: []string{"305898636@qq.com"}, Subject: "Yunbay Error", Content: fmt.Sprintln("async_deliver_seller fail! err=", e)})
	}
}
func Man_KtRewardSellerCheck(c *gin.Context) {
	checker_name, err := util.GetHeaderString(c, "X-Yf-Maner")
	if checker_name == "" || err != nil {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var args datecheck
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	now := time.Now()
	// 释放ybt
	db := db.GetTxDB(c)
	if err := deliverSellerReward(db, args.Date); err != nil {
		s := fmt.Sprintf("Man_KtRewardSellerCheck fail! err=%v", err)
		glog.Error(s)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		util.PublishMsg(common.MQMail{Receiver: []string{"305898636@qq.com"}, Subject: "Yunbay Error", Content: s})
		util.SendDingTextTalk(s, []string{"15818717950"})
		return
	}
	glog.Infof("Man_KtRewardSellerCheck ok! tick=%v", time.Since(now).String())
	util.SendDingTextTalk(fmt.Sprintf("kt:%v 发放完毕 共耗时:%v", args.Date, time.Since(now).String()), nil)
	yf.JSON_Ok(c, gin.H{})
}

// 发放商家奖励ybt及kt
func deliverSellerReward(db *db.PsqlDB, date string) (err error) {

	bzj_id := conf.Config.SystemAccounts["bzj_account"]
	if bzj_id == 0 {
		glog.Error("DeliverSellerReward fail! bzj_id=0")
		err = fmt.Errorf("DeliverSellerReward fail! bzj_id=0")
		return
	}
	// 获取保证金帐号昨日分红明细记录
	var bzj_bonus common.KtBonusDetail
	if err = db.Find(&bzj_bonus, "user_id=? and date=?", bzj_id, date).Error; err != nil {
		glog.Error("DeliverSellerReward fail! err=", err)
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	// 从订单挖矿表中查找每一笔发货未转给商家的订单的商家奖励ybt 注:有可能发货的订单是以前的
	var ors []common.Ordereward
	if err = db.Find(&ors, "date<=? and valid=? and seller_status=?", date, common.STATUS_OK, common.STATUS_INIT).Error; err != nil {
		glog.Error("DeliverSellerReward fail! err=", err)
		return
	}

	mybt := make(map[int64]float64) // 共需转给商家的商家奖励的ybt
	mkt := make(map[int64]float64)  // 共需转给商家的商家奖励ybt产生的kt收益

	ok_ids := []int64{}

	for _, v := range ors {
		ok_ids = append(ok_ids, v.OrderId)
		mybt[bzj_id] += -math.Abs(v.SellerYbt)        // 从保证金帐号将该笔订单产生的商家奖励ybt及对应的分红转给商家用户
		mybt[v.SellerUserId] += math.Abs(v.SellerYbt) // 该商家奖励的ybt发放给商家

		// 将该笔ybt产生的kt分红也转给商家
		if v.SellerYbt > bzj_bonus.BonusYbt {
			s := fmt.Sprintln("DeliverSellerReward fail! seller_ybt:", v.SellerYbt, " > bzj_account ybt:", bzj_bonus.BonusYbt)
			glog.Error(s)
			err = fmt.Errorf(s)
			return
		}
		// 保证金帐号昨日有分红
		if bzj_bonus.BonusYbt > 0 && bzj_bonus.KtBonus > 0 {
			kt := (v.SellerYbt / bzj_bonus.BonusYbt) * bzj_bonus.KtBonus
			mkt[bzj_id] += -kt        // 从保证金帐号将该笔订单产生的商家奖励ybt的分红转给商家用户
			mkt[v.SellerUserId] += kt // 该商家奖励的ybt产生的分红发放给商家
		}
	}

	if len(ok_ids) > 0 {
		today := time.Now().Format("2006-01-02")
		now := time.Now().Unix()

		us := []common.UserAssetDetail{}
		for k, v := range mybt {
			us = append(us, common.UserAssetDetail{UserId: k, Type: common.CURRENCY_YBT, TransactionType: common.YBT_TRANSACTION_SELLER, Amount: v, Date: today})
		}
		for k, v := range mkt {
			us = append(us, common.UserAssetDetail{UserId: k, Type: common.CURRENCY_KT, TransactionType: common.KT_TRANSACTION_PROFIT, Amount: v, Date: today})
		}

		// 先将订单挖矿表的商家奖励状态置为已转
		ret := db.Model(&common.Ordereward{}).Where("date<=? and valid=? and seller_status=?", date, common.STATUS_OK, common.STATUS_INIT).Update(map[string]interface{}{"seller_status": common.STATUS_OK, "update_time": now})
		if ret.Error != nil || ret.RowsAffected != int64(len(ok_ids)) {
			glog.Error("DeliverSellerReward fail! RowsAffected=", ret.RowsAffected, " len(ok_ids)=", len(ok_ids), " err=", err)
			return
		}

		// 添加资产明细记录
		for _, v := range us {
			if err = db.Save(&v).Error; err != nil {
				glog.Error("DeliverSellerReward fail! user_asset_detail err=", err)
				return
			}
		}

		// 更新商家的每日ybt和kt的分红记录
		bonusKt := common.BonusKtDetail{}
		for k, v := range mkt {
			sql := fmt.Sprintf("update %v set update_time=%v, ybt=ybt+%v, kt=kt+%v where user_id=%v and date='%v'", bonusKt.TableName(), now, mybt[k], v, k, date)
			if err = db.Exec(sql).Error; err != nil {
				glog.Error("deliverSellerReward fail! update BonusKtDetail err=", err)
			}
		}

		bonusYbt := common.BonusYbtDetail{}
		// 更新每日ybt分红记录
		for k, v := range mybt {
			sql := fmt.Sprintf("update %v set update_time=%v, total_ybt=total_ybt+%v, infos= infos || CONCAT('{\"consume\":', COALESCE(infos->>'consume', '0')::double precision + %v, '}')::jsonb where user_id=%v and date='%v'", bonusYbt.TableName(), now, v, v, k, date)
			if err = db.Exec(sql).Error; err != nil {
				glog.Error("deliverSellerReward fail! update BonusYbtDetail err=", err)
			}
		}

		dao.RefrenshUserBonus(date)
	}

	return
}
