package task

import (
	"yunbay/ybcron/util"
	"time"

	"github.com/jie123108/glog"
)

func YBAsset_ReleaseYbtKt() {
	yester_day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	if err := util.ReleaseYbt(yester_day); err != nil {
		glog.Error("ReleaseYbt fail! err=", err)
		//return
	}
	if err := util.ReleaseKt(yester_day); err != nil {
		glog.Error("ReleaseKt fail! err=", err)
		//return
	}

	glog.Infof("YBAsset_ReleaseYbtKt success!")

	// 执行ybt空投回收
	YBT_Recover()

	// 解冻ybt定期冻结
	YBT_UnlockFixYbt()
	return
}
