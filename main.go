package main

import (
	"encoding/gob"
	"github.com/beego/beego/v2/adapter/toolbox"
	"github.com/beego/beego/v2/server/web"
	"ziyoubiancheng/mbook/models"
	_ "ziyoubiancheng/mbook/routers"
	_ "ziyoubiancheng/mbook/sysinit"
	"ziyoubiancheng/mbook/utils/pagecache"
)

func init() {
	// 由于session内部采用gob来注册存储的对象，
	// 因此采用非memory引擎时，需要注册这些对象(结构体)
	gob.Register(models.Member{})
}

func main() {
	//beego.Run()

	// 执行定时任务，每2秒清理依次缓存
	task := toolbox.NewTask("clear_expired_cache", "*/2 * * * * *", func() error {
		//fmt.Println("------delete cache------")
		pagecache.ClearExpiredFiles()
		return nil
	})

	// 添加到任务列表
	toolbox.AddTask("mbook_task", task)
	// 开始执行任务
	toolbox.StartTask()
	// 结束释放任务
	defer toolbox.StopTask()

	web.Run()
}
