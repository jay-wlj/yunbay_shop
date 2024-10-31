package conf

import (


	"github.com/jay-wlj/gobaselib"
	//"strings"
)

//APIConfig is config of server
type ApiConfig struct {
	Imports []string `yaml:"imports"`
	Debug   bool

	Nsqladdr   string
	Maxnsqd    int
	Email      EmailSt
	Consumers  map[string]Consumer `yaml:"consumers"`
	Servers    map[string]string
	ServerHost map[string]string `yaml:"serverhost"`
	AppKeys    map[string]string
}

type EmailSt struct {
	Sender string
}

type Consumer struct {
	Channels            []string `yaml:"channels"`
	DefaultRequeueDelay *string  `yaml:"default_requeue_delay"`
	MaxAttempts         *uint16  `yaml:"max_attempts"`
	Concurrenct         *uint16  `yaml:"concurrent"`
}

// Config is global variation
var Config ApiConfig
var _ base.IConf = Config // 确保实现IConf接口
func (t ApiConfig) GetImports() []string {
	return t.Imports
}

func LoadConfig(file string) (*ApiConfig, error) {
	err := base.LoadConf(file, &Config)
	if err != nil {
		return nil, err
	}
	return &Config, err
}
