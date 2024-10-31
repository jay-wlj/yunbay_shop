
package util

import (
	"encoding/json"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	yf "github.com/jay-wlj/gobaselib/yf"
	."yunbay/ybasset/conf"
	"github.com/jie123108/glog"
	"github.com/json-iterator/go"
	"github.com/gin-gonic/gin"
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
	//glog.Infof("request [%s] status:%d cache:%v", res.ReqDebug, res.StatusCode, res.Cached)

	if res.StatusCode == 401 {
		fmt.Println("signature [",headers["x-bitex-sign"],"]  [[\n", signature, "\n]]")
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

	if res.StatusCode == 401 || res.Reason == "INVALID_SIGN" {
		fmt.Println("signature [",headers["x-bitex-sign"],"]  [[\n", signature, "\n]]")
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


func Chain_SignCheck(c *gin.Context) bool {
	// 测试工具使用。
	if Config.Server.Debug && !Config.Server.CheckSign {
		return true
	}
	
	body, _ := base.GetPostJsonData(c)

	c.Request.ParseForm()	
	headers := c.Request.Header

	req_signs := headers["X-Bitex-Sign"]
	req_ts := headers["X-Bitex-Ts"]
	req_nonce := headers["X-Bitex-Nonce"]
	if len(req_signs) != 1 || len(req_ts) != 1 || len(req_nonce) != 1{
		glog.Errorf("find %d Sign value..", len(req_signs))
		c.JSON(401, gin.H{"ok": false, "reason": "INVALID_SIGN"})
		c.Abort()
		return false
	}
	req_sign := req_signs[0]
	ts := req_ts[0]
	nonce := req_nonce[0]
	// glog.Errorf("req_sign: %s", req_sign)

	if v, ok := Config.Server.Ext["chainsign"]; ok {
		if b, ok := v.(bool); ok && !b {
			return true
		}
	}

	signature, SignStr := signtx(body, ts, nonce, Config.ThirdAccount["chain"].Secret)
	if signature != req_sign {
		glog.Errorf("req_sign: [%s] != calc_sign: [%s] \nSignStr [[%s]]", req_sign,
			signature, SignStr)
		glog.Infof("req body len: %d", len(body))
		if Config.Server.Debug && len(body) < 100 {
			glog.Infof("body: [[%v]]", string(body))
		}
		c.JSON(401, gin.H{"ok": false, "reason": "INVALID_SIGN", "SignStr": SignStr})
		c.Abort()
		return false
	}
	return true
}