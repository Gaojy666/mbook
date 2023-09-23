package controllers

import "strings"

type DocumentController struct {
	BaseController
}

// 图书目录&详情页
func (c *DocumentController) Index() {
	token := c.GetString("token")
	identify := c.Ctx.Input.Param(":key")
	if identify == "" {
		c.Abort("404")
	}
	tab := strings.ToLower(c.GetString("tab"))

	bookResult := c.getBookData(identify, token)
}
