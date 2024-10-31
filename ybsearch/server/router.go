package server

import (
	base "github.com/jay-wlj/gobaselib"
	//"github.com/jay-wlj/gobaselib/db"

	"github.com/jay-wlj/gobaselib/db"
	yf "github.com/jay-wlj/gobaselib/yf"
	"runtime"

	//"yunbay/ybsearch/common"
	"yunbay/common"
	. "yunbay/ybsearch/conf"
	api_cli "yunbay/ybsearch/server/client"
	api_man "yunbay/ybsearch/server/man"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

type router_path = map[string]common.RouterInfo

var mRounter map[string]router_path // 路由组

// var v1router = make(map[string]common.RouterInfo)
// var manrouter = make(map[string]common.RouterInfo)

func handlerwrap(c *gin.Context) {
	url := base.GetUri(c)
	glog.Infof("url:%v NumCgoCall:%v NumGoroutine:%v", url, runtime.NumCgoCall(), runtime.NumGoroutine())

	var routerinfo common.RouterInfo
	var ok bool
	for _, v := range mRounter {
		if routerinfo, ok = v[url]; ok {
			break
		}
	}

	if !ok {
		glog.Errorf("err handler not found url:%v ", url)
		yf.JSON(c, 404, false, yf.ERR_SERVER_ERROR)
		return
	}

	routerinfo.Handler(c)

	// 获取事务db
	var sqldb *db.PsqlDB
	if conn, exist := c.Get("sqldao"); exist {
		if db, ok := conn.(*db.PsqlDB); ok {
			sqldb = db
		}
	}

	if sqldb != nil {
		tx := yf.GetRespTx(c)
		if tx {
			err := sqldb.Commit().Error
			if err != nil {
				glog.Errorf("commit err! %v", err)
			}
		} else {
			sqldb.Rollback()
		}
	}
	return
}

func routerRegister(r *gin.Engine, g *gin.RouterGroup, routerinfos map[string]common.RouterInfo) {
	for _, routerinfo := range routerinfos {
		switch routerinfo.Op {
		case common.HTTP_GET:
			g.GET(routerinfo.Url, handlerwrap)
		case common.HTTP_POST:
			g.POST(routerinfo.Url, handlerwrap)
		}
	}
}

func ignoresignlist(mp map[string]bool, urlprefix string, routerinfos map[string]common.RouterInfo) {
	var url string
	for _, routerinfo := range routerinfos {
		url = urlprefix
		url += routerinfo.Url

		//glog.Errorf("url:%v ignoresignlist:%v", url, !routerinfo.Checksign)
		if !routerinfo.Checksign {
			mp[url] = true
		}
	}
}

func needtokenlist(mp map[string]bool, urlprefix string, routerinfos map[string]common.RouterInfo) {
	var url string
	for _, routerinfo := range routerinfos {
		url = urlprefix
		url += routerinfo.Url
		if routerinfo.Checktoken {
			mp[url] = true
		}
	}
}

func IgnoreSignList(mp map[string]bool) {
	for k, v := range mRounter {
		ignoresignlist(mp, k, v)
	}

	mp["/debug/vars"] = true

	return
}

func NeedTokenList(mp map[string]bool) {
	for k, v := range mRounter {
		needtokenlist(mp, k, v)
	}
}

func InitRouter() {
	mRounter = make(map[string]router_path)
	mRounter["/v1"] = make(router_path)
	mRounter["/man"] = make(router_path)

	api_cli.InitRouter("/v1", mRounter["/v1"])
	api_man.InitRouter("/man", mRounter["/man"])
}

func Load(config *ApiConfig, middleware ...gin.HandlerFunc) *gin.Engine {
	r := gin.Default()

	InitRouter()

	yf.SignConfig.Debug = config.Server.Debug
	yf.SignConfig.CheckSign = config.Server.CheckSign
	yf.SignConfig.AppKeys = config.AppKeys
	IgnoreSignList(yf.SignConfig.IgnoreSignList)

	yf.TokenConfig.AccountServer = config.Servers["account"]
	yf.TokenConfig.Debug = config.Server.Debug
	NeedTokenList(yf.TokenConfig.NeedTokenList)

	//r.Use(yf.Cors)
	r.Use(yf.Sign_Check)
	r.Use(yf.Token_Check)
	r.Use(middleware...)

	for k, v := range mRounter {
		rp := r.Group(k)
		routerRegister(r, rp, v)
	}
	// v1 := r.Group("/v1")
	// RouterV1(r, v1)

	// man := r.Group("/man")
	// RouterMan(r, man)

	return r
}
