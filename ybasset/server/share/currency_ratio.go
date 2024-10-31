package share

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"yunbay/ybasset/common"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/cache"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/shopspring/decimal"

	"github.com/jie123108/glog"
)

var CURRENCY_RATIO_TIMEOUT time.Duration = 0 // 永久

type rmbRatioSt struct {
	sync.RWMutex
	ratios map[string]float64
}

func (t *rmbRatioSt) Assign(rmbs map[string]float64) {
	t.Lock()
	defer t.Unlock()
	for k, v := range rmbs {
		t.ratios[k] = v
	}
}
func (t *rmbRatioSt) Set(key string, ratio float64) {
	t.Lock()
	defer t.Unlock()
	t.ratios[key] = ratio
}
func (t *rmbRatioSt) Get(key string) (ratio float64) {
	t.RLock()
	defer t.RUnlock()
	var ok bool
	if ratio, ok = t.ratios[key]; !ok {
		ratio = 0
	}
	return
}
func (t *rmbRatioSt) GetAll() (ratios map[string]float64) {
	t.RLock()
	defer t.RUnlock()
	ratios = make(map[string]float64)
	for k, v := range t.ratios {
		ratios[k] = v
	}
	return
}

var rmbRatio rmbRatioSt

func init() {
	rmbRatio = rmbRatioSt{ratios: make(map[string]float64)}
}

// 从一种货币到另一种货币的兑换比例
func GetRatio(from_type, to_type string) (ratio float64) {
	//rmbs, _ := GetRmbRatio()
	from_type = strings.ToLower(from_type)
	to_type = strings.ToLower(to_type)
	switch to_type {
	case "cny": // cny
		//ratio = rmbs[from_type]
		ratio = rmbRatio.Get(from_type)
		break
	default:
		if from_type == to_type {
			ratio = 1
		} else if from_type == "cny" {
			if v := rmbRatio.Get(to_type); v > base.FLOAT_MIN {
				//if rmbs[to_type] > base.FLOAT_MIN {
				ratio = float64(1) / v
			}
		} else {
			//if rmbs[to_type] > base.FLOAT_MIN {
			if v := rmbRatio.Get(to_type); v > base.FLOAT_MIN {
				ratio = rmbRatio.Get(from_type) / v
			}
		}
	}
	return
}

// 其它货币转换到该货币的比例
func GetRatios(to_type string) (ratios map[string]float64) {
	ratios = make(map[string]float64)
	rmbs, _ := GetRmbRatio()

	switch to_type {
	case "cny":
		ratios, _ = GetRmbRatio()
		break
	default:
		if v := rmbs[to_type]; v > base.FLOAT_MIN {
			ratios["cny"] = float64(1) / v
			for k, v := range rmbs {
				if k != to_type {
					ratios[k] = v / rmbs[to_type]
				}
			}
		}
		break
	}
	return
}

// 将货币兑换比例添加到缓存中
func SetRatioCache(txType string, ratios map[string]float64) (err error) {
	cache, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
	}
	key := fmt.Sprintf("asset_ratio_%v", txType)
	for k, v := range ratios {
		err = cache.HSet(key, k, v, CURRENCY_RATIO_TIMEOUT)
		if err != nil {
			return
		}
	}
	return
}

func SetOneRatioCache(txType, from_type string, ratio float64) (err error) {
	cache, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
	}
	key := fmt.Sprintf("asset_ratio_%v", txType)
	err = cache.HSet(key, from_type, ratio, CURRENCY_RATIO_TIMEOUT)
	return
}
func GetRatioCache(to_type, from_type string) (ratio float64, err error) {
	cache, err := cache.GetWriter(common.RedisPub)
	// 优先从缓存里获取
	if err == nil {
		key := fmt.Sprintf("asset_ratio_%v", to_type)
		ratio, err = cache.HGetF64(key, from_type)
	}

	return
}

