package conf

import (
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
)

type ApiConfig struct {
	Imports    []string `yaml:"imports"`
	LogLevel   string
	Servers    map[string]string
	ServerHost map[string]string `yaml:"serverhost"`
	Server     SecServer         `yaml:"serverinfo"`
	// CommonRedis       RedisServer               `yaml:"common_redis"`
	// ApiRedis          RedisServer               `yaml:"api_redis"`
	Redis             map[string]cache.RedisCfg `yaml:"redis"`
	AppKeys           map[string]string
	RsaPublicKeyFile  string
	RsaPrivateKeyFile string
	IpipFile          string
	Orders            OrdersConf
	Email             map[string]string
	SmsText           map[string]string      `yaml:"smstext"`
	AppCfgPath        map[string]string      `yaml:"app_cfg"`
	OfPay             OfPay                  `yaml:"ofpay"`
	Ext               map[string]interface{} `yaml:"ext"`
}

type SecServer struct {
	Listen    string
	CheckSign bool
	Debug     bool
	PSQLUrl   string
	MQUrls    []string `yaml:"mqurls"`
	Secret    string
	Test      bool `yaml:"test"`
}

type OfPay struct {
	Host      string `yaml:"host"`
	AppId     string `yaml:"app_id"`
	AppPws    string `yaml:"app_pws"`
	AppSecret string `yaml:"app_secret"`
	RetUrl    string `yaml:"ret_url"`
	Test      bool
	Ofcard    map[string]string `yaml:"ofcard"`
}

type RedisServer struct {
	Addr     string
	Password string
	DBIndex  int    `yaml:"dbindex"`
	Timeout  string `yaml:"timeout"`
}

type OrdersConf struct {
	AutoCancelTime string `yaml:"auto_cancel_time"`
	AutoFinishTime string `yaml:"auto_finish_time"`
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
