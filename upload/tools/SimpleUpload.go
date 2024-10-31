package main

import (
	"flag"
	"fmt"
	"time"
	"yunbay/upload/util"
)

func main() {
	var filename string
	var host string
	var app_id string
	var app_key string
	var id string
	var resize string
	var target string
	var timeout time.Duration
	is_test := true

	flag.StringVar(&filename, "img", "test.jpg", "Image Filename(.jpg|.png|.gif)")
	flag.StringVar(&app_id, "appid", "upload", "App Id")
	flag.StringVar(&app_key, "appkey", "super", "App Key")
	flag.StringVar(&id, "id", "", "Resource Id")
	flag.StringVar(&resize, "resize", resize, "Resize, .eg: 300x400, 300x0, 0x400")
	flag.StringVar(&target, "target", target, "target file path")

	flag.DurationVar(&timeout, "timeout", time.Minute*5, "upload timeout, like: 300ms, 10s, 2m, 1h")
	flag.StringVar(&host, "host", "http://127.0.0.1:2000", "指定上传服务地址")
	flag.Parse()
	if app_key == "super" {
		app_key = "69c5c1c89f9f6093559af661bc4e4df1"
	}

	res := util.UploadFile(host, filename, app_id, app_key, id, resize, target, timeout, is_test)

	fmt.Printf("response headers -------------- status: %v\n", res.Headers)
	fmt.Printf("response -------------- status: %d\n", res.StatusCode)
	fmt.Println(string(res.RawBody))
	if res.StatusCode != 200 {
		fmt.Println("reason:", res.Reason)
		fmt.Println("Error:", res.Error)
		fmt.Println("req_debug:", res.ReqDebug)
	}

}
