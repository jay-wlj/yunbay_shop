package conf

import (
	"fmt"

	base "github.com/jay-wlj/gobaselib"

	"github.com/jay-wlj/gobaselib/cache"

	"github.com/jie123108/glog"
	"github.com/shopspring/decimal"
)

type ApiConfig struct {
	Imports         []string `yaml:"imports"`
	LogLevel        string
	Debug           bool
	Test            bool
	AppKeys         map[string]string
	Redis           map[string]cache.RedisCfg `yaml:"redis"`
	Rebat           RebatConfig
	Crons           map[string]string
	PsqlUrl         map[string]string `yaml:"psqlurl"`
	MQUrls          []string          `yaml:"mqurls"`
	Servers         map[string]string
	ServerHost      map[string]string `yaml:"serverhost"`
	Mining          MiningConfig
	AirdopCfg       AidropConfig `yaml:"air_drop"`
	ProjectYbtAllot []YbtAllot   `yaml:"project_ybt_allot"`
	//LastBonusKtUserId int64 `yaml:"lastbonuskt_account"`
	ThirdAccount    map[string]ThirdAccount    `yaml:"third_plat"`
	SystemAccounts  map[string]int64           `yaml:"system_accounts"`
	Orders          map[string]string          `yaml:"orders"`
	ConversionRatio map[string]decimal.Decimal `yaml:"conversion_ratio"`
}

type RedisServer struct {
	Addr     string
	Password string
	DBIndex  int    `yaml:"dbindex"`
	Timeout  string `yaml:"timeout"`
}

type AidropConfig struct {
	Timeout int64 `yaml:"timeout"`
}

type RebatConfig struct {
	//IssueYbt float64 `yaml:"issue_ybt"`
	DestroyedYbt          float64 `yaml:"destoryed_ybt"`
	BuyerRebatPercent     float64 `yaml:"buyer_rebat"`
	SellerRebatPercent    float64 `yaml:"seller_rebat"`
	RecommendRebatPercent float64 `yaml:"re_rebat"`
	// Recommend2RebatPercent float64 `yaml:"re2_rebat"`
	YunbayRebat RebatYunbayConfig
}
type RebatYunbayConfig struct {
	//IssueYbt float64 `yaml:"issue_ybt"`
	Forever float64
	Fix     float64
	FixDays int64
}

type MiningConfig struct {
	Standard    float64
	Coefficient float64
	Powy        float64
	OnlineTime  string  `yaml:"online_time"`
	TotalIssue  float64 `yaml:"total_issue"`
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
	Ext        map[string]string `yaml:"ext`
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

	// 检验配置合法性
	if rebat_percent := (Config.Rebat.BuyerRebatPercent + Config.Rebat.SellerRebatPercent + Config.Rebat.RecommendRebatPercent); rebat_percent >= float64(1.0) {
		glog.Error("反利比率配置>1 rebat_percent:", rebat_percent)
		panic("反利比率配置不等于1")
		return nil, err
	}

	if Config.Mining.Standard == float64(0) {
		panic("挖矿基准难度不可为0")
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
