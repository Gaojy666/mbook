package main

import (
	"fmt"
	"github.com/beego/beego/v2/adapter/toolbox"
	"github.com/beego/beego/v2/server/web"
	_ "ziyoubiancheng/mbook/routers"
	_ "ziyoubiancheng/mbook/sysinit"
	"ziyoubiancheng/mbook/utils/pagecache"
)

func init() {
}
func main() {
	//beego.Run()
	web.BConfig.WebConfig.Session.SessionOn = true

	task := toolbox.NewTask("clear_expired_cache", "*/2 * * * * *", func() error {
		fmt.Println("------delete cache------")
		pagecache.ClearExpiredFiles()
		return nil
	})

	toolbox.AddTask("mbook_task", task)
	toolbox.StartTask()
	defer toolbox.StopTask()

	web.Run()
}
