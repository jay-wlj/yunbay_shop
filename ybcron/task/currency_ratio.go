package task

import (
	"encoding/json"
	"errors"
	"fmt"
	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"
	"regexp"
	"strings"
	"time"
	"yunbay/ybasset/common"
	"yunbay/ybcron/conf"

	"github.com/shopspring/decimal"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func Currency_Update() {
	db, err := db.InitPsqlDb(conf.Config.PsqlUrl["asset"], conf.Config.Debug)
	if err != nil {
		fmt.Println("Currency_Update fail! err=", err)
	}

	var vs []common.CurrencyRate
	if err = db.Find(&vs, "auto=? and digital=?", true, false).Error; err != nil {
		glog.Error("sina_spider fail! err=", err)
		return
	}

	source := "sina"
	mr := make(map[string]decimal.Decimal)
	if mr, err = sina_spider(db, vs); err != nil {
		glog.Error("Currency_Update fail! sina_spider err=", err)

		mr = ip138_spider(vs) // 改从ip138接口获取汇率
		source = "ip138"
	}

	if len(mr) > 0 {
		// 对于真实货币 调整当前汇率
		for k, v := range mr {
			mr[k] = v.Mul(decimal.NewFromFloat(1.01)) // 微调当前汇率增加1%
		}
		if err = updateRatio(db, vs, mr, source); err != nil {
			glog.Error("Currency_Update fail! updateRatio err=", err)
		}
	}

	vs = vs[0:0]
	if err = db.Find(&vs, "auto=? and digital=?", true, true).Error; err != nil {
		glog.Error("sina_spider fail! err=", err)
		return
	}

	// 数字货币汇率
	if mr, err = yunex_spider(db, vs); err != nil {
		glog.Error("Currency_Update fail! updateRatio err=", err)
	}
	if len(mr) > 0 {
		if err = updateRatio(db, vs, mr, "yunex"); err != nil {
			glog.Error("Currency_Update fail! updateRatio err=", err)
		}
	}
	return
}

// 更新汇率数据
func updateRatio(db *gorm.DB, vs []common.CurrencyRate, mr map[string]decimal.Decimal, source string) (err error) {

	now := time.Now().Unix()
	for i, v := range vs {
		if p, ok := mr[v.Key]; ok {
			if p.IsZero() {
				continue
			}

			// 相对于上一次的波动阀值绝对值不得大于百分比
			threshold, _ := decimal.NewFromString("0.05")
			str_threshold := "cny_threshold"
			if v.Digital {
				str_threshold = "kt_threshold"
			}
			if p, ok := conf.Config.ConversionRatio[str_threshold]; ok {
				threshold = p
			}
			// 波动阀值大于0.05
			if !v.Ratio.IsZero() && v.Ratio.Sub(p).Abs().GreaterThan(v.Ratio.Mul(threshold)) {
				//if v.Ratio != 0 && math.Abs(v.Ratio-p) > (0.05*v.Ratio) { // 波动阀值大于0.05
				glog.Error("updateRatio old:", v.Ratio, " new:", p)
				continue
			}

			// 货币波动不能太大
			if v.Ratio.Sub(p).Abs().GreaterThanOrEqual(decimal.NewFromFloat32(0.0001)) {
				vs[i].Ratio = p // 更新汇率
				vs[i].UpdateTime = now
			}
		}
	}

	ch, _ := cache.GetWriter("pub")
	key := "currency_ratio"
	for _, v := range vs {
		if v.UpdateTime < now { // 无更新
			continue
		}
		v.Source = source
		if err = db.Save(&v).Error; err != nil {
			glog.Error("sina_spider fail! err=", err)
			return
		}
		if ch != nil {
			ch.HSet(key, v.Key, fmt.Sprintf("%v", v.Ratio), 0)
		}
	}
	return
}

type okJson struct {
	Ok     int     `json:"ok"`
	Reason string  `json:"reason"`
	Data   yunexSt `json:"data"`
}

type yunexSt struct {
	CurPrice decimal.Decimal `json:"cur_price"`
	MaxPrice decimal.Decimal `json:"max_price"`
	MinPrice decimal.Decimal `json:min_price"`
	CoinPair string          `json:"coin_pair"`
}

// 从云网更新汇率
func yunex_spider(db *gorm.DB, vs []common.CurrencyRate) (mr map[string]decimal.Decimal, err error) {
	mr = make(map[string]decimal.Decimal)
	for _, v := range vs {
		switch v.From {
		case "kt", "ybt", "snet":
			switch v.To {
			case "kt", "ybt", "snet":
				uri := fmt.Sprintf("https://a.yunex.io/api/market/trade/info?symbol=%v_%v", v.From, v.To)
				rep := base.HttpGet(uri, nil, base.DefTimeOut)
				if err = rep.Error; err != nil {
					glog.Error("yunex_spider fail! err=", err)
					continue
				}
				var ret okJson
				if err = json.Unmarshal(rep.RawBody, &ret); err != nil {
					glog.Error("yunex_spider fail! err=", err)
					continue
				}
				if 1 != ret.Ok {
					err = errors.New(ret.Reason)
					glog.Error("yunex_spider fail! err=", err, " res=", ret)
					continue
				}
				if p, ok := conf.Config.ConversionRatio[v.Key]; ok {
					ret.Data.CurPrice = ret.Data.CurPrice.Mul(p) // 调整当前交易价格比例
				}
				mr[v.From+v.To] = ret.Data.CurPrice
			}
		}
	}
	return
}

// 从新浪接口爬取汇率
func sina_spider(db *gorm.DB, vs []common.CurrencyRate) (mr map[string]decimal.Decimal, err error) {
	keys := []string{}
	for _, v := range vs {
		if v.Digital {
			continue
		}
		keys = append(keys, fmt.Sprintf("fx_s%v%v", v.From, v.To))
	}
	list_keys := base.StringSliceToString(keys, ",")
	uri := fmt.Sprintf("http://hq.sinajs.cn/rn=%v&list=%v", time.Now().Unix(), list_keys)
	res := base.HttpGet(uri, nil, 10*time.Second)
	if res.StatusCode != 200 {
		glog.Error("sina_spider fail! err=", err, " uri=", uri)
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}
	s := string(res.RawBody)
	mr = parse_js(s)
	return
}

func parse_js(s string) (mr map[string]decimal.Decimal) {
	reg, _ := regexp.Compile("var ([a-zA-Z_0-9]+)=([^;]*);")
	ms := reg.FindAllString(s, -1)

	mr = make(map[string]decimal.Decimal)
	for _, v := range ms {
		nkey := strings.Index(v, "=")
		str_key := v[:nkey]
		//str_key = strings.TrimLeft(str_key, "var hq_str_fx_")
		str_key = strings.Replace(str_key, "var hq_str_fx_s", "", -1)
		str_val := v[nkey+1:]
		vals := strings.Split(str_val, ",")
		if len(vals) > 1 {
			mr[str_key], _ = decimal.NewFromString(vals[1])
		}
	}
	return
}

// 爬取ip138的货币汇率
func ip138_spider(vs []common.CurrencyRate) (mr map[string]decimal.Decimal) {
	mr = make(map[string]decimal.Decimal)
	for _, v := range vs {
		if v.Digital {
			continue
		}
		ratio, err := getip138Ratio(v.From, v.To)
		if err == nil {
			mr[v.Key] = ratio
		}
	}
	return mr
}

func getip138Ratio(from, to string) (ratio decimal.Decimal, err error) {
	uri := fmt.Sprintf("http://qq.ip138.com/hl.asp?from=%s&to=%s&q=1", strings.ToUpper(from), strings.ToUpper(to))
	resp := base.HttpGet(uri, nil, 10*time.Second)
	if resp.StatusCode != 200 {
		glog.Error("ip138_spider fail! err=", err, " uri=", uri)
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}
	s := string(resp.RawBody)
	var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(resp.RawBody)
	s = string(decodeBytes)
	// 正则查找汇率
	reg, _ := regexp.Compile(`<table class=\"rate\">[\s\S]*</table>`)
	s = reg.FindString(s)
	s = strings.Replace(s, "\r\n", "", -1)
	reg, _ = regexp.Compile(`</td></tr><tr><td>(.*)</td><td>(.*)</td><td>(.*)</td></tr></table>`)

	ms := reg.FindStringSubmatch(s)
	if len(ms) >= 3 {
		ratio, err = decimal.NewFromString(ms[2])
	}
	return
}
