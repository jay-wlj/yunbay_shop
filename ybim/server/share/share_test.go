package share

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"sync"

	"github.com/jay-wlj/gobaselib/melody"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

func TestWebsocket(t *testing.T) {
	r := gin.Default()
	ws := melody.New()

	// websocket 链接
	r.GET("/ws", func(c *gin.Context) {
		ws.HandleRequest(c.Writer, c.Request)
	})

	ws.HandleMessage(func(c *melody.Session, msg []byte) {
		s := string(msg)
		glog.Info("recv msg:", s)

		//ws.BroadcastMultiple([]*Session{c}, "vsedgdfg")
		ws.BroadcastMultiple([]byte("nihao"), []*melody.Session{c})
		//ws.PubMessage([]string{c.Id}, "nihao")

	})
	r.Run(":2004")
}

type s struct {
	exit chan bool
	msg  chan interface{}
	open bool
	sync.RWMutex
}

func (t *s) closed() bool {
	t.RLock()
	defer t.RUnlock()
	return !t.open
}
func (t *s) close() {
	if t.closed() {
		return
	}
	runtime.Gosched()
	t.exit <- true
}

func (t *s) run() {
	t.Lock()
	t.open = true
	t.Unlock()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

loop:
	for {

		select {
		case <-t.exit:
			fmt.Println("handle exit")
			t.Lock()
			t.open = false
			t.Unlock()
			break loop
		case m := <-t.msg:
			// TODO handle the msg
			fmt.Println("recv msg:", m)
		case <-ticker.C:
			fmt.Println("tick")
		}
	}

	fmt.Println("exit loop")
}
func TestCh(t *testing.T) {
	s := s{exit: make(chan bool)}
	go s.run()

	for i := 0; i < 10; i++ {
		go s.close() // 测试并发关闭
	}

	for i := 0; i < 10; i++ {
		go func() {
			if s.closed() {
				return
			}
			s.msg <- "sdf"
		}()
	}
	//go s.close() // 测试并发关闭

	ch := make(chan bool)
	ch <- true
}

func TestIm(t *testing.T) {
	m := make(map[int]string)
	m[0] = "sdf"
	m[1] = "sdfdb"
	fmt.Println(m)
}
