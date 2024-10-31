package common

import (

)

type MQUserId struct {
	UserId int64 `json:"user_id"`
}

type MQMtPic struct {
	UserId int64 `json:"user_id"`
	Pid int64 `json:"pid"`
}


//UsrID is the message struct
type MQUrl struct {
	Methond string `json:"method"`
	AppKey string `json:"appkey"`
	Uri string `json:"uri"`
	Headers map[string]string `json:"headers"`
	Data interface{} `json:"data"`
	Timeout int `json:"timeout"`
	MaxTrys int16 `json:"maxtrys"`	// 
	Delay string `json:"delay"` // 间隔多久重新排队 ns
	ResponseBody string `json:"response_body"`
}

func (MQUrl)Topic() string {
	return "mqurl"
}

type MQMail struct {
	Receiver []string `json:"receivers"` 	 	// 接收人邮件
	Sender string `json:"sender"`			// 发送人邮件
	Subject string `json:"subject"`    		// 标题
	Content string `json:"content"`    		// 内容
	Html	string `json:"html"`			// html内容
}

func (MQMail)Topic() string {
	return "sendmail"
}