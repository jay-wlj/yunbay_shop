package man

import (
	"yunbay/ybapi/common"
	"yunbay/ybapi/util"

	"github.com/gin-gonic/gin"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"github.com/jie123108/glog"

	//"github.com/lib/pq"
	"github.com/jinzhu/gorm"
)

type businesstatus struct {
	UserId int64 `json:"user_id" binding:"gt=0"`
	Status int   `json:"status"`
}

func Business_Status(c *gin.Context) {

	var args businesstatus
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}

	db := db.GetTxDB(c).Model(common.Business{}).Where("user_id=?", args.UserId).Updates(map[string]interface{}{"status": args.Status})
	if err := db.Error; err != nil {
		glog.Error("Business_Status  fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	if err := util.SetUserType(args.UserId, args.Status); err != nil {
		glog.Error("SetUserType  fail! err", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	yf.JSON_Ok(c, gin.H{})
}

func Business_List(c *gin.Context) {
	status, _ := base.CheckQueryIntDefaultField(c, "status", -1)
	page, _ := base.CheckQueryIntDefaultField(c, "page", 1)
	page_size, _ := base.CheckQueryIntDefaultField(c, "page_size", 10)

	db := db.GetDB().ListPage(page, page_size).Order("create_time desc")
	if status >= 0 {
		db = db.Where("status=?", status)
	}

	vs := []common.Business{}
	if err := db.Find(&vs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Business_Info fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	list_ended := true
	if len(vs) == page_size {
		list_ended = false
	}
	yf.JSON_Ok(c, gin.H{"list": vs, "list_ended": list_ended})
}

type bamountSt struct {
	UserId int64   `json:"user_id" binding:"gt=0"`
	Type   int     `json:"type"`
	Amount float64 `json:"amount"  binding:"gt=0"`
	Rebat  float64 `json:"rebat"`
}

func Business_AmountUpdate(c *gin.Context) {
	var req []bamountSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}

	// mAmount := make(map[int64]float64)
	// mRebat := make(map[int64]float64)
	// for _, v := range req {
	// 	mAmount[v.UserId] = v.Amount
	// 	mRebat[v.UserId] = v.Rebat
	// }
	db := db.GetTxDB(c)
	for _, v := range req {
		switch v.Type {
		case common.CURRENCY_YBT:
			if err := db.Model(&common.Business{}).Where("user_id=?", v.UserId).Updates(map[string]interface{}{"total_ybtflow": gorm.Expr("total_ybtflow + ?", v.Amount)}).Error; err != nil {
				glog.Error("Business_AmountUpdate fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			break
		case common.CURRENCY_KT:
			if err := db.Model(&common.Business{}).Where("user_id=?", v.UserId).Updates(map[string]interface{}{"total_tradeflow": gorm.Expr("total_tradeflow + ?", v.Amount), "total_rebat": gorm.Expr("total_rebat + ?", v.Rebat)}).Error; err != nil {
				glog.Error("Business_AmountUpdate fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			break
		}

	}

	yf.JSON_Ok(c, gin.H{})
}
