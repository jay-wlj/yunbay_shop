package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybapi/conf"
	"yunbay/ybapi/dao"
	"yunbay/ybapi/util"

	"github.com/gin-gonic/gin"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jay-wlj/gobaselib/yf"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

var app_cfg map[string]interface{}

func init() {
	app_cfg = make(map[string]interface{})
}

func Pub_GetConf(c *gin.Context) {
	country := util.GetCountry(c)
	strCountry := strconv.Itoa(country)
	ver, _ := base.CheckQueryIntField(c, "ver")

	if app_cfg[strCountry] == nil {
		if cfg_path := conf.Config.AppCfgPath[strCountry]; cfg_path != "" {
			data, err := ioutil.ReadFile(cfg_path)
			if err != nil {
				glog.Error("Pub_GetConf fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			var cfg map[string]interface{}
			if err := json.Unmarshal(data, &cfg); err != nil {
				glog.Error("Pub_GetConf fail! err=", err)
				yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
				return
			}
			app_cfg[strCountry] = cfg
		}
	}

	cfg := app_cfg[strCountry].(map[string]interface{})
	if ver > 0 {
		if v, ok := cfg["ver"].(int); ok && ver <= v {
			yf.JSON_Ok(c, yf.DATA_NOT_MOTIFIED)
			return
		}
	}

	yf.JSON_Ok(c, app_cfg[strCountry])
}

func Feedback_Add(c *gin.Context) {
	userid := c.GetInt64("user_id")
	args := common.Feedback{UserId: userid}
	if ok := util.UnmarshalReq(c, &args); !ok {
		return
	}

	if args.Email == "" || args.Title == "" || args.Info == "" {
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	if ok, _ := regexp.MatchString("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z0-9]{2,6}$", args.Email); !ok {
		yf.JSON_Fail(c, yf.ERR_EMAIL_INVALID)
		return
	}
	now := time.Now().Unix()
	args.CreateTime = now
	args.UpdateTime = now
	args.Id = 0

	var str_affix string
	if len(args.Affix) > 0 {
		for _, v := range args.Affix {
			str_affix += fmt.Sprintf("<a href=%v>%v</a> ", v, v)
		}
	}
	// 发送用户反馈邮件
	html := fmt.Sprintf("<p>帐号id: %v<br/></p>\n<p>邮箱: <a href=\"mailto:%v\">%v</a><br/></p>\n<p>标题: %v</p>\n<p>描述: %v</p>\n<p>附件: %v</p>\n", args.UserId, args.Email, args.Email, args.Title, args.Info, str_affix)
	v := common.MQMail{Sender: args.Email, Receiver: []string{"service@yunbay.com"}, Subject: "用户问题反馈", Html: html}
	if err := util.PublishMsg(v); err != nil {
		glog.Error("Feedback_Add fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	if html, ok := conf.Config.Email["fankui"]; ok {
		// 发送自动回复邮件
		v := common.MQMail{Receiver: []string{args.Email}, Subject: "收到用户问题反馈", Html: html}
		if err := util.PublishMsg(v); err != nil {
			glog.Error("Feedback_Add fail! err=", err)
			yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
			return
		}
	}

	// if err := db.GetTxDB(c).Create(&args).Error; err != nil {
	// 	glog.Error("Feedback_Add fail! err=", err)
	// 	yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
	// 	return
	// }
	// yf.JSON_Ok(c, gin.H{"id": args.Id})
	yf.JSON_Ok(c, gin.H{})
}

func Banner_List(c *gin.Context) {
	platform, _ := util.GetPlatformVersionByContext(c)
	pos := 1
	if platform == "android" || platform == "ios" {
		pos = 1
	}
	v := dao.Banner{}
	results, err := v.Get(pos)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Banner_List fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}

	yf.JSON_Ok(c, gin.H{"list": results.Content})
}

func Upgrade_Check(c *gin.Context) {
	platform, version := util.GetPlatformVersionByContext(c)
	if platform == "" || version == "" {
		glog.Error("Upgrade_Check fail! err=ERR_ARGS_INVALID")
		yf.JSON_Fail(c, yf.ERR_ARGS_INVALID)
		return
	}
	channel, _ := util.GetHeaderString(c, "X-Yf-Channel")
	country := util.GetCountry(c)
	ver := base.Version4ToInt(version)
	var v common.Upgrade
	db := db.GetDB().Where("status=1 and platform=? and country=?", platform, country)
	if channel != "" {
		db = db.Where(fmt.Sprintf("(array_length(channels,1) is null or channels::text[] @> ARRAY['%v'])", channel))
	}
	err := db.Order("id desc").Find(&v).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound { // 未开启不需要升级
			yf.JSON_Ok(c, gin.H{})
			return
		}
		glog.Error("Upgrade_Check fail! err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	// 开启了，版本号为最新版，不需要升级
	versionInteger := base.Version4ToInt(v.Version)
	if ver >= versionInteger {
		yf.JSON_Ok(c, gin.H{})
		return
	}
	// 强升版本判断，如果在强升版本中，一定需要升级
	var isForce bool = false
	var m common.Mandatory
	binary, err := json.Marshal(v.Mandatory)
	if err != nil {
		glog.Error("Upgrade_Check fail! Error json.Marshal->mandatory; err=", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	json.Unmarshal(binary, &m)
	// fmt.Println("mandatory json string:", string(binary))
	for _, sectionVersion := range m.Section {
		if (ver >= base.Version4ToInt(sectionVersion.Start)) && (ver <= base.Version4ToInt(sectionVersion.End)) {
			isForce = true
			break
		}
	}
	if isForce == false {
		for _, specifyVersion := range m.Specify {
			if ver == base.Version4ToInt(specifyVersion) {
				isForce = true
				break
			}
		}
	}
	if isForce {
		v.Type = 2
		yf.JSON_Ok(c, v)
		return
	}
	switch v.Mold {
	case 2: // 低于设定版本升级
		if (len(v.Upversions) > 0) && (ver > base.Version4ToInt(v.Upversions[0])) {
			yf.JSON_Ok(c, gin.H{})
			return
		}
	case 3: // 指定版本升级
		for _, v1 := range v.Upversions {
			if base.Version4ToInt(v1) == ver {
				yf.JSON_Ok(c, v)
				return
			}
		}
		yf.JSON_Ok(c, gin.H{})
		return
	}
	yf.JSON_Ok(c, v)
}

// 获取app端最新的安装包
func App_Download(c *gin.Context) {
	var vs []common.Upgrade
	db := db.GetDB()
	// err := db.Raw("select * from(select *,row_number() over(partition by platform order by update_time desc)rn from upgrade)t where rn=1 and platform like 'web_%'").Scan(&vs).Error
	country := util.GetCountry(c)
	sql := "(select * from upgrade where platform='web_ios' and country=? order by update_time desc limit 1)"
	sql = sql + " UNION	(select * from upgrade where platform='web_android' and country=? order by update_time desc limit 1)"
	err := db.Raw(sql, country, country).Scan(&vs).Error
	if err != nil {
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	for key, value := range vs {
		value.Platform = strings.Trim(value.Platform, "web_")
		vs[key] = value
	}
	yf.JSON_Ok(c, vs)
}

type PaidAffiche struct {
	Rebate       float64 `json:"rebate"`
	Username     string  `json:"username"`
	ProductTitle string  `json:"product_title"`
	UpdatedTime  int64   `json:"updated_time"`
}

// 折扣专区购买公告
func RebatePaidAfficheList(c *gin.Context) {
	redisCache, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("redis cache connect is fail! error = ", err)
		yf.JSON_Fail(c, yf.ERR_SERVER_ERROR)
		return
	}
	var afficheList []string
	afficheList, _ = redisCache.LRange("paid_affiche_list", 0, 5).Result()
	stringList := strings.Replace(fmt.Sprint(afficheList), " ", ",", -1)
	var paidAfficheList []PaidAffiche
	json.Unmarshal([]byte(stringList), &paidAfficheList)
	var longTime int64 = 80
	var num int = 0
	returnData := []PaidAffiche{}
	nowTime := time.Now().Unix()
	for _, v := range paidAfficheList {
		if (nowTime - v.UpdatedTime) > longTime {
			break
		} else {
			// glog.Error(num)
			returnData = append(returnData, v)
			num++
		}
	}
	// 判断长度是否超过5，超过后，后面的记录清空
	redisCache.LTrim("paid_affiche_list", 0, 5)
	yf.JSON_Ok(c, returnData)
}
