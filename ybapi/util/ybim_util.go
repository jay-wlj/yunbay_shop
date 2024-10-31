package util

import (
	"yunbay/ybapi/common"
)

type msg struct {
	Type int `json:"type"`
	//Content string      `json:"content" binding:"required,max=5000"`
	Action string                 `json:"action"`
	Data   interface{}            `json:"data"`
	Ack    bool                   `json:"ack,omitempty"` // 是否需要回复此消息
	To     []int64                `json:"to"`
	Filter map[string]interface{} `json:"filter,omitempty"` // 只针对此map进行过滤发送消息
}

type LotterysNotify struct{}

// 发送中将消息
func (t *LotterysNotify) NotifyRetMsg(id int64, status int, content string, uids []int64) (err error) {
	ext := make(map[string]interface{})
	ext["lotterys_id"] = id
	ext["status"] = status
	ext["content"] = content
	m := msg{Type: 0, Action: "lotterys_notify", To: uids, Data: ext, Ack: true}
	v := common.MQUrl{Methond: "post", AppKey: "ybim", Uri: "/man/msg/send", Data: m}
	return PublishMsg(v)
}

// 发送hash更新消息
func (t *LotterysNotify) NotifyHash(to, lotterys_record_id int64, num_hash string) (err error) {
	ext := make(map[string]interface{})
	ext["id"] = lotterys_record_id
	ext["num_hash"] = num_hash
	m := msg{Type: 0, Action: "lotterys_hash", To: []int64{to}, Data: ext, Ack: false} // 无需回复
	v := common.MQUrl{Methond: "post", AppKey: "ybim", Uri: "/man/msg/send", Data: m}
	return PublishMsg(v)
}

// 发送状态消息
func (t *LotterysNotify) NotifyStatus(lotterys_id int64, sold, status int) (err error) {
	ext := make(map[string]interface{})
	ext["id"] = lotterys_id
	ext["sold"] = sold
	ext["status"] = status
	m := msg{Type: 0, Action: "lotterys_status", To: []int64{}, Data: ext, Ack: false} // 无需回复
	filters := make(map[string]interface{})
	filters["platform"] = "web"
	m.Filter = filters // 只针对web平台发送消息
	v := common.MQUrl{Methond: "post", AppKey: "ybim", Uri: "/man/msg/send", Data: m}
	return PublishMsg(v)
}
