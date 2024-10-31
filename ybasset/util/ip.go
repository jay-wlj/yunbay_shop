package util


import (
	"github.com/ipipdotnet/datx-go"
	."yunbay/ybasset/conf"
	"github.com/jie123108/glog"
	"sync"
)


var (
	g_city *datx.City
	g_city_once sync.Once
)



func LoadIpData(){
	var err error
	g_city, err = datx.NewCity(Config.IpipFile)
	if err != nil {
		glog.Error("path:", Config.IpipFile)
		panic("ipipvip.datx not exist!")
	}
}

func GetCitysByIp(ip string) (citys []string) {
	g_city_once.Do(LoadIpData)

	citys = []string{}
    if g_city != nil {
		var e error
		citys, e = g_city.Find(ip)
		if e != nil {
			glog.Error(e)
		}
		return 
	}
	return 
}