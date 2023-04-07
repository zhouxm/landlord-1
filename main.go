package main

import (
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	_ "landlord/routers"

	"github.com/beego/beego/v2/server/web"
)

func main() {
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	logs.Info(web.AppConfig.Strings("httpport"))
	web.BConfig.RouterCaseSensitive = false
	tree := web.PrintTree()
	methods := tree["Data"].(web.M)
	for k, v := range methods {
		fmt.Printf("%s => %v\n", k, v)
	}
	if web.BConfig.RunMode == "dev" {
		web.BConfig.WebConfig.DirectoryIndex = true
		web.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	web.Run()
}
