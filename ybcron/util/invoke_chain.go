
package util

import (
	"encoding/json"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	yf "github.com/jay-wlj/gobaselib/yf"
	."yunbay/ybcron/conf"
	"github.com/jie123108/glog"
	"github.com/json-iterator/go"
	"time"
	"math/rand"
)

const (
	letters string ="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	//secret string = "abcdefghijklmn"
)

func RandomSample(letters string, n int) string {
	b := make([]byte, n)
	llen := len(letters)
	for i := range b {
		b[i] = letters[rand.Intn(llen)]
	}
	return string(b)
}

func signtx(body []byte, ts, nocne, secret string)(sign, s string) {
	s = string(body) + ts + nocne + secret
	sign = yf.Md5hex([]byte(s))
	return
}

func post_chain(uri string, app_id string, headers map[string]string, data interface{}, datakey string, datastu interface{}, notuserproxy bool, timeout time.Duration)(err error){	
	uri = Config.Servers[app_id] + uri
	if headers == nil {
		headers = make(map[string]string)
	}	
	headers["Host"] = Config.ServerHost[app_id]
	headers["x-bitex-ts"] = fmt.Sprintf("%v", time.Now().Unix())
	headers["x-bitex-nonce"] = RandomSample(letters, 8)
	
	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}

	body, _ := json.Marshal(data)
	var signature string
	headers["x-bitex-sign"], signature = signtx(body, headers["x-bitex-ts"], headers["x-bitex-nonce"], Config.ThirdAccount["chain"].Secret)
	res := base.HttpPostJson(uri, body, headers, timeout)	
	//res := base.HttpGetJson(uri, headers, timeout)
	glog.Infof("request [%s] status:%d cache:%v", res.ReqDebug, res.StatusCode, res.Cached)

	if res.StatusCode == 401 || res.Reason == "INVALID_SIGN" {
		fmt.Printf("signature [",headers["x-bitex-sign"],"]  [[\n", signature, "\n]]")
	}
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


func get_chain(uri string, app_id string, datakey string, datastu interface{}, notuserproxy bool, timeout time.Duration)(err error){
	uri = Config.Servers[app_id] + uri

	headers := make(map[string]string)
	headers["Host"] = Config.ServerHost[app_id]
	headers["x-bitex-ts"] = fmt.Sprintf("%v", time.Now().Unix())
	headers["x-bitex-nonce"] = RandomSample(letters, 8)

	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}

	body := []byte{}
	var signature string
	headers["x-bitex-sign"], signature = signtx(body, headers["x-bitex-ts"], headers["x-bitex-nonce"], Config.ThirdAccount["chain"].Secret)
	res := base.HttpGetJson(uri, headers, timeout)	
	//res := base.HttpGetJson(uri, headers, timeout)
	glog.Infof("request [%s] status:%d cache:%v", res.ReqDebug, res.StatusCode, res.Cached)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", headers["x-bitex-sign"], signature)
	}


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
