package conf

import (
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
)

type ApiConfig struct {
	Imports    []string `yaml:"imports"`
	LogLevel   string
	Servers    map[string]string
	ServerHost map[string]string         `yaml:"serverhost"`
	Server     SecServer                 `yaml:"serverinfo"`
	Redis      map[string]cache.RedisCfg `yaml:"redis"`
	AppKeys    map[string]string
	Ext        map[string]interface{} `yaml:"ext"`
	Amazon     map[string]string
	Upload     Upload
}

type Upload struct {
	ImageQuality int
	Bucket       string
	UrlPrefix    string
	LocalDir     string
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
type Exts map[string]interface{}

func (t Exts) GetInt64(keys ...string) (v int64, ok bool) {
	l := len(keys)
	if 0 == l {
		return 0, false
	}

	var val Exts
	if val, ok = t.GetMap(keys[0 : l-1]...); !ok {
		return
	}

	v, ok = val[keys[l-1]].(int64)
	return
}

func (t Exts) GetString(keys ...string) (v string, ok bool) {
	l := len(keys)
	if 0 == l {
		return "", false
	}

	var val Exts
	if val, ok = t.GetMap(keys[0 : l-1]...); !ok {
		return
	}
	v, ok = val[keys[l-1]].(string)
	return
}

func (t Exts) GetMap(keys ...string) (val Exts, ok bool) {
	val = t
	for _, key := range keys {
		// 此处断言必须是map[string]interface{}
		if val, ok = t[key].(map[string]interface{}); !ok {
			return
		}
	}
	return
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
