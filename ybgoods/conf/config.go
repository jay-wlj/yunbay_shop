package conf

import (
	"github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
)

type ApiConfig struct {
	Imports    []string `yaml:"imports"`
	LogLevel   string
	Servers    map[string]string
	ServerHost map[string]string         `yaml:"serverhost"`
	Server     SecServer                 `yaml:"serverinfo"`
	MQUrls     []string                  `yaml:"mqurls"`
	Redis      map[string]cache.RedisCfg `yaml:"redis"`
	AppKeys    map[string]string
	Coins      map[int][]string
}

type SecServer struct {
	Listen    string
	CheckSign bool
	Debug     bool
	PSQLUrl   string
	Secret    string
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
