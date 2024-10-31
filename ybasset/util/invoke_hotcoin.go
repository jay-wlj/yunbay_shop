
package util

import (
	"strings"
	//"strings"
	"net/url"
	"fmt"
	"encoding/json"
	base "github.com/jay-wlj/gobaselib"
	yf "github.com/jay-wlj/gobaselib/yf"
	."yunbay/ybasset/conf"
	"github.com/jie123108/glog"
	"github.com/json-iterator/go"
	"github.com/gin-gonic/gin"
	"time"
	//"math/rand"
)

func signgethotcoin(data []byte, ts, nocne, secret string)(sign, s string) {	
	s = string(data) + ts + nocne + secret
	sign = yf.Md5hex([]byte(s))
	return
}

// func signtx(body []byte, ts, nocne string)(sign, s string) {
// 	s = string(body) + ts + nocne + secret
// 	sign = yf.Md5hex([]byte(s))
// 	return
// }

func post_hotcoin(uri string, app_id string, headers map[string]string, data interface{}, datakey string, datastu interface{}, notuserproxy bool, timeout time.Duration)(err error){	
	uri = Config.Servers[app_id] + uri
	if headers == nil {
		headers = make(map[string]string)
	}	
	headers["Host"] = Config.ServerHost[app_id]
	headers["x-ts"] = fmt.Sprintf("%v", time.Now().Unix())
	headers["x-nonce"] = RandomSample(letters, 8)
	headers["Content-Type"] = "application/json"
	
	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}

	body, _ := json.Marshal(data)
	var signature string
	headers["x-sign"], signature = signgethotcoin(body, headers["x-ts"], headers["x-nonce"], Config.ThirdAccount["hotcoin"].Secret)
	res := base.HttpPost(uri, body, headers, timeout)	
	//res := base.HttpGetJson(uri, headers, timeout)
	//glog.Infof("request [%s] status:%d ", res.ReqDebug, res.StatusCode)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", headers["x-sign"], signature)
	}

	if datastu != nil {
		var buf []byte
		buf = res.RawBody
		if err != nil {
			glog.Errorf("Marshal(%v) failed! err: %v", buf, err)
			return
		}
		//glog.Error("getresinfos body=%v", string(buf))
		//err = json.Unmarshal(buf, datastu)
		err = jsoniter.Unmarshal(buf, datastu)
	
		if err != nil {
			glog.Errorf("getresinfos Unmarshal fail! err=%v buf:[%v]", err, string(buf))
		}
		s := string(buf)
		if strings.Index(s, "INVALID_SIGN") > -1 {
			fmt.Println("signature [",headers["x-sign"],"]  [[\n", signature, "\n]]")
		}
	}

	if res.StatusCode != 200 {
		glog.Errorf("request [%s] status:%d body:%v", res.ReqDebug, res.StatusCode, string(res.RawBody))
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}
	return
}


func get_hotcoin(uri string, app_id string, datakey string, datastu interface{}, notuserproxy bool, timeout time.Duration)(err error){
	uri = Config.Servers[app_id] + uri

	headers := make(map[string]string)
	headers["Host"] = Config.ServerHost[app_id]
	headers["x-ts"] = fmt.Sprintf("%v", time.Now().Unix())
	headers["x-nonce"] = RandomSample(letters, 8)
	headers["Content-Type"] = "application/json"

	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}
	
	var data []byte
	r, _ := url.Parse(uri)
	if r != nil {
		data = []byte(r.Query().Encode())	
	}
	var signature string
	headers["x-sign"], signature = signgethotcoin(data, headers["x-ts"], headers["x-nonce"], Config.ThirdAccount["hotcoin"].Secret)
	res := base.HttpGet(uri, headers, timeout)	
	//res := base.HttpGetJson(uri, headers, timeout)
	glog.Infof("request [%s] status:%d", res.ReqDebug, res.StatusCode)

	if res.StatusCode == 401 {
		fmt.Println("signature [",headers["x-sign"],"]  [[\n", signature, "\n]]")
	}

	if datastu != nil && len(res.RawBody) > 0 {
		var buf []byte
		buf = res.RawBody
		if err != nil {
			glog.Errorf("Marshal(%v) failed! err: %v", buf, err)
			return
		}

		err = jsoniter.Unmarshal(buf, datastu)
		s := string(buf)
		if err != nil {		
			glog.Errorf("getresinfos Unmarshal fail! err=%v buf:[%v]\n", err, s)
		}
		if strings.Index(s, "INVALID_SIGN") > -1 {
			fmt.Println("signature [",headers["x-sign"],"]  [[\n", signature, "\n]]")
		}
	}
	
	if res.StatusCode != 200 {
		glog.Errorf("request [%s] status:%d body:%v", res.ReqDebug, res.StatusCode, string(res.RawBody))
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}

	return
}


func HotCoin_SignCheck(c *gin.Context) bool {
	body, _ := base.GetPostJsonData(c)

	c.Request.ParseForm()	
	headers := c.Request.Header
	// hs, _ := json.Marshal(headers)
	// fmt.Println("headers:", string(hs))

	req_signs := headers["X-Sign"]
	req_ts := headers["X-Ts"]
	req_nonce := headers["X-Nonce"]
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
	// 测试工具使用。
	if Config.Server.Debug && !Config.Server.CheckSign {
		return true
	}

	if c.Request.Method == "GET" {
		//body = []byte(c.Request.URL.Query().Encode())	// url的参数按key顺序排好的参数串
		body = []byte(c.Request.URL.RawQuery)	// url的原生参数串
		//glog.Info("url query encod:", c.Request.URL.Query().Encode())
		//glog.Info("url rawquery:", c.Request.URL.RawQuery)
	}
	signature, SignStr := signtx(body, ts, nonce, Config.ThirdAccount["hotcoin"].Secret)
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