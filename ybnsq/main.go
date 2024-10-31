package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"yunbay/ybnsq/conf"
	"yunbay/ybnsq/server"

	"github.com/jie123108/glog"
)

func main() {
	var configfilename string
	flag.StringVar(&configfilename, "config", "./ybnsq/conf/config.yml", "ini config filename")

	flag.Parse()
	defer glog.Flush()

	_, err := conf.LoadConfig(configfilename)
	if err != nil {
		panic(err)
	}
	//server.InitPsqlDb(Config.Dburl)
	//defer server.CloseDb()

	server.StartServer()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}
