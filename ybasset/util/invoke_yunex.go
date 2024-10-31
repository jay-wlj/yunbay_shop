
package util

import (
	//"strconv"
	"strings"
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
	"bytes"
)

func signyunex(body []byte, vals url.Values, ts, nocne, secret string)(sign, s string) {	
	params := vals.Encode()
	if len(body) > 0 {		
		params += string(body)		
	} 			

	s = params + ts + nocne + secret
	sign = yf.Sha256hex([]byte(s))
	return
}

type okJson struct {
	Error error 
	StatusCode int
	Ok     bool                   `json:"ok"`
	Reason interface{}            `json:"reason"`
	Data   map[string]interface{} `json:"data"`
}

func okJsonParse(ret *base.Resp)(res *okJson) {
	res = &okJson{StatusCode:ret.StatusCode}	
	decoder := json.NewDecoder(bytes.NewBuffer(ret.RawBody))
	decoder.UseNumber()
	err := decoder.Decode(&res)
	if err != nil {
		glog.Errorf("Invalid json [%s] err: %v", string(ret.RawBody), err)
		res.Error = err
		res.Reason = yf.ERR_SERVER_ERROR
		res.StatusCode = 500
		return res
	}
	if !res.Ok && res.Reason != "" && res.Error == nil {
		if s, ok := res.Reason.(string); ok {
			res.Error = fmt.Errorf(s)
		} else {
			res.Error = fmt.Errorf(yf.ERR_SERVER_ERROR)
		}		
	}

	return res
}

