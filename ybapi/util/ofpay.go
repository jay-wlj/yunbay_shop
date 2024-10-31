package util

import (
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
	"yunbay/ybapi/common"
	"yunbay/ybapi/conf"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/yf"

	"github.com/axgle/mahonia"

	"golang.org/x/net/html/charset"

	"github.com/jie123108/glog"
)

type ofPay struct {
}

var (
	g_ofpay      *ofPay
	g_ofpay_once sync.Once
)

const (
	timeout = 5 * time.Second
)

const (
	ERR_OFPAY_OK                = 1
	ERR_OFPAY_MCH               = 1001
	ERR_OFPAY_IP                = 1002
	ERR_OFPAY_MD5               = 1003
	ERR_OFPAY_DISABLE           = 1004
	ERR_OFPAY_AMOUNT_EXCEED     = 1005
	ERR_OFPAY_PRICE_EXCEED      = 1007
	ERR_OFPAY_NO_ARGS           = 1008
	ERR_OFPAY_NO_DATA           = 1010
	ERR_OFPAY_DISABLE_RECHARGE  = 11
	ERR_OFPAY_NOT_MORE          = 12
	ERR_OFPAY_ORDERID_NOT_EXIST = 2000

	GAME_STATE_OK   = 1
	GAME_STATE_FAIL = 9
)

type OfPayUnmarsh interface {
	UnMarshl(s string) error
}

type telCheckResp struct {
	Code    int
	Reason  string
	Address string
}

func (t *telCheckResp) UnMarshl(s string) (err error) {
	as := strings.Split(s, "#")
	if len(as) >= 2 {
		t.Code, err = strconv.Atoi(as[0])
		t.Reason = as[1]
		t.Address = as[2]
	} else {
		glog.Error("telCheckResp fail! s=", s)
		err = errors.New(s)
	}
	if t.Code != ERR_OFPAY_OK {
		err = errors.New(t.Reason)
	}
	return
}

func GetOfpay() *ofPay {
	g_ofpay_once.Do(func() {
		g_ofpay = &ofPay{}
	})
	return g_ofpay
}

func (t *ofPay) getHost() string {
	return conf.Config.OfPay.Host
}

func (t *ofPay) httpPost(uri string, args map[string]string, r interface{}) (err error) {

	resp := base.HttpReqInternal("POST", uri, nil, args, nil, timeout)
	if err = resp.Error; err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("resp status code is %v", resp.StatusCode)
		return
	}

	s := string(resp.RawBody)

	if r != nil {
		switch m := r.(type) {
		case *int:
			*m, err = strconv.Atoi(s)
		case OfPayUnmarsh:
			s := string(resp.RawBody)
			if strings.Contains(resp.Headers.Get("Content-Type"), "charset=gbk") {
				dec := mahonia.NewDecoder("GBK")
				s = dec.ConvertString(s)
			}

			err = m.UnMarshl(s)

		case xml.Unmarshaler:
			dec := xml.NewDecoder(strings.NewReader(s))
			err = dec.Decode(m)
		default:
			if args["format"] == "json" {
				err = json.Unmarshal(resp.RawBody, r)
				return
			}
			// xml解析
			dec := xml.NewDecoder(strings.NewReader(s))
			// 添加从指定编码格式转到utf-8的编码器
			dec.CharsetReader = func(c string, i io.Reader) (io.Reader, error) {
				return charset.NewReaderLabel(c, i)
			}
			err = dec.Decode(r)
		}

	} else {
		s := string(resp.RawBody)
		if strings.Contains(resp.Headers.Get("Content-Type"), "charset=gbk") {
			dec := mahonia.NewDecoder("GBK")
			s = dec.ConvertString(s)
		}

		fmt.Println("body=", string(s))

	}
	return
}

func (t *ofPay) getargs() yf.StringMap {
	args := make(yf.StringMap)

	f := conf.Config.OfPay
	args.SetString("userid", f.AppId)
	args.SetString("userpws", f.AppPws)
	if f.RetUrl != "" {
		args.SetString("ret_url", f.RetUrl)
	}

	return args
}

// 检测是否可充值
func (t *ofPay) Tel_Check(tel string, price int) (err error) {
	uri := t.getHost() + "/telcheck.do"
	a := t.getargs()
	a.SetString("phoneno", tel)
	a.SetInt("price", price)

	//var s string
	var v telCheckResp
	err = t.httpPost(uri, a, &v)
	return
}

type OfOrder struct {
	common.OfOrder

	//OrderId string `xml:"orderid"`
	// OrderCash  decimal.Decimal `xml:"ordercash"`
	// CardName   string          `xml:"cardname"`
	// Sporder_id int64           `xml:"sporder_id"`
	// GameUserid string          `xml:"game_userid"`
	// GameState  int             `xml:"game_state"`
}

// 话费充值
func (t *ofPay) Tel_Recharge(order_id int64, tel string, price int64) (v OfOrder, err error) {
	uri := t.getHost() + "/onlineorder.do"
	a := t.getargs()
	a.SetString("cardid", "140101")
	a.SetInt64("cardnum", price)
	a.SetInt64("sporder_id", order_id)
	a.SetString("sporder_time", time.Now().Format("20060102150405"))
	a.SetString("game_userid", tel)
	a.SetString("version", "6.0")
	if ret_url := conf.Config.OfPay.RetUrl; ret_url != "" {
		a.SetString("ret_url", ret_url)
	}

	a["md5_str"] = t.sign(a.GetString("cardid"), a.GetString("cardnum"), a.GetString("sporder_id"), a.GetString("sporder_time"), a.GetString("game_userid"))

	err = t.httpPost(uri, a, &v)
	return
}

// 话费充值
func (t *ofPay) Query(order_id int64) (v int, err error) {
	uri := t.getHost() + "/api/query.do"
	a := t.getargs()
	a.SetInt64("spbillid", order_id)

	err = t.httpPost(uri, a, &v)
	return
}

// 查询订单详情
func (t *ofPay) QueryInfo(order_id int64) (v OfOrder, err error) {
	uri := t.getHost() + "/queryOrderInfo.do"

	a := t.getargs()
	a.SetInt64("sporder_id", order_id)
	a.SetString("version", "6.0")
	a.SetString("format", "json")

	a["md5_str"] = t.sign(a.GetString("sporder_id"))

	err = t.httpPost(uri, a, &v)
	return
}

func (t *ofPay) sign(params ...string) string {
	f := conf.Config.OfPay

	ss := strings.Builder{}
	ss.WriteString(f.AppId)
	ss.WriteString(f.AppPws)
	ss.WriteString(base.StringSliceToString(params, ""))
	ss.WriteString(f.AppSecret)
	return fmt.Sprintf("%X", md5.Sum([]byte(ss.String())))
}

// 卡密提取
func (t *ofPay) CardWidthdraw(order_id int64, cardid string, num int) (v OfOrder, err error) {
	uri := t.getHost() + "/order.do"
	a := t.getargs()
	a.SetString("cardid", cardid)
	a.SetInt("cardnum", num)
	a.SetInt64("sporder_id", order_id)
	a.SetString("sporder_time", time.Now().Format("20060102150405"))
	//a.SetString("game_userid", tel)
	a.SetString("version", "6.0")
	a.SetString("format", "json")
	if ret_url := conf.Config.OfPay.RetUrl; ret_url != "" {
		a.SetString("ret_url", ret_url)
	}

	a["md5_str"] = t.sign(a.GetString("cardid"), a.GetString("cardnum"), a.GetString("sporder_id"), a.GetString("sporder_time"))
	err = t.httpPost(uri, a, &v)
	return
}
