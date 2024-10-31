package util

import (
	"fmt"
	"yunbay/ybeos/conf"
	"github.com/jay-wlj/gobaselib/yf"
	"yunbay/ybeos/common"
)

var s_nsqd *yf.Nsq

func PublishMsg(v yf.NsqMsg) error {
	if s_nsqd == nil {
		if len(conf.Config.Server.MQUrls) == 0 {
			return fmt.Errorf("conf.Config.Server.MQUrls len is 0!")			
		}
		s_nsqd = yf.NewNsq(conf.Config.Server.MQUrls)
	}
	return s_nsqd.PublishMsg(v)
}

type dingAtSt struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll bool `json:"isAtAll"`
}
type dingtextSt struct {
	Content string `json:"content"`
}
type dingTalkText struct {
	Msgtype string `json:"msgtype"` 
	Text dingtextSt `json:"text"`
	At dingAtSt `json:"at"`
}
func SendDingTextTalk(content string, atMobiles []string) error {
	// 忽略掉测试环境
	if conf.Config.Server.Test {
		return nil
	}
	v := dingTalkText{Msgtype:"text", Text:dingtextSt{content}, At:dingAtSt{AtMobiles:atMobiles}}
	msg := common.MQUrl{MaxTrys:1, Methond:"post", Data:v, Uri:"https://oapi.dingtalk.com/robot/send?access_token=d8d99ddf7cdf8cc2c4bc135a4d26a599826533b780b2166586a638ea5ba8ecd1"}
	msg.ResponseBody = `{"errmsg":"ok","errcode":0}`
	return PublishMsg(msg)
}

func SendDingTextTalkToMe(content string) error {
	// 忽略掉测试环境
	if conf.Config.Server.Test {
		return nil
	}
	v := dingTalkText{Msgtype:"text", Text:dingtextSt{content}, At:dingAtSt{AtMobiles:[]string{"15818717950"}}}
	msg := common.MQUrl{MaxTrys:1, Methond:"post", Data:v, Uri:"https://oapi.dingtalk.com/robot/send?access_token=d8d99ddf7cdf8cc2c4bc135a4d26a599826533b780b2166586a638ea5ba8ecd1"}
	msg.ResponseBody = `{"errmsg":"ok","errcode":0}`
	return PublishMsg(msg)
}