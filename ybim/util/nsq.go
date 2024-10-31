package util

import (
	"yunbay/ybim/conf"
	"github.com/jay-wlj/gobaselib/yf"
)

var s_nsqd *yf.Nsq

func PublishMsg(v yf.NsqMsg) error {
	if s_nsqd == nil {
		s_nsqd = yf.NewNsq(conf.Config.Server.MQUrls)
	}
	return s_nsqd.PublishMsg(v)
}
