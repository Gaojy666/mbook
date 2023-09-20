package sysinit

import (
	"github.com/beego/beego/v2/server/web"
	"path/filepath"
	"strings"
)

// 系统初始化，做一些静态路径和必要变量的设置
func sysinit() {
	// 书会上传一些静态资源，会自动上传到uploads目录
	uploads := filepath.Join("./", "uploads")
	// 这意味着当用户访问 /uploads 路径时，Beego框架会查找并返回mbook/uploads目录下的静态文件
	web.BConfig.WebConfig.StaticDir["/uploads"] = uploads

	// 注册前端使用函数
	registerFunctions()

}

// 由于view层调用一些后端的函数
// 需要将一些函数注册到beego的FunctionMap中。
// 需要添加的函数较多，全部放入registerFunctions()中
func registerFunctions() {
	// 将cdnjs函数中对应的js全路径返回给前端页面
	// 传入的p是数据库给定的一个相对目录
	// cdn
	web.AddFuncMap("cdnjs", func(p string) string {
		//获取了配置项名为 "cdnjs" 的值，并将其赋值给变量 cdn。
		//如果配置项存在，则 cdn 变量将被赋值为配置项的值；
		//如果配置项不存在，则 cdn 变量将被赋值为空字符串。
		cdn := web.AppConfig.DefaultString("cdnjs", "")
		// 判断p中是否以/打头，并且cdn是否以/为结尾
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasSuffix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})

	web.AddFuncMap("cdncss", func(p string) string {
		cdn := web.AppConfig.DefaultString("cdncss", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasPrefix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})

}
