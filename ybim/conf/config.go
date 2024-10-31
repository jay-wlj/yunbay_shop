package conf

import (
	base "github.com/jay-wlj/gobaselib"
)

type ApiConfig struct {
	Imports     []string `yaml:"imports"`
	LogLevel    string
	Servers     map[string]string
	ServerHost  map[string]string `yaml:"serverhost"`
	Server      SecServer         `yaml:"serverinfo"`
	CommonRedis RedisServer       `yaml:"common_redis"`
	IMRedis     RedisServer       `yaml:"im_redis"`
	AppKeys     map[string]string
	IMKey       string `yaml:"im_key"`
	IMSecret    string `yaml:"im_secret"`
	IMEanble    bool   `yaml:"im_enalbe"`
}

type SecServer struct {
	Listen    string
	CheckSign bool
	Debug     bool
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
