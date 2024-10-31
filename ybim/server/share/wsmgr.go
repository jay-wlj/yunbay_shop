package share

import (
	"encoding/json"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/melody"
	"reflect"
	"sync"
	"yunbay/ybim/common"
	"yunbay/ybim/conf"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

const (
	MSG_TYPE_NOTIFY  = iota // 文本消息
	MSG_TYPE_CONTROL        // 控制消息
)

const (
	MSG_ACK       string = "ack"
	MSG_BRAODCAST string = "broadcast"
)

type WsMgr struct {
	*melody.Melody
	ucs    map[int64]*melody.Session
	rwmutx *sync.RWMutex
}

var (
	once    sync.Once
	g_wsmgr *WsMgr
)

type Msg struct {
	Id     int64       `json:"id"`            // 消息id
	Type   int         `json:"type"`          // 消息类型
	Action string      `json:"action"`        // 动作
	Data   interface{} `json:"data"`          // 动作数据
	Ack    bool        `json:"ack,omitempty"` // 是否需要回复此消息
}

type MsgReq struct {
	Msg
	To     []int64                `json:"to" binding:"required"` // 指定用户发送，空为广播消息
	Filter map[string]interface{} `json:"filter"`                 // 广播过滤，To为空时才有效
}

func (m *MsgReq) Send() (err error) {
	return GetWsMgr().PublishMsg(*m)
}

func GetWsMgr() *WsMgr {
	once.Do(func() {
		if g_wsmgr == nil {
			g_wsmgr = &WsMgr{Melody: melody.New(), ucs: make(map[int64]*melody.Session), rwmutx: &sync.RWMutex{}}
		}
	})

	g_wsmgr.HandleMessage(g_wsmgr.OnMessage)
	g_wsmgr.HandleConnect(g_wsmgr.OnConnect)
	g_wsmgr.HandleDisconnect(g_wsmgr.OnDisconnectConnect)
	g_wsmgr.HandleSentMessage(g_wsmgr.OnSentMessage)
	return g_wsmgr
}

func (t *WsMgr) Close() {
	t.Melody.Close()

}

func (t *WsMgr) OnMessage(s *melody.Session, msg []byte) {
	var v Msg
	if err := json.Unmarshal(msg, &v); err != nil {
		glog.Error("recv unknown msg:", string(msg))
		return
	}
	switch v.Action {
	case MSG_ACK: // 确认消息
		if user_id, ok := s.GetI("user_id"); ok {
			// 更新该用户已收到消息 TODO  有待优化 放协程处理
			go func() {
				if err := db.GetDB().Model(&common.IMgs{}).Where("id =? ", v.Id).Updates(base.Maps{"uids": gorm.Expr("array_remove(uids, ?)", user_id), "ok_uids": gorm.Expr("array_append(ok_uids, ?)", user_id)}).Error; err != nil {
					glog.Error("OnMessage update fail! user_id:", user_id, " id:", v.Id, " err=", err)
				}
				if conf.Config.Server.Debug {
					if uid, ok := conf.Config.Server.Ext["mgr_uid"].(int); ok {
						if s := t.get(int64(uid)); s != nil {
							m := make(map[string]interface{})
							m["user_id"] = user_id
							m["msg"] = string(msg)
							buf, _ := json.Marshal(m)
							t.BroadcastOne(buf, s)
						}
					}
				}
			}()
		}
	case MSG_BRAODCAST: //广播消息
		var m MsgReq
		buf, err := json.Marshal(v.Data)
		if err != nil {
			return
		}
		fmt.Println(string(buf))
		if err = json.Unmarshal(buf, &m); err != nil {
			return
		}

		t.PublishMsg(m)

	default:
		break
	}
}

// 有新链接进来
func (t *WsMgr) OnConnect(s *melody.Session) {
	//t.BroadcastOne(g_wsmgr.ok_msg("{}"), s)
	if user_id, ok := s.GetI("user_id"); ok {
		t.rwmutx.Lock()
		t.ucs[user_id] = s
		t.rwmutx.Unlock()

		go func() {
			// 确保需要应答的消息发送出去，待优化
			ms := []common.IMgs{}
			err := db.GetDB().Find(&ms, "(msg->>'ack')::bool is true and array_position(uids, ?) is not null and array_position(ok_uids, ?) is null", user_id, user_id).Error
			if err != nil {
				glog.Error("OnConnect fail! err=", err)
				return
			}
			// 发送用户未接收的消息
			ok_ids := []int64{}
			for _, v := range ms {
				v.Msg["id"] = v.Id // 消息id
				buf, _ := json.Marshal(v.Msg)
				if err = t.BroadcastOne(buf, s); err != nil {
					break
				}
				ok_ids = append(ok_ids, v.Id)
			}
			// 更新已发送的消息
			if len(ok_ids) > 0 {
				db.GetDB().Model(&common.IMgs{}).Where("id in(?)", ok_ids).Updates(base.Maps{"uids": gorm.Expr("array_remove(uids, ?)", user_id), "ok_uids": gorm.Expr("array_append(ok_uids, ?)", user_id)})
			}
		}()
	}

}

func (t *WsMgr) OnDisconnectConnect(s *melody.Session) {
	//t.BroadcastOne(g_wsmgr.ok_msg("{}"), s)
	if user_id, ok := s.GetI("user_id"); ok {
		t.rwmutx.Lock()
		defer t.rwmutx.Unlock()
		delete(t.ucs, user_id)
	}
}

// 消息发送成功
func (t *WsMgr) OnSentMessage(s *melody.Session, msg []byte) {
	//t.BroadcastOne(g_wsmgr.ok_msg("{}"), s)

}

func (t *WsMgr) get(uid int64) *melody.Session {
	t.rwmutx.RLock()
	defer t.rwmutx.RUnlock()
	return t.ucs[uid]
}

// 发送持久消息
func (t *WsMgr) PublishMsg(msg MsgReq) (err error) {

	v := common.IMgs{Type: msg.Type, Msg: base.StructToMap(msg.Msg), Uids: msg.To}

	// TODO 这儿有点问题 广播消息是否需要储存？
	if err = db.GetDB().Save(&v).Error; err != nil {
		glog.Error("wsmgr PublishMsg fail! err=", err)
		return
	}
	msg.Id = v.Id // 消息id
	body, _ := json.Marshal(msg.Msg)
	if 0 == len(msg.To) {
		if len(msg.Filter) == 0 {
			return t.Broadcast(body)
		}

		// 广播过滤
		return t.BroadcastFilter(body, func(s *melody.Session) bool {
			for k, v := range msg.Filter {
				// 过滤不相关的链接
				if val, ok := s.Get(k); !ok || !reflect.DeepEqual(v, val) {
					return false
				}
			}
			return true
		})
	}

	// 指定用户发送消息
	for _, id := range msg.To {
		if v := t.get(id); v != nil {
			if e := v.Write(body); e != nil {
				glog.Error("PublishMsg fail! user_id=", id)
			}
		}
	}

	return
}
