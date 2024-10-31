package conf

import (
	"fmt"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"

	"github.com/jie123108/glog"
)

type ApiConfig struct {
	Imports           []string `yaml:"imports"`
	LogLevel          string
	Server            SecServer
	AppKeys           map[string]string `yaml:"appkeys"`
	RsaPublicKeyFile  string
	RsaPrivateKeyFile string
	IpipFile          string
	KtRatio           map[string]float64 `yaml:"ktratio"`
	RmbRatio          map[string]float64 `yaml:"rmbratio"`
	Servers           map[string]string
	ServerHost        map[string]string         `yaml:"serverhost"`
	Redis             map[string]cache.RedisCfg `yaml:"redis"`
	Drawfees          []Fee
	Switch            []ChargeSwitch          `yaml:"chargeswitch"`
	RewardYbt         map[string]float64      `yaml:"reward_ybt"`
	ProjectYbtAllot   []YbtAllot              `yaml:"project_ybt_allot"`
	ThirdAccount      map[string]ThirdAccount `yaml:"third_plat"`
	SystemAccounts    map[string]int64        `yaml:"system_accounts"`
	Alipay            AlipaySt                `yaml:"alipay"`
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

type Fee struct {
	Type          int      `json:"type"`
	Feetype       int      `json:"feetype"`
	Val           float64  `json:"val"`
	Min           float64  `json:"min"` // 可提币最小数量
	DayMaxPercent float64  `json:"daymaxpercent"`
	Max           *float64 `json:"max,omitempty"` // 可提币最大数量
}

type ChargeSwitch struct {
	Type     int  `json:"type"`
	Recharge bool `json:"recharge"`
	Withdraw bool `json:"withdraw"`
}

type YbtAllot struct {
	Type    int
	UserId  int64 `yaml:"user_id"`
	Percent float64
	Forever float64
	Fix     float64
	FixDays int64
	Users   []UserAllot
}
type UserAllot struct {
	UserId  int64 `yaml:"user_id"`
	Percent float64
}

type ThirdAccount struct {
	Key        string            `yaml:"key"`
	Secret     string            `yaml:"secret"`
	UserId     int64             `yaml:"user_id"`
	BonusId    int64             `yaml:"bonus_id"`
	WithDrawId int64             `yaml:"withdarw_id"`
	Ext        map[string]string `yaml:"ext"`
}

type AlipaySt struct {
	Appid     string
	Public    string
	Private   string
	NotifyUrl string
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

	// 判断数据合法性
	pya := Config.ProjectYbtAllot

	var percent float64 = 0
	for _, v := range pya {
		if v.Fix+v.Forever > 1+base.FLOAT_MIN {
			panic("config ProjectYbtAllot percent != 1")
			glog.Error("config ProjectYbtAllot percent != 1")
			return nil, err
		}
		if len(v.Users) > 0 {
			var percent2 float64 = 0
			for _, u := range v.Users {
				percent2 += u.Percent
			}
			if !base.IsEqual(percent2, 1) {
				s := fmt.Sprintln("config ProjectYbtAllot percent != 1, percent2=", percent2)
				panic(s)
				glog.Error(s)
				return nil, err
			}
		}
		percent += v.Percent
	}

	if !base.IsEqual(percent, 1) {
		s := fmt.Sprintln("config ProjectYbtAllot percent != 1, percent=", percent)
		panic(s)
		glog.Error(s)
		return nil, err
	}

	return &Config, err
}
