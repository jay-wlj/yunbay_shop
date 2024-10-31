package dao

import (
	"errors"
	"github.com/jay-wlj/gobaselib/cache"
	"yunbay/yborder/common"
	"yunbay/yborder/util"

	"github.com/jie123108/glog"
)

type UniqueKey string

const (
	RandStrings string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	UNIQUE_KEYS string = "unique_keys"
)

func GetUniqueKey() (s string, err error) {
	s = util.RandomSample(RandStrings, 12)

	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}
	for {
		// 确保唯一key没有被使用
		if !ch.SIsMember(UNIQUE_KEYS, s).Val() {
			break
		}
		s = util.RandomSample(RandStrings, 12)
	}
	return
}

func SaveUniqueKey(s string) (err error) {
	var ch *cache.RedisCache
	ch, err = cache.GetReader(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
		return
	}
	if 0 == ch.SAdd(UNIQUE_KEYS, s).Val() {
		err = errors.New("key exists")
	}
	return
}
