package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	"ziyoubiancheng/mbook/models"
)

type HomeController struct {
	BaseController
}

// 首页
func (c *HomeController) Index() {
	// 将分类的所有信息拿到以后给home/list.html做渲染
	// 拿到所有的status为1的分类(第一个-1表示所有)
	if cates, err := new(models.Category).GetCates(-1, 1); err == nil {
		c.Data["Cates"] = cates
	} else {
		logs.Error(err.Error())
	}
	c.TplName = "home/list.html"
}

func (c *HomeController) Index2() {
	c.TplName = "home/list.html"
}
