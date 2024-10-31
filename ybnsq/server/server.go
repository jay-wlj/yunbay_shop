package server

import (
	"github.com/jie123108/glog"
	"time"
	"yunbay/ybnsq/conf"
	nsq "github.com/nsqio/go-nsq"
    _ "github.com/jinzhu/gorm/dialects/postgres"
)

// DBConn is global variration of db



type consumerParams struct {
	Topic string
	Channel string
	Address string
	DefaultRequeueDelay *string
	MaxAttempts *uint16
	Concurrenct *uint16
	Handler nsq.Handler
}
//InitConsumer is init function of nsq's consumer
func RegisterConsumer(v consumerParams) {
	cfg := nsq.NewConfig()
		
	if v.DefaultRequeueDelay != nil {
		if delay, err := time.ParseDuration(*v.DefaultRequeueDelay); err == nil {
			cfg.DefaultRequeueDelay = delay		// 默认排队时长
		} else {
			glog.Error("InitConsumer ParseDuration err=", err)
		}
	}
	if v.MaxAttempts != nil {
		cfg.MaxAttempts = *v.MaxAttempts		// 重试次数
	}
	
    cfg.MaxInFlight=conf.Config.Maxnsqd				// //注意MaxInFlight的设置，默认只能接受一个节点
	cfg.LookupdPollInterval = time.Second          //设置重连时间
	c, err := nsq.NewConsumer(v.Topic, v.Channel, cfg) // 新建一个消费者
	if err != nil {
		panic(err)
	}
	
	c.SetLogger(nil, 0) //屏蔽系统日志
	
	cocurrenct := 1
	if v.Concurrenct != nil && *v.Concurrenct>0 {
		cocurrenct = int(*v.Concurrenct)
	}
	c.AddConcurrentHandlers(v.Handler, cocurrenct)		// 添加消费者接口	
	
	//建立NSQLookupd连接
	if err := c.ConnectToNSQLookupd(v.Address); err != nil {
		panic(err)
	}
	glog.Infof("RegisterConsumer topic:%v success", v.Topic)
}


func getTopicHandler(topic string) (nsq.Handler) {
	switch topic {
	case "useradd":
		return &UserRecommend{}
	case "sendmail":
		return &MailGun{}
	case "mqurl":
		return &MQUrl{}
	}
	return nil
}



//StartServer is init and start the server
func StartServer() {

	for k, v := range conf.Config.Consumers {
		for _, ch := range v.Channels {
			if ch == "" {
				glog.Error("topic ", k, " channel is empty!")
				continue
			}
			p := consumerParams{Topic:k, Channel:ch, Address:conf.Config.Nsqladdr, DefaultRequeueDelay:v.DefaultRequeueDelay, MaxAttempts:v.MaxAttempts, Handler:getTopicHandler(k)}
			RegisterConsumer(p)
		}		
	}
}
