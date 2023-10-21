package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/server/web"
	"time"
	"ziyoubiancheng/mbook/models"
	"ziyoubiancheng/mbook/utils"
)

type ElasticsearchController struct {
	BaseController
}

func (c *ElasticsearchController) Search() {
	c.TplName = "search/search.html"
}

func (c *ElasticsearchController) Result() {
	// 获取关键词
	wd := c.GetString("wd")
	if wd == "" {
		c.Redirect(web.URLFor("ElasticsearchController.Search"), 302)
	}
	c.Data["Wd"] = wd

	// 搜章节 & 图书
	tab := c.GetString("tab", "doc")
	c.Data["Tab"] = tab

	// page&size
	page, _ := c.GetInt("page", 1)
	if page < 1 {
		page = 1
	}
	size := 10

	// 开始搜索
	now := time.Now()
	if tab == "doc" {
		ids, count, err := models.ElasticSearchDocument(wd, size, page)
		c.Data["totalRows"] = count
		if err == nil && len(ids) > 0 {
			c.Data["Docs"], _ = models.NewDocumentSearch().GetDocsById(ids)
		}
	} else {
		ids, count, err := models.ElasticSearchBook(wd, size, page)
		c.Data["totalRows"] = count
		if err == nil && len(ids) > 0 {
			c.Data["Books"], _ = models.NewBook().GetBooksByIds(ids)
		}
	}

	if c.Data["totalRows"].(int) > size {
		// 有分页信息
		urlSuffix := fmt.Sprintf("&tab=%v&wd=%v", tab, wd)
		html := utils.NewPaginations(4, c.Data["totalRows"].(int), size, page, web.URLFor("ElasticsearchController.Result"), urlSuffix)
		c.Data["PageHtml"] = html
	} else {
		c.Data["PageHtml"] = ""
	}

	c.Data["SpendTime"] = fmt.Sprintf("%.3f", time.Since(now).Seconds())
	c.TplName = "search/result.html"

}
