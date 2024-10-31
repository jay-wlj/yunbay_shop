package server

import (
	"encoding/json"
	"yunbay/ybnsq/util"

	"github.com/jie123108/glog"
	nsq "github.com/nsqio/go-nsq"
)

//UsrID is the message struct
type UsrID struct {
	UID int64 `json:"user_id"` //UserID
}

//ConsumerRecommend is NSQ Consumer struct
type UserRecommend struct{}

//HandleMessage is function of message
func (*UserRecommend) HandleMessage(msg *nsq.Message) (err error) {
	glog.Info("[useradd]receive:", msg.NSQDAddress, " ,attempts:", msg.Attempts, " message:", string(msg.Body))

	var args util.UserAdd

	err = json.Unmarshal(msg.Body, &args)
	if err != nil {
		msg.Finish() // 此消息body非法 丢弃
		glog.Error("UserRecommend args invalid! finish msg, err=", err)
		return err
	}
	glog.Infof("user_id: %v broker_id: %d", args.UserId, args.BrokerId)

	//insertRecommend(DBConn, Recommddata.fromuser, Recommddata.mid)
	//db := GetDefaultDb().Begin()
	if err = HandUser(args); err != nil {
		//db.Rollback()
		glog.Error("HandUrl fail! try again! trys:", msg.Attempts)
		msg.Requeue(-1) // 保证不丢弃该消息，直到处理成功
		return
	}

	//db.Commit()
	msg.Finish() // 此消息已完成处理 丢弃
	return
}

func HandUser(args util.UserAdd) (err error) {
	//now := time.Now().Unix()
	//if args.FromInviteId > 0 {
	// 添加用户邀请记录
	if err = util.AddInvite(args); err != nil {
		glog.Error("AddInvite fail! err=", err)
		return
	}
	//}

	// 生成用户资金记录
	if err = util.AddUserAsset(args.UserId, args.BrokerId); err != nil {
		glog.Error("AddUserAsset fail! err=", err)
		return
	}

	// // 注册im信息 失败无需返回
	// if err1 := util.RegisterIM(args.UserId); err1 != nil {
	// 	glog.Error("RegisterIM fail! user_id=", args.UserId, " err=", err)
	// }

	err = nil
	return
}
