package client

import (
	"fmt"
	"io/ioutil"
	"strings"
	"yunbay/upload/conf"
	"yunbay/upload/server/share"
	"yunbay/upload/util"

	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
)

func UploadSimple(c *gin.Context) {

	body, _ := c.GetRawData()
	req := share.UploadParams{
		AppId:          c.GetHeader("X-Yf-Appid"),
		Test:           conf.Config.Server.Test || c.GetHeader("X-Yf-Test") == "1",
		Rid:            c.GetHeader("X-Yf-Rid"),
		Hash:           c.GetHeader("X-Yf-Hash"),
		ContentType:    c.GetHeader("Content-Type"),
		FileName:       c.GetHeader("X-Yf-Filename"),
		EnlargeSmaller: c.GetHeader("x-yf-enarge_smaller") == "true",
	}

	if resize := c.GetHeader("X-Yf-Resize"); resize != "" {
		if _, err := fmt.Sscanf(resize, "%dx%d", &req.Width, &req.Height); err != nil {
			glog.Error("X-Yf-resize invalid", resize, " err=", err)
			yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
			return
		}
	}

	r, err := req.Upload(body)
	if err != nil {
		glog.Error("UploadSimple fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}

	yf.JSON_Ok(c, r)
}

func WebForm(c *gin.Context) {

	req := share.UploadParams{
		AppId:          c.GetHeader("X-Yf-Appid"),
		Hash:           c.GetHeader("x-yf-hash"),
		Test:           conf.Config.Server.Test || c.GetHeader("X-Yf-Test") == "1",
		Rid:            c.GetHeader("X-Yf-Rid"),
		ContentType:    c.GetHeader("Content-Type"),
		FileName:       c.GetHeader("X-Yf-Filename"),
		EnlargeSmaller: c.GetHeader("x-yf-enarge_smaller") == "true",
	}

	if strings.Index(req.ContentType, "multipart/form-data") == -1 {
		glog.Error("unsupport content-type ", req.ContentType)
		yf.JSON_Fail(c, "content-type is not multipart/form-data")
		return
	}

	if resize := c.GetHeader("X-Yf-Resize"); resize != "" {
		if _, err := fmt.Sscanf(resize, "%dx%d", &req.Width, &req.Height); err != nil {
			glog.Error("X-Yf-resize invalid", resize, " err=", err)
			yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
			return
		}
	}

	body := []byte{}
	{
		handler, err := c.FormFile("file")
		//file, handler, err := r.FormFile("file")
		if err != nil {
			glog.Error("From file err=", err)
			return
		}
		file, e := handler.Open()
		if err = e; err != nil {
			glog.Error("From file err=", err)
			return
		}
		defer file.Close()
		body, _ = ioutil.ReadAll(file)
		req.ContentType = handler.Header.Get("Content-Type")
		// if strings.HasSuffix(handler.Filename, ".ipa") {
		// 	args.ContentType = "application/vnd.iphone"
		// }
	}
	req.Hash = util.Sha1hex(body) // web端暂不判断hash

	r, err := req.Upload(body)
	if err != nil {
		glog.Error("UploadSimple fail! err=", err)
		yf.JSON_Fail(c, err.Error())
		return
	}

	yf.JSON_Ok(c, r)
}
