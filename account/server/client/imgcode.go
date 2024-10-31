package client

import (
	"time"
	"yunbay/account/util"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"

	"github.com/mojocn/base64Captcha"
)

type ImgageCode struct{}

var imgcode_redis *cache.RedisCache
var expires int

func init() {
}

func InitImagecodeRedis(expire int) (err error) {
	expires = expire
	imgcode_redis, err = cache.GetWriter("imgcode")
	if err != nil {
		panic("InitImagecodeRedis fail!")
	}
	return
}

func saveImgCode(key string, code string) (err error) {
	keyid := "img:" + key
	expires := time.Duration(int(time.Second) * expires)
	err = imgcode_redis.Set(keyid, code, expires)
	return
}

func getImgCode(key string) (code string, err error) {
	keyid := "img:" + key
	code, err = imgcode_redis.Get(keyid)
	return
}

func delImgCode(key string) {
	keyid := "img:" + key
	imgcode_redis.Del(keyid)
	return
}

// 检测图片验证码是否正确
func ImgCodeCheck(key, val string) bool {
	idkey, err := getImgCode(key)
	if err != nil && err != cache.ErrNotExist {
		glog.Error("GetCode(", key, ") failed! err:", err)
		return false
	}
	verifyResult := base64Captcha.VerifyCaptcha(idkey, val)
	if !verifyResult {
		glog.Error("VerifyCaptcha fail:", idkey)
		return false
	}
	delImgCode(key)
	return true
}

// @router /get [get]
func (t *ImgageCode) GetImgCode(c *gin.Context) {
	key, _ := base.CheckQueryStringField(c, "key")
	platform, _ := base.CheckQueryStringField(c, "platform")
	width, _ := base.CheckQueryIntDefaultField(c, "width", 0)
	height, _ := base.CheckQueryIntDefaultField(c, "height", 0)

	w := 72
	h := 32
	if platform == "" {
		platform, _ = util.GetPlatformVersionByContext(c)
	}
	if platform == "web" || platform == "h5" {
		w = 92
		h = 46
	}
	if width > 0 {
		w = width
	}
	if height > 0 {
		h = height
	}
	var config base64Captcha.ConfigCharacter
	config = base64Captcha.ConfigCharacter{
		Height: h,
		Width:  w,
		//const CaptchaModeNumber:数字,CaptchaModeAlphabet:字母,CaptchaModeArithmetic:算术,CaptchaModeNumberAlphabet:数字字母混合.
		Mode:               base64Captcha.CaptchaModeNumber,
		ComplexOfNoiseText: base64Captcha.CaptchaComplexLower,
		ComplexOfNoiseDot:  base64Captcha.CaptchaComplexLower,
		IsShowHollowLine:   true,
		IsShowNoiseDot:     false,
		IsShowNoiseText:    false,
		IsShowSlimeLine:    false,
		IsShowSineLine:     false,
		CaptchaLen:         4,
	}

	//GenerateCaptcha 第一个参数为空字符串,包会自动在服务器一个随机种子给你产生随机uiid.
	//id, digitCap := base64Captcha.GenerateCaptcha("1234", config)
	//code := util.RandomSample("0123456789", 4)
	idkey, digitCap := base64Captcha.GenerateCaptcha("", config)

	data := digitCap.BinaryEncoding()
	//base64Png := base64Captcha.CaptchaWriteToBase64Encoding(digitCap)
	glog.Info("code:%v", idkey, "%v")
	saveImgCode(key, idkey)

	//c.Ctx.Output.Header("Content-Type", "image/png; charset=utf-8")
	//c.Ctx.Output.Body(data)

	c.Data(200, "image/png; charset=utf-8", data)
	return
}

type imgCode struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// @router /check [post]
func (t *ImgageCode) CheckImgCode(c *gin.Context) {
	var args imgCode
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}

	if ok := ImgCodeCheck(args.Key, args.Value); ok {
		yf.JSON_Fail(c, yf.ERR_IMGCODE_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{})
}
