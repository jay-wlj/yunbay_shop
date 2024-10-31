package common

type MQMail struct {
	Receiver []string `json:"receivers"` // 接收人邮件
	Sender   string   `json:"sender"`    // 发送人邮件
	Subject  string   `json:"subject"`   // 标题
	Content  string   `json:"content"`   // 内容
	Html     string   `json:"html"`      // html内容
}

func (MQMail) Topic() string {
	return "sendmail"
}

//UsrID is the message struct
type MQUrl struct {
	Methond string            `json:"method"`
	AppKey  string            `json:"appkey"`
	Uri     string            `json:"uri"`
	Headers map[string]string `json:"headers"`
	Data    interface{}       `json:"data"`
	Timeout int               `json:"timeout"`
	MaxTrys int               `json:"maxtrys"` //
	ResponseBody string            `json:"response_body"`
}

func (MQUrl) Topic() string {
	return "mqurl"
}
