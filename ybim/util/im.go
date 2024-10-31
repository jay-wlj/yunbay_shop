package util

import (
	"strconv"
	"fmt"
	"encoding/json"
	"github.com/jie123108/glog"
)

type codeSt struct {
	Code int `json:"code"`
}
type errSt struct {
	codeSt
	Desc string `json:"desc"`
}

type imToken struct {
	Token string	`json:"token"`
	Accid string	`json:"accid"`
	Name string		`json:"name"`
}

func GetIMToken(user_id int64, username, avatar, birth string, gender int, ex map[string]interface{})(accid string, token string, err error) {
	uri := "/nimserver/user/create.action"
	m := make(map[string]string)
	m["accid"] = fmt.Sprintf("%v", user_id)
	m["name"] = username
	m["icon"] = avatar
	m["bith"] = birth
	m["gender"] = fmt.Sprintf("%v", gender)
	if ex != nil {
		if exs, e := json.Marshal(ex); e == nil {
			m["ex"] = string(exs)
		}		
	}
	var v imToken
	err = post_im(uri, "netease", nil, m, "info", &v, false, EXPIRE_RES_INFO)
	if err != nil {		
		var e errSt
		if err1 := json.Unmarshal([]byte(err.Error()), &e); err1 == nil {
			// 该用户已经注册 但没有保存token了 则需要刷新token
			if e.Code == 414 && e.Desc == "already register" {	
				accid, token, err = RefleshIMToken(user_id)
				if err == nil {	// 刷新token后 需要更新一次用户信息
					UpdateIMUInfo(user_id, username, avatar, birth, gender, ex)
				}
			}
		}
		return
	}
	accid = v.Accid
	token = v.Token
	return
}

func RefleshIMToken(user_id int64)(accid string, token string, err error) {
	uri := "/nimserver/user/refreshToken.action"
	m := make(map[string]string)
	m["accid"] = fmt.Sprintf("%v", user_id)
	var v imToken
	err = post_im(uri, "netease", nil, m, "info", &v, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("GetIMToken fail! err=", err)
		return
	}
	accid = v.Accid
	token = v.Token
	return
}

// 更新用户头像等信息
func UpdateIMUInfo(user_id int64, username, avatar, birth string, gender int, ex map[string]interface{}) (err error) {
	uri := "/nimserver/user/updateUinfo.action"
	m := make(map[string]string)
	m["accid"] = fmt.Sprintf("%v", user_id)
	m["name"] = username
	if username == "" {
		m["name"] = fmt.Sprintf("%v", user_id)
	}
	
	m["icon"] = avatar
	m["bith"] = birth
	m["gender"] = fmt.Sprintf("%v", gender)
	if ex != nil {
		if exs, e := json.Marshal(ex); e == nil {
			m["ex"] = string(exs)
		}		
	}
	err = post_im(uri, "netease", nil, m, "", nil, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("UpdateIMUInfo fail! err=", err)
		return
	}
	return	
}

// 更新用户头像等信息
func QueryIMUInfo(user_ids []int64) (ret []interface{}, err error) {
	uri := "/nimserver/user/getUinfos.action"
	us := []string{}
	for _, v := range user_ids {
		us = append(us, fmt.Sprintf("%v", v))
	}
	uss, _ := json.Marshal(us)
	m := make(map[string]string)
	m["accids"] = string(uss)
	println("accids:", string(uss))
	err = post_im(uri, "netease", nil, m, "uinfos", &ret, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("UpdateIMUInfo fail! err=", err)
		return
	}
	return	
}


type MsgSt struct {
	Type int `json:"type"`
	From int64 `json:"from"`
	To int64 `json:"to" binding:"required"`
	Ope int `json:"ope"`	
	Content string `json:"content" binding:"required,max=5000"`
	PushContent string `json:"push_content" binding:"omitempty,max=150"`
	Payload string	`json:"payload"`
	Ext interface{} `json:"ext"`
}

type msgRet struct {
	Msgid int64	`json:"msgid"`
	Antispam bool `json:"antispam"`	
}
func SendMsg(v MsgSt) (msgid int64, err error) {
	uri := "/nimserver/msg/sendMsg.action"
	m := make(map[string]string)
	m["type"] = strconv.Itoa(v.Type)
	m["from"] = fmt.Sprintf("%v", v.From)
	m["to"] = fmt.Sprintf("%v", v.To)
	m["ope"] = strconv.Itoa(v.Ope)
	m["body"] = fmt.Sprintf("{\"msg\":\"%v\"}", v.Content)
	if v.PushContent != "" {
		m["pushcontent"] = v.PushContent
		if v.Payload != "" {
			m["payload"] = v.Payload
		}
		opt := make(map[string]interface{})
		opt["push"] = true
		ob, _ := json.Marshal(opt)
		m["options"] = string(ob)
	}
	if v.Ext != nil {
		eb, _ := json.Marshal(v.Ext)
		m["ext"] = string(eb)
	}
	var ret msgRet
	err = post_im(uri, "netease", nil, m, "data", &ret, false, EXPIRE_RES_INFO)
	if err != nil {
		glog.Error("SendMsg fail! err=", err)
		return
	}
	msgid = ret.Msgid
	return
}