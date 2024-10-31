package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"yunbay/ybnsq/util"

	"github.com/jie123108/glog"
	"github.com/nsqio/go-nsq"
)

//UsrID is the message struct
type UrlParams struct {
	Methond      string            `json:"method"`
	AppKey       string            `json:"appkey"`
	Uri          string            `json:"uri"`
	Headers      map[string]string `json:"headers"`
	Data         interface{}       `json:"data"`
	Timeout      int               `json:"timeout"`
	MaxTrys      int16             `json:"maxtrys"` // default 3 times
	ResponseBody string            `json:"response_body"`
	Delay        *string           `json:"delay"` // default -1  ns
	// Host string `json:"host"`
}

//ConsumerRecommend is NSQ Consumer struct
type MQUrl struct{}

//HandleMessage is function of message
func (*MQUrl) HandleMessage(msg *nsq.Message) (err error) {
	glog.Info("[MQUrl]receive:", msg.NSQDAddress, " message:", string(msg.Body))

	args := UrlParams{MaxTrys: 3}

	err = json.Unmarshal(msg.Body, &args)
	if err != nil {
		msg.Finish() // 此消息body非法 丢弃
		glog.Error("MQUrl args invalid! finish msg, err=", err)
		return err
	}
	glog.Info("url: ", args.Uri, " body: ", args.Data, " maxtrys:", args.MaxTrys)

	//db := GetDefaultDb().Begin()
	if err = HandUrl(args); err != nil {
		//db.Rollback()
		var delay time.Duration = -1
		if args.Delay != nil {
			if d, e := time.ParseDuration(*args.Delay); e == nil {
				delay = d
			}
		}
		if args.MaxTrys > int16(0) {
			if msg.Attempts >= uint16(args.MaxTrys) {
				msg.Finish() // 大于尝试次数则丢弃，直到大于尝试次数
				glog.Error("HandUrl fail! trys arrivate maxtrys:", msg.Attempts, " discard this msg!")
				return
			} else {
				msg.Requeue(delay)
			}
		} else if args.MaxTrys == -1 {
			msg.Requeue(delay) // 保证不丢弃该消息，直到处理成功
		}
		glog.Error("HandUrl fail! try again! trys:", msg.Attempts, " maxtrys:", args.MaxTrys, " delay:", delay)
		return
	}
	//db.Commit()
	msg.Finish() // 此消息已完成处理 丢弃
	return
}

func HandUrl(args UrlParams) (err error) {
	//now := time.Now().Unix()
	timeout := util.EXPIRE_RES_INFO
	if args.Timeout > 0 {
		timeout = time.Duration(args.Timeout) * time.Second
	}
	switch strings.ToLower(args.Methond) {
	case "get":
		if err = util.Get(args.Uri, args.AppKey, "", nil, args.ResponseBody, false, timeout); err != nil {
			glog.Error("HandUrl post_info fail! err=", err)
			return
		}
		break
	case "post":
		if err = util.Post(args.Uri, args.AppKey, args.Headers, args.Data, "", nil, args.ResponseBody, false, timeout); err != nil {
			glog.Error("HandUrl post_info fail! err=", err)
			return
		}
		break
	default:
		glog.Error("HandUrl post_info fail! not support method=", args.Methond)
		err = fmt.Errorf("appkey is empty!")
		break
	}
	return
}
