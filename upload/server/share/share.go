package share

import (
	// Imports the Google Cloud Storage client package
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"yunbay/upload/conf"
	"yunbay/upload/models"
	"yunbay/upload/util"

	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"github.com/jie123108/glog"
)

// [Content-Type] = {文件扩展名，路径前缀}
var g_contentType map[string][]string
var g_ext_contentType map[string]string
var g_url_prefix string

func init() {
	g_contentType = make(map[string][]string)
	// 图片
	g_contentType["image/gif"] = []string{"gif", "img"}
	g_contentType["image/jpeg"] = []string{"jpg", "img"}
	g_contentType["image/jpg"] = []string{"jpg", "img"}
	g_contentType["image/pjpeg"] = []string{"jpg", "img"}
	g_contentType["image/png"] = []string{"png", "img"}
	g_contentType["image/x-png"] = []string{"png", "img"}
	g_contentType["image/x-png"] = []string{"png", "img"}
	g_contentType["image/bmp"] = []string{"bmp", "img"}

	// 视频
	g_contentType["video/mp4"] = []string{"mp4", "video"}
	g_contentType["video/x-matroska"] = []string{"mkv", "video"}
	g_contentType["video/x-msvideo"] = []string{"avi", "video"}
	g_contentType["application/vnd.rn-realmedia-vbr"] = []string{"rmvb", "video"}
	g_contentType["video/3gpp"] = []string{"3gp", "video"}
	g_contentType["video/x-flv"] = []string{"flv", "video"}
	g_contentType["video/mpeg"] = []string{"mpg", "video"}
	g_contentType["video/quicktime"] = []string{"mov", "video"}
	g_contentType["video/x-ms-wmv"] = []string{"wmv", "video"}

	// 文本文件
	g_contentType["text/plain"] = []string{"txt", "file"}
	g_contentType["application/octet-stream"] = []string{"bin", "file"}
	g_contentType["application/vnd.android.package-archive"] = []string{"apk", "file"}
	g_contentType["application/vnd.iphone"] = []string{"ipa", "file"}

	g_ext_contentType = make(map[string]string)
	for contentType, varr := range g_contentType {
		ext := varr[0]
		g_ext_contentType[ext] = contentType
	}
	g_url_prefix = "https://file.yunbay.com/"
}

/**
 * prefix, suffix
 */

func getPrefixSuffix(contentType string) (string, string) {
	info := g_contentType[contentType]
	if len(info) == 0 {
		glog.Error("Content Type [", contentType, "] not support!")
		return "", "ERR_CONTENT_TYPE_INVALID"
	}

	return info[1], info[0]
}

func getContentTypeByExt(ext string) string {
	return g_ext_contentType[ext]
}

// 判断是不是bmp文件
func is_bmp(body []byte) bool {
	var BMP byte = 66
	if len(body) > 0 && body[0] == BMP {
		return true
	}
	return false
}

func SaveFileToLocal(filename string, body []byte) error {
	LocalDir := conf.Config.Upload.LocalDir
	if LocalDir == "" {
		return errors.New("not define upload local_dir")
	}

	fullfilename := path.Join(LocalDir, filename)
	glog.Info("write file [", fullfilename, "] to disk...")
	if util.IsExist(fullfilename) {
		glog.Info("file [", filename, "] is exist in disk!")
	} else {
		dir := path.Dir(fullfilename)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			glog.Error("MkdirAll(", dir, ") failed! err: ", err)
			return err
		}
		err = ioutil.WriteFile(fullfilename, body, os.ModePerm)
		if err != nil {
			glog.Error("write file [", filename, "] failed! err: ", err)
			return err
		}
		glog.Info("write file [", fullfilename, "] to disk success")
	}
	return nil
}

//TODO: 大文件分块上传。

// GetAccountById retrieves Account by Id. Returns error if
// Id doesn't exist

type UploadParams struct {
	AppId       string `validate:"required"`
	ContentType string `validate:"required"`
	Hash        string
	Rid         string
	Width       int
	Height      int
	FileName    string
	//Reader         *bytes.Reader
	//BodyLen        int
	Test           bool
	EnlargeSmaller bool
}

