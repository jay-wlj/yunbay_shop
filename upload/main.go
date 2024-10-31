package main

import (
	"flag"
	"yunbay/upload/conf"
	"yunbay/upload/server"

	"github.com/jie123108/glog"
)

func main() {

	var configfilename string

	flag.StringVar(&configfilename, "config", "./conf/config.yml", "ini config filename")

	flag.Parse()
	defer glog.Flush()
	// glog.Errorf("################### Build Time: %s ###################", base.BuildTime)

	config, err := conf.LoadConfig(configfilename)
	if err != nil {
		return
	}

	glog.Infof("init pprof ------------")
	//	go func() {
	//		res := http.ListenAndServe(":8910", nil)
	//		glog.Errorf("init pprof stop, res:%v", res)
	//	}()
	err = server.StartServer(config)
}
