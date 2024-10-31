package man

import (
	"fmt"
	"time"

	//"time"
	"yunbay/ybasset/common"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"

	//base "github.com/jay-wlj/gobaselib"
	"yunbay/ybasset/server/share"
	"yunbay/ybasset/util"
)

// 提币申请审核接口
func Yunex_SnapBonusAccount(c *gin.Context) {

	yf.JSON_Ok(c, gin.H{})
}

type thirdbonus struct {
	ThirdBonusId int64  `json:"third_bonusid"`
	Date         string `json:"date"`
}

// 发放第三方平台yunex分红接口
func Third_BonusKt(c *gin.Context) {
	var args thirdbonus
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}
	plat := share.GetThirdPlatFromBonusId(args.ThirdBonusId)
	now := time.Now()
	// 发放分红 注意此处不需要tx回滚操作
	db := db.GetTxDB(c)

	switch plat {
	case "yunex": // yunex平台分红
		if err := bonus_yunex(db.DB, args.ThirdBonusId, args.Date); err != nil {
			glog.Error("Yunex_BonusKt fail! err=", err)
			//yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			util.PublishMsg(common.MQMail{Receiver: []string{"305898636@qq.com"}, Subject: "Yunex_BonusKt Error", Content: fmt.Sprintf("bonus_yunex fail! err=%v", err)})
			//return
		}
	default:
		glog.Error("Third_BonusKt invalid bonus_id:", args.ThirdBonusId)
		return
	}

	glog.Infof("Third_BonusKt ok! tick=%v", time.Since(now).String())
}

// 发放yunex平台用户分红
func bonus_yunex(db *gorm.DB, tid int64, date string) (err error) {
	var vs []common.ThirdBonus
	// 获取当日没有发放的第三方用户分红数据
	if err = db.Find(&vs, "tid=? and date=? and status in(?)", tid, date, []int{common.STATUS_FAIL, common.STATUS_INIT}).Error; err != nil {
		glog.Error("bonus_yunex fail! err=", err)
		return
	}
	if len(vs) == 0 {
		return
	}
	bs := []util.YunexKtBonus{}
	for _, v := range vs {
		bs = append(bs, util.YunexKtBonus{ToUid: v.Uid, Symbol: "KT", Amount: v.Kt, Date: v.Date, OrderId: v.Id})
	}

	var fs []util.YunexKtBonusRet
	var ys []util.YunexKtBonus

	sids := []int64{} // 分红成功
	fids := []int64{} // 分红失败
	for {
		// 每次只发送<=100条数据
		if len(bs) > 100 {
			ys = bs[:100]
			bs = bs[100:]
		} else {
			ys = bs
		}
		if fs, err = util.BonusYunexKt(ys); err != nil {
			glog.Error("util.BonusYunexKt fail! err=", err)
			break
		}
		for _, v := range ys {
			sids = append(sids, v.OrderId)
		}
		for _, v := range fs {
			fids = append(fids, v.OrderId)
		}
	}

	now := time.Now().Unix()
	// 先保存成功状态
	if err = db.Model(&common.ThirdBonus{}).Where("tid=? and date=? and id in(?)", tid, date, sids).Update(map[string]interface{}{"status": common.STATUS_OK, "update_time": now}).Error; err != nil {
		glog.Errorf("bonus_yunex update success status fail! fids:%v", sids)
		return
	}
	// 再保存失败状态
	if err = db.Model(&common.ThirdBonus{}).Where("tid=? and date=? and id in(?)", tid, date, fids).Update(map[string]interface{}{"status": common.STATUS_FAIL, "update_time": now}).Error; err != nil {
		glog.Errorf("bonus_yunex update fail status fail! fids:%v", fids)
		return
	}
	return
}
