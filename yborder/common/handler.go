package common

import (
	"github.com/gin-gonic/gin"
)

const (
	HTTP_GET  = 1
	HTTP_POST = 2
)

type RouterInfo struct {
	Op         int
	Url        string
	Checksign  bool
	Checktoken bool
	Handler    gin.HandlerFunc
	//RouterHandler RouterHandlerFunc
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
