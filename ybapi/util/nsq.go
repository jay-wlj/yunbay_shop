package util

import (
	"github.com/jay-wlj/gobaselib/yf"
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybapi/conf"
)

var s_nsqd *yf.Nsq

func PublishMsg(v yf.NsqMsg) error {
	if s_nsqd == nil {
		s_nsqd = yf.NewNsq(conf.Config.Server.MQUrls)
	}
	return s_nsqd.PublishMsg(v)
}

func DeferedPublishMsg(v yf.NsqMsg, delay time.Duration) error {
	if s_nsqd == nil {
		s_nsqd = yf.NewNsq(conf.Config.Server.MQUrls)
	}
	return s_nsqd.DeferedPublishMsg(v, delay)
}

func AsyncPublishMsg(v yf.NsqMsg) {
	if s_nsqd == nil {
		s_nsqd = yf.NewNsq(conf.Config.Server.MQUrls)
	}
	s_nsqd.AsyncPublishMsg(v)
}

type dingAtSt struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}
type dingtextSt struct {
	Content string `json:"content"`
}
type dingTalkText struct {
	Msgtype string     `json:"msgtype"`
	Text    dingtextSt `json:"text"`
	At      dingAtSt   `json:"at"`
}

func SendDingTextTalk(content string, atMobiles []string) error {
	// 忽略掉测试环境
	if conf.Config.Server.Test {
		return nil
	}
	v := dingTalkText{Msgtype: "text", Text: dingtextSt{content}, At: dingAtSt{AtMobiles: atMobiles}}
	msg := common.MQUrl{MaxTrys: 1, Methond: "post", Data: v, Uri: "https://oapi.dingtalk.com/robot/send?access_token=d8d99ddf7cdf8cc2c4bc135a4d26a599826533b780b2166586a638ea5ba8ecd1"}
	msg.ResponseBody = `{"errmsg":"ok","errcode":0}`
	return PublishMsg(msg)
}

func SendDingTextTalkToMe(content string) error {
	// 忽略掉测试环境
	if conf.Config.Server.Test {
		return nil
	}
	v := dingTalkText{Msgtype: "text", Text: dingtextSt{content}, At: dingAtSt{AtMobiles: []string{"15818717950"}}}
	msg := common.MQUrl{MaxTrys: 1, Methond: "post", Data: v, Uri: "https://oapi.dingtalk.com/robot/send?access_token=d8d99ddf7cdf8cc2c4bc135a4d26a599826533b780b2166586a638ea5ba8ecd1"}
	msg.ResponseBody = `{"errmsg":"ok","errcode":0}`
	return PublishMsg(msg)
}
