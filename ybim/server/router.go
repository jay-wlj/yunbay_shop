package server

import (
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	yf "github.com/jay-wlj/gobaselib/yf"
	"runtime"
	"yunbay/ybim/common"
	. "yunbay/ybim/conf"
	api_cli "yunbay/ybim/server/client"
	api_man "yunbay/ybim/server/man"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

var v1router = make(map[string]common.RouterInfo)
var manrouter = make(map[string]common.RouterInfo)

func handlerwrap(c *gin.Context) {
	url := base.GetUri(c)
	glog.Infof("url:%v NumCgoCall:%v NumGoroutine:%v", url, runtime.NumCgoCall(), runtime.NumGoroutine())
	routerinfo, ok := v1router[url]
	if !ok {
		routerinfo, ok = manrouter[url]
	}
	if !ok {
		glog.Errorf("err handler not found url:%v v2router:%v maprouter:%v", url, v1router, manrouter)
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

func InitRouter() {
	api_cli.InitClientRouter("/v1", v1router)
	api_man.InitManagerRouter("/man", manrouter)
}

func IgnoreSignList(mp map[string]bool) {
	ignoresignlist(mp, "/v1", v1router)
	ignoresignlist(mp, "/man", manrouter)

	mp["/debug/vars"] = true

	return
}

func NeedTokenList(mp map[string]bool) {
	needtokenlist(mp, "/v1", v1router)
	needtokenlist(mp, "/man", manrouter)
}

func RouterV1(r *gin.Engine, g *gin.RouterGroup) {
	routerRegister(r, g, v1router)
}

func RouterMan(r *gin.Engine, g *gin.RouterGroup) {
	routerRegister(r, g, manrouter)
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

	v1 := r.Group("/v1")
	RouterV1(r, v1)

	man := r.Group("/man")
	RouterMan(r, man)

	return r
}
