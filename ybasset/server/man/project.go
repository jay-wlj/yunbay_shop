package man

import (
	"fmt"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"time"
	"yunbay/ybasset/common"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"

	//"github.com/jinzhu/gorm"
	//base "github.com/jay-wlj/gobaselib"
	"yunbay/ybasset/conf"
	"yunbay/ybasset/util"
)

func getUserIdByProjectType(project_type int) (user_id int64) {
	for _, v := range conf.Config.ProjectYbtAllot {
		if v.Type == project_type {
			user_id = v.UserId
			break
		}
	}
	return
}
func getTeamsAsset() (v common.UserAsset, err error) {
	user_id := getUserIdByProjectType(common.PROJECT_TYPE_TEAMS)
	if err = db.GetDB().Find(&v, "user_id=?", user_id).Error; err != nil {
		glog.Error("getTeamsAsset fail! err=", err)
		return
	}
	return
}

func getUserAssetByProjectType(project_type int) (v common.UserAsset, err error) {
	user_id := getUserIdByProjectType(project_type)
	if user_id == 0 {
		glog.Error("getUserIdByProjectType fail! user_id=0")
		err = fmt.Errorf("user_id is 0, project_type=%v", project_type)
		return
	}
	if err = db.GetDB().Find(&v, "user_id=?", user_id).Error; err != nil {
		glog.Error("getTeamsAsset fail! err=", err)
		return
	}
	return
}

type teamsRewardSt struct {
	List        []RewarSt `json:"list" binding:"required"`
	Type        int       `json:"type"`
	ProjectType int       `json:"project_type"`
}

// 从团队激励帐户中扣除相应数量ybt分配到项目人员的帐户
func Project_Reward(c *gin.Context) {
	maner, err := util.GetHeaderString(c, "X-Yf-Maner")
	if maner == "" || err != nil {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	var req teamsRewardSt
	if ok := util.UnmarshalReq(c, &req); !ok {
		return
	}
	ua, err := getUserAssetByProjectType(req.ProjectType)
	if err != nil {
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 查询可用ybt是否足额
	var total_amount float64 = 0
	for _, v := range req.List {
		amount, _ := v.Amount.Float64()
		total_amount += amount
	}
	var transaction_type int = common.YBT_TRANSACTION_PROJECT
	switch req.Type {
	case common.CURRENCY_YBT:
		if ua.NormalYbt < total_amount {
			glog.Error("Project_TeamsReward fail! temas account normal_ybt not more! normal_ybt=", ua.NormalKt, " need ybt:", total_amount)
			yf.JSON_Fail(c, common.ERR_YBT_NOT_MORE)
			return
		}
	case common.CURRENCY_KT:
		transaction_type = common.KT_TRANSACTION_PROJECT
		if ua.NormalKt < total_amount {
			glog.Error("Project_TeamsReward fail! temas account normal_kt not more! normal_kt=", ua.NormalKt, " need kt:", total_amount)
			yf.JSON_Fail(c, common.ERR_KT_NOT_MORE)
			return
		}
	default:
		glog.Error("type is not support!")
		yf.JSON_Fail(c, common.ERR_TYPE_NOT_SUPPORT)
		return
	}

	// 释放ybt
	today := time.Now().Format("2006-01-02")

	// 先将总额从团队激励中扣除
	vs := []common.UserAssetDetail{common.UserAssetDetail{UserId: ua.UserId, Amount: -total_amount, Type: req.Type, TransactionType: transaction_type, Date: today}}
	for _, v := range req.List {
		amount, _ := v.Amount.Float64()
		vs = append(vs, common.UserAssetDetail{UserId: v.UserId, Amount: amount, Type: req.Type, TransactionType: transaction_type, Date: today})
	}

	// 生成资产流水记录
	db := db.GetTxDB(c)
	for _, v := range vs {
		if err := db.Save(&v).Error; err != nil {
			glog.Error("Project_TeamsRewardYbt fail! UserAssetDetail insert fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}

	yf.JSON_Ok(c, gin.H{})
}
