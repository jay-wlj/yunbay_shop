package conf

import (
	base "github.com/jay-wlj/gobaselib"
)

type ApiConfig struct {
	Imports     []string `yaml:"imports"`
	LogLevel    string
	Server      SecServer
	CommonRedis RedisServer       `yaml:"common_redis"`
	ApiRedis    RedisServer       `yaml:"api_redis"`
	AppKeys     map[string]string `yaml:"appkeys"`
	Servers     map[string]string
	ServerHost  map[string]string `yaml:"serverhost"`
	EOSConf     map[string]string `yaml:"eosconf"`
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
