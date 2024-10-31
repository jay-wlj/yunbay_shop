
package utils

import (
	"encoding/json"
	"fmt"
	//base "github.com/jay-wlj/gobaselib"
	yf "github.com/jay-wlj/gobaselib/yf"
	."yunbay/ybapi/conf"
	"github.com/jie123108/glog"
	"github.com/json-iterator/go"
	//"strings"
	"time"
)

const (
	EXPIRE_RES_INFO time.Duration = 10 * time.Second
	
)

func PostInfo(uri string, app_id string, headers map[string]string, data interface{}, datakey string, datastu interface{}, notuserproxy bool, timeout ...time.Duration)(err error){	
	uri = Config.Servers[app_id] + uri
	appkey := Config.AppKeys[app_id]
	if headers == nil {
		headers = make(map[string]string)
	}
	
	headers["Host"] = Config.ServerHost[app_id]
	headers["X-YF-AppId"] = app_id
	headers["X-YF-rid"] = "1"
	headers["X-YF-Platform"] = "man"
	headers["X-YF-Version"] = "1.0.1"

	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}
	tt := EXPIRE_RES_INFO
	if timeout != nil && len(timeout) > 0 {
		tt = timeout[0]
	}

	body, _ := json.Marshal(data)
	res := yf.YfHttpPost(uri, body, headers, tt, appkey)
	//res := base.HttpGetJson(uri, headers, timeout)
	//glog.Infof("request [%s] status:%d cache:%v", res.ReqDebug, res.StatusCode, res.Cached)

	if res.StatusCode != 200 {
		glog.Errorf("request [%s] status:%d", res.ReqDebug, res.StatusCode)
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}

	if !res.Ok {
		glog.Errorf("rquest [%s] failed! reason: %s", res.ReqDebug, res.Reason)
		err = fmt.Errorf(res.Reason)
		return
	}

	if datastu != nil {
		var buf []byte
		if datakey != ""{
			buf, err = json.Marshal(res.Data[datakey])
		} else {
			buf, err = json.Marshal(res.Data)
		}
		if err != nil {
			glog.Errorf("Marshal(%v) failed! err: %v", res.Data[datakey], err)
			return
		}
		//glog.Error("getresinfos body=%v", string(buf))
		//err = json.Unmarshal(buf, datastu)
		err = jsoniter.Unmarshal(buf, datastu)
	
		if err != nil {
			glog.Errorf("getresinfos Unmarshal fail! err=%v buf:[%v]", err, string(buf))
		}
	}

	return
}


func GetInfo(uri string, app_id string, datakey string, datastu interface{}, notuserproxy bool, timeout ...time.Duration)(err error){
	uri = Config.Servers[app_id] + uri
	appkey := Config.AppKeys[app_id]
	headers := make(map[string]string)
	headers["Host"] = Config.ServerHost[app_id]
	headers["X-YF-AppId"] = app_id
	headers["X-YF-rid"] = "1"
	headers["X-YF-Platform"] = "man"
	headers["X-YF-Version"] = "1.0.1"

	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}

	tt := EXPIRE_RES_INFO
	if timeout != nil && len(timeout) > 0 {
		tt = timeout[0]
	}
	res := yf.YfHttpGet(uri, headers, tt, appkey)
	//res := base.HttpGetJson(uri, headers, timeout)
	glog.Infof("request [%s] status:%d cache:%v", res.ReqDebug, res.StatusCode, res.Cached)

	if res.StatusCode != 200 {
		glog.Errorf("request [%s] status:%d", res.ReqDebug, res.StatusCode)
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}

	if !res.Ok {
		glog.Errorf("rquest [%s] failed! reason: %s", res.ReqDebug, res.Reason)
		err = fmt.Errorf(res.Reason)
		return
	}

	if datastu != nil {
		var buf []byte
		if datakey != ""{
			buf, err = json.Marshal(res.Data[datakey])
		} else {
			buf, err = json.Marshal(res.Data)
		}
		if err != nil {
			glog.Errorf("Marshal(%v) failed! err: %v", res.Data[datakey], err)
			return
		}
		// glog.Error("getresinfos body=%v", string(buf))
		// err = json.Unmarshal(buf, datastu)
		err = jsoniter.Unmarshal(buf, datastu)
	

		if err != nil {
			glog.Errorf("getresinfos Unmarshal fail! err=%v buf:[%v]", err, string(buf))
		}
	}

	return
}