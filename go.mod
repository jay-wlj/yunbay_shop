module yunbay

go 1.12

require (
	cloud.google.com/go v0.37.4
	github.com/aws/aws-sdk-go v1.26.4
	github.com/axgle/mahonia v0.0.0-20180208002826-3358181d7394
	github.com/bwmarrin/snowflake v0.3.0
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/eoscanada/eos-go v0.0.0-20181211220314-714ac3c6c8c4
	github.com/facebookgo/ensure v0.0.0-20160127193407-b4ab57deab51 // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20150612182917-8dac2c3c4870 // indirect
	github.com/gin-gonic/gin v1.6.3
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.3.0 // indirect
	github.com/ipipdotnet/datx-go v0.0.0-20181123035258-af996d4701a0
	github.com/jay-wlj/gobaselib v1.1.1
	github.com/jay-wlj/wxpay v1.0.7
	github.com/jayden211/retag v0.0.0-20180725100701-84849874234d
	github.com/jie123108/glog v0.0.0-20160701133742-ca74c069d4e1
	github.com/jie123108/imaging v1.1.0
	github.com/jinzhu/gorm v1.9.11
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.9
	github.com/lib/pq v1.2.0
	github.com/mojocn/base64Captcha v0.0.0-20190801020520-752b1cd608b2
	github.com/nsqio/go-nsq v1.0.7
	github.com/pkg/errors v0.9.1 // indirect
	github.com/robfig/cron v1.2.0
	github.com/shopspring/decimal v0.0.0-20191009025716-f1972eb1d1f5
	github.com/smartwalle/alipay v1.0.2
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/testify v1.6.1
	github.com/tealeg/xlsx v1.0.5
	github.com/tidwall/gjson v1.3.2 // indirect
	github.com/tidwall/sjson v1.0.4 // indirect
	github.com/yunge/sphinx v0.0.0-20150804231640-7962b7621b64
	go.uber.org/zap v1.12.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/net v0.1.0
	golang.org/x/text v0.4.0
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/mailgun/mailgun-go.v1 v1.1.1
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

//replace github.com/jay-wlj/gobaselib => ../gobaselib
