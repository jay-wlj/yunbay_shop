package task


import (
	//"github.com/jie123108/glog"
    "fmt"
    "net/smtp"
    "strings"
)


func MainSend(body string){
	auth := smtp.PlainAuth("", "305898636@qq.com", "password", "smtp.qq.com")
	to := []string{"weilijian@yunfan.com"}
	nickname := "test"
	user := "305898636@qq.com"
	subject := "Yunbay Error"
	content_type := "Content-Type: text/plain; charset=UTF-8"
	msg := []byte("To: " + strings.Join(to, ",") + "\r\nFrom: " + nickname +
		"<" + user + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	err := smtp.SendMail("smtp.qq.com:25", auth, user, to, msg)
	if err != nil {
		fmt.Printf("send mail error: %v", err)
	}
}
