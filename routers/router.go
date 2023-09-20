package routers

import (
	"github.com/beego/beego/v2/server/web"
	"ziyoubiancheng/mbook/controllers"
)

func init() {
	//web.Router("/", &controllers.MainController{})
	// 首页&分类
	// 通过get方法请求到controller中的Index方法里面
	web.Router("/", &controllers.HomeController{}, "get:Index")
	web.Router("/explore", &controllers.ExploreController{}, "get:Index")
}