func post_yunex(uri string, app_id string, headers map[string]string, data interface{}, datakey string, datastu interface{}, notuserproxy bool, timeout time.Duration)(err error){	
	uri = Config.Servers[app_id] + uri
	if headers == nil {
		headers = make(map[string]string)
	}	
	if v, ok := Config.ServerHost[app_id]; ok {
		headers["Host"] = v
	}
	a := Config.ThirdAccount["yunex"]
	headers["-x-ts"] = fmt.Sprintf("%v", time.Now().Unix())
	headers["-x-nonce"] = RandomSample(letters, 8)
	headers["-x-key"] = a.Key
	headers["Content-Type"] = "application/json"
	
	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}

	vals := url.Values{}
	body, _ := json.Marshal(data)
	var signature string
	headers["-x-sign"], signature = signyunex(body, vals, headers["-x-ts"], headers["-x-nonce"], a.Secret)
	ret := base.HttpPost(uri, body, headers, timeout)	
	res := okJsonParse(ret)
	//res := base.HttpGetJson(uri, headers, timeout)
	//glog.Infof("request [%s] status:%d ", ret.ReqDebug, ret.StatusCode)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", headers["x-sign"], signature)
	}
	if res.StatusCode != 200 {
		glog.Errorf("request [%s] status:%d", ret.ReqDebug, ret.StatusCode)
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}
	if !res.Ok {
		glog.Errorf("rquest [%s] failed! reason: %s", ret.ReqDebug, res.Reason)
		switch res.Reason.(type) {
		case string:
			err = fmt.Errorf(res.Reason.(string))
		default:
			err = fmt.Errorf(fmt.Sprintf("%v", res.Reason))
		}	
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


func get_yunex(uri string, app_id string, datakey string, datastu interface{}, notuserproxy bool, timeout time.Duration)(err error){
	uri = Config.Servers[app_id] + uri

	headers := make(map[string]string)
	if v, ok := Config.ServerHost[app_id]; ok {
		headers["Host"] = v
	}
	a := Config.ThirdAccount["yunex"]
	headers["-x-ts"] = fmt.Sprintf("%v", time.Now().Unix())
	headers["-x-nonce"] = RandomSample(letters, 8)
	headers["-x-key"] = a.Key
	headers["Content-Type"] = "application/json"

	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}
	
	r, _ := url.Parse(uri)
	var signature string
	headers["-x-sign"], signature = signyunex(nil, r.Query(), headers["-x-ts"], headers["-x-nonce"], a.Secret)
	ret := base.HttpGet(uri, headers, timeout)	
	res := okJsonParse(ret)
	//res := base.HttpGetJson(uri, headers, timeout)
	glog.Infof("request [%s] status:%d", ret.ReqDebug, ret.StatusCode)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", headers["x-sign"], signature)
	}

	if res.StatusCode != 200 {
		glog.Errorf("request [%s] status:%d", ret.ReqDebug, ret.StatusCode)
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}
	if !res.Ok {
		glog.Errorf("rquest [%s] failed! reason: %s", ret.ReqDebug, res.Reason)
		switch res.Reason.(type) {
		case string:
			err = fmt.Errorf(res.Reason.(string))		
		default:
			err = fmt.Errorf(fmt.Sprintf("%v", res.Reason))
		}	
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


/**
 * @params string method 请求方式，取值：GET、POST
 */
func RequestYunExApi(uri string, method string, app_id string, headers map[string]string, data interface{}, datakey string, datastu interface{}, notuserproxy bool, timeout time.Duration) (err error) {
	uri = Config.Servers[app_id] + uri
	if headers == nil {
		headers = make(map[string]string)
	}
	if v, ok := Config.ServerHost[app_id]; ok {
		headers["Host"] = v
	}

	headers["Content-Type"] = "application/json"
	headers["x-bitex-ts"] = fmt.Sprintf("%v", time.Now().Unix())
	headers["x-bitex-nonce"] = RandomSample(letters, 12)
	if notuserproxy {
		headers["X-Not-Use-Proxy"] = "true"
	}
	var signString string
	var body []byte
	if strings.ToUpper(method) == "POST" { // POST请求，需要传递data数据
		body, _ = json.Marshal(data)
		signString = string(body)
	} else { // GET请求，uri中包括请求参数
		r, _ := url.Parse(uri)
		signString = r.Query().Encode()
	}

	secret := Config.ThirdAccount["yunex"].Ext["transfer_secret"]
	// 第二套签名算法。对数据进行签名（因为比较简单就没有单独函数出来）
	signString += headers["x-bitex-ts"] + headers["x-bitex-nonce"] + secret  // TODO  app secrete从配置读取
	//println("sign string:", signString)
	signString = yf.Md5hex([]byte(signString))

	headers["x-bitex-sign"] = signString
	var response *base.Resp
	if (method == "POST") {
		response = base.HttpPost(uri, body, headers, timeout)
	} else {
		response = base.HttpGet(uri, headers, timeout)
	}
	res := okJsonParse(response)
	if !res.Ok {
		glog.Errorf("rquest [%s] failed! reason: %s", response.ReqDebug, res.Reason)
		switch res.Reason.(type) {
		case string:
			err = fmt.Errorf(res.Reason.(string))		
		default:
			err = fmt.Errorf(fmt.Sprintf("%v", res.Reason))
		}	
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

func Yunex_SignCheck(c *gin.Context) bool {
	// 测试工具使用。
	if Config.Server.Debug && !Config.Server.CheckSign {
		return true
	}

	body, _ := base.GetPostJsonData(c)
	c.Request.ParseForm()	
	headers := c.Request.Header

	req_signs := headers["X-Sign"]
	req_ts := headers["X-Ts"]
	req_nonce := headers["X-Nonce"]
	req_keys := headers["X-Key"]
	if len(req_signs) != 1 || len(req_ts) != 1 || len(req_nonce) != 1 || len(req_keys) != 1 {
		glog.Errorf("Yunex_SignCheck find %d Sign value..", len(req_signs))
		c.JSON(401, gin.H{"ok": false, "reason": "INVALID_SIGN"})
		c.Abort()
		return false
	}
	req_sign := req_signs[0]
	ts := req_ts[0]
	nonce := req_nonce[0]
	key := req_keys[0]

	// 判断时间戳
	req_time, err := base.StringToInt64(ts)
	if err != nil || time.Now().Unix()-req_time > 600 {
		c.JSON(401, gin.H{"ok": false, "reason": "INVALID_TIME"})
		c.Abort()
		return false
	}


	// if c.Request.Method == "GET" {		
	// 	body = []byte(c.Request.URL.Query().Encode())
	// }
	params := c.Request.URL.RawQuery
	if len(body) > 0 {	
		params += string(body)	
	} 
	//signature, SignStr := signyunex(body, vals, ts, nonce, Config.AppKeys[key])
	signature, SignStr := signtx([]byte(params), ts, nonce, Config.AppKeys[key])
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
