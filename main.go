package main

import (
	"github.com/beego/beego/v2/server/web"
	_ "ziyoubiancheng/mbook/routers"
	_ "ziyoubiancheng/mbook/sysinit"
)

func init() {
}
func main() {
	//beego.Run()
	web.BConfig.WebConfig.Session.SessionOn = true

	web.Run()
}
