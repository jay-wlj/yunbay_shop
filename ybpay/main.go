package main

import (
	_ "expvar"
	"flag"
	_ "net/http/pprof"
	"runtime"
	"yunbay/ybpay/conf"
	"yunbay/ybpay/server"

	"github.com/jie123108/glog"
	"github.com/json-iterator/go/extra"
)

// func rlimit_init() {
// 	var rlim syscall.Rlimit
// 	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
// 	if err != nil {
// 		fmt.Println("get rlimit error: " + err.Error())
// 		//os.Exit(1)
// 	}
// 	fmt.Printf("limit before cur:%v max:%v\n", rlim.Cur, rlim.Max)
// 	rlim.Cur = 50000
// 	rlim.Max = 50000
// 	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
// 	if err != nil {
// 		fmt.Println("set rlimit error: " + err.Error())
// 		//os.Exit(1)
// 		return
// 	}
// 	fmt.Printf("limit after cur:%v max:%v\n", rlim.Cur, rlim.Max)
// }

func main() {
	//extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)	// 统一设置命名风格 如: UserId->user_id
	extra.RegisterFuzzyDecoders()
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

	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	//util.ReloadRsaKey()
	err = server.StartServer(config)

}
