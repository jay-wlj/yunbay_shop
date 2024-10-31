package util

import (
	"yunbay/ybgoods/conf"
	"github.com/jay-wlj/gobaselib/yf"
)

var s_nsqd *yf.Nsq

func PublishMsg(v yf.NsqMsg) error {
	if s_nsqd == nil {
		s_nsqd = yf.NewNsq(conf.Config.MQUrls)
	}
	return s_nsqd.PublishMsg(v)
}

func AsyncPublishMsg(v yf.NsqMsg) {
	if s_nsqd == nil {
		s_nsqd = yf.NewNsq(conf.Config.MQUrls)
	}
	s_nsqd.AsyncPublishMsg(v)
}