func GetRmbRatio() (rmbs map[string]float64, err error) {
	rmbs = rmbRatio.GetAll()
	cache, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("GetDefaultCache fail! err=", err)
	} else {
		// 从redis缓存里更新
		for k, _ := range rmbs {
			if f, err := cache.HGetF64("asset_ratio_rmb", k); err == nil {
				rmbRatio.Set(k, f)
			}
		}
	}
	rmbs = rmbRatio.GetAll()
	return
}

// 同步币种兑换比例到缓存中
func SyncRatioToCacheFromConfig(rmbs map[string]float64) (err error) {
	// 同步其它币种到rmb的兑换比例
	//rmbs := rmbRatio.GetAll()
	if rmbs != nil {
		rmbRatio.Assign(rmbs)
	} else {
		rmbs = rmbRatio.GetAll()
	}

	if err = SetRatioCache("cny", rmbs); err != nil {
		glog.Error("SyncRatioToCacheFromConfig fail! err=", err)
		return
	}
	for k, _ := range rmbs {
		if k != "cny" {
			ratios := GetRatios(k)
			if err = SetRatioCache(k, ratios); err != nil {
				glog.Error("SyncRatioToCacheFromConfig fail! err=", err)
				return
			}
		}
	}
	return
}

// 其它币种相对rmb汇率有变化
func UpdateRatio(from_type string, ratio float64) (err error) {
	//conf.Config.RmbRatio[strings.ToLower(from_type)] = ratio
	rmbRatio.Set(strings.ToLower(from_type), ratio)
	// 刷新缓存汇率
	return SyncRatioToCacheFromConfig(nil)
}

type RatioSt struct {
	From  string          `json:"from"`
	To    string          `json:"to"`
	Type  int             `json:"type"`
	Ratio decimal.Decimal `json:"ratio"`
}

// 币种兑换
func (t *RatioSt) UpdateRatio() (err error) {
	key := t.From + t.To
	if t.Type > 0 {
		key += fmt.Sprintf("_%v", t.Type)
	}

	t.Ratio = t.Ratio.Round(4) // 保留四位小数

	db := db.GetTxDB(nil)
	v := common.CurrencyRate{Key: key, From: t.From, To: t.To, Digital: true, Ratio: t.Ratio, Auto: true}
	db.DB = db.Set("gorm:insert_option", fmt.Sprintf("ON CONFLICT (key) DO update set ratio=%v", t.Ratio))
	if err = db.Save(&v).Error; err != nil {
		glog.Error("UpdateRatio fail! err=", err)
		db.Rollback()
		return
	}

	// 更新缓存
	ch, e := cache.GetWriter("pub")
	if err = e; err != nil {
		glog.Error("UpdateRatio fail! err=", err)
		db.Rollback()
		return
	}
	if ch != nil {
		ch.HSet("currency_ratio", key, t.Ratio.String(), 0)
	}

	db.Commit()
	return
}

func (t *RatioSt) Get() (ratio decimal.Decimal, err error) {
	cache, err := cache.GetWriter(common.RedisPub)
	if err != nil {
		glog.Error("RatioSt Get fail! err=", err)
		return
	}

	key := "currency_ratio"
	field := t.From + t.To
	if t.Type > 0 {
		field += fmt.Sprintf("_%v", t.Type)
	}
	var val string
	if val, err = cache.HGet(key, field); err == nil {
		ratio, err = decimal.NewFromString(val)
	}
	if err != nil {
		glog.Error("CurrencyRatio HGetF64 fail! key=", field, " err=", err)
		// 从数据库里获取
		var r common.CurrencyRate
		db := db.GetDB()
		if err = db.Find(&r, "key=?", field).Error; err != nil {
			glog.Error("CurrencyRatio fail! key=", field, " err=", err)
			return
		}
		ratio = r.Ratio
		cache.HSet(key, field, ratio.String(), 0)
	}
	return
}
