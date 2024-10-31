package conf

import (

	"github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"

)

type ApiConfig struct {
	Imports    []string `yaml:"imports"`
	LogLevel   string
	Server     SecServer
	Redis      map[string]cache.RedisCfg `yaml:"redis"`
	AppKeys    map[string]string         `yaml:"appkeys"`
	Servers    map[string]string
	ServerHost map[string]string `yaml:"serverhost"`
	Alipay     AlipaySt          `yaml:"alipay"`
	Weixin     WeixinSt          `yaml:"weixin"`
	BankCfg    BankSt            `yaml:"bank"`
	//ProjectRebat RebatProjectConfig	`yaml:"project_ybt_allot"`
}

type SecServer struct {
	Listen    string
	CheckSign bool
	Debug     bool
	Test      bool
	PSQLUrl   string
	MQUrls    []string `yaml:"mqurls"`
	Secret    string
	Ext       map[string]interface{} `yaml:"ext"`
}

type RedisServer struct {
	Addr     string
	Password string
	DBIndex  int    `yaml:"dbindex"`
	Timeout  string `yaml:"timeout"`
}

type AlipaySt struct {
	Appid       string
	Public      string
	Private     string
	NotifyUrl   string `yaml:"notify_url"`
	ProductCode string `yaml:"product_code"`
}

type WeixinSt struct {
	Appid           string
	Appkey          string
	Mchid           string
	Sanbox          bool
	NotifyUrl       string `yaml:"notify_url"`
	RefundNotifyUrl string `yaml:"refund_notify_url"`
}

type BankSt struct {
	DefaultIcon string            `yaml:"default_icon"`
	Banks       map[string]string `yaml:"support_banks"`
}

var Config ApiConfig
var _ base.IConf = Config // 确保实现IConf接口
func (t ApiConfig) GetImports() []string {
	return t.Imports
}
func (this *ApiConfig) GetSignKey(appid string) string {
	return this.AppKeys[appid]
}

func LoadConfig(file string) (*ApiConfig, error) {
	err := base.LoadConf(file, &Config)
	if err != nil {
		return nil, err
	}
	return &Config, err
}