type UploadRsp struct {
	Url    string `json:"url"`
	Rid    string `json:"rid"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func (t UploadParams) is_bmp(body []byte) bool {
	var BMP byte = 66
	if len(body) > 0 && body[0] == BMP {
		return true
	}
	return false
}

func (t *UploadParams) Upload(body []byte) (rsp *UploadRsp, err error) {
	// 验参
	if err = yf.Valid(t); err != nil {
		glog.Error("Upload fail! err= ", err)
		return
	}

	app_id := t.AppId
	if t.is_bmp(body) {
		glog.Error("request img_type is bmp, not support")
		err = errors.New("ERR_FILE_NOT_SUPPORT")
		return
	}

	/**** 原图一样不处理直接返回 ****/
	prefix, suffix := getPrefixSuffix(t.ContentType)
	if prefix == "" {
		err = errors.New(suffix)
		return
	}

	// body_reader := bytes.NewReader(body)
	// file, _ := os.OpenFile("/opt/1.png", os.O_WRONLY|os.O_CREATE, 0666)
	// defer file.Close()
	// io.Copy(file, body_reader)

	// 效验hash
	hash := util.Sha1hex(body)
	if t.Hash != hash {
		glog.Error("request hash is ", t.Hash, ", calc_hash is ", hash)
		err = errors.New("ERR_HASH_INVALID")
		return
	}

	filename := t.FileName
	if filename == "" {
		filename = fmt.Sprintf("%s/%s/%s/%s/%s.%s", app_id, prefix, hash[0:2], hash[2:4], hash, suffix)
	}

	if t.Test {
		filename = "test/" + filename
	}

	width, height := 0, 0
	width_new, height_new := t.Width, t.Height

	if prefix == "img" {
		size, err := util.GetImageSize(body)
		if err != nil {
			glog.Error("Get Image[", filename, "]'s size failed! err: ", err)
		} else {
			width = size.Width
			height = size.Height
		}
	}

	if prefix != "file" {
		// 需要调整图像
		if (width_new != width || height_new != height) && width_new > 0 && height_new > 0 {
			glog.Info(fmt.Sprintf("image [%s] Request Size: %dx%d != ServerFetchSize: %dx%d", filename, width, height, width_new, height_new))
			quality := conf.Config.Upload.ImageQuality
			//filename := c.Ctx.Request.RequestURI

			body_new, e := util.ResizeBytesImgToBytes(body, filename, width_new, height_new, t.EnlargeSmaller, quality)
			if err = e; err != nil {
				glog.Error(fmt.Sprintf("ResizeBytesImgToBytes(filename: %s, width: %d, height: %d, enlarge: %v) failed! err: %v",
					filename, width_new, height_new, t.EnlargeSmaller, err))
				return
			}
			body = body_new
			width, height = width_new, height_new
			// 重新计算HASH.
			hash = util.Sha1hex(body)
		}

		// if width_new == size.Width && height_new == size.Height {
		// 	//这里存储的的是原始图片.代码后面一个存储是存储裁剪后的文件.
		// 	// url, reason := SaveFileToS3(app_id, rid_, hash, filename, contentType, body_reader, body_len, width, height, 0, "")
		// 	url, reason := SaveFileToGoogleStorage(app_id, rid_, hash, filename, contentType, body_reader, body_len, width, height, 0, "")
		// 	if url == "" {
		// 		c.ServeJsonFail(reason)
		// 		return
		// 	}
		// 	glog.Info("write file ", filename, " ok. url is: ", url)

		// 	c.ServeJsonOkEx(util.H{"url": url, "rid": rid_})
		// 	return
		// }
	}
	rid_ := t.Rid
	if rid_ == "" {
		rid_ = util.HashToRid(app_id, hash)
		glog.Info("appid: ", app_id, ", hash: ", hash, " ==> rid: ", rid_)
	} else {
		rid_ = util.IdToRid(app_id, rid_)
	}

	resInfo, err := models.GetUploadfileByRid(rid_)
	if err == nil && resInfo != nil {
		glog.Info("-------- file [hash=", hash, "] is uploaded! -------- ")
		rsp = &UploadRsp{Url: resInfo.Path, Rid: resInfo.Rid, Width: resInfo.Width, Height: resInfo.Height}
		return
	}

	filename = fmt.Sprintf("%s/%s/%s/%s/%s.%s", app_id, prefix, hash[0:2], hash[2:4], hash, suffix)

	if t.Test {
		filename = "test/" + filename
	}

	if err = util.SaveFileToGoogleStorage(conf.Config.Upload.Bucket, filename, t.ContentType, bytes.NewReader(body)); err != nil {
		return
	}
	url := g_url_prefix + filename
	glog.Info("write file ", filename, " ok. url is: ", url)

	// 先保存数据
	info := &models.Uploadfile{Path: url, AppId: app_id, Rid: rid_, Hash: hash, Size: int(len(body)), Width: width, Height: height}
	db := db.GetDB()
	if err = db.Save(&info).Error; err != nil {
		glog.Error("uploadfile.Upsert(", info, ") failed! err: ", err)
	}

	rsp = &UploadRsp{Url: url, Rid: rid_, Width: width, Height: height}
	return
}
