package conf

import (
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
)

type ApiConfig struct {
	Imports    []string `yaml:"imports"`
	LogLevel   string
	Server     SecServer                 `yaml:"server"`
	Redis      map[string]cache.RedisCfg `yaml:"redis"`
	AppKeys    map[string]string         `yaml:"appkeys"`
	Servers    map[string]string
	ServerHost map[string]string `yaml:"serverhost"`
}

func (t ApiConfig) GetImports() []string {
	return t.Imports
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

var Config ApiConfig
var _ base.IConf = Config // 确保实现IConf接口

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
