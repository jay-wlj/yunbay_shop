package common

import (
	"reflect"
	"strings"

	"github.com/jie123108/glog"

	"github.com/gin-gonic/gin"
)

const (
	HTTP_GET     = 1
	HTTP_POST    = 2
	HTTP_OPTIONS = 3
)

type RouterInfo struct {
	Op         int
	Url        string
	Checksign  bool
	Checktoken bool
	//Handler    gin.HandlerFunc
	HandlerObj interface{}
	//RouterHandler RouterHandlerFunc
}

func (t *RouterInfo) Handler(c *gin.Context) {
	switch v := t.HandlerObj.(type) {
	case gin.HandlerFunc:
		v(c)
	case func(*gin.Context):
		v(c)
	default:
		rv := reflect.ValueOf(t.HandlerObj)
		if rv.Kind() == reflect.Ptr {
			if mv := rv.MethodByName(t.getMethodName()); mv.IsValid() {
				mv.Call([]reflect.Value{reflect.ValueOf(c)})
				return
			}
		}
		glog.Error("Handler fail! rounterinfo=", *t)
	}
}

func (t *RouterInfo) getMethodName() string {
	rn := strings.LastIndex(t.Url, "/")
	name_str := t.Url[rn+1:]
	rq := strings.Index(name_str, "?")
	if rq >= 0 {
		name_str = name_str[:rq]
	}
	// 换成跎峰命名
	name_str = strFirstToUpper(name_str)
	return name_str
}

/**
 * 字符串首字母转化为大写 ios_bbbbbbbb -> IosBbbbbbbbb
 */
func strFirstToUpper(str string) string {
	ss := strings.Split(str, "_")
	for i, _ := range ss {
		sb := []byte(ss[i])
		if len(sb) > 0 {
			sb[0] = strings.ToUpper(string(sb[0]))[0]
		}
		ss[i] = string(sb)
	}
	var method string
	for _, v := range ss {
		method += v
	}
	return method
}

func Routerlistadd(ver string, routerinfos map[string]RouterInfo, routerinfo RouterInfo) map[string]RouterInfo {
	url := ver
	url += routerinfo.Url
	routerinfos[url] = routerinfo
	return routerinfos
}

func Routeraddlist(ver string, routerinfos map[string]RouterInfo, infos []RouterInfo) map[string]RouterInfo {
	for _, routerinfo := range infos {
		url := ver
		url += routerinfo.Url
		routerinfos[url] = routerinfo
	}

	return routerinfos
}
