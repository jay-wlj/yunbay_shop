
package util

import (
	"encoding/json"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"net/url"
	yf "github.com/jay-wlj/gobaselib/yf"
	."yunbay/ybim/conf"
	"github.com/jie123108/glog"
	"github.com/json-iterator/go"
	//"strings"
	"bytes"
	"time"
	"math/rand"
)

const (
	letters string ="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

)

func RandomSample(letters string, n int) string {
	b := make([]byte, n)
	llen := len(letters)
	for i := range b {
		b[i] = letters[rand.Intn(llen)]
	}
	return string(b)
}

func signim(nocne, curtime string)(sign string) {
	s := Config.IMSecret + nocne + curtime
	sign = yf.Sha1hex([]byte(s))
	return
}



func post_im(uri string, app_id string, headers map[string]string, data map[string]string, datakey string, datastu interface{}, notuserproxy bool, timeout time.Duration)(err error){	
	uri = Config.Servers[app_id] + uri
	if headers == nil {
		headers = make(map[string]string)
	}
	
	if headers == nil {
		headers = make(map[string]string)
	}	
	headers["Host"] = Config.ServerHost[app_id]
	headers["AppKey"] = Config.IMKey
	nonce := RandomSample(letters, 8)
	curtime := fmt.Sprintf("%v", time.Now().Unix())
	headers["Nonce"] = nonce
	headers["CurTime"] = curtime
	headers["CheckSum"] = signim(nonce, curtime)	

	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}

	urlval := url.Values{}
	for k, v := range data {
		urlval.Add(k, v)
	}
	body := []byte(urlval.Encode())
	headers["Content-Type"] = "application/x-www-form-urlencoded;charset=utf-8"
	res := base.HttpPost(uri, body, headers, timeout)

	glog.Infof("request [%s] status:%d ", res.ReqDebug, res.StatusCode)

	if res.StatusCode != 200 {
		glog.Errorf("request [%s] status:%d", res.ReqDebug, res.StatusCode)
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}
	var v map[string]interface{}	
	decoder := json.NewDecoder(bytes.NewBuffer(res.RawBody))
	decoder.UseNumber()

	if err = decoder.Decode(&v); err != nil {
		glog.Errorf("rquest [%s] failed! reason: %s", res.ReqDebug)
		err = fmt.Errorf("ERR_SYSTEM_ERROR")
		return
	}
	if c, ok := v["code"]; ok {
		code, _ := c.(json.Number).Int64()
		if code != 200 {
			err = fmt.Errorf(string(res.RawBody))
			return
		}
	}

	if datastu != nil {
		var buf []byte
		if datakey != ""{
			buf, err = json.Marshal(v[datakey])
		} else {
			buf, err = json.Marshal(v)
		}
		if err != nil {
			glog.Errorf("Marshal(%v) failed! err: %v", v[datakey], err)
			return
		}
		//println("getresinfos datakey",datakey, "body=", string(buf), "data:", string(res.RawBody))
		//err = json.Unmarshal(buf, datastu)
		err = jsoniter.Unmarshal(buf, datastu)
	
		if err != nil {
			glog.Errorf("getresinfos Unmarshal fail! err=%v buf:[%v]", err, string(buf))
		}
	}

	return
}
