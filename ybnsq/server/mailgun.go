package server

import (
	"gopkg.in/mailgun/mailgun-go.v1"
	"github.com/nsqio/go-nsq"
	"github.com/jie123108/glog"
	"encoding/json"
	"yunbay/ybnsq/conf"
	"regexp"
)

// Your available domain names can be found here:
// (https://app.mailgun.com/app/domains)
var yourDomain string = "yunbay.com" // e.g. mg.yourcompany.com

// The API Keys are found in your Account Menu, under "Settings":
// (https://app.mailgun.com/app/account/security)

// starts with "key-"
var privateAPIKey string = "84237d0189822ce5d99e9179e8b9dbf0-770f03c4-6869fc8d"

// starts with "pubkey-"
var publicValidationKey string = "pubkey-ace932534fa29362441b1e5ab5ba7313"



func SendMail(subject, body, html, sender string, recivers []string)(err error) {
    // Create an instance of the Mailgun Client

	//sender := "chenli@yunfan.com"    
	if sender == "" {
		sender = conf.Config.Email.Sender
	}
	
	mg := mailgun.NewMailgun(yourDomain, privateAPIKey, publicValidationKey)
    return sendMessage(mg, sender, subject, body, html, recivers)
}

func sendMessage(mg mailgun.Mailgun, sender, subject, body, html string, recipient []string)(err error) {
	message := mg.NewMessage(sender, subject, body, recipient...)
	if html != ""{
		message.SetHtml(html)
	}	
	resp, id, err1 := mg.Send(message)
	err = err1
	

	glog.Infof("ID: %s Resp: %s\n", id, resp)
	return
}


//RecommendMID is the message struct
type MailParams struct {
	Receiver []string `json:"receivers"` 	 	// 接收人邮件
	Sender string `json:"sender"`			// 发送人邮件
	Subject string `json:"subject"`    		// 标题
	Content string `json:"content"`    		// 内容
	Html	string `json:"html"`			// html内容
}

//ConsumerRecommend is NSQ Consumer struct
type MailGun struct{}



//HandleMessage is function of message
func (*MailGun) HandleMessage(msg *nsq.Message)(err error) {
	glog.Info("[sendmail]receive:", msg.NSQDAddress, " message:", string(msg.Body))

	var args MailParams

	err = json.Unmarshal(msg.Body, &args)
	if err != nil {
		msg.Finish() // 此消息body非法 丢弃
		glog.Error("MailGun args invalid! finish msg, err=", err)
		return
	}

	recivers := []string{}
	match_sender := true	
	
	for _, v := range args.Receiver {
		if mathch, _ := regexp.MatchString("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z0-9]{2,6}$", v); mathch {
			recivers = append(recivers, v)			
		} else {
			glog.Error("MailGun filter invalid email address! address:", v)
		}
	}

	//mathch, _ := regexp.MatchString("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z0-9]{2,6}$", args.Receiver)
	if args.Sender != "" {
		match_sender, _ = regexp.MatchString("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z0-9]{2,6}$", args.Sender)
	}
	
	if len(recivers) == 0 || !match_sender || args.Subject == "" {
		msg.Finish() // 此消息body非法 丢弃
		glog.Error("MailGun args invalid! finish msg, args=", args)
		return
	}

	//insertRecommend(DBConn, Recommddata.fromuser, Recommddata.mid)
	if err = SendMail(args.Subject, args.Content, args.Html, args.Sender, recivers); err != nil {	
		glog.Error("mailgun fail! try again! trys:", msg.Attempts)
		msg.Requeue(-1)	// 保证不丢弃该消息，直到处理成功
		return
	}

	msg.Finish() // 此消息已完成处理 丢弃
	return 
}

