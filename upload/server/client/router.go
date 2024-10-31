package client

import (
	"github.com/jay-wlj/gobaselib/yf"
)

func InitRouter() (routers []yf.RouterInfo) {

	routers = []yf.RouterInfo{

		{yf.HTTP_POST, "/upload/simple", true, false, UploadSimple}, // 上传文件
		{yf.HTTP_POST, "/upload/web/form", false, false, WebForm},   // 上传文件(multipart/form-data形式提交)

	}
	return
}
