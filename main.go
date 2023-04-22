package main

import (
	"flag"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"landlord/controllers"
	_ "landlord/routers"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/beego/beego/v2/server/web"
)

func profile() {
	var profile = flag.String("profile", "", "write memory profile to this file")
	go func() {
		defer pprof.StopCPUProfile()
		if *profile != "" {
			f, err := os.Create(*profile)
			if err != nil {
				logs.Error(err)
			}
			file, err := os.Create("./cpu.pprof")
			if err != nil {
				fmt.Printf("create cpu pprof failed, err:%v\n", err)
				return
			}
			err = pprof.StartCPUProfile(file)
			err = pprof.WriteHeapProfile(f)
			if err != nil {
				logs.Error(err)
			}
		}
	}()
}

func setLoggerFile() {
	err := logs.SetLogger(logs.AdapterFile, `{"filename":"landlord.log","level":7,"maxLines":0,"maxsize":0,"daily":true,"maxDays":10,"color":true}`)
	if err != nil {
		logs.Error(err)
		os.Exit(0)
	}
}
func main() {
	logs.SetLogFuncCall(true)
	//logs.EnableFuncCallDepth(true)
	//logs.SetLogFuncCallDepth(3)
	tree := web.PrintTree()
	methods := tree["Data"].(web.M)
	for k, v := range methods {
		fmt.Printf("%s => %v\n", k, v)
	}

	//err := config.InitGlobalInstance("yaml", "conf.yaml")
	//if err != nil {
	//	logs.Critical("An error occurred:", err)
	//	panic(err)
	//}
	if web.BConfig.RunMode == "dev" {
		web.BConfig.WebConfig.DirectoryIndex = true
		web.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	runtime.GOMAXPROCS(runtime.NumCPU()) //设置 P的数目  M P G
	web.ErrorController(&controllers.ErrorController{})
	web.Run()
}
