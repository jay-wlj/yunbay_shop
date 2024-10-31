package task

import (
	"github.com/jie123108/glog"
	//"github.com/jie123108/glog"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"math"
	"time"
	"yunbay/ybcron/conf"
)

func getDifficultByTime(tm time.Time) (difficult float64, period int64) {
	online_time, err := time.Parse("2006-01-02 15:04:05", conf.Config.Mining.OnlineTime)
	if err != nil {
		panic(fmt.Sprintf("getDifficultByTime fail! err=%v\n", err))
		return
	}
	fperiod := math.Floor(tm.Sub(online_time).Hours() / 24.0) // 计算已有多少个周期 向下取整
	period, _ = base.StringToInt64(base.Float64ToString(fperiod))
	difficult = getdifficult()
	return
}

// 获取挖矿难度
// func getdifficult(period int64)(difficult float64) {
// 	if period < 1 {
// 		period = 1
// 	}
// 	var day_per float64
// 	v := conf.Config.Mining
// 	day_per = float64(1.0) + float64(period-1) * v.Coefficient
// 	difficult = v.Standard * math.Pow(day_per, v.Powy)
// 	return
// }

func getdifficult() (difficult float64) {
	v := conf.Config.Mining
	ybt, err := GetYbt()
	if err != nil {
		panic(fmt.Sprintf("GetYbt err=%v", err))
	}
	// 计算已释放百分比
	unlockper := (ybt.UnlockMinepool + ybt.UnlockReward + ybt.UnlockProject) / (ybt.Reward + ybt.Minepool + ybt.Project)
	difseq := 1 / math.Pow(1-unlockper, v.Powy) // 难度系数
	difficult = v.Standard * difseq             // 挖矿难度
	glog.Info("current ybt:", ybt, "total ybt:", ybt.Reward+ybt.Minepool+ybt.Project, "total unlock:", ybt.UnlockMinepool+ybt.UnlockReward+ybt.UnlockProject, "difficult:", difficult)
	return
}

func getdifficultbyrelease(release_ybt float64) (difficult float64) {
	ybt, err := GetYbt()
	if err != nil {
		panic(fmt.Sprintf("GetYbt err=%v", err))
	}
	v := conf.Config.Mining
	// 计算已释放百分比
	unlockper := release_ybt / (ybt.Reward + ybt.Minepool + ybt.Project)
	difseq := 1 / math.Pow(1-unlockper, v.Powy) // 难度系数
	difficult = v.Standard * difseq             // 挖矿难度
	//glog.Info("current ybt:",ybt, "total ybt:", ybt.Reward+ybt.Minepool+ybt.Project, "total unlock:",ybt.UnlockMinepool+ybt.UnlockReward+ybt.UnlockProject, "difficult:", difficult)
	return
}
