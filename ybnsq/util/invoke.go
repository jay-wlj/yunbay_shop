
package util

import (
	"encoding/json"
	"fmt"
	//base "github.com/jay-wlj/gobaselib"
	yf "github.com/jay-wlj/gobaselib/yf"
	."yunbay/ybnsq/conf"
	"github.com/jie123108/glog"
	"github.com/json-iterator/go"
	//"strings"
	"time"
)

const (
	EXPIRE_RES_INFO time.Duration = 10 * time.Second
	
)

func post_info(uri string, app_id string, headers map[string]string, data interface{}, datakey string, datastu interface{}, responsebody string, notuserproxy bool, timeout time.Duration)(err error){	
	uri = Config.Servers[app_id] + uri
	appkey := Config.AppKeys[app_id]
	if headers == nil {
		headers = make(map[string]string)
	}
	
	if v, ok := Config.ServerHost[app_id]; ok {
		headers["Host"] = v
	}
	headers["X-YF-AppId"] = app_id
	headers["X-YF-rid"] = "1"
	headers["X-YF-Platform"] = "man"
	headers["X-YF-Version"] = "1.0.1"

	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}


	body, _ := json.Marshal(data)
	res := yf.YfHttpPost(uri, body, headers, timeout, appkey)
	//res := base.HttpGetJson(uri, headers, timeout)
	//glog.Infof("request [%s] status:%d cache:%v", res.ReqDebug, res.StatusCode, res.Cached)

	if res.StatusCode != 200 {
		glog.Errorf("request [%s] status:%d", res.ReqDebug, res.StatusCode)
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}

	if responsebody == "" {
		if !res.Ok {
			glog.Errorf("rquest [%s] failed! reason: %s", res.ReqDebug, res.Reason)
			err = fmt.Errorf(res.Reason)
			return
		}
	} else {
		// 返回body和期望目标不一样 则认为失败
		if string(res.RawBody) != responsebody {
			glog.Errorf("rquest [%s] failed! expected body: [%s] now body: [%s]\n", res.ReqDebug, responsebody, string(res.RawBody))
			err = fmt.Errorf("not expected body")
			return
		}
	}


	if datastu != nil {
		var buf []byte
		if datakey != ""{
			buf, err = json.Marshal(res.Data[datakey])
		} else if responsebody != "" {
			buf = []byte(res.RawBody)
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


func get_info(uri string, app_id string, datakey string, datastu interface{}, responsebody string, notuserproxy bool, timeout time.Duration)(err error){
	uri = Config.Servers[app_id] + uri
	appkey := Config.AppKeys[app_id]
	headers := make(map[string]string)
	if v, ok := Config.ServerHost[app_id]; ok {
		headers["Host"] = v
	}
	
	headers["X-YF-AppId"] = app_id
	headers["X-YF-rid"] = "1"
	headers["X-YF-Platform"] = "man"
	headers["X-YF-Version"] = "1.0.1"

	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}

	res := yf.YfHttpGet(uri, headers, timeout, appkey)
	//res := base.HttpGetJson(uri, headers, timeout)
	glog.Infof("request [%s] status:%d cache:%v", res.ReqDebug, res.StatusCode, res.Cached)

	if res.StatusCode != 200 {
		glog.Errorf("request [%s] status:%d", res.ReqDebug, res.StatusCode)
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}
	if responsebody == "" {
		if !res.Ok {
			glog.Errorf("rquest [%s] failed! reason: %s", res.ReqDebug, res.Reason)
			err = fmt.Errorf(res.Reason)
			return
		}
	} else {
		// 返回body和期望目标不一样 则认为失败
		if string(res.RawBody) != responsebody {
			glog.Errorf("rquest [%s] failed! expected body: [%s] now body: [%s]\n", res.ReqDebug, responsebody, string(res.RawBody))
			err = fmt.Errorf("not expected body")
			return
		}
	}

	if datastu != nil {
		var buf []byte
		if datakey != ""{
			buf, err = json.Marshal(res.Data[datakey])
		} else if responsebody != "" {
			buf = []byte(res.RawBody)
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


func Post(uri string, app_id string, headers map[string]string, data interface{}, datakey string, datastu interface{}, responsebody string, notuserproxy bool, timeout time.Duration)(err error){	
	return post_info(uri, app_id, headers, data, datakey, datastu, responsebody, notuserproxy, timeout)
}


func Get(uri string, app_id string, datakey string, datastu interface{}, responsebody string, notuserproxy bool, timeout time.Duration)(err error){
	return get_info(uri, app_id, datakey, datastu, responsebody, notuserproxy, timeout)
}
